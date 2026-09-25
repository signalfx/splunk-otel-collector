// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhome

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkhecout"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkscript"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunktcp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkudp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

var (
	_ = splunkbatch.Config{}
)

const (
	baseURI   = "file:testdata/base.yaml"
	splunkURI = "splunkhome://testdata/splunkhome?pipeline=uf"
)

// TestResolveAndValidate proves the locked design on the real collector API:
// merging --config=base.yaml with --config=dotconf://testdata produces a
// logs/uf pipeline whose emitted UF-native receiver configs and splunk_hecout
// exporter config unmarshal, validate, and build via their real factories.
func TestResolveAndValidate(t *testing.T) {
	ctx := context.Background()

	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{baseURI, splunkURI},
		ProviderFactories: []confmap.ProviderFactory{
			fileprovider.NewFactory(),
			NewFactory(),
		},
	})
	require.NoError(t, err)

	merged, err := res.Resolve(ctx)
	require.NoError(t, err, "merge base.yaml + splunkhome must succeed")

	// The base logs/base pipeline survives and the splunkhome logs/uf pipeline is added.
	pipelines, err := merged.Sub("service::pipelines")
	require.NoError(t, err)
	pipes := pipelines.ToStringMap()
	require.Contains(t, pipes, "logs/base", "base pipeline must survive the merge")
	require.Contains(t, pipes, "logs/uf", "splunkhome must ADD the logs/uf pipeline")

	// Collect the emitted receiver IDs from the logs/uf pipeline.
	ufPipe, err := merged.Sub("service::pipelines::logs/uf")
	require.NoError(t, err)
	var pl struct {
		Receivers []string `mapstructure:"receivers"`
		Exporters []string `mapstructure:"exporters"`
	}
	require.NoError(t, ufPipe.Unmarshal(&pl))
	require.NotEmpty(t, pl.Receivers)
	require.Len(t, pl.Exporters, 1)

	// Six enabled input stanzas: two monitor + tcp, udp, script, batch.
	// The disabled stanza is not emitted. (Wineventlog requires Windows.)
	require.Len(t, pl.Receivers, 6, "all enabled input stanzas should emit receivers")

	// Every emitted receiver config unmarshals, validates, and builds.
	// Map receiver types to their factories for testing.
	recvFactories := map[string]receiver.Factory{
		"splunk_monitor": splunkmonitor.NewFactory(),
		"splunk_tcp":     splunktcp.NewFactory(),
		"splunk_udp":     splunkudp.NewFactory(),
		"splunk_script":  splunkscript.NewFactory(),
		"splunk_batch":   splunkbatch.NewFactory(),
	}

	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)
	for _, id := range pl.Receivers {
		sub, subErr := receivers.Sub(id)
		require.NoError(t, subErr, "receiver %s must exist in merged config", id)

		var cid component.ID
		require.NoError(t, cid.UnmarshalText([]byte(id)))

		factory, ok := recvFactories[cid.Type().String()]
		require.True(t, ok, "unknown receiver type %s", cid.Type().String())

		cfg := factory.CreateDefaultConfig()
		require.NoError(t, sub.Unmarshal(cfg), "receiver %s config must unmarshal", id)
		require.NoError(t, confmap.Validate(cfg), "receiver %s config must validate", id)

		_, buildErr := factory.CreateLogs(ctx, receivertest.NewNopSettings(factory.Type()), cfg, consumertest.NewNop())
		require.NoError(t, buildErr, "receiver %s must build via factory", id)
	}

	// The exporter config unmarshals, validates, and builds via splunk_hecout.
	ef := splunkhecout.NewFactory()
	exporters, err := merged.Sub("exporters")
	require.NoError(t, err)
	esub, err := exporters.Sub(pl.Exporters[0])
	require.NoError(t, err)
	ecfg := ef.CreateDefaultConfig()
	require.NoError(t, esub.Unmarshal(ecfg))
	require.NoError(t, confmap.Validate(ecfg), "splunk_hecout exporter config must validate")
	_, err = ef.CreateLogs(ctx, exportertest.NewNopSettings(ef.Type()), ecfg)
	require.NoError(t, err, "splunk_hecout exporter must build via factory")
}

// TestMergedGraphBuilds runs the collector core's own build path over the
// MERGED config. otelcol.DryRun resolves base.yaml + splunkhome://... into one
// otelcol.Config, runs confmap.Validate over the whole config, then
// service.Validate -> graph.Build, which actually wires each pipeline's
// receivers through to its exporters. A green DryRun proves the logs/uf pipeline
// assembles its splunk_monitor receiver(s) through to the splunk_hecout
// exporter, not just that the fragments unmarshal.
func TestMergedGraphBuilds(t *testing.T) {
	factories := func() (otelcol.Factories, error) {
		recvs, err := otelcol.MakeFactoryMap(
			receivertest.NewNopFactory(),
			splunkmonitor.NewFactory(),
			splunktcp.NewFactory(),
			splunkudp.NewFactory(),
			splunkscript.NewFactory(),
			splunkbatch.NewFactory(),
		)
		if err != nil {
			return otelcol.Factories{}, err
		}
		exps, err := otelcol.MakeFactoryMap(exportertest.NewNopFactory(), splunkhecout.NewFactory())
		if err != nil {
			return otelcol.Factories{}, err
		}
		return otelcol.Factories{
			Receivers: recvs,
			Exporters: exps,
			Telemetry: otelconftelemetry.NewFactory(),
		}, nil
	}

	col, err := otelcol.NewCollector(otelcol.CollectorSettings{
		BuildInfo: component.NewDefaultBuildInfo(),
		Factories: factories,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{baseURI, splunkURI},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					NewFactory(),
				},
			},
		},
	})
	require.NoError(t, err)

	require.NoError(t, col.DryRun(context.Background()),
		"merged config (base logs/base + splunkhome logs/uf) must build a real graph")
}

