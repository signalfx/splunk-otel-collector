# Volume Quota Processor

An OpenTelemetry Collector processor that measures the volume of spans or traces associated per service (using `service.name`) and applies volume quota, making recommendations for sampling.

The processor decorates spans with the `sampling.priority` attribute that can be used by the probabilistic sampler to sample spans.

## What it does

This processor sits in your OTel pipeline and monitors the rate of emission of spans or traces per service.

When the throughput rate of spans or traces is larger than the limits configured (either total or per service, or specifically for some services), the processor annotates spans with a `sampling.priority attribute`.

The throughput is computed by counting spans and traces over an epoch (default: 1 minute).

2 possible use cases exist:
* If a sudden increase in emission occurs inside an epoch, the processor will offer to drop any spans past the maximum number of spans per epoch.
* Additionally, the processor will use a configurable lookback period (default: 5) to set a conservative sampling rate based on the average of spans.

### Examples:

#### Sudden increase
A sudden peak of spans is registered inside an epoch of a minute. At 20s into the epoch, it has recorded 2000 spans - the limit. Any new span inside the minute will now decorate with `sampling.priority` of 0.

#### Sustained increase
After the sudden increase, the next epoch starts. During the last epoch, 8000 spans were recorded with a limit of 2000.

The processor therefore recommends a `sampling.priority` attribute set to 25.

#### Continuous increase
The number of spans now doubles every epoch while the limit is set to 2000 spans per epoch.

We use a look back period of 5 epochs to determine the starting rate.

| Epoch | Spans received | Lookback count | Lookback average count | Starting rate | Spans set to drop immediately |
|-------|----------------|----------------|------------------------|---------------|-------------------------------|
| 0     | -              | 0              | 0                      | 100           | 0                             |
| 1     | 2000           | 0              | 0                      | 100           | 0                             |
| 2     | 4000           | 2000           | 2000                   | 100           | 2000                          |
| 3     | 8000           | 6000           | 3000                   | 66            | 4969                          |
| 4     | 16000          | 14000          | 4666                   | 42            | 11238                         |

### Continuous decrease

Things are going back to normal, and traffic is divided by 2 every epoch then sets at 1000.

We use a look back period of 5 epochs to determine the starting rate.

| Epoch | Spans received | Lookback count | Lookback average count | Starting rate | Spans set to drop immediately |
|-------|----------------|----------------|------------------------|---------------|-------------------------------|
| 5     | 8000           | 30000          | 6000                   | 33            | 1939                          |
| 6     | 4000           | 38000          | 7600                   | 26            | 0                             |
| 7     | 2000           | 40000          | 8000                   | 25            | 0                             |
| 8     | 1000           | 38000          | 7600                   | 26            | 0                             |
| 9     | 1000           | 31000          | 6200                   | 32            | 0                             |
| 10    | 1000           | 16000          | 3200                   | 62            | 0                             |
| 11    | 1000           | 9000           | 1800                   | 100           | 0                             |
| 12    | 1000           | 6000           | 1200                   | 100           | 0                             |
| 13    | 1000           | 5000           | 1000                   | 100           | 0                             |
| 14    | 1000           | 5000           | 1000                   | 100           | 0                             |
| 15    | 1000           | 5000           | 1000                   | 100           | 0                             |

## Config

| Key                     | Description                                                               | Default   |
|-------------------------|---------------------------------------------------------------------------|-----------|
| `epoch`                 | Duration in seconds of a measurement period                               | 60s       |
| `lookback`.             | Number of previous epochs to consider when setting starting sampling rate | 5         |
| `global_limits::spans`  | Max number of spans per epoch                                             | 0 (unset) |
| `global_limits::traces` | Max number of traces per epoch                                            | 0 (unset) |
| `limits::spans`         | map of `service.name` to max number of spans per epoch                    | `{}`      |
| `limits::traces`        | map of `service.name` to max number of traces per epoch                   | `{}`      |

Example:
```yaml
volume_quota:
  global_limits:
    spans: 100000
    traces: 1000
  limits:
    spans:
      my.java.service: 100000000
      my.node.app: 100
    traces:
      my.dotnet.service: 1000000
```

With this example configuration:
* `my.java.service` can send up to 100000000 spans a second, but only up to 1000 traces a second (global limit applies)
* `my.dotnet.service` has a higher limit of traces throughput of 1000000, but only up to 100000 spans a second.

