# Security

## Exposed endpoints

By default, the Collector exposes the following endpoints:

- `http(s)://<collectorFQDN>:13133/` Health endpoint useful for load balancer monitoring
- `http(s)://<collectorFQDN>:[14250|14268]` Jaeger [gRPC|Thrift HTTP] receiver
- `http(s)://localhost:55679/debug/[tracez|pipelinez]` zPages monitoring
- `http(s)://<collectorFQDN>:4317` OpenTelemetry gRPC receiver
- `http(s)://<collectorFQDN>:6060` HTTP Forwarder used to receive Smart Agent `apiUrl` data
- `http(s)://<collectorFQDN>:7276` SignalFx Infrastructure Monitoring gRPC receiver
- `http(s)://localhost:8888/metrics` Prometheus metrics for the Collector
- `http(s)://<collectorFQDN>:9411/api/[v1|v2]/spans` Zipkin JSON (can be set to proto) receiver
- `http(s)://<collectorFQDN>:9943/v2/trace` SignalFx APM receiver

Receivers can and should be disabled if not required for an environment.
