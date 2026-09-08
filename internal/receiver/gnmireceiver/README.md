# gNMI Receiver

The gNMI receiver ingests push-based streaming network telemetry from devices via the
[gNMI](https://github.com/openconfig/gnmi) `Subscribe` RPC, converting the streamed
updates into OpenTelemetry metrics.

| Status        |                                                                                                                                     |
| ------------- |-------------------------------------------------------------------------------------------------------------------------------------|
| Stability     | [development](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development): metrics |
| Distributions | [Splunk](https://github.com/signalfx/splunk-otel-collector)                                                                         |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@jkoronaAtCisco](https://www.github.com/jkoronaAtCisco)                                                                            |

> **Note:** This receiver is in early development and its configuration may change.

## Configuration

The receiver opens a long-lived gNMI `Subscribe` stream to each configured target. Each
stream can contain multiple path `subscriptions`.

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
          - path: /interfaces/interface/state/oper-status
            mode: on_change
            overrides:
              oper-status:
                type: gauge
                # closed set of values this leaf can take; see "Enum leaves" below
                enum_values: [UP, DOWN, TESTING]
```

### Target

Each target embeds the standard collector [gRPC client settings][configgrpc]
(`endpoint`, `tls`, `keepalive`, `headers`, `compression`, ...) plus the fields below.

| Field           | Default   | Description                                                        |
| --------------- | --------- | ------------------------------------------------------------------ |
| `endpoint`      | (required)| `host:port` of the gNMI target.                                    |
| `username`      |           | Sent as gNMI gRPC metadata (not an `Authorization` header). |
| `password`      |           | Sent as gNMI gRPC metadata alongside `username`. Redacted in logs. Requires `username`. |
| `encoding`      | `proto`   | gNMI encoding: `proto`, `json`, or `json_ietf`.                    |
| `redial`        | `10s`     | Delay before reconnecting after a session failure (min `1s`). Set to `0` to disable automatic reconnection. |
| `tls`           |           | Standard collector TLS client settings.                            |
| `subscriptions` | (required)| One or more path subscriptions (below).                            |

### Subscription

| Field                | Default   | Description                                                                 |
| -------------------- | --------- | --------------------------------------------------------------------------- |
| `path`               | (required)| gNMI path to subscribe to. Must be absolute (start with `/`).               |
| `origin`             |           | YANG model origin (e.g. `openconfig`).                                      |
| `mode`               | (required)| Per-path mode within a `STREAM` subscription: `sample`, `on_change`, or `target_defined`. The receiver does not support `ONCE` or `POLL` SubscriptionLists. |
| `sample_interval`    |           | Sampling period. Required and must be `> 0` when `mode` is `sample`; must not be set for other modes. |
| `heartbeat_interval` |           | Forces an update at this interval even if the value has not changed.        |
| `suppress_redundant` | `false`   | Skip sending unchanged values.                                              |
| `default`            |           | Metric `type`/`unit`/`enum_values` applied to leaves not matched by `overrides`. |
| `overrides`          |           | Map of leaf name → metric `type`/`unit`/`enum_values`, taking precedence over `default`. |

### Metric names and attributes

Metric names are the gNMI path elements joined with dots, prefixed with the model
origin when the target supplies one:

```
/interfaces/interface[name=eth0]/state/counters/in-octets
  -> interfaces.interface.state.counters.in-octets
```

Path keys are not part of the name; they become datapoint attributes (`name=eth0`
above). The target endpoint is recorded as the `server.address` resource attribute.

Two list elements in the same path can use the same key name (e.g. `index` on two
nested keyed lists). When a key name occurs only once it keeps its plain name; when
it occurs more than once, every occurrence of that name is prefixed with its owning
element name (e.g. `subinterface.index`) so the attributes don't collide or silently
overwrite one another.

Integer values (including `counter64`) are emitted as integer datapoints so large
counters keep full precision. Booleans are emitted as `1`/`0`.

`counter64` and other unsigned leaves that exceed `2^63-1` (larger than any real
interface counter is likely to reach, but possible on wraparound) cannot be
represented as a 64-bit signed integer datapoint. Rather than silently wrapping to a
negative value, the receiver falls back to a double for that single update — losing
precision above `2^53` but keeping the sign and magnitude correct — and logs the
occurrence.

Non-numeric (string) values cannot be represented as a metric value, so a configured
string leaf is emitted as an *info metric*: the name gains an `_info` suffix, the
datapoint value is always `1`, and the string is carried in the `value` attribute.
Unconfigured string leaves are dropped like any other unconfigured leaf.

An info metric reports only the current value, not every possible value — an
`oper-status` transition from `UP` to `DOWN` produces one new datapoint, not an
explicit `0` for the state that's no longer active:

```
interfaces.interface.state.oper-status_info{name="eth0"} 1  # value="UP"
interfaces.interface.state.oper-status_info{name="eth0"} 1  # value="DOWN"
```

For a leaf whose possible values are known in advance, `enum_values` (below) emits
the full state set and avoids this stale-timeseries problem.

Some targets encode numeric leaves as JSON strings rather than JSON numbers,
particularly over `json_ietf` (e.g. `"in-octets": "123"`, common for values that
don't fit safely in a JSON number). If the leaf is configured with a numeric `type`
(`sum` or `gauge`), the receiver parses the string as a number and emits it normally
instead of falling back to an info metric — configuration takes precedence over the
wire encoding. A string that does not parse as a number, or that has no numeric type
configured, still becomes an info metric.

JSON and JSON-IETF payloads are flattened into one metric per leaf. Nested object
keys extend the metric name; array elements keep the metric name and record their
position in an `index` attribute. YANG leaf-lists are flattened the same way, so each
element becomes a datapoint distinguished by its `index`.

### Enum leaves

A leaf with a known, closed set of possible values (e.g. `oper-status`) can be
declared with `enum_values` instead of being left as a plain info metric:

```yaml
overrides:
  oper-status:
    type: gauge
    enum_values: [UP, DOWN, TESTING]
```

Every update to that leaf then emits one datapoint per declared value — `1` for the
active value, `0` for the rest — as a single metric with an `_state` suffix and a
`state` attribute, rather than the `_info`/`value` pair used for unconfigured strings.
This makes transitions unambiguous: the same notification that reports `DOWN` also
reports `UP` going back to `0`.

```
interfaces.interface.state.counters.oper-status_state{name="eth0",state="UP"}      1
interfaces.interface.state.counters.oper-status_state{name="eth0",state="DOWN"}    0
interfaces.interface.state.counters.oper-status_state{name="eth0",state="TESTING"} 0

# after a transition to DOWN:
interfaces.interface.state.counters.oper-status_state{name="eth0",state="UP"}      0
interfaces.interface.state.counters.oper-status_state{name="eth0",state="DOWN"}    1
interfaces.interface.state.counters.oper-status_state{name="eth0",state="TESTING"} 0
```

Notes:

- **Values only, not types.** `enum_values` applies to leaves that arrive as strings.
  A leaf configured with `enum_values` that instead arrives as a number or boolean is
  unaffected and still emits its plain numeric metric.
- **Closed set.** A received value that is not in `enum_values` emits `0` for every
  declared value (no series is minted for the unrecognized value) and is logged as an
  error, so a stale enum list surfaces instead of silently growing cardinality.
- **Module prefix is stripped.** `json_ietf` targets often module-qualify identityref
  values (e.g. `openconfig-interfaces:UP`); the receiver strips everything up to and
  including the last `:` on both the configured and received values, so `UP` and
  `openconfig-interfaces:UP` are the same state. Matching is otherwise case-sensitive.
- **Unit** defaults to `1` (dimensionless) but can be overridden; `type` must be
  `gauge` — `enum_values` is rejected at startup with `type: sum`.
- Cardinality scales with the number of declared values: an enum leaf with N values
  produces N datapoints per update, per matched path.

### Metric type, unit, and enum resolution

gNMI carries a value's data type but not whether it is a counter or a gauge, nor its
unit or possible values — that information lives in the YANG schema. The receiver
resolves it from configuration instead:

- `type`: `sum` (emitted as a monotonic Sum) or `gauge` (emitted as a Gauge).
- `unit`: metric unit, ideally [UCUM] (e.g. `By`, `1`, `By/s`).
- `enum_values`: closed set of string values, for leaves emitted as a state metric
  (see [Enum leaves](#enum-leaves)).

For each leaf the receiver applies the matching `overrides` entry if present, otherwise
`default`. **A leaf matched by neither is dropped** (no metric is emitted). A subscription
must therefore define at least one of `default` or `overrides`.

### Credentials

Credentials are sent as gNMI AAA gRPC metadata rather than an `Authorization`
header. Per [section 3.1 of the specification][spec-auth], the metadata must contain a
username and may include a password, and the two cases mean different things:

- `username` **and** `password` — the target authenticates *and* authorizes the RPC.
- `username` only — the target authorizes the RPC.

The receiver therefore omits the `password` metadata key entirely when no password is
configured, rather than sending an empty one (which would be an authentication attempt
with a blank password). Setting `password` without `username` is rejected at startup.

### Notes on gNMI specification behavior

- **`sample_interval` must be explicit.** The [specification][spec-stream] gives
  `sample_interval: 0` the meaning "use the target's lowest supported interval". This
  receiver instead requires an explicit interval `> 0` for `sample` mode, so that the
  collection rate is always a deliberate choice rather than whatever maximum rate the
  device will produce.
- **`sample_interval` is rejected for other modes.** The specification states that a target
  should reject `target_defined` when a sample interval is supplied, so the receiver fails
  configuration validation rather than sending an interval that would be rejected or
  silently ignored.
- **Default `encoding` is `proto`.** The specification designates JSON as the minimum
  encoding a target must support, so `proto` may not be available everywhere. The receiver
  does not perform `Capabilities` negotiation, so set `encoding` explicitly if the target
  does not support `proto`.

[configgrpc]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/configgrpc/README.md
[UCUM]: https://ucum.org/ucum
[spec-stream]: https://openconfig.net/docs/gnmi/gnmi-specification/#35152-stream-subscriptions
[spec-auth]: https://openconfig.net/docs/gnmi/gnmi-specification/#31-session-security-authentication-and-rpc-authorization
