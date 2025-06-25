# Splunk OpenTelemetry Collector Salt Formula

This formula installs and configures Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Linux
Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2, 2023 (**Note:** Log collection with Fluentd not currently supported for Amazon Linux 2023.)
- CentOS / Red Hat: 8, 9
- Oracle: 8, 9
- Debian: 9, 10, 11
- SUSE: 15 (**Note:** Only for collector versions v0.34.0 or higher. Log collection with fluentd not currently supported.)
- Ubuntu: 16.04, 18.04, 20.04, 22.04

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://docs.splunk.com/observability/en/gdi/opentelemetry/exposed-endpoints.html) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Usage

All the attributes can be configured in pillar `splunk-otel-collector`.

```yaml
splunk-otel-collector:
  splunk_access_token: "MY_ACCESS_TOKEN"
  splunk_realm: "SPLUNK_REALM"
  splunk_repo_base_url: https://splunk.jfrog.io/splunk
  splunk_otel_collector_config: '/etc/otel/collector/agent_config.yaml'
  splunk_service_user: splunk-otel-collector
  splunk_service_group: splunk-otel-collector
```

## This Salt Formula accepts the following attributes:

### Collector

- `splunk_access_token` (**Required**): The Splunk access token to
  authenticate requests.

- `splunk_realm`: Which realm to send the data to. The `SPLUNK_REALM`
  environment variable will be set with this value for the Splunk OTel 
  Collector service. (**default:** `us0`)

- `splunk_hec_token`: Set the Splunk HEC authentication token if different than
  `splunk_access_token`. The `SPLUNK_HEC_TOKEN` environment 
  variable will be set with this value for the collector service. (**default:**
  `splunk_access_token`)

- `splunk_ingest_url`: The Splunk ingest URL, e.g.
  `https://ingest.us0.signalfx.com`. The `SPLUNK_INGEST_URL` environment 
  variable will be set with this value for the collector service. (**default:**
  `https://ingest.{{ splunk_realm }}.signalfx.com`)

- `splunk_api_url`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`.
  The `SPLUNK_API_URL` environment variable will be set with this value for the
  collector service. (**default:** `https://api.{{ splunk_realm }}.signalfx.com`)

- `collector_version`: Version of the collector package to install, e.g.
  `0.25.0`. (**default:** `latest`)

- `splunk_otel_collector_config`: Splunk OTel Collector config YAML file. Can be set to 
  `/etc/otel/collector/gateway_config.yaml` to install the collector in gateway
  mode. (**default:** `/etc/otel/collector/agent_config.yaml`)

- `splunk_otel_collector_config_source`: Source path to a Splunk OTel Collector config YAML 
  file on your control host that will be uploaded and set in place of
  `splunk_otel_collector_config` in remote hosts. To use custom collector config add the config file into salt dir, 
  e.g. `salt://templates/agent_config.yaml`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_otel_collector_config` will be used)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/pkg/extension/smartagentextension/README.md).
  The default path is provided by the collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service. (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle`)

- `splunk_collectd_dir`: The path to the collectd config directory for the
  Smart Agent bundle. The default path is provided by the collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node. The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the collector service.
  (**default:** `/usr/lib/splunk-otel-collector/agent-bundle`)

- `splunk_service_user` and `splunk_service_group` (Linux only): Set the user/group
  ownership for the collector service. The user/group will be created if they
  do not exist. (**default:** `splunk-otel-collector`)

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the Splunk OTel 
  Collector. (**default:** `512`)

