// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	otelcolextension "go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pipeline"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/cpu"
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"

	"github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension"
)

func cleanUp() func() {
	jh := "JAVA_HOME"
	existing, ok := os.LookupEnv(jh)
	os.Unsetenv(jh)
	return func() {
		configureEnvironmentOnce = sync.Once{}
		if ok {
			os.Setenv(jh, existing)
		}
	}
}

func newReceiverCreateSettings(name string, t *testing.T) otelcolreceiver.Settings {
	return otelcolreceiver.Settings{
		ID: component.MustNewIDWithName("smartagent", name),
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.NewNop(),
			TracerProvider: nooptrace.NewTracerProvider(),
			MeterProvider:  noopmetric.NewMeterProvider(),
		},
	}
}

var (
	expectedCPUMetrics = map[string]pmetric.MetricType{
		"cpu.idle":                 pmetric.MetricTypeSum,
		"cpu.interrupt":            pmetric.MetricTypeSum,
		"cpu.nice":                 pmetric.MetricTypeSum,
		"cpu.num_processors":       pmetric.MetricTypeGauge,
		"cpu.softirq":              pmetric.MetricTypeSum,
		"cpu.steal":                pmetric.MetricTypeSum,
		"cpu.system":               pmetric.MetricTypeSum,
		"cpu.user":                 pmetric.MetricTypeSum,
		"cpu.utilization":          pmetric.MetricTypeGauge,
		"cpu.utilization_per_core": pmetric.MetricTypeGauge,
		"cpu.wait":                 pmetric.MetricTypeSum,
	}

	partialSettingsID = component.MustNewIDWithName(typeStr, "partial_settings")
	extraSettingsID   = component.MustNewIDWithName(typeStr, "extra")
)

func newConfig(monitorType string, intervalSeconds int) Config {
	return Config{
		MonitorType: monitorType,
		monitorConfig: &cpu.Config{
			MonitorConfig: saconfig.MonitorConfig{
				Type:            monitorType,
				IntervalSeconds: intervalSeconds,
				ExtraDimensions: map[string]string{
					"required_dimension": "required_value",
				},
			},
			ReportPerCPU: true,
		},
	}
}

func TestSmartAgentReceiver(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 10)
	consumer := new(consumertest.MetricsSink)
	receiver := newReceiver(newReceiverCreateSettings("valid", t), cfg)
	receiver.registerMetricsConsumer(consumer)

	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))

	assert.EqualValues(t, "smartagentvalid", cfg.monitorConfig.MonitorConfigCore().MonitorID)
	monitor, isMonitor := receiver.monitor.(*cpu.Monitor)
	require.True(t, isMonitor)

	monitorOutput := monitor.Output
	_, isOutput := monitorOutput.(*output)
	assert.True(t, isOutput)

	assert.Eventuallyf(t, func() bool {
		// confirm single occurrence of total metrics as sanity in lieu of
		// out of scope cpu monitor verification.
		seenTotalMetric := map[string]bool{}

		allMetrics := consumer.AllMetrics()
		for _, m := range allMetrics {
			resourceMetrics := m.ResourceMetrics()
			for i := 0; i < resourceMetrics.Len(); i++ {
				resourceMetric := resourceMetrics.At(i)
				instrumentationLibraryMetrics := resourceMetric.ScopeMetrics()
				for j := 0; j < instrumentationLibraryMetrics.Len(); j++ {
					instrumentationLibraryMetric := instrumentationLibraryMetrics.At(j)
					metrics := instrumentationLibraryMetric.Metrics()
					for k := 0; k < metrics.Len(); k++ {
						metricInst := metrics.At(k)
						name := metricInst.Name()
						dataType := metricInst.Type()
						expectedDataType := expectedCPUMetrics[name]
						require.NotEqual(t, pmetric.MetricTypeEmpty, expectedDataType, "received unexpected none type for %s", name)
						assert.Equal(t, expectedDataType, dataType)
						var attributes pcommon.Map
						switch dataType {
						case pmetric.MetricTypeGauge:
							dg := metricInst.Gauge()
							for l := 0; l < dg.DataPoints().Len(); l++ {
								dgdp := dg.DataPoints().At(l)
								attributes = dgdp.Attributes()
								val := dgdp.DoubleValue()
								assert.NotEqual(t, val, 0, "invalid value of MetricTypeGauge metric %s", name)
							}
						case pmetric.MetricTypeSum:
							ds := metricInst.Sum()
							for l := 0; l < ds.DataPoints().Len(); l++ {
								dsdp := ds.DataPoints().At(l)
								attributes = dsdp.Attributes()
								val := dsdp.DoubleValue()
								assert.NotEqual(t, val, 0, "invalid value of MetricTypeSum metric %s", name)
							}
						default:
							t.Errorf("unexpected type %#v for metric %s", metricInst.Type(), name)
						}

						labelVal, ok := attributes.Get("required_dimension")
						require.True(t, ok)
						assert.Equal(t, "required_value", labelVal.Str())

						systemType, ok := attributes.Get("system.type")
						require.True(t, ok)
						assert.Equal(t, "cpu", systemType.Str())

						// mark metric as having been seen
						cpuNum, _ := attributes.Get("cpu")
						seenName := fmt.Sprintf("%s%s", name, cpuNum.Str())
						assert.False(t, seenTotalMetric[seenName], "unexpectedly repeated metric: %v", seenName)
						seenTotalMetric[seenName] = true
					}
				}
			}
		}
		return len(allMetrics) > 0
	}, 5*time.Second, 1*time.Millisecond, "failed to receive expected cpu metrics")

	metrics := consumer.AllMetrics()
	assert.Greater(t, len(metrics), 0)
	assert.NoError(t, receiver.Shutdown(context.Background()))
}

