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
  <a href="https://github.com/signalfx/splunk-otel-collector/actions/workflows/build-and-test.yml?query=branch%3Amain">
    <img alt="Build Status" src="https://img.shields.io/github/workflow/status/signalfx/splunk-otel-collector/build-and-test?style=for-the-badge">
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
# Splunk OpenTelemetry Collector

Splunk OpenTelemetry Collector is a distribution of the [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector). It
provides a unified way to receive, process, and export metric, trace, and log
data for [Splunk Observability Cloud](https://www.splunk.com/en_us/observability.html):

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
Splunk OpenTelemetry Collector can be configured to send data to them via the
[`splunk_hec`
exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter).

> :construction: This project is currently in **BETA** ([what does beta mean?](docs/beta-definition.md))

## Getting Started

The following resources are available:

- [Architecture](docs/architecture.md): How the Collector can be deployed
- [Components](docs/components.md): What the Collector supports with links to documentation
- [Monitoring](docs/monitoring.md): How to ensure the Collector is healthy
- [Security](docs/security.md): How to ensure the Collector is secure
- [Sizing](docs/sizing.md): How to ensure the Collector is properly sized
- [Troubleshooting](docs/troubleshooting.md): How to resolve common issues

All you need to get started is:

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Agent or Gateway mode](docs/agent-vs-gateway.md)
- [Confirm exposed
  ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints)
  to make sure your environment doesn't have conflicts and that firewalls are
  configured properly. Ports can be changed in the collector's configuration.

This distribution is supported on and packaged for a variety of platforms including:

- Kubernetes
  - [Helm (recommended)](https://github.com/signalfx/splunk-otel-collector-chart)
  - [Operator (in alpha)](https://github.com/signalfx/splunk-otel-collector-operator)
  - [YAML](https://github.com/signalfx/splunk-otel-collector-chart/tree/main/rendered)
- [HashiCorp Nomad](./deployments/nomad)
- Linux
  - [Installer script (recommended)](./docs/getting-started/linux-installer.md)
  - Configuration management
    - [Ansible](https://galaxy.ansible.com/signalfx/splunk_otel_collector)
    - [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
  - Platform as a Service
    - [Heroku](https://github.com/signalfx/splunk-otel-collector-heroku#getting-started)
  - [Manual](./docs/getting-started/linux-manual.md) including DEB/RPM packages, Docker, and binary
- Windows
  - [Installer script (recommended)](./docs/getting-started/windows-installer.md)
  - Configuration management
    - [Ansible](https://galaxy.ansible.com/signalfx/splunk_otel_collector)
    - [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
  - [Manual](./docs/getting-started/windows-manual.md) including MSI with GUI and Powershell, Chocolatey, and Docker

You can consult additional use cases in the [examples](./examples) directory.

## Advanced Configuration

A variety of default configuration files are provided:

- [OpenTelemetry
  Collector](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/otelcol/config/collector)
  see `full_config_linux.yaml` for a commented configuration with links to full
  documentation. `agent_config.yaml` is the recommended starting
  configuration for most environments.
- [Fluentd](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd)
  applicable to Helm or installer script installations only. See the `*.conf`
  files as well as the `conf.d` directory. Common sources including filelog,
  journald, and Windows event viewer are included.

In addition, the following components can be configured:

- Configuration sources
  - [Environment variables](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/envvarconfigsource)
  - [Etcd2](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/etcd2configsource)
  - [Include](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/includeconfigsource)
  - [Vault](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/vaultconfigsource)
  - [Zookeeper](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/zookeeperconfigsource)
- SignalFx Smart Agent
  - [Extension](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/extension/smartagentextension)
    offering Collectd and Python extensions
  - [Receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/smartagentreceiver)
    offering the complete set of Smart Agent monitors
  - Information about migrating from the SignalFx Smart Agent can be found
    [here](docs/signalfx-smart-agent-migration.md)

## Upgrade guidelines

The following changes need to be done to configuration files for Splunk OTel Collector for specific
version upgrades. We provide automated scripts included in the bundle that cover backward
compatibility on the fly, but configuration files will not be overridden, so you need to update them
manually before the backward compatibility is dropped. For every configuration update use
[the default agent config](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/agent_config.yaml)
as a reference.

### From 0.35.0 to 0.36.0

- Configuration parameter "`exporters` -> `otlp` -> `insecure`" is moved to
  "`exporters` -> `otlp` -> `tls` -> `insecure`".
  
  More details: https://github.com/open-telemetry/opentelemetry-collector/pull/4063/.
  
  Configuration part for `otlp` exporter should look like this:

  ```yaml
  exporters:
    otlp:
      endpoint: "${SPLUNK_GATEWAY_URL}:4317"
      tls:
        insecure: true
  ```

### From 0.34.0 to 0.35.0

- `ballast_size_mib` parameter moved from `memory_limiter` processor to `memory_ballast` extension
  as `size_mib`.
  
  More details: https://github.com/signalfx/splunk-otel-collector/pull/567.

  Remove `ballast_size_mib` parameter from `memory_limiter` and make sure that it's added to
  `memory_ballast` extension as `size_mib` parameter instead:

  ```yaml
  extensions:
    memory_ballast:
      size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
  ```

### Using Upstream OpenTelemetry Collector

It is possible to use the upstream OpenTelemetry Collector instead of this
distribution. The following features are not available upstream at this time:

- Packaging
  - Installer scripts for Linux and Windows
  - Configuration management via Ansible or Puppet
- Configuration sources
- Several SignalFx Smart Agent capabilities

:warning: Splunk only provides best-effort support for upstream OpenTelemetry

In order to use the upstream OpenTelemetry Collector:

- Use the
  [contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)
  distribution as commercial exporters must reside in contrib
- Properly configuration the Collector

An example configuration for upstream, that ensures [infrastructure
correlation](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/apm-infra-correlation.md)
is properly configured, is available
[here](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/upstream_agent_config.yaml).
<!--PRODUCT_DOCS-->

## License

[Apache Software License version 2.0](./LICENSE).

>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk SignalFx](https://www.splunk.com/en_us/investor-relations/acquisitions/signalfx.html) for more information.
