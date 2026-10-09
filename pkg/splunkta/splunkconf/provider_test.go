// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"context"
	"runtime"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkscript"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunktcp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkudp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkwineventlog"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const (
	baseURI  = "file:testdata/base.yaml"
	confURI  = "splunkconf://testdata/etc?pipeline=uf"
	confRoot = "testdata/etc"
)

// The component IDs the fixtures resolve to. Asserting the exact set pins the ID
// rule: the type comes from the stanza scheme and the name is a readable slug
// plus a hash of the raw stanza name, so any change to either shows up here.
var (
	wantReceivers = []string{
		"splunk_batch/batch-var-lib-batch-txt-3fccdf42",
		"splunk_monitor/monitor-var-log-bar-with-space-c81f7357",
		"splunk_monitor/monitor-var-log-foo-31463e8f",
		"splunk_script/script-bin-collect-sh-55e5ff8f",
		"splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09",
		"splunk_udp/udp-5515-016ce446",
		"splunk_wineventlog/wineventlog-application-9a1a225b",
	}
	wantExporters = []string{"splunk_hec/hecout-d6a24a0d"}
)

func receiverFactories() map[string]receiver.Factory {
	return map[string]receiver.Factory{
		splunkmonitor.TypeStr:     splunkmonitor.NewFactory(),
		splunktcp.TypeStr:         splunktcp.NewFactory(),
		splunkudp.TypeStr:         splunkudp.NewFactory(),
		splunkscript.TypeStr:      splunkscript.NewFactory(),
		splunkbatch.TypeStr:       splunkbatch.NewFactory(),
		splunkwineventlog.TypeStr: splunkwineventlog.NewFactory(),
	}
}

func resolve(t *testing.T, uris ...string) *confmap.Conf {
	t.Helper()
	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: uris,
		ProviderFactories: []confmap.ProviderFactory{
			fileprovider.NewFactory(),
			NewFactory(),
		},
	})
	require.NoError(t, err)

	merged, err := res.Resolve(context.Background())
	require.NoError(t, err)
	return merged
}

func pipelineIDs(t *testing.T, merged *confmap.Conf, pipeline string) (recvs, exps []string) {
	t.Helper()
	sub, err := merged.Sub("service::pipelines::" + pipeline)
	require.NoError(t, err)
	var pl struct {
		Receivers []string `mapstructure:"receivers"`
		Exporters []string `mapstructure:"exporters"`
	}
	require.NoError(t, sub.Unmarshal(&pl))
	return pl.Receivers, pl.Exporters
}

// The emitted fragment has to merge additively with a hand-written config and
// then survive the build path every other component goes through: unmarshal into
// the real factory's config, validate, and construct.
func TestResolveAndBuild(t *testing.T) {
	ctx := context.Background()
	merged := resolve(t, baseURI, confURI)

	pipelines, err := merged.Sub("service::pipelines")
	require.NoError(t, err)
	require.Contains(t, pipelines.ToStringMap(), "logs/base", "the base pipeline must survive the merge")
	require.Contains(t, pipelines.ToStringMap(), "logs/uf", "the .conf pipeline must be added, not replace the base")

	recvIDs, expIDs := pipelineIDs(t, merged, "logs/uf")
	require.Equal(t, wantReceivers, recvIDs)
	require.Equal(t, wantExporters, expIDs)

	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)
	factories := receiverFactories()
	for _, id := range recvIDs {
		t.Run(id, func(t *testing.T) {
			var cid otelcomponent.ID
			require.NoError(t, cid.UnmarshalText([]byte(id)))
			f, ok := factories[cid.Type().String()]
			require.True(t, ok, "emitted a type with no registered factory")

			sub, subErr := receivers.Sub(id)
			require.NoError(t, subErr)
			cfg := f.CreateDefaultConfig()
			require.NoError(t, sub.Unmarshal(cfg), "emitted config must unmarshal into the wrapper")
			require.NoError(t, confmap.Validate(cfg))

			r, buildErr := f.CreateLogs(ctx, receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
			// The underlying wineventlog receiver is build-tagged and its factory
			// refuses off Windows, so the expectation is platform-dependent.
			if cid.Type().String() == splunkwineventlog.TypeStr && runtime.GOOS != "windows" {
				require.ErrorContains(t, buildErr, "wineventlog is not supported outside Windows environments")
				require.Nil(t, r)
				return
			}
			require.NoError(t, buildErr)
			require.NotNil(t, r)
		})
	}

	exporters, err := merged.Sub("exporters")
	require.NoError(t, err)
	ef := splunkhecexporter.NewFactory()
	esub, err := exporters.Sub(expIDs[0])
	require.NoError(t, err)
	ecfg := ef.CreateDefaultConfig()
	require.NoError(t, esub.Unmarshal(ecfg))
	require.NoError(t, confmap.Validate(ecfg))
	e, err := ef.CreateLogs(ctx, exportertest.NewNopSettings(ef.Type()), ecfg)
	require.NoError(t, err)
	require.NotNil(t, e)
}

