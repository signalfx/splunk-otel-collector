# Examples

This folder contains examples showcasing how the collector can integrate and complement existing environments.

## [Splunk HEC example](./splunk-hec)

This example showcases how the collector can send data to a Splunk Enterprise deployment over the Docker logging driver.

[Read more...](./splunk-hec)

## [Splunk HEC metrics](./splunk-hec-metrics)

This example showcases how the collector can send data to a Splunk Enterprise deployment from a Telegraf instance.

[Read more...](./splunk-hec-metrics)

## [Prometheus Federation](./prometheus-federation)

This example showcases how the agent works with Splunk Enterprise and an existing Prometheus deployment.

[Read more...](./prometheus-federation)

## [Nomad](./nomad)

The demo job deploys the Splunk OpenTelemetry Collector as `agent` and `gateway`, `load
generators`, to collect metrics and traces and export them using the `SignalFx` exporter.

[Read more...](./nomad)

## [Splunk HEC traces](./splunk-hec-traces)

This example showcases how the collector can send traces to a Splunk Enterprise deployment from a program emitting traces.

[Read more...](./splunk-hec-traces)

## [Filelog receiver with Splunk Enterprise](./otel-logs-splunk)

This example showcases how the collector can follow a file and send its contents to Splunk Enterprise.