# gNMI Receiver

The gNMI receiver ingests push-based streaming network telemetry from devices via the
[gNMI](https://github.com/openconfig/gnmi) `Subscribe` RPC, converting the streamed
updates into OpenTelemetry metrics.

| Status        |                                                                                                                                     |
| ------------- |-------------------------------------------------------------------------------------------------------------------------------------|
| Stability     | [development](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development): metrics |
| Distributions | [Splunk](https://github.com/signalfx/splunk-otel-collector)                                                                         |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@jkoronaAtCisco](https://www.github.com/jkoronaAtCisco)                                                                            |

> **Note:** This receiver is in early development. Configuration is defined, but the
> subscription/streaming engine and metric conversion are still being added, so the
> receiver does not yet produce telemetry.

## Configuration

The receiver connects to one or more gNMI `targets`. Each target opens a `Subscribe` 
stream and can carry multiple path `subscriptions`.

```yaml
receivers:
  gnmi:
    targets:
      - endpoint: 10.0.0.1:57400
        username: admin
        password: ${env:GNMI_PASSWORD}
        encoding: json_ietf        # proto (default) | json | json_ietf
        redial: 30s                # delay before reconnecting after a failure
        tls:
          insecure: true           # lab only; omit for TLS
        subscriptions:
          - path: /interfaces/interface/state/counters
            origin: openconfig
            mode: sample            # sample | on_change | target_defined
            sample_interval: 10s    # required when mode is "sample"
            # default type/unit applied to every leaf under this path...
            default:
              type: sum             # sum (monotonic) | gauge
              unit: "1"
            # ...unless a leaf name is listed in overrides:
            overrides:
              in-octets:
                type: sum
                unit: By
              out-octets:
                type: sum
                unit: By
```

### Target

Each target embeds the standard collector [gRPC client settings][configgrpc]
(`endpoint`, `tls`, `keepalive`, `headers`, `compression`, ...) plus the fields below.

| Field           | Default   | Description                                                        |
| --------------- | --------- | ------------------------------------------------------------------ |
| `endpoint`      | (required)| `host:port` of the gNMI target.                                    |
| `username`      |           | Sent as gNMI gRPC metadata (not an `Authorization` header).        |
| `password`      |           | Sent as gNMI gRPC metadata. Redacted in logs.                      |
| `encoding`      | `proto`   | gNMI encoding: `proto`, `json`, or `json_ietf`.                    |
| `redial`        | `10s`     | Delay before reconnecting after a session failure (min `1s`).      |
| `tls`           |           | Standard collector TLS client settings.                            |
| `subscriptions` | (required)| One or more path subscriptions (below).                            |

### Subscription

| Field                | Default   | Description                                                                 |
| -------------------- | --------- | --------------------------------------------------------------------------- |
| `path`               | (required)| gNMI path to subscribe to.                                                  |
| `origin`             |           | YANG model origin (e.g. `openconfig`).                                      |
| `mode`               | (required)| `sample`, `on_change`, or `target_defined`.                                 |
| `sample_interval`    |           | Sampling period; required and must be `> 0` when `mode` is `sample`.        |
| `heartbeat_interval` |           | Forces an update at this interval even if the value has not changed.        |
| `suppress_redundant` | `false`   | Skip sending unchanged values.                                              |
| `default`            |           | Metric `type`/`unit` applied to leaves not matched by `overrides`.          |
| `overrides`          |           | Map of leaf name → metric `type`/`unit`, taking precedence over `default`.  |

### Metric type and unit resolution

gNMI carries a value's data type but not whether it is a counter or a gauge, nor its
unit — that information lives in the YANG schema. The receiver resolves it from
configuration instead:

- `type`: `sum` (emitted as a monotonic Sum) or `gauge` (emitted as a Gauge).
- `unit`: metric unit, ideally [UCUM] (e.g. `By`, `1`, `By/s`).

For each leaf the receiver applies the matching `overrides` entry if present, otherwise
`default`. **A leaf matched by neither is dropped** (no metric is emitted). A subscription
must therefore define at least one of `default` or `overrides`.

[configgrpc]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configgrpc/README.md
[UCUM]: https://ucum.org/ucum
