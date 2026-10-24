// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/exporter/splunkhecout"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkscript"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunktcp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkudp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkwineventlog"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
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
		"splunk_monitor/monitor-var-log-other-log-44426e37",
		"splunk_script/script-bin-collect-sh-55e5ff8f",
		"splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09",
		"splunk_udp/udp-1-5516-3417bfdb",
		"splunk_udp/udp-5515-016ce446",
		"splunk_wineventlog/wineventlog-application-9a1a225b",
	}
	wantExporters = []string{"splunk_hecout/hecout-d6a24a0d"}
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

// providerFactory registers the same mappers internal/settings does, which is
// the only way a stanza reaches the fragment.
func providerFactory() confmap.ProviderFactory {
	return splunkconf.NewFactory(
		splunkconf.WithInputMappers(
			splunkmonitor.NewFactory(),
			splunktcp.NewFactory(),
			splunkudp.NewFactory(),
			splunkscript.NewFactory(),
			splunkbatch.NewFactory(),
			splunkwineventlog.NewFactory(),
		),
		splunkconf.WithOutputMappers(splunkhecout.NewFactory()),
	)
}

func resolve(t *testing.T, uris ...string) *confmap.Conf {
	t.Helper()
	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: uris,
		ProviderFactories: []confmap.ProviderFactory{
			fileprovider.NewFactory(),
			providerFactory(),
		},
	})
	require.NoError(t, err)

	merged, err := res.Resolve(context.Background())
	require.NoError(t, err)
	return merged
}

// writeConfTree writes the named .conf files to a fresh etc/system/local and
// returns the config root to point a splunkconf:// URI at.
func writeConfTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "etc")
	local := filepath.Join(root, "system", "local")
	require.NoError(t, os.MkdirAll(local, 0o700))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(local, name), []byte(content), 0o600))
	}
	return root
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
	ef := splunkhecout.NewFactory()
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
				// app_dir is a filesystem path, so it carries the platform separator.
				"app_dir":  filepath.Join("testdata", "etc", "apps", "splunk_ta_demo"),
				"interval": "60",
				"index":    "app_scripts",
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
			// A bracketed IPv6 literal is a valid UF listen target, and its own
			// colons must not be read as the port separator. listen_address is
			// unbracketed because the receiver re-encodes it with net.JoinHostPort.
			name: "udp bracketed IPv6 literal keeps its address and port",
			id:   "splunk_udp/udp-1-5516-3417bfdb",
			want: map[string]any{
				"listen_address": "::1",
				"port":           5516,
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
			why:  "an unprefixed stanza carries global settings and declares no input",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for id := range ids {
				require.NotContains(t, id, tt.slug, tt.why)
			}
		})
	}
}

// A system receiver gets the system props/transforms set.
func TestPropsTransformsParity(t *testing.T) {
	dirs := tabuilder.SystemConfRootDirs(confRoot)
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

// propNames is the set of props stanza names a receiver's emitted config
// carries, which is what has to be scoped to the declaring app.
func propNames(t *testing.T, receivers *confmap.Conf, id string) []string {
	t.Helper()
	sub, err := receivers.Sub(id)
	require.NoError(t, err)
	var cfg struct {
		Rest  map[string]any `mapstructure:",remain"`
		Props []conf.Prop    `mapstructure:"props"`
	}
	require.NoError(t, sub.Unmarshal(&cfg))
	names := make([]string, 0, len(cfg.Props))
	for _, prop := range cfg.Props {
		names = append(names, prop.Name)
	}
	return names
}

// An app's receivers resolve props against that app's own layers merged with
// system, which is what splunkd gives them. Resolving once across the whole
// configuration root instead applies one app's props to another app's data.
func TestPropsScopedToDeclaringApp(t *testing.T) {
	merged := resolve(t, confURI)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)

	const (
		demoOnly  = "demo_sourcetype"
		otherOnly = "other_sourcetype"
		systemOne = "syslog"
	)

	for _, tt := range []struct {
		name    string
		id      string
		wants   string
		missing string
	}{
		{
			name:    "app receiver gets its own props",
			id:      "splunk_script/script-bin-collect-sh-55e5ff8f",
			wants:   demoOnly,
			missing: otherOnly,
		},
		{
			name:    "sibling app receiver gets its own, not the other app's",
			id:      "splunk_monitor/monitor-var-log-other-log-44426e37",
			wants:   otherOnly,
			missing: demoOnly,
		},
		{
			name:    "system receiver gets no app props at all",
			id:      "splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09",
			wants:   systemOne,
			missing: demoOnly,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			names := propNames(t, receivers, tt.id)
			require.Contains(t, names, systemOne, "system props reach every receiver")
			require.Contains(t, names, tt.wants)
			require.NotContains(t, names, tt.missing)
		})
	}
}

