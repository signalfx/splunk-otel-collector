// Copyright 2021, OpenTelemetry Authors
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

// +build linux

package smartagentreceiver

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/uptime"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/extension/healthcheckextension"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/smartagentextension"
)

var cleanUp = func() {
	configuredCollectd = false
}

func TestSmartAgentReceiverDefaultCollectdConfig(t *testing.T) {
	t.Cleanup(cleanUp)
	cfg := Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: fmt.Sprintf("%s/%s", typeStr, "valid"),
		},
		monitorConfig: &uptime.Config{
			MonitorConfig: config.MonitorConfig{
				Type: "collectd/uptime",
			},
		},
	}
	r := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())

	// Test whether default config is correctly loaded from smartagent extension factory.
	// Note, that the below invocation of Start method is bound to fail since collectd
	// is not available on the target host. However, in this case, we're only interested
	// in the correctness of the derived collectd config.
	r.Start(context.Background(), componenttest.NewNopHost())
	r.Shutdown(context.Background())
	require.Equal(t, &config.CollectdConfig{
		DisableCollectd:      false,
		Timeout:              40,
		ReadThreads:          5,
		WriteThreads:         2,
		WriteQueueLimitHigh:  500000,
		WriteQueueLimitLow:   400000,
		LogLevel:             "notice",
		IntervalSeconds:      0,
		WriteServerIPAddr:    "127.9.8.7",
		WriteServerPort:      0,
		ConfigDir:            "/var/run/signalfx-agent/collectd",
		BundleDir:            os.Getenv(constants.BundleDirEnvVar),
		HasGenericJMXMonitor: true,
		InstanceName:         "",
		WriteServerQuery:     "",
	}, r.getCollectdConfig())
}

func TestSmartAgentReceiverCollectdConfigOverrides(t *testing.T) {
	t.Cleanup(cleanUp)
	cfg := Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: fmt.Sprintf("%s/%s", typeStr, "valid"),
		},
		monitorConfig: &uptime.Config{
			MonitorConfig: config.MonitorConfig{
				Type: "collectd/uptime",
			},
		},
	}
	r := NewReceiver(zap.NewNop(), cfg, consumertest.NewMetricsNop())
	host := &mockHost{
		smartagentextensionConfig: &smartagentextension.Config{
			CollectdConfig: smartagentextension.CollectdConfig{
				Timeout:             10,
				ReadThreads:         1,
				WriteThreads:        4,
				WriteQueueLimitHigh: 5,
				WriteQueueLimitLow:  1,
				LogLevel:            "info",
				IntervalSeconds:     5,
				WriteServerIPAddr:   "127.9.8.7",
				WriteServerPort:     0,
				ConfigDir:           "/etc/collectd/collectd.conf",
				BundleDir:           "/opt/bin/collectd/",
			},
		},
	}

	// Test whether config overrides are correctly loaded from smartagent extension factory.
	// Note, that the below invocation of Start method is bound to fail since collectd
	// is not available on the target host. However, in this case, we're only interested
	// in the correctness of the derived collectd config.
	r.Start(context.Background(), host)
	r.Shutdown(context.Background())
	require.Equal(t, &config.CollectdConfig{
		DisableCollectd:      false,
		Timeout:              10,
		ReadThreads:          1,
		WriteThreads:         4,
		WriteQueueLimitHigh:  5,
		WriteQueueLimitLow:   1,
		LogLevel:             "info",
		IntervalSeconds:      5,
		WriteServerIPAddr:    "127.9.8.7",
		WriteServerPort:      0,
		ConfigDir:            "/etc/collectd/collectd.conf",
		BundleDir:            "/opt/bin/collectd/",
		HasGenericJMXMonitor: true,
		InstanceName:         "",
		WriteServerQuery:     "",
	}, r.getCollectdConfig())
}

type mockHost struct {
	smartagentextensionConfig *smartagentextension.Config
}

func (m *mockHost) ReportFatalError(err error) {
}

func (m *mockHost) GetFactory(kind component.Kind, componentType configmodels.Type) component.Factory {
	return nil
}

func (m *mockHost) GetExtensions() map[configmodels.Extension]component.ServiceExtension {
	m.smartagentextensionConfig.ExtensionSettings = configmodels.ExtensionSettings{
		TypeVal: "smartagent",
		NameVal: "smartagent",
	}

	randomExtensionConfig := &healthcheckextension.Config{}
	return map[configmodels.Extension]component.ServiceExtension{
		m.smartagentextensionConfig: getExtension(smartagentextension.NewFactory(), m.smartagentextensionConfig),
		randomExtensionConfig:       getExtension(healthcheckextension.NewFactory(), randomExtensionConfig),
	}
}

func getExtension(f component.ExtensionFactory, cfg configmodels.Extension) component.ServiceExtension {
	e, err := f.CreateExtension(context.Background(), component.ExtensionCreateParams{}, cfg)
	if err != nil {
		panic(err)
	}
	return e
}

func (m *mockHost) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	return nil
}
