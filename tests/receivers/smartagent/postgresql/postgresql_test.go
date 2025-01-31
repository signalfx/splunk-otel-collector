// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build smartagent_integration

package tests

import (
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPostgresReceiverProvidesAllMetrics(t *testing.T) {
	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithEnv(
		map[string]string{"POSTGRES_DB": "test_db", "POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres"},
	).WithExposedPorts("5432:5432").WithName("postgres-server").WithNetworks(
		"postgres",
	).WillWaitForPorts("5432").WillWaitForLogs("database system is ready to accept connections")

	client := testutils.NewContainer().WithContext(path.Join(".", "testdata", "client")).WithEnv(
		map[string]string{"POSTGRES_SERVER": "postgres-server"},
	).WithName("postgres-client").WithNetworks("postgres").WillWaitForLogs("Beginning psql requests")
	containers := []testutils.Container{server, client}

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector("all_metrics_config.yaml")
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "all.yaml"))
	require.NoError(t, err)
	lastIndex := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		var err error
		newIndex := len(tc.OTLPReceiverSink.AllMetrics()) - 1
		for i := newIndex; i >= lastIndex; i-- {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			if m.MetricCount() == expected.MetricCount() {
				err = pmetrictest.CompareMetrics(expected, m,
					pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
					pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
					pmetrictest.IgnoreResourceAttributeValue("server.port"),
					pmetrictest.IgnoreResourceAttributeValue("service.name"),
					pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
					pmetrictest.IgnoreResourceAttributeValue("service_version"),
					pmetrictest.IgnoreMetricAttributeValue("service_version"),
					pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
					pmetrictest.IgnoreMetricAttributeValue("queryid"),
					pmetrictest.IgnoreMetricAttributeValue("table"),
					pmetrictest.IgnoreSubsequentDataPoints(),
					pmetrictest.IgnoreTimestamp(),
					pmetrictest.IgnoreStartTimestamp(),
					pmetrictest.IgnoreMetricDataPointsOrder(),
					pmetrictest.IgnoreScopeMetricsOrder(),
					pmetrictest.IgnoreMetricsOrder(),
					pmetrictest.IgnoreScopeVersion(),
					pmetrictest.IgnoreResourceMetricsOrder(),
					pmetrictest.IgnoreMetricValues(),
				)
				if err == nil {
					return
				}
			}
		}
		lastIndex = newIndex
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)
}