// Spot-check the emitted fields that carry a design decision, rather than every
// key: the decisions are what a future change is likely to undo.
func TestEmittedFields(t *testing.T) {
	merged := resolve(t, confURI)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)

	for _, tt := range []struct {
		want map[string]any
		name string
		id   string
	}{
		{
			// monitor's target is a scalar path, not a list: a stanza has exactly
			// one target. blacklist is unmodeled so it round-trips verbatim rather
			// than being re-encoded into an exclude list.
			name: "monitor target is scalar and blacklist survives verbatim",
			id:   "splunk_monitor/monitor-var-log-foo-31463e8f",
			want: map[string]any{
				"path":       "/var/log/foo",
				"index":      "main",
				"sourcetype": "syslog",
				"blacklist":  `\.gz$`,
			},
		},
		{
			// app_dir is what lets a relative script target resolve against the app
			// that declared it instead of the collector's working directory.
			name: "script carries the declaring app dir",
			id:   "splunk_script/script-bin-collect-sh-55e5ff8f",
			want: map[string]any{
				"script_filename": "./bin/collect.sh",
				"app_dir":         "testdata/etc/apps/splunk_ta_demo",
				"interval":        "60",
				"index":           "app_scripts",
			},
		},
		{
			// UF's port-only form ([udp://5515]) means every interface.
			name: "udp port-only form binds every interface",
			id:   "splunk_udp/udp-5515-016ce446",
			want: map[string]any{
				"listen_address": "0.0.0.0",
				"port":           5515,
			},
		},
		{
			// [WinEventLog://...] is the only input kind whose canonical spelling
			// is not lowercase, and was previously dropped entirely.
			name: "WinEventLog stanza maps despite its mixed-case kind",
			id:   "splunk_wineventlog/wineventlog-application-9a1a225b",
			want: map[string]any{
				"event_log_name": "Application",
				"index":          "wineventlog",
				"current_only":   "1",
			},
		},
		{
			name: "tcp keeps an unmodeled param for the Extra passthrough",
			id:   "splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09",
			want: map[string]any{
				"listen_address": "0.0.0.0",
				"port":           5514,
				"queueSize":      "500KB",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := receivers.Sub(tt.id)
			require.NoError(t, err)
			got := sub.ToStringMap()
			for k, want := range tt.want {
				require.Equal(t, want, got[k], "key %q", k)
			}
		})
	}
}

// Stanzas the provider must not emit. The exact ID set in TestResolveAndBuild
// already fixes the count; this records why each of these is excluded.
func TestStanzasNotEmitted(t *testing.T) {
	merged := resolve(t, confURI)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)
	ids := receivers.ToStringMap()

	for _, tt := range []struct {
		name string
		slug string
		why  string
	}{
		{
			name: "disabled stanza",
			slug: "monitor-var-log-disabled",
			why:  "disabled = 1 gates emission",
		},
		{
			name: "unsupported kind",
			slug: "tcp-ssl",
			why:  "[tcp-ssl:<port>] has no mapping yet",
		},
		{
			name: "unprefixed stanza name",
			slug: "my-bare-script",
			why:  "resolves under bin/<os>_<arch>/, so routing it through splunk_script would move the executable",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for id := range ids {
				require.NotContains(t, id, tt.slug, tt.why)
			}
		})
	}
}

// Every receiver gets the whole merged props/transforms set, matching how
// splunk_inputs feeds the same set to each sub-receiver.
func TestPropsTransformsParity(t *testing.T) {
	dirs := tabuilder.ConfRootDirs(confRoot)
	wantProps, err := tabuilder.ReadProps(dirs)
	require.NoError(t, err)
	wantTransforms, err := tabuilder.ReadTransforms(dirs)
	require.NoError(t, err)
	require.NotEmpty(t, wantProps, "fixture must carry props for this to mean anything")
	require.NotEmpty(t, wantTransforms)

	merged := resolve(t, confURI)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)

	sub, err := receivers.Sub("splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09")
	require.NoError(t, err)
	cfg := splunktcp.NewFactory().CreateDefaultConfig().(*splunktcp.Config)
	require.NoError(t, sub.Unmarshal(cfg))

	require.Equal(t, wantProps, cfg.Props)
	require.Equal(t, wantTransforms, cfg.Transforms)
}

