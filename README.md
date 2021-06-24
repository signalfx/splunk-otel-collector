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

<!--PRODUCT_DOCS-->
# Splunk OpenTelemetry Connector

Splunk OpenTelemetry Connector is a distribution of the [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector). It
provides a unified way to receive, process, and export metric, trace, and log
data for [Splunk Observability Cloud](https://www.observability.splunk.com/):

- [Splunk APM](https://www.splunk.com/en_us/software/splunk-apm.html) via the
  [`sapm`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/sapmexporter).
  The [`otlphttp`
  exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)
  can be used with a [custom
  configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/otlp_config_linux.yaml).
  More information available
  [here](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-opentelemetry-collector.html).
- [Splunk Infrastructure
  Monitoring](https://www.splunk.com/en_us/software/infrastructure-monitoring.html)
  via the [`signalfx`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter).
  More information available
  [here](https://docs.signalfx.com/en/latest/otel/imm-otel-collector.html).
- [Splunk Log Observer](https://www.splunk.com/en_us/form/splunk-log-observer.html) via
  the [`splunk_hec`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter).

While it is recommended to use [Splunk
Forwarders](https://www.splunk.com/en_us/products/splunk-enterprise/features/forwarders.html)
to send data to [Splunk
Cloud](https://www.splunk.com/en_us/software/splunk-cloud.html) or [Splunk
Enterprise](https://www.splunk.com/en_us/software/splunk-enterprise.html),
Splunk OpenTelemetry Connector can be configured to send data to them via the
[`splunk_hec`
exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter).

> :construction: This project is currently in **BETA** ([what does beta mean?](docs/beta-definition.md))

## Getting Started

All you need to get started is:

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Agent or Gateway mode](docs/agent-vs-gateway.md)

- [Double-check exposed
  ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints)
  to make sure your environment doesn't have conflicts. Ports can be changed in
  the collector's configuration.

This distribution is supported on and packaged for a variety of platforms including:

- Kubernetes
  - [Helm (recommended)](https://github.com/signalfx/splunk-otel-collector-chart)
  - [YAML](https://github.com/signalfx/splunk-otel-collector-chart/tree/main/rendered)
- Linux
  - [Installer script (recommended)](./docs/getting-started/linux-installer.md)
  - Configuration management
    - [Ansible](https://galaxy.ansible.com/signalfx/splunk_otel_collector)
    - [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
  - [Manual](./docs/getting-started/linux-manual.md) including DEB/RPM packages, Docker, and binary
- Windows
  - [Installer script (recommended)](./docs/getting-started/windows-installer.md)
  - [Manual](./docs/getting-started/windows-manual.md) including MSI with GUI and Powershell

You can consult additional use cases in the [examples](./examples) directory.

## More Information

- [Architecture](docs/architecure.md) - way in which the Connector can be deployed
- [Components](docs/components.md) - what the Connector supports with links to documentation
- [Monitoring](docs/monitoring.md) - how to ensure the Connector is healthy
- [Security](docs/security.md) - how to ensure the Connector is secure
- [Sizing](docs/sizing.md) - how to ensure the Connector is properly sized
- [Troubleshooting](docs/troubleshooting.md) - how to resolve common issues

Information about migrating from the SignalFx Smart Agent can be found
[here](docs/signalfx-smart-agent-migration.md).
<!--PRODUCT_DOCS-->

## License

[Apache Software License version 2.0](./LICENSE).