// A literal $ is legal in a .conf value, and the config resolver reads a lone $
// as a reference, so the provider emits it doubled and the resolver gives back
// the original value.
func TestDollarInConfValues(t *testing.T) {
	uri := "splunkconf://" + writeConfTree(t, map[string]string{
		"inputs.conf":  "[monitor:///var/log/app$dollar.log]\nblacklist = \\.gz$done\n",
		"outputs.conf": "[hecout:hec]\nuri = https://localhost:8088\nhttpEventCollectorToken = secret$value\n",
	})
	// The ID is spelled out rather than recomputed: an external test should not
	// reach for the unexported name rule it is meant to pin.
	recvID := splunkmonitor.TypeStr + "/monitor-var-log-app-dollar-log-73ffe9d8"

	retrieved, err := providerFactory().Create(confmap.ProviderSettings{}).Retrieve(context.Background(), uri, nil)
	require.NoError(t, err)
	raw, err := retrieved.AsRaw()
	require.NoError(t, err)
	emitted := raw.(map[string]any)["receivers"].(map[string]any)[recvID].(map[string]any)
	require.Equal(t, "/var/log/app$$dollar.log", emitted["path"])

	merged := resolve(t, uri)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)
	recv, err := receivers.Sub(recvID)
	require.NoError(t, err)
	require.Equal(t, "/var/log/app$dollar.log", recv.ToStringMap()["path"])
	require.Equal(t, `\.gz$done`, recv.ToStringMap()["blacklist"])

	exporters, err := merged.Sub("exporters")
	require.NoError(t, err)
	exps := exporters.ToStringMap()
	require.Len(t, exps, 1)
	for _, cfg := range exps {
		require.Equal(t, "secret$value", cfg.(map[string]any)["token"])
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
	const validInputs = "[monitor:///var/log/foo]\n"
	const validOutputs = "[hecout]\nuri = https://localhost:8088\n"
	const malformed = "[stanza]\nno delimiter here\n"

	for _, tt := range []struct {
		files  map[string]string
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
			name:   "undecodable query",
			uri:    "splunkconf://testdata/etc?pipeline=%zz",
			errMsg: "invalid query",
		},
		{
			// testdata itself holds neither conf file, and only the output side
			// still refuses to resolve.
			name:   "config root with neither inputs nor outputs",
			uri:    "splunkconf://testdata",
			errMsg: "no output stanzas found in outputs.conf",
		},
		{
			name:   "malformed inputs.conf",
			files:  map[string]string{"inputs.conf": malformed},
			errMsg: "read inputs.conf",
		},
		{
			name:   "malformed props.conf",
			files:  map[string]string{"inputs.conf": validInputs, "props.conf": malformed},
			errMsg: "read props.conf",
		},
		{
			name:   "malformed transforms.conf",
			files:  map[string]string{"inputs.conf": validInputs, "transforms.conf": malformed},
			errMsg: "read transforms.conf",
		},
		{
			name:   "malformed outputs.conf",
			files:  map[string]string{"inputs.conf": validInputs, "outputs.conf": malformed},
			errMsg: "read outputs.conf",
		},
		{
			name:   "input stanza with no kind",
			files:  map[string]string{"inputs.conf": "[://nokind]\n", "outputs.conf": validOutputs},
			errMsg: "parse stanza",
		},
		{
			name:   "output stanza with no kind",
			files:  map[string]string{"inputs.conf": validInputs, "outputs.conf": "[://nokind]\n"},
			errMsg: "parse output stanza",
		},
		{
			name:   "output kind with no mapping",
			files:  map[string]string{"inputs.conf": validInputs, "outputs.conf": "[tcpout:primary]\nserver = idx1:9997\n"},
			errMsg: "no supported output stanzas found",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			uri := tt.uri
			if tt.files != nil {
				uri = "splunkconf://" + writeConfTree(t, tt.files)
			}
			p := providerFactory().Create(confmap.ProviderSettings{})
			_, err := p.Retrieve(context.Background(), uri, nil)
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

// A tree with nothing to read reads from nop instead of refusing to start,
// because an idle agent is a normal state. The outputs still resolve.
func TestNoUsableInputsReadsFromNop(t *testing.T) {
	const outputs = "[hecout]\nuri = https://localhost:8088\n"

	for _, tt := range []struct {
		name   string
		inputs string
	}{
		{name: "inputs.conf absent"},
		{name: "no stanzas", inputs: "# nothing here\n"},
		{name: "every stanza disabled", inputs: "[monitor:///var/log/foo]\ndisabled = 1\n"},
		{name: "only unmapped kinds", inputs: "[http://tok]\ntoken = abc\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"outputs.conf": outputs}
			if tt.inputs != "" {
				files["inputs.conf"] = tt.inputs
			}
			resolved := resolve(t, "splunkconf://"+writeConfTree(t, files))

			recvIDs, expIDs := pipelineIDs(t, resolved, "logs/uf")
			require.Equal(t, []string{"nop"}, recvIDs)
			require.Equal(t, []string{"splunk_hecout/hecout-d6a24a0d"}, expIDs)

			receivers, err := resolved.Sub("receivers")
			require.NoError(t, err)
			require.Equal(t, map[string]any{"nop": map[string]any{}}, receivers.ToStringMap())
		})
	}
}

// s2sConfig stands in for an exporter config whose type lives outside this repo.
type s2sConfig struct {
	Server string `mapstructure:"server"`
}

// s2sFactory maps [tcpout] the way the closed-source S2S exporter does, which is
// the registration path WithOutputMappers exists for.
type s2sFactory struct {
	exporter.Factory
}

func newS2SFactory() s2sFactory {
	return s2sFactory{Factory: exporter.NewFactory(
		otelcomponent.MustNewType("splunk_s2s"),
		func() otelcomponent.Config { return &s2sConfig{} },
	)}
}

func (s2sFactory) MapOutputs(outputs []splunkconf.Output) (map[string]otelcomponent.Config, error) {
	instances := map[string]otelcomponent.Config{}
	for _, o := range outputs {
		if o.Name.Kind != "tcpout" {
			continue
		}
		cfg := &s2sConfig{}
		if p := o.Configuration.Stanza.Params.Get("server"); p != nil {
			cfg.Server = p.Value
		}
		instances[o.Configuration.Stanza.Name] = cfg
	}
	return instances, nil
}

// An exporter registered from outside this package reaches outputs.conf through
// the same path the built-in [hecout] mapping does, and its own config type
// survives the marshal/unmarshal round trip the emitted fragment forces.
func TestRegisteredOutputMapper(t *testing.T) {
	uri := "splunkconf://" + writeConfTree(t, map[string]string{
		"inputs.conf":  "[monitor:///var/log/foo]\n",
		"outputs.conf": "[hecout]\nuri = https://localhost:8088\n\n[tcpout:primary]\nserver = idx1:9997\n",
	})

	f := newS2SFactory()
	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{uri},
		ProviderFactories: []confmap.ProviderFactory{splunkconf.NewFactory(
			splunkconf.WithInputMappers(splunkmonitor.NewFactory()),
			splunkconf.WithOutputMappers(splunkhecout.NewFactory(), f),
		)},
	})
	require.NoError(t, err)
	resolved, err := res.Resolve(context.Background())
	require.NoError(t, err)

	const s2sID = "splunk_s2s/tcpout-primary-bc98bd6f"
	_, expIDs := pipelineIDs(t, resolved, "logs/uf")
	require.Equal(t, []string{"splunk_hecout/hecout-d6a24a0d", s2sID}, expIDs)

	exporters, err := resolved.Sub("exporters")
	require.NoError(t, err)
	sub, err := exporters.Sub(s2sID)
	require.NoError(t, err)
	cfg := f.CreateDefaultConfig()
	require.NoError(t, sub.Unmarshal(cfg))
	require.Equal(t, &s2sConfig{Server: "idx1:9997"}, cfg)
}

