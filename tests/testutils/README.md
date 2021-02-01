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
    ilms:
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
  - ilms:
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

Using `LoadResourceMetrics("my_yaml.path")` you can create an analogous `ResourceMetrics` instance.
Using `PDataToResourceMetrics(myReceivedPDataMetrics)` you can use the assertion helpers to determine if your expected
`ResourceMetrics` are the same as those received in your test case. `FlattenResourceMetrics()` is a good way to "normalize"
metrics received over time to ensure that only unique datapoints are represented, and that all unique Resources and
Instrumentation Libraries have a single item.
