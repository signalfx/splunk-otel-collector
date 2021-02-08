# Tests Utilities

The `testutils` package provides an internal test format for Collector data, and helpers to help assert its integrity
from arbitrary components.

### Resource Metrics

`ResourceMetrics` are at the core of the internal metric data format for these tests and are intended to be defined
in yaml files or by converting from obtained `pdata.Metrics` items.

```yaml
resource_metrics:
  - attributes:
      a_resource_attribute: a_value
      another_resource_attribute: another_value
    instrumentation_library_metrics:
      - instrumentation_library:
          name: a library
          version: some version
      - metrics:
          - name: my.int.gauge
            type: IntGauge
            description: my.int.gauge description
            unit: ms
            labels:
              my_label: my_label_value
              my_other_label: my_other_label_value
            value: 123
          - name: my.double.sum
            type: DoubleNonmonotonicDeltaSum
            labels:
              label_one: value_one
              label_two: value_two
            value: -123.456
  - instrumentation_library_metrics:
      - instrumentation_library:
          name: an instrumentation library from a different resource without attributes
        metrics:
          - name: my.double.gauge
            type: DoubleGauge
            labels:
              another: label
            value: 456.789
          - name: my.double.gauge
            type: DoubleGauge
            labels:
              another: label
            value: 567.890
      - instrumentation_library:
          name: another instrumentation library
          version: this_library_version
        metrics:
          - name: another_int_gauge
            type: IntGauge
            value: 456
```

Using `LoadResourceMetrics("my_yaml.path")` you can create an equivalent `ResourceMetrics` instance to what your yaml file specifies.
Using `PDataToResourceMetrics(myReceivedPDataMetrics)` you can use the assertion helpers to determine if your expected
`ResourceMetrics` are the same as those received in your test case. `FlattenResourceMetrics()` is a good way to "normalize"
metrics received over time to ensure that only unique datapoints are represented, and that all unique Resources and
Instrumentation Libraries have a single item.

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