func TestUnclaimedOutputStanzaIsSkipped(t *testing.T) {
	uri := "splunkconf://" + writeConfTree(t, map[string]string{
		"inputs.conf":  "[monitor:///var/log/foo]\n",
		"outputs.conf": "[hecout]\nuri = https://localhost:8088\n\n[tcpout:primary]\nserver = idx1:9997\n",
	})

	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{uri},
		ProviderFactories: []confmap.ProviderFactory{splunkconf.NewFactory(
			splunkconf.WithInputMappers(splunkmonitor.NewFactory()),
			splunkconf.WithOutputMappers(splunkhecout.NewFactory()),
		)},
	})
	require.NoError(t, err)
	resolved, err := res.Resolve(context.Background())
	require.NoError(t, err)

	_, expIDs := pipelineIDs(t, resolved, "logs/uf")
	require.Equal(t, []string{"splunk_hecout/hecout-d6a24a0d"}, expIDs)
}

func TestShutdownIsANoop(t *testing.T) {
	require.NoError(t, providerFactory().Create(confmap.ProviderSettings{}).Shutdown(context.Background()))
}

// A stanza param must not overwrite a key the mapper already emitted: the
// stanza name is the only source of the target.
func TestParamDoesNotOverrideEmittedKey(t *testing.T) {
	uri := "splunkconf://" + writeConfTree(t, map[string]string{
		"inputs.conf":  "[monitor:///var/log/foo]\npath = /var/log/elsewhere\n",
		"outputs.conf": "[hecout]\nuri = https://localhost:8088\n",
	})

	receivers, err := resolve(t, uri).Sub("receivers")
	require.NoError(t, err)
	recv, err := receivers.Sub(splunkmonitor.TypeStr + "/monitor-var-log-foo-31463e8f")
	require.NoError(t, err)
	require.Equal(t, "/var/log/foo", recv.ToStringMap()["path"])
}

