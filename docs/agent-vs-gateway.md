# Agent versus Gateway

Splunk OpenTelemetry Connector provides a single binary and two deployment methods:

- Agent: A Collector instance running with the application or on the same host
  as the application (e.g. binary, sidecar, or DaemonSet). Used to collect host
  and application metrics as well as offload application instrumentation
  processing and exporting. Recommended for all installations.
- Gateway: One or more Collector instances running as a standalone service
  (e.g. container or deployment) typically per cluster, datacenter, or region.
  Used to collect data from multiple Agents and/or serverless functions.
  Optional component for specific use cases.

Also review the [architecture](architecture.md).
