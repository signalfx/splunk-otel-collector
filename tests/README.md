# Splunk OpenTelemetry Collector Integration Tests and Utilities

To assist in vetting and validating the upstream and Splunk Collector distributions of the Collector, this library
provides a set of integration tests and associated utilities.  The general testing pattern this project is geared toward
is:

1. Building the Collector (`make otelcol` or `make all`)
1. Defining your expected [resource metric content](./testutils/README.md#resource-metrics) as a yaml file
([see example](testutils/telemetry/testdata/metrics/resource-metrics.yaml))
1. Spin up your target resources as [docker containers](./testutils/README.md#test-containers).
1. Stand up an in-memory [OTLP metrics receiver and sink](./testutils/README.md#otlp-metrics-receiver-sink) capable of detecting if/when desired data are received.
1. Spin up your Collector [as a subprocess](./testutils/README.md#collector-process) or [as a container](./testutils/README.md#collector-container) configured to report to this OTLP receiver.
  
...but if you are interested in something else enhancements and contributions are a great way to ensure this library
is more useful overall.

**NOTE** At this time, integration tests generally target collector containers (`SPLUNK_OTEL_COLLECTOR_IMAGE` env var),
and test coverage for the subprocess is best effort only, unless the test cases explicitly maintain one.
The collector process targets are generally for test development without requiring frequent rebuilds of a local docker image.

```go
package example_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

func TestMyExampleComponent(t *testing.T) {
	expectedResourceMetrics, err := telemetry.LoadResourceMetrics(
		filepath.Join(".", "testdata", "metrics", "my_resource_metrics.yaml"),
	)
	require.NoError(t, err)
	require.NotNil(t, expectedResourceMetrics)

	// combination OTLP Receiver, consumertests.MetricsSink, and consumertests.LogsSink
	otlp, err := testutils.NewOTLPReceiverSink().WithEndpoint("localhost:23456").Build()
	require.NoError(t, err)
	require.NoError(t, otlp.Start())

	defer func() {
		require.NoError(t, otlp.Shutdown())
	}()

	myContainer := testutils.NewContainer().WithImage("someTarget").Build()
	err = myContainer.Start(context.Background())
	require.NoError(t, err)

	// running collector subprocess that uses the provided config set to export OTLP to our test receiver
	myCollector, err := testutils.NewCollectorProcess().WithConfigPath(filepath.Join(".", "testdata", "config.yaml")).Build()
	require.NoError(t, err)
	err = myCollector.Start()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, myCollector.Shutdown() )
	}()

	require.NoError(t, otlp.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
```