- `gomemlimit`: `splunk_ballast_size_mib` is deprecated and removed. For Splunk Otel Collector version `0.97.0` or greater, `GOMEMLIMIT` env var is introduced. The default is set to 90% of the `SPLUNK_TOTAL_MEM_MIB`. For more information regarding the usage, please follow the instructions ([here](https://github.com/signalfx/splunk-otel-collector?tab=readme-ov-file#from-0961-to-0970)).  (**default:** 90% of `splunk_memory_total_mib`, otherwise)

- `splunk_listen_interface`: The network interface the collector receivers will listen
  on. (**default:** `127.0.0.1` for agent config, `0.0.0.0` otherwise)

- `collector_additional_env_vars`: Dictionary of additional environment
  variables from the collector configuration file for the collector service
  (**default:** `{}`). For example, if the collector configuration file
  includes references to `${MY_CUSTOM_VAR1}` and `${MY_CUSTOM_VAR2}`, specify
  the following to allow the collector service to expand these variables:
  ```yaml
  collector_additional_env_vars:
    MY_CUSTOM_VAR1: value1
    MY_CUSTOM_VAR2: value2
  ```
  The variables/values will be added to the
  `/etc/otel/collector/splunk-otel-collector.conf` systemd environment file.

### Fluentd

> **_NOTE:_**  Fluentd support has been deprecated and will be removed in a future release.
> Please refer to [deprecation documentation](../../docs/deprecations/fluentd-support.md) for more information.

- `install_fluentd`: Whether to install/manage fluentd and dependencies for log
  collection. The dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](
  https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `False`)

- `td_agent_version`: Version of [td-agent](
  https://td-agent-package-browser.herokuapp.com/) (fluentd package) that will
  be installed (**default:** `4.3.0`)

- `splunk_fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf`)

- `splunk_fluentd_config_source`: Source path to a fluentd config file on your 
  control host that will be uploaded and set in place of `splunk_fluentd_config` on
  remote hosts. To use custom fluentd config add the config file into salt dir, 
  e.g. `salt://templates/td_agent.conf` (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_fluentd_config` will be used)

### Auto Instrumentation (Linux Only)

**Note:** The application(s) on the node need to be started/restarted separately
after installation/configuration in order for any changes to take effect.

- `install_auto_instrumentation`: Whether to install/manage [Splunk
  OpenTelemetry Auto Instrumentation](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation).
  When set to `True`, the `splunk-otel-auto-instrumentation` deb/rpm package
  will be downloaded and installed from the Collector repository. (**default:**
  `False`)

- `auto_instrumentation_version`: Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`. The
  minimum supported version is `0.48.0`. The minimum supported version for
  Node.js auto instrumentation is `0.87.0`. The minimum supported version for
  .NET auto instrumentation is `0.99.0`. (**default:** `latest`)

- `auto_instrumentation_sdks`: List of Splunk OpenTelemetry Auto
  Instrumentation SDKs to install, configure, and activate. (**default:**
  `['java', 'nodejs', 'dotnet']`)

  Currently, the following values are supported:
  - `java`: [Splunk OpenTelemetry for Java](https://github.com/signalfx/splunk-otel-java)
  - `nodejs`: [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js)
  - `dotnet`: [Splunk OpenTelemetry for .NET](https://github.com/signalfx/splunk-otel-dotnet) (x86_64/amd64 only)

  **Note:** This formula does not manage the installation/configuration of
  Node.js, `npm`, or Node.js applications. If `nodejs` is included in this
  option, Node.js and `npm` are required to be pre-installed on the node in
  order to install and activate the Node.js SDK.

- `auto_instrumentation_systemd`: By default, the `/etc/ld.so.preload` file on
  the node will be configured for the
  `/usr/lib/splunk-instrumentation/libsplunk.so` [shared object library](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation)
  provided by the `splunk-otel-auto-instrumentation` package to activate and
  configure auto instrumentation system-wide for all supported applications.
  Alternatively, set this option to `True` to activate and configure auto
  instrumentation ***only*** for supported applications running as `systemd`
  services. If this option is set to `True`,
  `/usr/lib/splunk-instrumentation/libsplunk.so` will not be added to
  `/etc/ld.so.preload`. Instead, the
  `/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf`
  `systemd` drop-in file will be created and configured for environment
  variables based on the default and specified options. (**default:** `False`)

- `auto_instrumentation_ld_so_preload`: Configure this variable to include
  additional library paths, e.g. `/path/to/my.library.so`, to
  `/etc/ld.so.preload`. (**default:** `None`)

- `auto_instrumentation_java_agent_path`: Path to the [Splunk OpenTelemetry
  Java agent](https://github.com/signalfx/splunk-otel-java). The default path
  is provided by the `splunk-otel-auto-instrumentation` package. If the path is
  changed from the default value, the path should be an existing file on the
  node. (**default:** `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`)

- `auto_instrumentation_npm_path`: If the `auto_instrumentation_sdks` option
  includes `nodejs`, the Splunk OpenTelemetry for Node.js SDK will be installed
  with the `npm` command. Use this option to specify a custom path on the
  node for `npm`, for example `/my/custom/path/to/npm`. (**default:** `npm`)

  **Note:** This role does not manage the installation/configuration of
  Node.js or `npm`.

- `auto_instrumentation_resource_attributes`: Configure the OpenTelemetry auto
  instrumentation resource attributes, e.g.
  `deployment.environment=prod,my.key=value` (comma-separated `key=value` pairs.).
  (**default:** `None`)

- `auto_instrumentation_service_name`: Explicitly set the service name for
  ***all*** instrumented applications on the node, e.g. `my.service`. By
  default, the service name is automatically generated for each instrumented
  application. (**default:** `None`)

- **DEPRECATED** `auto_instrumentation_generate_service_name`: Set this option
  to `False` to prevent the preloader from setting the `OTEL_SERVICE_NAME`
  environment variable. Only applicable if `auto_instrumentation_version` is <
  `0.87.0`. (**default:** `True`)

- **DEPRECATED**` auto_instrumentation_disable_telemetry`: Enable or disable
  the preloader from sending the `splunk.linux-autoinstr.executions` metric to
  the local collector. Only applicable if `auto_instrumentation_version` is <
  `0.87.0`. (**default:** `False`)

- `auto_instrumentation_enable_profiler`: Enable or disable AlwaysOn CPU
  Profiling. (**default**: `False`)

- `auto_instrumentation_enable_profiler_memory`: Enable or disable AlwaysOn
  Memory Profiling. (**default:** `False`)

- `auto_instrumentation_enable_metrics`: Enable or disable exporting
  instrumentation metrics. (**default**: `False`)

- `auto_instrumentation_otlp_endpoint`: Set the OTLP endpoint for captured
  traces, metrics, and logs. The value will be set to the
  `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. Only applicable if
  `auto_instrumentation_version` is `latest` or >= `0.87.0`. (**default:**
  `""`, i.e. defer to the default `OTEL_EXPORTER_OTLP_ENDPOINT` value for
  each activated SDK)

- `auto_instrumentation_otlp_endpoint_protocol`: Set the protocol for the OTLP
  endpoint, for example `grpc` or `http/protobuf`. The value will be set to the
  `OTEL_EXPORTER_OTLP_PROTOCOL` environment variable. Only applicable if
  `auto_instrumentation_version` is `latest` or >= `0.104.0`. (**default:**
  `""`, i.e. defer to the default `OTEL_EXPORTER_OTLP_PROTOCOL` value for
  each activated SDK)

- `auto_instrumentation_metrics_exporter`: Comma-separated list of exporters
  for collected metrics by all activated SDKs, for example `otlp,prometheus`.
  Set the value to `none` to disable collection and export of metrics. The
  value will be set to the `OTEL_METRICS_EXPORTER` environment variable. Only
  applicable if `auto_instrumentation_version` is `latest` or >= `0.104.0`.
  (**default:** `""`, i.e. defer to the default `OTEL_METRICS_EXPORTER`
  value for each activated SDK)

- `auto_instrumentation_logs_exporter`: Set the exporter for collected logs by
  all activated SDKs, for example `otlp`. Set the value to `none` to disable
  collection and export of logs. The value will be set to the
  `OTEL_LOGS_EXPORTER` environment variable. Only applicable if
  `auto_instrumentation_version` is `latest` or >= `0.108.0`. (**default:**
  `""`, i.e. defer to the default `OTEL_LOGS_EXPORTER` value for each
  activated SDK)