// inputs.conf globals configure the instance stanzas rather than declaring
// inputs of their own, so they reach the emitted receivers and emit nothing
// themselves. [monitor] is not one: UF ignores it, per the global-stanzas
// parity case, so it stays an ordinary stanza and is reported unsupported.
func TestGlobalStanzasReachEmittedReceivers(t *testing.T) {
	uri := "splunkconf://" + writeConfTree(t, map[string]string{
		"inputs.conf": "host = from_top\n" +
			"\n[default]\nsourcetype = from_default\n" +
			"\n[monitor:///var/log/a]\n" +
			"\n[monitor:///var/log/b]\nsourcetype = own_st\n",
		"outputs.conf": "[hecout]\nuri = https://localhost:8088\nhttpEventCollectorToken = tok\n",
	})

	merged := resolve(t, uri)
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)

	ids := receivers.ToStringMap()
	require.Len(t, ids, 2, "only the two instance stanzas are inputs: %v", ids)

	for _, tt := range []struct {
		id             string
		wantSourcetype string
	}{
		{id: "splunk_monitor/monitor-var-log-a-77c17b38", wantSourcetype: "from_default"},
		{id: "splunk_monitor/monitor-var-log-b-82d2526a", wantSourcetype: "own_st"},
	} {
		require.Contains(t, ids, tt.id)
		sub, subErr := receivers.Sub(tt.id)
		require.NoError(t, subErr)
		cfg := sub.ToStringMap()
		require.Equal(t, "from_top", cfg["host"], "a pre-stanza global reaches the receiver")
		require.Equal(t, tt.wantSourcetype, cfg["sourcetype"], "the instance's own value outranks [default]")
	}
}
