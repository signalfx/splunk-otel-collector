# HTTP Sink Exporter

This exporter makes span data available via a HTTP endpoint. The endpoint
accepts requests for spans with specific characteristics and blocks until the exporter
receives such spans or the request times out. Once the requested spans are detected, they're
returned back to the client as JSON. 

This exporter returns data as [JSON encoding](https://developers.google.com/protocol-buffers/docs/proto3#json)
using [Jaeger protocol](https://github.com/jaegertracing/jaeger-idl/tree/master/proto/api_v2).

Please note that there is no guarantee that exact field names will remain stable.
This intended for primarily for testing observability pipelines without setting up backends.

Supported pipeline types: traces.

## Getting Started

The following settings are required:

- `endpoint` (defaults to `0.0.0.0:8378`).

Example:

```yaml
exporters:
  httpsink:
    endpoint: "0.0.0.0:8378"
```
