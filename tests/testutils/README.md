# Tests Utilities

The `testutils` package provides an internal test format for Collector data, and helpers to help assert its integrity
from arbitrary components.

### Test Containers

The Testcontainers project is a popular testing resource for easy container creation and usage for a number of languages
including [Go](https://github.com/testcontainers/testcontainers-go).  The `testutils` package provides a helpful [container
builder and wrapper library](./container.go) to avoid needing direct Docker api usage:

```go
import "github.com/signafx/splunk-otel-collector/tests/testutils"

myContainerFromImageName := testutils.NewContainer().WithImage(
	"my-docker-image:123.4-alpine",
).WithEnvVar("MY_ENV_VAR", "ENV_VAR_VALUE",
).WithExposedPorts("12345:12345").WillWaitForPorts("12345",
).WillWaitForLogs(
    "My container is running and ready for interaction"
).Build()

// After building, `myContainerFromImageName` implements the testscontainer.Container interface
err := myContainerFromImageName.Start(context.Background())

myContainerFromBuildContext := testutils.NewContainer().WithContext(
    "./directory_with_dockerfile_and_resources",
).WithEnv(map[string]string{
    "MY_ENV_VAR_1": "value1",
    "MY_ENV_VAR_2": "value2",
    "MY_ENV_VAR_3": "value3",
}).WithExposedPorts("23456:23456", "34567:34567",
).WillWaitForPorts("23456", "34567",
).WillWaitForLogs(
    "My container is running.", "My container is ready for interaction"
).Build()

err = myContainerFromBuildContext.Start(context.Background())
```

### OTLP Metrics Receiver Sink

The `OTLPReceiverSink` is a helper type that will easily stand up an in memory OTLP Receiver with
`consumertest.MetricsSink` functionality.  It will listen to the configured gRPC endpoint that running Collector
processes can be configured to reach and provides an `AssertAllMetricsReceived()` test method to confirm that expected
`ResourceMetrics` are received within the specified window.

```go
import "github.com/signafx/splunk-otel-collector/tests/testutils"

otlp, err := testutils.NewOTLPReceiverSink().WithEndpoint("localhost:23456").Build()
require.NoError(t, err)

defer func() {
    require.Nil(t, otlp.Shutdown())
}()

require.NoError(t, otlp.Start())

require.NoError(t, otlp.AssertAllMetricsReceived(t, expectedResourceMetrics, 10*time.Second))
```

### Collector Process

The `CollectorProcess` is a helper type that will run the desired Collector executable as a subprocess using whatever 
config you provide.  If an executable path isn't specified via builder method, the first `bin/otelcol` match walking up
your current directory tree will be used, which can be helpful to ease test development and execution.

You can also specify the desired command line arguments using `builder.WithArgs()` if not simply running with the
specified config.

```go
import "github.com/signafx/splunk-otel-collector/tests/testutils"

collector, err := testutils.NewCollectorProcess().WithPath("my_otel_collector_path",
).WithConfigPath("my_config_path").WithLogger(logger).WithLogLevel("debug").Build()

err = collector.Start()
require.NoError(t, err)
defer func() { require.NoError(t, collector.Shutdown()) }()

// Also able to specify other arguments for feature, performance, and soak testing.
// path will be first `bin/otelcol` match in a parent directory
collector, err = testutils.NewCollectorProcess().WithArgs("--tested-feature", "--etc").Build()
```

### Collector Container

The `CollectorContainer` is an equivalent helper type to the `CollectorProcess` but will run a container in host network
mode for an arbitrary Collector image and tag using the config you provide.  If an image is not specified it will use a default
of `"otelcol:latest"` (the default local image tag built via `make docker-otelcol`).

```go
import "github.com/signafx/splunk-otel-collector/tests/testutils"

collector, err := testutils.NewCollectorContainer().WithImage("quay.io/signalfx/splunk-otel-collector:latest",
).WithConfigPath("my_config_path").Build()

err = collector.Start()
require.NoError(t, err)
defer func() { require.NoError(t, collector.Shutdown()) }()
```

### Testcase

All the above test utilities can be easily configured by the `Testcase` helper to avoid unnecessary boilerplate in
resource creation and cleanup.  The associated OTLPReceiverSink for each `Testcase` will have an OS-provided
endpoint that can be rendered via the `"${OTLP_ENDPOINT}"` environment variable in your tested config. `testutils`
provides a general `AssertAllMetricsReceived()` function that utilizes this type to stand up all the necessary resources
associated with a test and assert that all expected metrics are received:

If the `SPLUNK_OTEL_COLLECTOR_IMAGE` environment variable is set and not empty its value will be used to start a
`CollectorContainer` instead of a subprocess.

```go
import "github.com/signafx/splunk-otel-collector/tests/testutils"

func MyTest(t *testing.T) {
    containers := []testutils.Container{
        testutils.NewContainer().WithImage("my_docker_image"),
        testutils.NewContainer().WithImage("my_other_docker_image"),
    }

    // will implicitly create a Testcase with OTLPReceiverSink listening at $OTLP_ENDPOINT,
    // ./testdata/expected.yaml golden file, CollectorProcess with
    // ./testdata/my_collector_config.yaml config, and build and start all specified containers before calling
	testutils.RunMetricsCollectionTest(t, "my_collector_config.yaml", "expected.yaml",
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
		),
	)
}
```
