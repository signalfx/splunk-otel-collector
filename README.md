# Splunk distribution of OpenTelemetry Collector

The Splunk distribution of [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector) provides
a binary that can be deployed as a standalone service (also known as a gateway)
that can receive, process and export trace, metric and log data. This
distribution is supported by Splunk.

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
- [Splunk Cloud](https://www.splunk.com/en_us/software/splunk-cloud.html) or
  [Splunk
  Enterprise](https://www.splunk.com/en_us/software/splunk-enterprise.html) via
  the [`splunk_hec`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/splunkhecexporter).

> :construction: This project is currently in **BETA**.

## Getting Started

The Collector is supported on and packaged for a variety of platforms including:

- Kubernetes
  - Helm (coming soon!)
  - [YAML](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/exporter/sapmexporter/examples/signalfx-k8s.yaml)
- Linux
  - [Installer script (recommended)](./docs/getting-started/linux-installer.md)
  - [Standalone](./docs/getting-started/linux-standalone.md)
- Windows
  - Install script (coming soon!)
  - [Standalone](./docs/getting-started/windows-standalone.md)

You can consult additional use cases in the [examples](./examples) directory.

## Supported Components

### Receivers

| Name             | Status |
| :--:             | :----: |
| carbon           | Alpha  |
| collectd         | Alpha  |
| fluentforward    | Beta   |
| hostmetrics      | Beta   |
| jaeger           | Beta   |
| k8scluster       | Beta   |
| kubletstats      | Beta   |
| opencensus       | Beta   |
| otlp             | Beta   |
| sapm             | Beta   |
| signalfx         | Beta   |
| simpleprometheus | Beta   |
| statsd           | Alpha  |
| splunkhec        | Beta   |
| zipkin           | Beta   |

### Processors

| Name              | Status |
| :--:              | :----: |
| attributes        | Beta   |
| batch             | Beta   |
| filter            | Beta   |
| memorylimiter     | Beta   |
| resource          | Beta   |
| span              | Beta   |
| k8s               | Beta   |
| metricstransform  | Beta   |
| resourcedetection | Beta   |

### Exporters

| Name             | Status |
| :--:             | :----: |
| file             | Beta   |
| logging          | Beta   |
| otlp             | Beta   |
| sapm             | Beta   |
| signalfx         | Beta   |
| splunkhec        | Beta   |

### Extensions

| Name          | Status |
| :--:          | :----: |
| fluentbit     | Beta   |
| healthcheck   | Beta   |
| httpforwarder | Beta   |
| observer/host | Beta   |
| observer/k8s  | Beta   |
| pprof         | Beta   |
| zpages        | Beta   |

## Sizing

The OpenTelemetry Collector can be scaled up or out as needed. Sizing is based
on the amount of data per data source and requires 1 CPU core per:

- Traces: 10,000 spans per second
- Metrics: 20,000 data points per second

If a Collector handles both trace and metric data then both must be accounted
for when sizing. For example, 5K spans per second plus 10K data points per
second would require 1 CPU core.

The recommendation is to use a ratio of 1:2 for CPU:memory and to allocate at
least a CPU core per Collector. Multiple Collectors can deployed behind a
simple round-robin load balancer. Each Collector runs independently, so scale
increases linearly with the number of Collectors you deploy.

> The Collector does not persist data to disk so no disk space is required.

## Monitoring

The default configuration automatically scrapes the Collector's own metrics and
sends the data using the `signalfx` exporter. A built-in dashboard provides
information about the health and status of Collector instances.

## Troubleshooting

See the [Collector troubleshooting
documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/troubleshooting.md).

## License

[Apache Software License version 2.0](./LICENSE).
