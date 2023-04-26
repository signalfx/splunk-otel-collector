# HTTP Sink Exporter

**This component is for testing purposes only. It does not redact or filter any telemetry content it exposes and should not be used with production data.**

This exporter makes span data available via a HTTP endpoint. The endpoint
accepts requests for spans or metrics with specific characteristics and blocks until the exporter
receives matching data or the request times out. Once the requested data is detected, it is
returned back to the client as JSON. 

Spans are returned as [JSON encoding](https://developers.google.com/protocol-buffers/docs/proto3#json)
using [Jaeger protocol](https://github.com/jaegertracing/jaeger-idl/tree/master/proto/api_v2). We plan to
switch this to OTLP in future.

Metrics are returned as JSON encoding using the [OTLP protocol](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/metrics/v1/metrics.proto).

Please note that there is no guarantee that exact field names will remain stable.
This is intended primarily for automatic and manual testing (and occasional debugging) observability pipelines without setting up backends.

Supported pipeline types: traces, metrics.

## Getting Started

The following settings are required:

- `endpoint` (defaults to `0.0.0.0:8378`).

Example:

```yaml
exporters:
  httpsink:
    endpoint: "0.0.0.0:8378"
```

## Example usage:

- Splunk Otel Python uses this to implement [end to end tests](https://github.com/signalfx/splunk-otel-python/tree/main/tests/integration).
- Splunk Otel JS uses this to [implement tests](https://github.com/signalfx/splunk-otel-js/tree/666f4406e29a8dee2dadbfd1efc0cc1cb6154755/test/examples).
- Used internally at Splunk for occasional debugging.