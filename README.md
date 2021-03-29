---

<p align="center">
  <strong>
    <a href="#getting-started">Getting Started</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="CONTRIBUTING.md">Getting Involved</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/signalfx-smart-agent-migration.md">Migrating from Smart Agent</a>
  </strong>
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/signalfx/splunk-otel-collector">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/signalfx/splunk-otel-collector?style=for-the-badge">
  </a>
  <a href="https://circleci.com/gh/signalfx/splunk-otel-collector">
    <img alt="Build Status" src="https://img.shields.io/circleci/build/github/signalfx/splunk-otel-collector?style=for-the-badge">
  </a>
  <a href="https://codecov.io/gh/signalfx/splunk-otel-collector/branch/main/">
    <img alt="Codecov Status" src="https://img.shields.io/codecov/c/github/signalfx/splunk-otel-collector?style=for-the-badge">
  </a>
  <a href="https://github.com/signalfx/splunk-otel-collector/releases">
    <img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/signalfx/splunk-otel-collector?include_prereleases&style=for-the-badge">
  </a>
  <img alt="Beta" src="https://img.shields.io/badge/status-beta-informational?style=for-the-badge">
</p>

<p align="center">
  <strong>
    <a href="docs/architecture.md">Architecture</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/components.md">Components</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/monitoring.md">Monitoring</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/security.md">Security</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/sizing.md">Sizing</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/troubleshooting.md">Troubleshooting</a>
  </strong>
</p>

---

# Splunk Distribution of OpenTelemetry Collector

The Splunk Distribution of [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector) provides
a binary that can receive, process and export trace, metric and log data.

**Installations that use this distribution can receive direct help from
Splunk's support teams.** Customers are free to use the core OpenTelemetry OSS
components (several do!) and we will provide best effort guidance to them for
any issues that crop up, however only the Splunk distributions are in scope for
official Splunk support and support-related SLAs.

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

## License

[Apache Software License version 2.0](./LICENSE).