func TestStripMonitorTypePrefix(t *testing.T) {
	assert.Equal(t, "nginx", stripMonitorTypePrefix("collectd/nginx"))
	assert.Equal(t, "cpu", stripMonitorTypePrefix("cpu"))
}

func TestStartReceiverWithUnknownMonitorType(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("notamonitortype", 1)
	receiver := newReceiver(newReceiverCreateSettings("invalid", t), cfg)
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	assert.EqualError(t, err,
		"failed creating monitor \"notamonitortype\": unable to find MonitorFactory for \"notamonitortype\"",
	)
}

func TestStartAndShutdown(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 1)
	receiver := newReceiver(newReceiverCreateSettings("valid", t), cfg)
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = receiver.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestOutOfOrderShutdownInvocations(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 1)
	receiver := newReceiver(newReceiverCreateSettings("valid", t), cfg)

	err := receiver.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestMultipleInstancesOfSameMonitorType(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 1)
	fstRcvr := newReceiver(newReceiverCreateSettings("valid", t), cfg)

	ctx := context.Background()

	require.NoError(t, fstRcvr.Start(ctx, componenttest.NewNopHost()))
	require.NoError(t, fstRcvr.Shutdown(ctx))

	sndRcvr := newReceiver(newReceiverCreateSettings("valid", t), cfg)
	assert.NoError(t, sndRcvr.Start(ctx, componenttest.NewNopHost()))
	assert.NoError(t, sndRcvr.Shutdown(ctx))
}

func TestInvalidMonitorStateAtShutdown(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 1)
	receiver := newReceiver(newReceiverCreateSettings("valid", t), cfg)
	receiver.monitor = new(any)

	err := receiver.Shutdown(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid monitor state at Shutdown(): (*interface {})")
}

func TestConfirmStartingReceiverWithInvalidMonitorInstancesDoesntPanic(t *testing.T) {
	t.Cleanup(cleanUp())
	tests := []struct {
		name           string
		monitorFactory func() any
		expectedError  string
	}{
		{"anonymous struct", func() any { return struct{}{} }, ""},
		{"anonymous struct pointer", func() any { return &struct{}{} }, ""},
		{"nil interface pointer", func() any { return new(any) }, ": invalid struct instance: (*interface {})"},
		{"nil", func() any { return nil }, ": invalid struct instance: <nil>"},
		{"boolean", func() any { return false }, ": invalid struct instance: false"},
		{"string", func() any { return "asdf" }, ": invalid struct instance: \"asdf\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			monitors.MonitorFactories["notarealmonitor"] = test.monitorFactory
			monitors.MonitorMetadatas["notarealmonitor"] = &monitors.Metadata{MonitorType: "notarealmonitor"}

			cfg := newConfig("notarealmonitor", 123)
			receiver := newReceiver(newReceiverCreateSettings("invalid", t), cfg)
			err := receiver.Start(context.Background(), componenttest.NewNopHost())
			require.Error(tt, err)
			assert.Contains(tt, err.Error(),
				fmt.Sprintf("failed creating monitor \"notarealmonitor\": unable to set output field of monitor%s", test.expectedError),
			)
		})
	}
}

func TestFilteringNoMetadata(t *testing.T) {
	t.Cleanup(cleanUp())
	monitors.MonitorFactories["fakemonitor"] = func() any { return struct{}{} }
	cfg := newConfig("fakemonitor", 1)
	receiver := newReceiver(newReceiverCreateSettings("valid", t), cfg)
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	require.EqualError(t, err, "failed creating monitor \"fakemonitor\": could not find monitor metadata of type fakemonitor")
}

