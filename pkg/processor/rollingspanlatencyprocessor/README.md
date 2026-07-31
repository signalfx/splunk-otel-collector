# Rolling Span Latency Processor

| Status                   |                  |
|--------------------------|------------------|
| Stability                | [in-development] |
| Supported pipeline types | traces           |
| Distributions            | [Splunk]         |

The Rolling Span Latency processor labels spans as `slow` or `very_slow` when their duration is
statistically anomalous relative to a rolling baseline for that span. It maintains a
time-aware exponentially weighted moving average (EWMA) of mean and variance per
`(resource attribute tuple, span name)` key. When a span's duration exceeds the baseline
mean by a configurable number of standard deviations, a span attribute is written.

## Configuration

All fields are optional. Defaults are shown in the example below.

| Field | Required | Description |
|---|---|---|
| `half_life` | No | EWMA decay period. The effective weight of any sample halves every `half_life`. Default: `2h`. |
| `slow_threshold` | No | Standard deviations above the mean at which a span is labeled `slow`. Default: `3.0`. |
| `very_slow_threshold` | No | Standard deviations above the mean at which a span is labeled `very_slow`. Must be greater than `slow_threshold`. Default: `4.0`. |
| `attribute_key` | No | Span attribute key written when a span is slow or very slow. Default: `latency.category`. |
| `resource_key_attributes` | No | Ordered list of resource attribute keys whose values are combined with the span name to form a baseline key. Spans sharing the same values share a single EWMA baseline. Default: `[service.namespace, service.name, deployment.environment.name]`. |
| `idle_timeout` | No | How long a baseline key must go without observations before being evicted from memory. Default: `8h`. |
| `eviction_interval` | No | How often the background eviction sweep runs. Default: `10m`. |
| `max_baselines` | No | Maximum number of baseline entries held in memory. `0` means unlimited. When the cap is reached, new keys are dropped and a warning is logged. Default: `0`. |
| `churn_warning_ratio` | No | Fraction of the active baseline count that, when exceeded by a single eviction sweep's evicted count, triggers a high-churn warning. Default: `0.5`. |
| `warmup_count` | No | Minimum number of observations required before a baseline is eligible for labeling. Default: `30`. |
| `min_stddev` | No | Minimum standard deviation (nanoseconds) used when scoring a span. Prevents near-zero variance from producing false positives. Default: `1000000` (1ms). |

### Example

```yaml
processors:
  rolling_span_latency:
    half_life: 2h
    slow_threshold: 3.0
    very_slow_threshold: 4.0
    attribute_key: latency.category
    resource_key_attributes:
      - service.namespace
      - service.name
      - deployment.environment.name
    idle_timeout: 8h
    eviction_interval: 10m
    max_baselines: 1000
    churn_warning_ratio: 0.5
    warmup_count: 30
    min_stddev: 1000000
```
