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

package tests

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCoreValidateDefaultConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	c, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			// deferring running service for exec
			c := collector.WithEnv(
				map[string]string{
					"SPLUNK_REALM":        "noop",
					"SPLUNK_ACCESS_TOKEN": "noop",
				},
			).WithArgs("-c", "trap exit SIGTERM ; echo ok ; while true; do : ; done")
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithEntrypoint("sh").WillWaitForLogs("ok")
			return cc
		},
	)

	defer shutdown()

	for _, config := range []string{"gateway", "agent"} {
		config := config
		t.Run(config, func(t *testing.T) {
			sc, stdout, stderr := c.Container.AssertExec(t, 15*time.Second,
				"sh", "-c", fmt.Sprintf("/otelcol --config /etc/otel/collector/%s_config.yaml validate", config),
			)
			assert.Zero(t, sc)
			require.Empty(t, stdout)
			require.False(t, regexp.MustCompile("(?i)error").MatchString(stderr))
		})
	}
}

func TestCoreValidateYamlProvider(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	config := `exporters:
  debug:
    verbosity: detailed
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu:
service:
  pipelines:
    metrics:
      exporters:
      - debug
      receivers:
      - hostmetrics
`

	c, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			// deferring running service for exec
			c := collector.WithEnv(
				map[string]string{
					"SPLUNK_CONFIG":      "",
					"SPLUNK_CONFIG_YAML": config,
				},
			).WithArgs("-c", "trap exit SIGTERM ; echo ok ; while true; do : ; done")
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithEntrypoint("sh").WillWaitForLogs("ok")
			return cc
		},
	)

	defer shutdown()

	sc, stdout, stderr := c.Container.AssertExec(t, 15*time.Second,
		"sh", "-c", "/otelcol validate",
	)
	assert.Zero(t, sc)
	require.Empty(t, stdout)
	require.False(t, regexp.MustCompile("(?i)error").MatchString(stderr))
}

func TestCoreValidateDetectsInvalidYamlProvider(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	config := `exporters:
  notreal:
    endpoint: an-endpoint
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
service:
  pipelines:
    metrics:
      exporters:
      - notreal
      receivers:
      - hostmetrics
`

	c, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			// deferring running service for exec
			c := collector.WithEnv(
				map[string]string{
					"SPLUNK_CONFIG":      "",
					"SPLUNK_CONFIG_YAML": config,
				},
			).WithArgs("-c", "trap exit SIGTERM ; echo ok ; while true; do : ; done")
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithEntrypoint("sh").WillWaitForLogs("ok")
			return cc
		},
	)

	defer shutdown()

	sc, stdout, stderr := c.Container.AssertExec(t, 15*time.Second,
		"sh", "-c", "/otelcol validate",
	)
	require.Equal(t, 1, sc)
	require.Empty(t, stdout)
	require.Contains(t, stderr, `'exporters' unknown type: \"notreal\" for id: \"notreal\"`)
}