func TestSmartAgentConfigProviderOverrides(t *testing.T) {
	t.Cleanup(cleanUp())
	cfg := newConfig("cpu", 1)
	observedLogger, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(observedLogger)
	rcs := newReceiverCreateSettings("valid", t)
	rcs.Logger = logger
	r := newReceiver(rcs, cfg)

	configs := getSmartAgentExtensionConfig(t)
	host := &mockHost{
		smartagentextensionConfig:      configs[0],
		smartagentextensionConfigExtra: configs[1],
	}

	require.NoError(t, r.Start(context.Background(), host))
	require.NoError(t, r.Shutdown(context.Background()))
	require.True(t, func() bool {
		for _, message := range logs.All() {
			if strings.HasPrefix(message.Message, "multiple smartagent extensions found, using ") {
				return true
			}
		}
		return false
	}())
	require.Equal(t, saconfig.CollectdConfig{
		Timeout:             10,
		ReadThreads:         1,
		WriteThreads:        4,
		WriteQueueLimitHigh: 5,
		WriteQueueLimitLow:  400000,
		LogLevel:            "notice",
		IntervalSeconds:     10,
		WriteServerIPAddr:   "127.9.8.7",
		WriteServerPort:     0,
		ConfigDir:           filepath.Join("/opt", "run", "collectd"),
		BundleDir:           "/opt/",
		InstanceName:        "",
		WriteServerQuery:    "",
	}, saConfig.Collectd)

	if runtime.GOOS == "windows" {
		require.NotEqual(t, filepath.Join("/opt", "jre"), os.Getenv("JAVA_HOME"))
	} else {
		require.Equal(t, filepath.Join("/opt", "jre"), os.Getenv("JAVA_HOME"))
	}

	require.Equal(t, "/proc", hostfs.HostProc())
	require.Equal(t, "/sys", hostfs.HostSys())
	require.Equal(t, "/run", hostfs.HostRun())
	require.Equal(t, "/var", hostfs.HostVar())
	require.Equal(t, "/etc", hostfs.HostEtc())
}

func TestJavaHomeRespected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-windows only")
	}
	t.Cleanup(cleanUp())
	os.Setenv("JAVA_HOME", "/existing/java/home")
	cfg := newConfig("cpu", 1)
	rcs := newReceiverCreateSettings("valid", t)
	r := newReceiver(rcs, cfg)

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, r.Shutdown(context.Background()))

	require.Equal(t, "/existing/java/home", os.Getenv("JAVA_HOME"))
}

func getSmartAgentExtensionConfig(t *testing.T) []*smartagentextension.Config {
	cfg, err := confmaptest.LoadConf(path.Join(".", "testdata", "extension_config.yaml"))
	require.NoError(t, err)

	cm, err := cfg.Sub(partialSettingsID.String())
	require.NoError(t, err)
	partialSettingsConfig := smartagentextension.NewFactory().CreateDefaultConfig().(*smartagentextension.Config)
	err = cm.Unmarshal(&partialSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, partialSettingsConfig)

	cm, err = cfg.Sub(partialSettingsID.String())
	require.NoError(t, err)
	extraSettingsConfig := smartagentextension.NewFactory().CreateDefaultConfig().(*smartagentextension.Config)
	err = cm.Unmarshal(&extraSettingsConfig)
	require.NoError(t, err)
	require.NotNil(t, extraSettingsConfig)

	return []*smartagentextension.Config{partialSettingsConfig, extraSettingsConfig}
}

type mockHost struct {
	smartagentextensionConfig      *smartagentextension.Config
	smartagentextensionConfigExtra *smartagentextension.Config
}

func (m *mockHost) ReportFatalError(error) {
}

func (m *mockHost) GetFactory(component.Kind, component.Type) component.Factory {
	return nil
}

func (m *mockHost) GetExtensions() map[component.ID]component.Component {
	exampleFactory := extensiontest.NewNopFactory()
	exampleID := component.MustNewID(exampleFactory.Type().String())
	randomExtensionConfig := exampleFactory.CreateDefaultConfig()
	return map[component.ID]component.Component{
		partialSettingsID: getExtension(smartagentextension.NewFactory(), partialSettingsID, m.smartagentextensionConfig),
		exampleID:         getExtension(exampleFactory, exampleID, randomExtensionConfig),
		extraSettingsID:   getExtension(smartagentextension.NewFactory(), extraSettingsID, m.smartagentextensionConfigExtra),
	}
}

func getExtension(f otelcolextension.Factory, id component.ID, cfg component.Config) component.Component {
	e, err := f.Create(context.Background(), otelcolextension.Settings{ID: id}, cfg)
	if err != nil {
		panic(err)
	}
	return e
}

func (m *mockHost) GetExporters() map[pipeline.Signal]map[component.ID]component.Component {
	return nil
}
