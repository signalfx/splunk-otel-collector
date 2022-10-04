# Changelog

## Unreleased

## v0.61.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.61.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.61.0) and the [opentelemetry-collector-contrib v0.61.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.61.0) releases.

### 💡 Enhancements 💡

- `signalfx` exporter: Drop datapoints with more than 36 dimensions [#14625](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/14625)
- Security updates for third-party dependencies

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Reduce severity of logged unsupported config fields warning [#2072](https://github.com/signalfx/splunk-otel-collector/pull/2072)

## v0.60.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.60.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.60.0) and the [opentelemetry-collector-contrib v0.60.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.60.0) releases.

### 💡 Enhancements 💡

- Update auto instrumentation java agent to [v1.16.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.16.0)
- Replace usage of Map.Insert* and Map.Update* with Map.Upsert (#1957)
- Refactor main flags as settings.Settings (#1952)
- Support installing with ansible and skipping restart of services (#1930)

## v0.59.1

### 💡 Enhancements 💡

- Upgrade to include changes from [SignalFx Smart Agent v5.23.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.23.0)
- Add `processlist` and `resourcedetection` to default config

## v0.59.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.59.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.59.0) and the [opentelemetry-collector-contrib v0.59.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.59.0) releases.

### 💡 Enhancements 💡

- Upgrade Golang to 1.19
- debug/configz: Address multiple confmap.Providers for service config and index debug/configz/initial by provider scheme.
- Add tar.gz distribution of Splunk Collector
- Update default gateway config to sync host metadata by default

## v0.58.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.58.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.58.0) and the [opentelemetry-collector-contrib v0.58.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.58.0) releases.

### 💡 Enhancements 💡

- Update auto instrumentation java agent to [v1.14.2](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.14.2)

## v0.57.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.57.2](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.57.2) and the [opentelemetry-collector-contrib v0.57.2](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.57.2) releases.

### 💡 Enhancements 💡

- Include [`sqlquery` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.57.2/receiver/sqlqueryreceiver/README.md)(#1833)
- Security updates for third-party dependencies

## v0.56.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.56.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.56.0) and the [opentelemetry-collector-contrib v0.56.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.56.0) releases.

### 💡 Enhancements 💡

- Add the `--collector-config` option to the Linux installer script to allow a custom config file path (#1806)
- Update auto instrumentation java agent to [v1.14.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.14.0)
- Update bundled Smart Agent to [v5.22.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.22.0)

### 🧰 Bug fixes 🧰

- `signalfx` exporter: Fix invalid error response message [#12654](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12654)

## v0.55.1

### 🧰 Bug fixes 🧰

- `pulsar` exporter: Removed pulsar producer name from config to avoid producer name conflict (#1782)

## v0.55.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.55.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.55.0) and the [opentelemetry-collector-contrib v0.55.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.55.0) releases.

### 💡 Enhancements 💡

- Update default `td-agent` version to 4.3.2 in the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) to support log collection with fluentd on Ubuntu 22.04
- Include [tail_sampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor) and [span_metrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/spanmetricsprocessor) in our distribution

### 🧰 Bug fixes 🧰

- Correct invalid environment variable expansion for ECS task metadata endpoints on EC2 (#1764)
- Adopt [metricstransformprocessor empty metrics fix](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12211)

## v0.54.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.54.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.54.0) and the [opentelemetry-collector-contrib v0.54.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.54.0) releases.

### 💡 Enhancements 💡

- Only use config server if env var unset (#1728)
- Update bundled Smart Agent to [v5.21.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.21.0)

### 🧰 Bug fixes 🧰

- Wrap log messages for windows support bundle (#1725)

## v0.53.1

### 🧰 Bug fixes 🧰

- Upgrade [`metricstransform` 
  processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor) 
  to pick up [migration from OpenCensus data model to 
  OTLP](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10817) that fixes a few issues with 
  the processor.

## v0.53.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.53.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.53.0) and the [opentelemetry-collector-contrib v0.53.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.53.0) releases.

### 🚀 New components 🚀

- [`k8sevents` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver)
  to collect Kubernetes events in OpenTelemetry semantics (#1641)
- **Experimental**: [`pulsar` exporter](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/exporter/pulsarexporter) to export metrics to Pulsar (#1683)

## v0.52.2

### 💡 Enhancements 💡

- Upgrade Golang to 1.18.3 (#1633)
- Support multiple `--config` command-line arguments (#1576)

### 🧰 Bug fixes 🧰

- [`kubeletstats` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver) introduced a regression in version 52.0 that can break metrics for Kubernetes pods and containers, pinning this receiver's version to v0.51.0 until the regression is resolved (#1638)

## v0.52.1

### 🚀 New components 🚀

- [`transform` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor) to modify telemetry based on configuration using the [Telemetry Query Language](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/processing.md#telemetry-query-language) (Alpha)

### 💡 Enhancements 💡

- Initial release of [Chef cookbook](https://supermarket.chef.io/cookbooks/splunk_otel_collector) for Linux and Windows

## v0.52.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.52.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.52.0) and the [opentelemetry-collector-contrib v0.52.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.52.0) releases.

### 💡 Enhancements 💡

- Add Ubuntu 22.04 support to the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible), [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet), and [Salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) (collector only; log collection with Fluentd [not currently supported](https://www.fluentd.org/blog/td-agent-v4.3.1-has-been-released))

## v0.51.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.51.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.51.0) and the [opentelemetry-collector-contrib v0.51.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.51.0) releases.

Additionally, this release includes [an update to the `resourcedetection` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10015) to support "cname" and "lookup" hostname sources.

### 🛑 Breaking changes 🛑

- Removed Debian 8 (jessie) support from the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (#1354), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible) (#1547), and [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet) (#1545)

### 💡 Enhancements 💡

- Added Debian 11 (bullseye) support to the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (#1354), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible) (#1547), [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet) (#1545), and [Salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) (#1546)
- Upgrade Golang to 1.18.2 (#1551)

## v0.50.1

### 💡 Enhancements 💡

- Security updates for third-party dependencies
- Update bundled Smart Agent to [v5.20.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.20.1)

## v0.50.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.50.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.50.0) and the [opentelemetry-collector-contrib v0.50.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.50.0) releases.

Additionally, this release includes [an update to `k8scluster` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/9523) that allows it to run on older k8s clusters (1.20-).

## v0.49.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.49.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.49.0) and the [opentelemetry-collector-contrib v0.49.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.49.0) releases.

### 🚀 New components 🚀

- [`syslog` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver) parses Syslogs from tcp/udp using the [opentelemetry-log-collection](https://github.com/open-telemetry/opentelemetry-log-collection) library

### 💡 Enhancements 💡

- Updated the [Migrating from SignalFx Smart Agent to Splunk Distribution of OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/signalfx-smart-agent-migration.md) documentation (#1489)
- Upgrade to Go 1.18.1 (#1464)
- Initial support for [Cloud Foundry Buildpack](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/buildpack) (#1404)
- Initial support for [BOSH Release](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/bosh) (#1480)
- Update bundled Smart Agent to [v5.20.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.20.0)

## v0.48.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.48.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.48.0) and the [opentelemetry-collector-contrib v0.48.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.48.0) releases.

### 🚀 New components 🚀

- [`cloudfoundry` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/cloudfoundryreceiver)
  to receive metrics from Cloud Foundry deployments and services.

## v0.47.1

### 🧰 Bug fixes 🧰

- Remove `signalfx` exporter from traces pipeline in default gateway config (#1393)
- Update `github.com/open-telemetry/opentelemetry-log-collection` to [v0.27.1](https://github.com/open-telemetry/opentelemetry-log-collection/releases/tag/v0.27.1) to fix logging pipeline issues after upgrading to Go 1.18 (#1418)

## v0.47.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.47.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.47.0) and the [opentelemetry-collector-contrib v0.47.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.47.0) releases.

### 🚀 New components 🚀

- [`tcplog` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/tcplogreceiver) to receive logs from tcp using the [opentelemetry-log-collection](https://github.com/open-telemetry/opentelemetry-log-collection) library

### 💡 Enhancements 💡

- Upgrade to Go 1.18 (#1380)

### 🧰 Bug fixes 🧰

- Update core version during build (#1379)
- Update SA event type to fix processlist (#1385)

## v0.46.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.46.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.46.0) and the [opentelemetry-collector-contrib v0.46.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.46.0) releases.

### 💡 Enhancements 💡

- Upgrade `hostmetrics` receiver dependency (#1341)
- Update Linux installer script to fail immediately if running on an unsupported Linux distribution (#1351)
- Update bundled Smart Agent to [v5.19.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.19.1)

### 🧰 Bug fixes 🧰

- As a bug fix for hosts number miscalculation in Splunk Observability Cloud, Splunk OpenTelemetry Collector running in 
  agent mode now is configured to override `host.name` attribute of all signals sent from instrumentation libraries by 
  default (#1307)

## v0.45.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.45.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.45.0) and the [opentelemetry-collector-contrib v0.45.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.45.1) releases.

### 💡 Enhancements 💡

- Upgrade golang to 1.17.7 (#1294)

### 🧰 Bug fixes 🧰

- Correct collectd/hadoopjmx monitor type in windows Smart Agent receiver config validation [#1254](https://github.com/signalfx/splunk-otel-collector/pull/1254)

## v0.44.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.44.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.44.0) and the [opentelemetry-collector-contrib v0.44.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.44.0) releases.

### 🚀 New components 🚀

- [`databricks` receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/databricksreceiver) to generate metrics about the operation of a Databricks instance (Alpha)

### 💡 Enhancements 💡

- Bump default `td-agent` version to 4.3.0 in installer scripts (#1205)
- Enable shared pipeline for profiling by default (#1181)
- Update bundled Smart Agent to [v5.19.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.19.0)

## v0.43.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.43.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.43.1) and the [opentelemetry-collector-contrib v0.43.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.43.0) releases.

### 🧰 Bug fixes 🧰

- Provide informative unsupported monitor error on Windows for Smart Agent receiver [#1150](https://github.com/signalfx/splunk-otel-collector/pull/1150)
- Fix Windows support bundle script if fluentd is not installed (#1162)

## v0.42.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.42.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.42.0) and the [opentelemetry-collector-contrib v0.42.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.42.0) releases.

### 🛑 Breaking changes 🛑

- This version adopts OpenTelemetry Core version 0.42.0, and in doing so the configuration parsing process has changed slightly. The Splunk OpenTelemetry Collector used to [evaluate user configuration twice](https://github.com/signalfx/splunk-otel-collector/issues/628) and this required escaping desired `$` literals with an additional `$` character to prevent unwanted environment variable expansion. This version no longer doubly evaluates configuration so any `$$` instances in your configuration as a workaround should be updated to `$`.  [Config source directives](./internal/configsource) that include an additional `$` are provided with a temporary, backward-compatible `$${config_source:value}` and `$$config_source:value` parsing rule controlled by `SPLUNK_DOUBLE_DOLLAR_CONFIG_SOURCE_COMPATIBLE` environment variable (default `"true"`) to migrate them to single `$` usage to continue supporting the updating configs from [#930](https://github.com/signalfx/splunk-otel-collector/pull/930) and [#935](https://github.com/signalfx/splunk-otel-collector/pull/935). This functionality will be removed in a future release (#1099)

### 🚀 New components 🚀

- [`docker_observer`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/dockerobserver) to detect and create container endpoints, to be used with the [`receiver_creator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator) (#1044)
- [`ecs_task_observer`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecstaskobserver) to detect and create ECS task container endpoints, to be used with the [`receiver_creator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator) (#1125)

### 💡 Enhancements 💡

- Initial [salt module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) for Linux (#1050)
- Update bundled Smart Agent to [v5.18.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.18.0)

### 🧰 Bug fixes 🧰

- [`smartagent` receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/smartagentreceiver) will now attempt to create _any_ monitor from a Receiver Creator instance, disregarding its provided `endpoint`. Previously would error out if a monitor did not accept endpoints ([#1107](https://github.com/signalfx/splunk-otel-collector/pull/1107))
- Remove `$$`-escaped `env` config source usage in ECS configs ([#1139](https://github.com/signalfx/splunk-otel-collector/pull/1139)).

## v0.41.1

- Upgrade golang to 1.17.6 (#1088)

## v0.41.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.41.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.41.0) and the [opentelemetry-collector-contrib v0.41.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.41.0) releases.

### 🚀 New components 🚀

- [`journald` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/journaldreceiver) to parse journald events from systemd journal

### 💡 Enhancements 💡

- Update bundled Smart Agent to [v5.17.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.17.1)
- Update OTLP HTTP receiver endpoint to use port 4318 in default configuration files (#1017)

## v0.40.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.40.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.40.0) and the [opentelemetry-collector-contrib v0.40.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.40.0) releases.

### 🚀 New components 🚀

- [mongodbatlas](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbatlasreceiver) receiver to receive metrics from MongoDB Atlas via their monitoring APIs (#997)
- [routing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/routingprocessor) processor to route logs, metrics or traces to specific exporters (#982)

## v0.39.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.39.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.39.0) and the [opentelemetry-collector-contrib v0.39.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.39.0) releases.

### 💡 Enhancements 💡

- Initial [Chocolatey package](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/windows-manual.md#chocolatey-installation) release
- Update bundled Smart Agent to [v5.16.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.16.0)

### 🧰 Bug fixes 🧰

- Fix token passthrough for splunkhec receiver/exporter ([#5435](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5435))
- Fix --set command line flag functionality (#939)

## v0.38.1

### 🧰 Bug fixes 🧰

- Fix evaluating env variables in ecs ec2 configs (#930)
- Correct certifi CA bundle removal from Smart Agent bundle (#933)
- Fix evaluating env variables in fargate config (#935)

## v0.38.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.38.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.38.0) and the [opentelemetry-collector-contrib v0.38.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.38.0) releases.

### 💡 Enhancements 💡

- Initial release of multi-arch manifest for amd64 and arm64 linux docker images (#866)
  - **Note:** The Smart Agent and Collectd bundle is only included with the amd64 image
- Enable otlp receiver in the gateway logs pipeline (#903)
- Update bundled Smart Agent to [v5.15.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.15.0)

## v0.37.1

### 💡 Enhancements 💡

- Initial release of [`migratecheckpoint`](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/migratecheckpoint) to migrate Fluentd's position file to Otel checkpoints
- Upgrade golang to v1.17.2 for CVE-2021-38297
- Upgrade `github.com/hashicorp/consul/api` to v1.11.0 for CVE-2021-37219
- Upgrade `github.com/hashicorp/vault` to v1.7.2 for CVE-2021-27400, CVE-2021-29653, and CVE-2021-32923
- Upgrade `github.com/jackc/pgproto3/v2` to v2.1.1
- Upgrade `go.etcd.io/etcd` to `go.etcd.io/etcd/client/v2` for CVE-2020-15114
- Remove test certs from the smart agent bundle (#861)
- Run the `otelcol` container process as non-root user in provided docker image (#864)

### 🧰 Bug fixes 🧰

- Temporarily downgrade `gopsutil` dep to avoid errors in k8s deployment (#877)

## v0.37.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.37.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.37.0) and the [opentelemetry-collector-contrib v0.37.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.37.1) releases. Due to go modules dep issues, the Collector Contrib release 0.37.0 has been retracted in favor of 0.37.1.

### 💡 Enhancements 💡

- `signalfx` exporter: Add support for per cpu metrics [#5756](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5756)
- Add [Hashicorp Nomad](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/nomad) support (#819)
- Add config converter function to unsquash Splunk HEC exporter tls fields (#832)
- Rename `k8s_tagger` processor config entries to [`k8sattributes`](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5384) (#848)
- Update bundled Smart Agent to [v5.14.2](https://github.com/signalfx/signalfx-agent/releases/tag/v5.14.2)

## v0.36.1

### 🚀 New components 🚀

- [`httpsink` exporter](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/exporter/httpsinkexporter) to make span data available via a HTTP endpoint
- Initial release of [`translatesfx`](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/translatesfx) to translate a SignalFx Smart Agent configuration file into a configuration that can be used by an OpenTelemetry Collector

### 🛑 Breaking changes 🛑

- Reorder detectors in default configs, moving the `system` detector to the 
  end of the list. Applying this change to a pre-existing config in an EC2
  or Azure deployment will change both the `host.name` dimension and the 
  resource ID dimension on some MTSes, possibly causing detectors to fire. 
  (#822)

### 💡 Enhancements 💡

- Add `--skip-collector-repo` and `--skip-fluentd-repo` options to the Linux installer script to skip apt/yum/zypper repo config (#801)
- Add `collector_msi_url` and `fluentd_msi_url` options to the Windows installer script to allow custom URLs for downloading MSIs (#803)
- Start collector service after deb/rpm install or upgrade if env file exists (#805)

### 🧰 Bug fixes 🧰

- Allow the version flag without environment variables (#800)
- Fix Linux installer to set `SPLUNK_MEMORY_TOTAL_MIB` in the environment file if `--ballast` option is specified (#807)

## v0.36.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.36.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.36.0) and the [opentelemetry-collector-contrib v0.36.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.36.0) releases.

### 💡 Enhancements 💡

- Handle backwards compatibility of internal ballast removal (#759)
- Update bundled Smart Agent to [v5.14.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.14.1)
- Automatically relocate removed OTLP exporter "insecure" field (#783)

### 🧰 Bug fixes 🧰

- Move Heroku buildpack to [https://github.com/signalfx/splunk-otel-collector-heroku](https://github.com/signalfx/splunk-otel-collector-heroku) (#755)
- Fix rpm installation conflicts with the Smart Agent rpm (#773)

## v0.35.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.35.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.35.0) and the [opentelemetry-collector-contrib v0.35.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.35.0) releases.

### 🚀 New components 🚀

- [`groupbyattrs` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)

### 💡 Enhancements 💡

- Update bundled Smart Agent to [v5.13.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.13.0) (#738)
- Add SUSE support to [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (collector only, log collection with Fluentd not yet supported) (#720)
- Add SUSE support to [puppet module](https://forge.puppet.com/modules/signalfx/splunk_otel_collector) (collector only, log collection with Fluentd not yet supported) (#737)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Properly parse receiver creator endpoints (#718)

## v0.34.1

### 💡 Enhancements 💡

- Automatically add `system.type` dimension to all `smartagent` receiver datapoints (#702)
- Include ECS EC2 config in docker images (#713)

## v0.34.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.34.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.34.0) and the [opentelemetry-collector-contrib v0.34.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.34.0) releases.

### 💡 Enhancements 💡

- Add [Amazon ECS EC2](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ecs/ec2) deployment support (#642)
- Enable `memory_ballast` extension in Fargate config (#675)
- Initial [support bundle PowerShell script](https://github.com/signalfx/splunk-otel-collector/blob/main/internal/buildscripts/packaging/msi/splunk-support-bundle.ps1); included in the Windows MSI (#654)
- Remove strict `libcap` dependency from the collector RPM (#676)
  - Allows installation on Linux distros without the `libcap` package.
  - If installing the collector RPM manually, `libcap` will now need to be installed separately as a prerequisite.  See [linux-manual.md](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-manual.md#deb-and-rpm-packages) for details.

### 🧰 Bug fixes 🧰

- Use system env vars for default paths in the Windows installer script (#667)

## v0.33.1

### 💡 Enhancements 💡

- Initial release of the `quay.io/signalfx/splunk-otel-collector-windows` [docker image for Windows](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/windows-manual.md#docker)
- Upgrade to Go 1.17 (#650)
- Update bundled Smart Agent to [v5.12.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.12.0)

## v0.33.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.33.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.33.0) and the [opentelemetry-collector-contrib v0.33.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.33.0) releases.

### 💡 Enhancements 💡

- `smartagent` receiver: `signalfx-forwarder` now works with `k8s_tagger` processor. (#590)
- Add [Fargate](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/fargate/README.md) deployment support
- Update bundled Smart Agent to [v5.11.4](https://github.com/signalfx/signalfx-agent/releases/tag/v5.11.4)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Set redirected logrus logger level (#593)

## v0.31.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.31.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.31.0) and the [opentelemetry-collector-contrib v0.31.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.31.0) releases.

### 🚀 New components 🚀

- [`file_storage` extension](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage)

### 🛑 Breaking changes 🛑

- Change default config server port to 55554 (#579)

### 💡 Enhancements 💡

- Add support for non-persisted journald in the default fluentd config (#516)
- Add `SPLUNK_CONFIG_YAML` env var support for storing configuration YAML (#462)
- Initial puppet support for windows (#524)
- Update to use the `memory_ballast` extension instead of the `--mem-ballast-size-mib` flag (#567)
- Add Heroku buildpack (#571)
- Set required URL and TOKEN env vars for agent config (#572)

### 🧰 Bug fixes 🧰

- Remove SAPM receiver from default configuration (#517)
- `zookeeper` config source: Remove config validation for zk endpoints (#533)
- Fix memory limit calculation for deployments with 20Gi+ of total memory (#558)
- Set path ownership on deb/rpm postinstall (#582)

## v0.29.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.29.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.29.0) and the [opentelemetry-collector-contrib v0.29.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.29.0) releases.

### 💡 Enhancements 💡

- Add OTLP to logs pipeline for agent (#495)
- Enable collecting in memory config locally by default (#497)
- Enable host metadata updates by default (#513)

## v0.28.1

- Update bundled Smart Agent to [v5.11.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.11.0) (#487)
- Document APM infra correlation (#458)
- Alpha translatesfx feature additions.

## v0.28.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.28.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.28.0) and the [opentelemetry-collector-contrib v0.28.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.28.0) releases.

### 💡 Enhancements 💡

- Initial puppet module for linux (#405)
- Add `include` config source (#419, #402, #397)
- Allow setting both `SPLUNK_CONFIG` and `--config` with priority given to `--config` (#450)
- Use internal pipelines for collector prometheus metrics (#469)

### 🧰 Bug fixes 🧰

- Correctly handle nil value on the config provider (#434)

## v0.26.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.26.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.26.0) and the [opentelemetry-collector-contrib v0.26.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.26.0) releases.

### 🚀 New components 🚀

- [kafkametrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver) receiver

### 💡 Enhancements 💡

- zookeeper config source (#318)
- etcd2 config source (#317)
- Enable primary cloud resource detection in the default agent config (#344)
- Unset exclusion and translations by default in gateway config (#350)
- Update bundled Smart Agent to [v5.10.2](https://github.com/signalfx/signalfx-agent/releases/tag/v5.10.2) (#354)
- Set PATH in the docker image to include Smart Agent bundled utilities (#313)
- Remove 55680 exposed port from the docker image (#371)
- Expose initial and effective config for debugging purposes (#325)
- Add a config source for env vars (#348)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Remove premature protection for Start/Stop, trust Service to start/stop once (#342)
- `smartagent` receiver and extension: Fix config parsing for structs and pointers to structs (#345)

## v0.25.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.25.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.25.0) and the [opentelemetry-collector-contrib v0.25.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.25.0) releases.

### 🚀 New components 🚀

- [filelog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) receiver (#289)
- [probabilisticsampler](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) processor (#300)

### 💡 Enhancements 💡

- Add the config source manager (#295, #303)

### 🧰 Bug fixes 🧰

- Correct Jaeger Thrift HTTP Receiver URL to /api/traces (#288)

## v0.24.3

### 💡 Enhancements 💡

- Add AKS resource detector (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3035)

### 🧰 Bug fixes 🧰

- Fallback to `os.Hostname` when FQDN is not available (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3099)

## v0.24.2

### 💡 Enhancements 💡

- Include smart agent bundle in docker image (#241)
- Use agent bundle-relative Collectd ConfigDir default (#263, #268)

### 🧰 Bug fixes 🧰

- Sanitize monitor IDs in SA receiver (#266, #269)

## v0.24.1

### 🧰 Bug fixes 🧰

- Fix HEC Exporter throwing 400s (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3032)

### 💡 Enhancements 💡
- Remove unnecessary hostname mapping in fluentd configs (#250)
- Add OTLP HTTP exporter (#252)
- Env variable NO_WINDOWS_SERVICE to force interactive mode on Windows (#254)

## v0.24.0

### 🛑 Breaking changes 🛑

- Remove opencensus receiver (#230)
- Don't override system resource attrs in default config (#239)
  - Detectors run as part of the `resourcedetection` processor no longer overwrite resource attributes already present.

### 💡 Enhancements 💡

- Support gateway mode for Linux installer (#187)
- Support gateway mode for windows installer (#231)
- Add SignalFx forwarder to default configs (#218)
- Include Smart Agent bundle in msi (#222)
- Add Linux support bundle script (#208)
- Add Kafka receiver/exporter (#201)

### 🧰 Bug fixes 🧰

## v0.23.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.23.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.23.0) and the [opentelemetry-collector-contrib v0.23.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.23.0) releases.

### 🛑 Breaking changes 🛑

- Renamed default config from `splunk_config_linux.yaml` to `gateway_config.yaml` (#170)

### 💡 Enhancements 💡

- Include smart agent bundle in amd64 deb/rpm packages (#177)
- `smartagent` receiver: Add support for logs (#161) and traces (#192)

### 🧰 Bug fixes 🧰

- `smartagent` extension: Ensure propagation of collectd bundle dir (#180)
- `smartagent` receiver: Fix logrus logger hook data race condition (#181)
