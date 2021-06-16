# Agent versus Gateway

Splunk OpenTelemetry Connector provides a single binary and two deployment
modes as outlined in the [architecture](architecture.md). The recommended
[getting started](https://github.com/signalfx/splunk-otel-collector#getting-started) installation steps cover deploying in
either mode and come with a [default
configuration](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/otelcol/config/collector).

> IMPORTANT: The configuration is different for each mode.

Configuration changes from the default that may be desirable include:

- [Removing unneeded
  components](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/agent_config.yaml#L123);
  see
  [security.md](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md)
- Configuring processors such as the [attributes
  processor](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/attributesprocessor)
- Exporting to gateway

## Agent

Agent mode should be used when Splunk OpenTelemetry Connector is running with
the application or on the same host as the application (e.g. binary, sidecar,
or DaemonSet). Application instrumentation should be configured to send data to
Splunk OpenTelemetry Connector running in agent mode. Doing so offloads
responsibilities from the application including batching, queuing, retry, etc.
In addition, agent mode allows for the collection of host and application
metrics as well as host and application metadata enrichment for metrics, spans,
and logs.

By default, Splunk OpenTelemetry Connector in agent mode is configured to send data
directly to Splunk Observability Cloud. Alternatively, it can be configured to
send to Splunk OpenTelemetry Connector in gateway mode. To send data to a
gateway, change the
[configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/agent_config.yaml):

- Update `extensions.http_forwarder.egress.endpoint` (supports `${SPLUNK_GATEWAY_URL}` environment variable)
- Update `exporters.otlp.endpoint` (supports `${SPLUNK_GATEWAY_URL}` environment variable)
- Update `service.pipelines.[traces|metrics|logs].exporters`

## Gateway

Gateway mode should be used when one or more Splunk OpenTelemetry Connectors
are running as a standalone service (e.g. container or deployment); typically
gateway mode is deployed per cluster, datacenter, or region. Splunk
OpenTelemetry Connector running in agent mode or serverless instrumentation can
be configured to send data to Splunk OpenTelemetry Connector running in gateway
mode. Doing so offers capabilities including increased buffer and retry as well
as egress and token management control.

Splunk OpenTelemetry Connector running in gateway mode as an optional component
for specific use cases.
