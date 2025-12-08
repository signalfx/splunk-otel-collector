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
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestTelegrafExecWithGoScript(t *testing.T) {
	collectorImage := testutils.GetCollectorImageOrSkipTest(t)
	platform := "amd64"
	if testutils.CollectorImageIsForArm(t) {
		platform = "arm64"
	}

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("",
		func(collector testutils.Collector) testutils.Collector {
			cc := collector.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithContext(
				path.Join(".", "testdata", "exec"),
			).WithBuildArgs(map[string]*string{
				"IMAGE_PLATFORM":              &platform,
				"SPLUNK_OTEL_COLLECTOR_IMAGE": &collectorImage,
			})
			return cc.WithArgs("--config", "/etc/config.yaml")
		})
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "all_expected.yaml"))
	require.NoError(t, err)
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		var selected *pmetric.Metrics
		for i := len(tc.OTLPReceiverSink.AllMetrics()) - 1; i >= 0; i-- {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			if m.MetricCount() == expected.MetricCount() {
				selected = &m
				break
			}
		}

		require.NotNil(tt, selected)

		err := pmetrictest.CompareMetrics(expected, *selected,
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
			pmetrictest.IgnoreResourceAttributeValue("server.port"),
			pmetrictest.IgnoreResourceAttributeValue("service.name"),
			pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
			pmetrictest.IgnoreResourceAttributeValue("service_version"),
			pmetrictest.IgnoreMetricAttributeValue("service_version"),
			pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
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
		assert.NoError(tt, err)
	}, 60*time.Second, 1*time.Second)
}
