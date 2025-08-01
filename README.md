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
  <a href="https://github.com/signalfx/splunk-otel-collector/actions/workflows/build-and-test.yml?query=branch%3Amain">
    <img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/signalfx/splunk-otel-collector/build-and-test.yml?branch=main&style=for-the-badge">
  </a>
  <a href="https://goreportcard.com/report/github.com/signalfx/splunk-otel-collector">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/signalfx/splunk-otel-collector?style=for-the-badge">
  </a>
  <a href="https://github.com/signalfx/splunk-otel-collector/releases">
    <img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/signalfx/splunk-otel-collector?include_prereleases&style=for-the-badge">
  </a>
</p>

<p align="center">
  <strong>
    <a href="docs/architecture.md">Architecture</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="docs/components.md">Components</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="https://docs.splunk.com/observability/en/gdi/opentelemetry/opentelemetry.html#otel-monitoring">Monitoring</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="https://docs.splunk.com/Observability/gdi/opentelemetry/security.html">Security</a>
    &nbsp;&nbsp;&bull;&nbsp;&nbsp;
    <a href="https://docs.splunk.com/Observability/gdi/opentelemetry/sizing.html">Sizing</a>
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
data for [Splunk Observability Cloud](https://www.splunk.com/en_us/products/observability.html):

- [Splunk APM](https://docs.splunk.com/Observability/apm/intro-to-apm.html#nav-Introduction-to-Splunk-APM) via the
  [`otlphttp`
  exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter).
  More information available
  [here](https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html#nav-Install-and-configure-Splunk-Distribution-of-OpenTelemetry-Collector).
- [Splunk Infrastructure
  Monitoring](https://docs.splunk.com/Observability/infrastructure/intro-to-infrastructure.html#nav-Introduction-to-Splunk-Infrastructure-Monitoring)
  via the [`signalfx`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter).
  More information available
  [here](https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html#nav-Install-and-configure-Splunk-Distribution-of-OpenTelemetry-Collector).

While it is recommended to use [Splunk
Forwarders](https://docs.splunk.com/Documentation/Splunk/latest/Data/Usingforwardingagents)
to send data to [Splunk
Cloud](https://www.splunk.com/en_us/products/splunk-cloud-platform.html) or [Splunk
Enterprise](https://www.splunk.com/en_us/products/splunk-enterprise.html),
Splunk OpenTelemetry Collector can be configured to send data to them via the
[`splunk_hec`
exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter).

## Current Status

- The Splunk Distribution of the OpenTelemetry Collector is production tested; it is in use by a number of customers in their production environments
- Customers that use our distribution can receive direct help from official Splunk support within SLA's
- Customers can use or migrate to the Splunk Distribution of the OpenTelemetry Collector without worrying about future breaking changes to its core configuration experience for metrics and traces collection (OpenTelemetry logs collection configuration is in beta). There may be breaking changes to the Collector's own metrics.

## Getting Started

The following resources are available:

- [Architecture](docs/architecture.md): How the Collector can be deployed
- [Components](docs/components.md): What the Collector supports with links to documentation
- [Monitoring](https://docs.splunk.com/observability/en/gdi/opentelemetry/opentelemetry.html#otel-monitoring): How to ensure the Collector is healthy
- [Security](https://docs.splunk.com/Observability/gdi/opentelemetry/security.html): How to ensure the Collector is secure
- [Sizing](https://docs.splunk.com/Observability/gdi/opentelemetry/sizing.html): How to ensure the Collector is properly sized
- [Troubleshooting](docs/troubleshooting.md): How to resolve common issues

All you need to get started is:

- [Splunk Access Token](https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Agent or Gateway mode](https://docs.splunk.com/Observability/gdi/opentelemetry/deployment-modes.html)
- [Confirm exposed
  ports](https://docs.splunk.com/Observability/gdi/opentelemetry/exposed-endpoints.html)
  to make sure your environment doesn't have conflicts and that firewalls are
  configured properly. Ports can be changed in the collector's configuration.

This distribution is supported on and packaged for a variety of platforms including:

- Kubernetes
  - [Helm (recommended)](https://github.com/signalfx/splunk-otel-collector-chart)
  - [Operator (in alpha)](https://github.com/signalfx/splunk-otel-collector-operator)
  - [YAML](https://github.com/signalfx/splunk-otel-collector-chart/tree/main/examples)
- [HashiCorp Nomad](./deployments/nomad)
- Linux
  - [Installer script](./docs/getting-started/linux-installer.md) (recommended for single-host demo/test environments)
  - Configuration management (recommended for multi-host production environments)
    - [Ansible](https://galaxy.ansible.com/ui/repo/published/signalfx/splunk_otel_collector/)
    - [Chef](https://supermarket.chef.io/cookbooks/splunk_otel_collector)
    - [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
    - [Salt](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt)
  - Platform as a Service
    - [Heroku](https://github.com/signalfx/splunk-otel-collector-heroku#getting-started)
  - [Manual](./docs/getting-started/linux-manual.md) including DEB/RPM packages, Docker, and binary
- Windows
  - [Installer script](./docs/getting-started/windows-installer.md) (recommended for single-host demo/test environments)
  - Configuration management (recommended for multi-host production environments)
    - [Ansible](https://galaxy.ansible.com/ui/repo/published/signalfx/splunk_otel_collector/)
    - [Chef](https://supermarket.chef.io/cookbooks/splunk_otel_collector)
    - [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
  - [Manual](./docs/getting-started/windows-manual.md) including MSI with GUI and Powershell, Chocolatey, and Docker

You can consult additional use cases in the [examples](./examples) directory.

## Advanced Configuration

A variety of default configuration files are provided:

- [OpenTelemetry
  Collector](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/otelcol/config/collector)
  see `full_config_linux.yaml` for a commented configuration with links to full
  documentation. The `logs_config_linux.yaml` is a good starting point for using
  the collector for collecting application logs on Linux environments.
  `agent_config.yaml` is the recommended starting configuration for most environments.
- [Fluentd](https://github.com/signalfx/splunk-otel-collector/tree/main/packaging/fpm/etc/otel/collector/fluentd)
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
  - [Extension](https://github.com/signalfx/splunk-otel-collector/tree/main/pkg/extension/smartagentextension)
    offering Collectd and Python extensions
  - [Receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/pkg/receiver/smartagentreceiver)
    offering the complete set of Smart Agent monitors
  - Information about migrating from the SignalFx Smart Agent can be found
    [here](docs/signalfx-smart-agent-migration.md)

The Splunk OpenTelemetry Collector provides a sensitive value-redacting, local config server listening at
`http://localhost:55554/debug/configz/effective` that is helpful in troubleshooting. To enable this feature please
set the `SPLUNK_DEBUG_CONFIG_SERVER` environment variable to `true`. To set the desired port to
listen to configure the `SPLUNK_DEBUG_CONFIG_SERVER_PORT` environment variable.

You can use the environment variable `SPLUNK_LISTEN_INTERFACE` and associated installer option to configure the network
interface on which the collector's receivers and telemetry endpoints will listen.
The default value of `SPLUNK_LISTEN_INTERFACE` is set to `127.0.0.1` for the default agent configuration and `0.0.0.0` otherwise.

## Upgrade guidelines

The following changes need to be done to configuration files for Splunk OTel Collector for specific
version upgrades. We provide automated scripts included in the bundle that cover backward
compatibility on the fly, but configuration files will not be overridden, so you need to update them
manually before the backward compatibility is dropped. For every configuration update use
[the default agent config](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/agent_config.yaml)
as a reference.

### From 0.117.0 to 0.118.0

- The deprecated syntax for config source expansion is no longer supported. 

  Strings like `$ENV` or `$include:/path/to/file.yaml` will no longer be expanded. Instead, use the
  `${env:ENV}` or `${include:/path/to/file.yaml}` syntax. There are only two symbols allowed after `$`: `{` and `$`.
  The collector will log an error and fail to start if it encounters a bare config source.

  Please update your configuration files to use the correct syntax.

### From 0.114.0 to 0.115.0

-  The sapm exporter still works as before but has been deprecated. Use the otlphttp exporter instead

1. Replace the `sapm` exporter with `otlphttp` exporter using the following configuration
    ```yaml
    otlphttp:
      traces_endpoint: "${SPLUNK_INGEST_URL}/v2/trace/otlp"
      headers:
        "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    ```
2. Update traces pipeline to use otlphttp exporter instead of sapm:
    ```
    service:
      pipelines:
        traces:
          exporters: [otlphttp, signalfx]
    ```


### From 0.96.1 to 0.97.0

- `memory_ballast` is no longer effective. The garbage collection is now controlled by the soft memory limit set to 90%
  of total memory (`SPLUNK_MEMORY_TOTAL_MIB` env var) by default.

  If you haven't customized the `memory_ballast`, just remove it from the configuration. 

  If you have customized it via `SPLUNK_BALLAST_SIZE_MIB` (or `extensions::memory_ballast::size_mib` config), you should
  remove the `memory_ballast` extension and use the [`GOMEMLIMIT`](https://pkg.go.dev/runtime) environment variable to set a custom soft memory limit:
  - To decrease frequency of garbage collections: set `GOMEMLIMIT` to a higher value than the default 90% of total memory.
  - To increase frequency of garbage collections: set `GOMEMLIMIT` to a lower value than the default 90% of total memory.

### From 0.68.0 to 0.69.0

- `gke` and `gce` resource detectors in `resourcedetection` processor are replaced with `gcp` resource detector. 
  If you have `gke` and `gce` detectors configured in the `resourcedetection` processor, please update your 
  configuration accordingly. More details: https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10347 

### From 0.41.0 to 0.42.0

- The Splunk OpenTelemetry Collector used to [evaluate user configuration 
  twice](https://github.com/signalfx/splunk-otel-collector/issues/628) and this required escaping of 
  each `$` symbol with `$$` to prevent unwanted environment variable expansion. The issue was fixed in 
  0.42.0 version. Any occurrences of `$$` in your configuration should be replaced with `$`.

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
- Properly configure the Collector for your particular metrics, traces, and logs use cases, as only a minimal default configuration is provided by the contrib release.

An example configuration for upstream, that ensures [infrastructure
correlation](https://docs.splunk.com/observability/en/metrics-and-metadata/relatedcontent-collector-apm.html)
is properly configured, is available
[here](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/upstream_agent_config.yaml).
<!--PRODUCT_DOCS-->

## License

[Apache Software License version 2.0](./LICENSE).

>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk SignalFx](https://www.splunk.com/en_us/about-splunk/acquisitions/signalfx.html) for more information.
