# Splunk Distribution of OpenTelemetry Collector

The Splunk Distribution of [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector) provides
a binary that can receive, process and export trace, metric and log data.

**Installations that use this distribution can receive direct help from Splunk's support teams.**
Customers are free to use the core OpenTelemetry OSS components (several do!) and we will provide best
effort guidance to them for any issues that crop up, however only the Splunk distributions are in
scope for official Splunk support and support-related SLAs.

The Collector currently supports:

- [Splunk APM](https://www.splunk.com/en_us/software/splunk-apm.html) via the
  [`sapm`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/sapmexporter).
  More information available
  [here](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-opentelemetry-collector.html).
- [Splunk Infrastructure
  Monitoring](https://www.splunk.com/en_us/software/infrastructure-monitoring.html)
  via the [`signalfx`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/signalfxexporter).
  More information available
  [here](https://docs.signalfx.com/en/latest/otel/imm-otel-collector.html).
- [Splunk Log Observer](https://www.splunk.com/en_us/form/splunk-log-observer.html) via
  the [`splunk_hec`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/splunkhecexporter).
- [Splunk Cloud](https://www.splunk.com/en_us/software/splunk-cloud.html) or
  [Splunk
  Enterprise](https://www.splunk.com/en_us/software/splunk-enterprise.html) via
  the [`splunk_hec`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/splunkhecexporter).

> :construction: This project is currently in **BETA**.

## Getting Started

The Collector provides a single binary and two deployment methods:

- Agent: A Collector instance running with the application or on the same host
  as the application (e.g. binary, sidecar, or daemonset).
- Gateway: One or more Collector instances running as a standalone service
  (e.g. container or deployment) typically per cluster, datacenter or region.

> Use of the Collector running as an agent is only supported for the [Splunk
Observability
Suite](https://www.splunk.com/en_us/form/splunk-observability-suite.html) at
this time. The [SignalFx Smart
Agent](https://github.com/signalfx/signalfx-agent) or [Splunk Universal
Forwarder](https://docs.splunk.com/Documentation/Forwarder/8.1.2/Forwarder/Abouttheuniversalforwarder)
should be used for all other products.

The Collector is supported on and packaged for a variety of platforms including:

- Kubernetes
  - [Helm (recommended)](https://github.com/signalfx/splunk-otel-collector-chart)
  - [YAML](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/exporter/sapmexporter/examples/signalfx-k8s.yaml)
- Linux
  - [Installer script (recommended)](./docs/getting-started/linux-installer.md)
  - [Standalone](./docs/getting-started/linux-standalone.md)
- Windows
  - [Installer script (recommended)](./docs/getting-started/windows-installer.md)
  - [Standalone](./docs/getting-started/windows-standalone.md)

You can consult additional use cases in the [examples](./examples) directory.

## Supported Components

The distribution offers support for the following components. Support is based
on the [OpenTelemetry maturity
matrix](https://github.com/open-telemetry/community/blob/47813530864b9fe5a5146f466a58bd2bb94edc72/maturity-matrix.yaml#L57).

### Beta

| Receivers        | Processors        | Exporters | Extensions    |
| :--------------: | :--------:        | :-------: | :--------:    |
| hostmetrics      | attributes        | file      | healthcheck   |
| jaeger           | batch             | logging   | httpforwarder |
| k8s_cluster      | filter            | otlp      | host_observer |
| kubeletstats     | k8s_tagger        | sapm      | k8s_observer  |
| opencensus       | memorylimiter     | signalfx  | pprof         |
| otlp             | metrictransform   | splunkhec | zpages        |
| receiver_creator | resource          |           |               |
| sapm             | resourcedetection |           |               |
| signalfx         | span              |           |               |
| simpleprometheus |                   |           |               |
| smartagent       |                   |           |               |
| splunkhec        |                   |           |               |
| zipkin           |                   |           |               |

### Alpha

| Receivers      | Processors | Exporters | Extensions |
| :-------:      | :--------: | :-------: | :--------: |
| carbon         |            |           |            |
| collectd       |            |           |            |
| fluentdforward |            |           |            |
| statsd         |            |           |            |

## Sizing

The OpenTelemetry Collector can be scaled up or out as needed. Sizing is based
on the amount of data per data source and requires 1 CPU core per:

- Traces: 10,000 spans per second
- Metrics: 20,000 data points per second

If a Collector handles both trace and metric data then both must be accounted
for when sizing. For example, 5K spans per second plus 10K data points per
second would require 1 CPU core.

The recommendation is to use a ratio of 1 CPU to 2 GB of memory. By default, the
Collector is configured to use 512 MB of memory.

> The Collector does not persist data to disk so no disk space is required.

### Agent

For Agent instances, scale up resources as needed. Typically only a single
agent runs per application or host so properly sizing the agent is important.
Multiple independent agents could be deployed on a given application or host
depending on the use-case. For example, a privileged agent could be deployed
alongside an unprivileged agent.

### Gateway

For Gateway instances, allocate at least a CPU core per Collector. Note that
multiple Collectors can deployed behind a simple round-robin load balancer for
availability and performance reasons. Each Collector runs independently, so
scale increases linearly with the number of Collectors you deploy.

The recommendation is to configure at least N+1 redundancy, which means a load
balancer and a minimum of two Collector instances should be configured
initially.

## Monitoring

The default configuration automatically scrapes the Collector's own metrics and
sends the data using the `signalfx` exporter. A built-in dashboard provides
information about the health and status of Collector instances.

## Security

Start by reviewing the [Collector security
documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/security.md).

For security information specific to this distribution, please review
[security.md](docs/security.md).

## Troubleshooting

For troubleshooting information specific to this distribution, please review
[troubleshooting.md](docs/troubleshooting.md).

## Migrating from the SignalFx Smart Agent

The Splunk Distribution of OpenTelemetry Collector is the next-generation
agent and gateway for Splunk APM and Splunk Infrastructure Monitoring.
As such, it is the replacement for the [SignalFx Smart
Agent](https://github.com/signalfx/signalfx-agent).

This distribution provides helpful components to assist current Smart Agent
users in their transition to OpenTelemetry Collector and ensure no functionality
loss.  The [Smart Agent
Receiver](./internal/receiver/smartagentreceiver/README.md), its associated
[extension](./internal/extension/smartagentextension/README.md), and other
Collector components provide a means of integrating all Smart Agent metric
monitors into your Collector pipelines.

A detailed overview of our suggested migration practices is provided in the
[migration guide](./docs/signalfx-smart-agent-migration.md).

## License

[Apache Software License version 2.0](./LICENSE).
