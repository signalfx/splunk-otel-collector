> The official Splunk documentation for this page is [Use Splunk Distribution of OpenTelemetry Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/resources.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

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
  (in this case, we will actively work to resolve the issue and open a tracking GitHub issues with details)
- The dependency has not released an updated version with the patch

## Exposed endpoints

By default, the Splunk OpenTelemetry Connector exposes several endpoints.
Endpoints will either be exposed:

- Locally (`localhost`): Within the service
- Publicly (`0.0.0.0`): On all network interfaces

The endpoints exposed depends on which mode the Splunk OpenTelemetry Connector
is configured in.

- `http(s)://0.0.0.0:13133/` Health endpoint useful for load balancer monitoring
- `http(s)://0.0.0.0:[6831|6832|14250|14268]/api/traces` Jaeger [gRPC|Thrift HTTP] receiver
- `http(s)://localhost:55554/debug/configz/[initial|effective]` in-memory configuration
- `http(s)://localhost:55679/debug/[tracez|pipelinez]` zPages monitoring
- `http(s)://0.0.0.0:4317` OpenTelemetry gRPC receiver
- `http(s)://0.0.0.0:6060` HTTP Forwarder used to receive Smart Agent `apiUrl` data
- `http(s)://localhost:8888/metrics` Prometheus metrics for the Collector
- `http(s)://localhost:8006` Fluent forward receiver
- `http(s)://0.0.0.0:9080` SignalFx forwarder receiver
- `http(s)://0.0.0.0:9411/api/[v1|v2]/spans` Zipkin JSON (can be set to proto) receiver
- `http(s)://0.0.0.0:9943/v2/trace` SignalFx APM receiver

Components, especially receivers, can and should be disabled if not required
for an environment.