func TestParseURI(t *testing.T) {
	for _, tt := range []struct {
		name         string
		uri          string
		wantRoot     string
		wantPipeline string
		errMsg       string
	}{
		{
			name:         "pipeline defaults to uf",
			uri:          "splunkconf://testdata/etc",
			wantRoot:     "testdata/etc",
			wantPipeline: "uf",
		},
		{
			name:         "pipeline query param overrides the default",
			uri:          "splunkconf://testdata/etc?pipeline=custom",
			wantRoot:     "testdata/etc",
			wantPipeline: "custom",
		},
		{
			// An absolute root leaves the host empty, so host+path has to be joined
			// for a relocated tree to work.
			name:         "absolute config root",
			uri:          "splunkconf:///opt/splunk/etc",
			wantRoot:     "/opt/splunk/etc",
			wantPipeline: "uf",
		},
		{
			name:   "wrong scheme",
			uri:    "file://testdata/etc",
			errMsg: `unexpected scheme "file"`,
		},
		{
			name:   "empty config root",
			uri:    "splunkconf://",
			errMsg: "empty config root",
		},
		{
			name:   "pipeline name outside the component-name charset",
			uri:    "splunkconf://testdata/etc?pipeline=Not/Valid",
			errMsg: `invalid pipeline "Not/Valid"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, pipeline, err := parseURI(tt.uri)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantRoot, root)
			require.Equal(t, tt.wantPipeline, pipeline)
		})
	}
}

// The pipeline name reaches the emitted fragment, not just parseURI.
func TestPipelineNameInFragment(t *testing.T) {
	merged := resolve(t, "splunkconf://testdata/etc?pipeline=custom")
	pipelines, err := merged.Sub("service::pipelines")
	require.NoError(t, err)
	require.Contains(t, pipelines.ToStringMap(), "logs/custom")
}

// A .conf tree the provider cannot map must fail config resolution. Returning a
// partial or empty pipeline would start a collector that silently collects less
// than the .conf asked for.
func TestRetrieveSurfacesErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		uri    string
		errMsg string
	}{
		{
			name:   "unusable uri",
			uri:    "splunkconf://",
			errMsg: "empty config root",
		},
		{
			// testdata itself has no system/local/inputs.conf.
			name:   "config root with no mappable inputs",
			uri:    "splunkconf://testdata",
			errMsg: "no supported input stanzas found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider{}.Retrieve(context.Background(), tt.uri, nil)
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestMapOutputs(t *testing.T) {
	for _, tt := range []struct {
		name   string
		merged conf.Map
		wantID string
		errMsg string
	}{
		{
			name: "hecout maps to the contrib exporter",
			merged: conf.Map{
				"hecout": {"uri": "https://hec:8088", "httpEventCollectorToken": "tok"},
			},
			wantID: "splunk_hec/hecout-d6a24a0d",
		},
		{
			// S2S needs a real exporter that does not exist upstream, so it arrives
			// with the output-mapper registration API rather than being guessed at.
			name: "tcpout has no mapping yet and must not be dropped silently",
			merged: conf.Map{
				"tcpout:primary": {"server": "idx1:9997"},
			},
			errMsg: `unsupported output stanza(s) [tcpout:primary] (kind "tcpout")`,
		},
		{
			name:   "no output stanzas at all",
			merged: conf.Map{},
			errMsg: conf.ErrNoOutputStanzas.Error(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := mapOutputs(tt.merged)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Len(t, out, 1)
			require.Equal(t, tt.wantID, out[0].id)
		})
	}
}

// The three monitor params that bind to typed wrapper fields rather than riding
// through Extra. TRUNCATE becomes an int, so it is the one that would break
// strict unmarshal if emitted as a string.
func TestEmitMonitorTypedParams(t *testing.T) {
	inputs := []conf.Input{
		{
			Configuration: conf.Configuration{
				Stanza: conf.Stanza{
					Name: "monitor:///var/log/typed",
					Params: conf.Params{
						{Name: "CHARSET", Value: "UTF-8"},
						{Name: "EVENT_BREAKER", Value: "BRK"},
						{Name: "TRUNCATE", Value: "1024"},
						{Name: "NOT_A_NUMBER", Value: "x"},
					},
				},
			},
		},
	}

	out, skipped, err := mapInputs(inputs, nil, nil)
	require.NoError(t, err)
	require.Empty(t, skipped)
	require.Len(t, out, 1)
	require.Equal(t, map[string]any{
		"path":          "/var/log/typed",
		"charset":       "UTF-8",
		"event_breaker": "BRK",
		"truncate":      1024,
		"NOT_A_NUMBER":  "x",
	}, out[0].cfg)
}

func TestParseListenAddress(t *testing.T) {
	for _, tt := range []struct {
		name     string
		target   string
		wantAddr string
		wantPort int
	}{
		{
			name:     "address and port",
			target:   "10.0.0.5:5514",
			wantAddr: "10.0.0.5",
			wantPort: 5514,
		},
		{
			name:     "port only means every interface",
			target:   "5514",
			wantAddr: "0.0.0.0",
			wantPort: 5514,
		},
		{
			name:     "empty address means every interface",
			target:   ":5514",
			wantAddr: "0.0.0.0",
			wantPort: 5514,
		},
		{
			// A zero port fails the wrapper's own Validate, so a malformed stanza
			// surfaces at config time rather than binding something arbitrary.
			name:     "out of range port",
			target:   "99999",
			wantAddr: "0.0.0.0",
			wantPort: 0,
		},
		{
			name:     "non-numeric port",
			target:   "0.0.0.0:nope",
			wantAddr: "0.0.0.0",
			wantPort: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			addr, port := parseListenAddress(tt.target)
			require.Equal(t, tt.wantAddr, addr)
			require.Equal(t, tt.wantPort, port)
		})
	}
}
