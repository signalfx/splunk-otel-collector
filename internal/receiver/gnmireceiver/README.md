# gNMI Receiver

The gNMI receiver ingests push-based streaming network telemetry from devices via the
[gNMI](https://github.com/openconfig/gnmi) `Subscribe` RPC, converting the streamed
updates into OpenTelemetry metrics.

| Status        |                                                          |
| ------------- |----------------------------------------------------------|
| Stability     | [development]: metrics                                   |
| Distributions | [Splunk]                                                 |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@jkoronaAtCisco](https://www.github.com/jkoronaAtCisco) |

> **Note:** This receiver is in early development. The current implementation is a
> skeleton only — configuration, subscription handling, and metric conversion are
> added in follow-up changes.

## Configuration

Configuration options will be documented here as they are implemented.