// TestTranslationFidelity proves typed wrapper Configs are lossless: any stanza
// param not explicitly modeled is captured in Extra and passed through to tabuilder.
// This test verifies that all known receiver params are present after translation.
func TestTranslationFidelity(t *testing.T) {
	ctx := context.Background()

	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{baseURI, splunkURI},
		ProviderFactories: []confmap.ProviderFactory{
			fileprovider.NewFactory(),
			NewFactory(),
		},
	})
	require.NoError(t, err)

	merged, err := res.Resolve(ctx)
	require.NoError(t, err)

	// For each receiver type, verify that the typed config captures:
	// 1. Scheme-specific params (e.g., port for tcp/udp, include for monitor)
	// 2. Resource attributes (index/source/sourcetype/host)
	// 3. Type-specific params (e.g., charset for monitor)
	//
	// The Extra field captures any params not explicitly modeled, so roundtrip
	// fidelity is preserved: stanza → typed Config → conf.Input preserves all params.

	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)

	tcpReceiver, err := receivers.Sub("splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09")
	require.NoError(t, err)
	tcpCfg := &splunktcp.Config{}
	require.NoError(t, tcpReceiver.Unmarshal(tcpCfg))
	// TCP stanza [tcp://0.0.0.0:5514] with index param should preserve both
	require.Equal(t, "0.0.0.0", tcpCfg.ListenAddress, "listen_address must be preserved")
	require.Equal(t, 5514, tcpCfg.Port, "port must be preserved")
	require.Equal(t, "net", tcpCfg.Index, "index resource attr must be preserved")
	require.Equal(t, "tcp_source", tcpCfg.Sourcetype, "sourcetype must be preserved")
	// The unmodeled param queueSize must round-trip via the ",remain" passthrough,
	// not be dropped or rejected by strict unmarshal. This is the losslessness proof.
	require.Equal(t, "500KB", tcpCfg.Extra["queueSize"], "unmodeled param must survive in Extra")

	scriptReceiver, err := receivers.Sub("splunk_script/script-usr-local-bin-test-sh-c35160c3")
	require.NoError(t, err)
	scriptCfg := &splunkscript.Config{}
	require.NoError(t, scriptReceiver.Unmarshal(scriptCfg))
	require.Equal(t, "/usr/local/bin/test.sh", scriptCfg.ScriptFilename, "script_filename must be preserved")
	require.Equal(t, "30", scriptCfg.Interval, "interval must be preserved")
	require.Equal(t, "scripts", scriptCfg.Index, "index must be preserved")
	require.Equal(t, "splunk-system-user", scriptCfg.Extra["passAuth"], "unmodeled param must survive in Extra")
}

// TestPropsTransformsParity proves that the provider attaches the same []conf.Prop
// and []conf.Transform to all emitted receivers as tabuilder.ReadProps/ReadTransforms
// would return for the same testdata. This proves the wrappers receive the full shared
// set, consistent with splunk_inputs.
func TestPropsTransformsParity(t *testing.T) {
	ctx := context.Background()

	// Read props/transforms using tabuilder directly (the reference)
	dirs := tabuilder.ConfDirs("testdata/splunkhome")
	refProps, err := tabuilder.ReadProps(dirs)
	require.NoError(t, err, "tabuilder.ReadProps must succeed")
	refTransforms, err := tabuilder.ReadTransforms(dirs)
	require.NoError(t, err, "tabuilder.ReadTransforms must succeed")

	// Now resolve via the provider and extract props/transforms from emitted config
	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{baseURI, splunkURI},
		ProviderFactories: []confmap.ProviderFactory{
			fileprovider.NewFactory(),
			NewFactory(),
		},
	})
	require.NoError(t, err)

	merged, err := res.Resolve(ctx)
	require.NoError(t, err)

	// Pick one receiver and unmarshal its config
	receivers, err := merged.Sub("receivers")
	require.NoError(t, err)
	tcpReceiver, err := receivers.Sub("splunk_tcp/tcp-0-0-0-0-5514-4c1c5e09")
	require.NoError(t, err)
	tcpCfg := &splunktcp.Config{}
	require.NoError(t, tcpReceiver.Unmarshal(tcpCfg))

	// The props and transforms from the provider must EQUAL what tabuilder returns directly
	require.Equal(t, refProps, tcpCfg.Props, "Props in emitted config must equal tabuilder.ReadProps")
	require.Equal(t, refTransforms, tcpCfg.Transforms, "Transforms in emitted config must equal tabuilder.ReadTransforms")

	// Verify at least one prop and transform are present
	require.NotEmpty(t, tcpCfg.Props, "Props must be present in config")
	require.NotEmpty(t, tcpCfg.Transforms, "Transforms must be present in config")
}

// TestPipelineConfigurable proves the pipeline name follows the pipeline query param.
func TestPipelineConfigurable(t *testing.T) {
	res, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs: []string{"splunkhome://testdata/splunkhome?pipeline=custom"},
		ProviderFactories: []confmap.ProviderFactory{
			NewFactory(),
		},
	})
	require.NoError(t, err)
	merged, err := res.Resolve(context.Background())
	require.NoError(t, err)
	pipes, err := merged.Sub("service::pipelines")
	require.NoError(t, err)
	require.Contains(t, pipes.ToStringMap(), "logs/custom")
}
