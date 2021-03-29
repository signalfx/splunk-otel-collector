# Security

Start by reviewing the [OpenTelemetry Collector security
documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/security.md).

## Reporting Security Issues

Please *DO NOT* report security vulnerabilities via public GitHub issue
reports. Please [report security issues here](
https://www.splunk.com/en_us/product-security/report.html).

## Dependencies

This project relies on a variety of [external
dependencies](https://github.com/signalfx/splunk-otel-collector/network/dependencies).
These dependencies are monitored by
[Dependabot](https://docs.github.com/en/code-security/supply-chain-security/configuring-dependabot-security-updates).
Dependencies are [checked
daily](https://github.com/signalfx/splunk-otel-collector/blob/main/.github/dependabot.yml)
and associated pull requests are opened automatically. Upgrading to the [latest
release](https://github.com/signalfx/splunk-otel-collector/releases)
is recommended to ensure you have the latest security updates. If a security
vulnerability is detected for a dependency of this project then either:

- You are running an older release
- A new release with the updates has not been cut yet
- The updated dependency has not been merged likely due to some breaking change
- The dependency has not released an updated version with the patch

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
