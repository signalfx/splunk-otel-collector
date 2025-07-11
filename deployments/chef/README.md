# Splunk OpenTelemetry Collector Cookbook

This cookbook installs and configures the Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux and Windows machines and sends
data to [Splunk Observability Cloud](
https://www.splunk.com/en_us/products/observability.html).

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://docs.splunk.com/observability/en/gdi/opentelemetry/exposed-endpoints.html) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat: 8, 9
- Oracle: 8, 9
- Debian: 9, 10, 11
- SUSE: 15 (**Note:** Only for Collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 18.04, 20.04, 22.04

## Windows

Currently, the following Windows versions are supported:

- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

On Windows, the collector is installed as a Windows service and its environment
variables are set at the service scope, i.e.: they are only available to the
collector service and not to the entire machine.

## Usage

This cookbook can be downloaded and installed from [Chef Supermarket](https://supermarket.chef.io/cookbooks/splunk_otel_collector).

To install the Collector, include the
`splunk_otel_collector::default` recipe in the `run_list`, and set the
attributes on the node's `run_state`. Below is an example to configure the
required `splunk_access_token` attribute and some optional attributes:
```yaml 
{
    "splunk-otel-collector": {
        "splunk_access_token": "<SPLUNK_ACCESS_TOKEN>",
        "splunk_realm": "<SPLUNK_REALM>",
    }
}
```

# This cookbook accepts the following attributes

### Collector

- `splunk_access_token` (**Required**): The [Splunk access token](
  https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
  to authenticate requests.

- `splunk_realm`: Which Splunk realm to send the data to, e.g. `us1`. The
  `SPLUNK_REALM` environment variable will be set with this value for the
  Collector service. This value will derive the `splunk_ingest_url` and
  `splunk_api_url` attribute values unless they are explicitly set.
  (**default:** `us0`)

- `splunk_ingest_url`: Explicitly set the Splunk ingest URL, e.g.
  `https://ingest.us0.signalfx.com`, instead of the URL derived from the
  `splunk_realm` attribute. The `SPLUNK_INGEST_URL` environment variable will
  be set with this value for the Collector service. (**default:**
  `https://ingest.{{ splunk_realm }}.signalfx.com`)

- `splunk_api_url`: Explicitly set the Splunk API URL, e.g.
  `https://api.us0.signalfx.com`, instead of the URL derived from `splunk_realm`
  attribute. The `SPLUNK_API_URL` environment variable will be set with this
  value for the Collector service. (**default:**
  `https://api.{{ splunk_realm }}.signalfx.com`)

- `collector_version`: Version of the Collector package to install, e.g.
  `0.34.0`. (**default:** `latest`)

- `collector_config_source`: Source path to the Collector config YAML file.
  This file will be copied to the `collector_config_dest` path on the node. See
  the [source attribute](https://docs.chef.io/resources/remote_file/) of the
  file resource for supported value types. The default source file is provided
  by the Collector package. (**default:**
  `/etc/otel/collector/agent_config.yaml` on Linux,
  `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent_config.yaml` on Windows)

- `collector_config_dest`: Destination path of the Collector config file on the
  node. The `SPLUNK_CONFIG` environment variable will be set with this value
  for the Collector service. (**default:**
  `/etc/otel/collector/agent_config.yaml` on Linux,
  `%PROGRAMDATA%\Splunk\OpenTelemetry Collector\agent_config.yaml` on Windows)

- `node['splunk_otel_collector']['collector_config']`: The Collector
  configuration object. Everything underneath this object gets directly
  converted to YAML and becomes the Collector config file. Using this option
  preempts `collector_config_source` functionality. (**default:** `{}`)

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the
  Collector. The `SPLUNK_MEMORY_TOTAL_MIB` environment variable will be set
  with this value for the Collector service. (**default:** `512`)

- `gomemlimit`: The `GOMEMLIMIT` environment variable is introduced for the Splunk Otel Collector version >=0.97.0, allowing the limitation of memory usage in the GO runtime. This feature can help enhance GC (Garbage Collection) related performance and prevent GC related Out of Memory (OOM) situations.

- `splunk_listen_interface`: The network interface the collector receivers
  will listen on (**default** `0.0.0.0`).

- `splunk_service_user` and `splunk_service_group` (Linux only): Set the
  user/group ownership for the Collector service. The user/group will be
  created if they do not exist. (**default:** `splunk-otel-collector`)

- `package_stage`: The Collector package repository stage to use.  Can be
  `release`, `beta`, or `test`. (**default:** `release`)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/pkg/extension/smartagentextension/README.md).
  The default path is provided by the Collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the Collector service. (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle` on Linux,
  `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent-bundle` on Windows)

- `splunk_collectd_dir`: The path to the collectd config directory for the
  Smart Agent bundle. The default path is provided by the Collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node. The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the Collector service.
  (**default:** `/usr/lib/splunk-otel-collector/agent-bundle` on Linux,
  `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent-bundle\run\collectd`
  on Windows)

- `collector_additional_env_vars`: Hash of additional environment variables
  from the collector configuration file for the collector service
  (**default:** `{}`).
  For example, if the collector configuration file includes references to
  `${MY_CUSTOM_VAR1}` and `${MY_CUSTOM_VAR2}`, specify the following to allow
  the collector service to expand these variables:
  ```ruby
  collector_additional_env_vars: {'MY_CUSTOM_VAR1' => 'value1', 'MY_CUSTOM_VAR2' => 'value2'}
  ```
  On Linux, the variables/values will be added to the
  `/etc/otel/collector/splunk-otel-collector.conf` systemd environment file.
  On Windows, the variables/values will be added to the `Environment` value under the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\splunk-otel-collector`
  registry key.

- `collector_command_line_args`: Additional command line arguments to pass to the
  collector service. On Linux the value will be set to the `OTELCOL_OPTIONS` environment
  variable for the collector service. On Windows, this option is only supported
  by versions `>= 0.127.0` (**default:** `''`).

### Fluentd

> **_NOTE:_**  Fluentd support has been deprecated and will be removed in a future release.
> Please refer to [deprecation documentation](../../docs/deprecations/fluentd-support.md) for more information.

- `with_fluentd`: Whether to install/manage Fluentd and dependencies for log
  collection. On Linux, the dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `false`)

- `fluentd_version`: Version of the [td-agent](
  https://www.fluentd.org/download) (Fluentd) package to install (**default:**
  `4.3.1` for all Linux distros and Windows)

- `fluentd_config_source`: Source path to the Fluentd config file. This file
  will be copied to the `fluentd_config_dest` path on the node. See the
  [source attribute](https://docs.chef.io/resources/remote_file/) of the file
  resource for supported value types. The default source file is provided by
  the Collector package. Only applicable if `with_fluentd` is set to true.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf` on Linux, 
  `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf` on Windows)

- `fluentd_config_dest` (Linux only): Destination path to the Fluentd config
  file on the node. Only applicable if `with_fluentd` is set to `true`.
  **Note**: On Windows, the path will always be set to
  `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf`. (**default:**
  `/etc/otel/collector/fluentd/fluent.conf`)

### Auto Instrumentation on Linux

**Note:** The application(s) on the node need to be restarted separately
after installation/configuration in order for any change to take effect.

- `with_auto_instrumentation`: Whether to install/manage Splunk OpenTelemetry
  Auto Instrumentation. When set to `true`, the
  `splunk-otel-auto-instrumentation` deb/rpm package will be downloaded and
  installed from the Collector repository. (**default:** `false`)

- `with_auto_instrumentation_sdks`: List of Splunk OpenTelemetry Auto
  Instrumentation SDKs to install, configure, and activate. (**default:**
  `%w(java nodejs dotnet)`)

  Currently, the following values are supported:
  - `java`: [Splunk OpenTelemetry for Java](https://github.com/signalfx/splunk-otel-java)
  - `nodejs`: [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js)
  - `dotnet`: [Splunk OpenTelemetry for .NET](https://github.com/signalfx/splunk-otel-dotnet) (x86_64/amd64 only)

  **Note:** This recipe does not manage the installation/configuration of
  Node.js, `npm`, or Node.js applications. If `nodejs` is included in this
  option, Node.js and `npm` are required to be pre-installed on the node in
  order to install and activate the Node.js SDK.

- `auto_instrumentation_version`: Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`. The
  minimum supported version is `0.48.0`. The minimum supported version for
  Node.js auto instrumentation is `0.87.0`. The minimum supported version for
  .NET auto instrumentation is `0.99.0`. (**default:** `latest`)

- `auto_instrumentation_systemd` (Linux only): By default, the
  `/etc/ld.so.preload` file on the node will be configured for the
  `/usr/lib/splunk-instrumentation/libsplunk.so` [shared object library](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation)
  provided by the `splunk-otel-auto-instrumentation` package to activate and
  configure auto instrumentation system-wide for all supported applications.
  Alternatively, set this option to `true` to activate and configure auto
  instrumentation ***only*** for supported applications running as `systemd`
  services. If this option is set to `true`,
  `/usr/lib/splunk-instrumentation/libsplunk.so` will not be added to
  `/etc/ld.so.preload`. Instead, the
  `/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf`
  `systemd` drop-in file will be created and configured for environment
  variables based on the default and specified options. (**default:** `false`)

- `auto_instrumentation_ld_so_preload` (Linux only): Configure this variable to
  include additional library paths, e.g. `/path/to/my.library.so`, to
  `/etc/ld.so.preload`. (**default:** ``)

- `auto_instrumentation_java_agent_path`: Path to the [Splunk OpenTelemetry
  Java agent](https://github.com/signalfx/splunk-otel-java). The default path
  is provided by the `splunk-otel-auto-instrumentation` package. If the path is
  changed from the default value, the path should be an existing file on the
  node. (**default:** `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`)

- `auto_instrumentation_npm_path`: If the `with_auto_instrumentation_sdks`
  option includes `nodejs`, the Splunk OpenTelemetry for Node.js SDK will be
  installed only if `npm` is found on the node with the
  `bash -c 'command -v npm'` shell command. Use this option to specify a
  custom path on the node for `npm`, for example `/my/custom/path/to/npm`.
  (**default:** `npm`)

  **Note:** This recipe does not manage the installation/configuration of
  Node.js or `npm`.

- `auto_instrumentation_resource_attributes`: Configure the OpenTelemetry auto
  instrumentation resource attributes, e.g.
  `deployment.environment=prod,my.key=value` (comma-separated `key=value` pairs.).
  (**default:** `''`)

- `auto_instrumentation_service_name`: Explicitly set the service name for
  ***all*** instrumented applications on the node, e.g. `my.service`. By
  default, the service name is automatically generated for each instrumented
  application. (**default:** `''`)

- **DEPRECATED** `auto_instrumentation_generate_service_name`: Set this option
  to `false` to prevent the preloader from setting the `OTEL_SERVICE_NAME`
  environment variable. Only applicable if `auto_instrumentation_version` is <
  `0.87.0`. (**default:** `true`)

- **DEPRECATED**` auto_instrumentation_disable_telemetry` (Linux only): Enable
  or disable the preloader from sending the `splunk.linux-autoinstr.executions`
  metric to the local collector. Only applicable if
  `auto_instrumentation_version` is < `0.87.0`. (**default:** `false`)

- `auto_instrumentation_enable_profiler` (Linux only): Enable or disable
  AlwaysOn CPU Profiling. (**default**: `false`)

- `auto_instrumentation_enable_profiler_memory` (Linux only): Enable or disable
  AlwaysOn Memory Profiling. (**default:** `false`)

- `auto_instrumentation_enable_metrics` (Linux only): Enable or disable
  exporting instrumentation metrics. (**default**: `false`)

- `auto_instrumentation_otlp_endpoint` (Linux only): Set the OTLP endpoint for
  captured traces. The value will be set to the `OTEL_EXPORTER_OTLP_ENDPOINT`
  environment variable. Only applicable if `auto_instrumentation_version` is
  `latest` or >= `0.87.0`. (**default:** `''`, i.e. defer to the default
  `OTEL_EXPORTER_OTLP_ENDPOINT` value for each activated SDK)

- `auto_instrumentation_otlp_endpoint_protocol` (Linux only): Set the protocol
  for the OTLP endpoint, for example `grpc` or `http/protobuf`. The value will
  be set to the `OTEL_EXPORTER_OTLP_PROTOCOL` environment variable. Only
  applicable if `auto_instrumentation_version` is `latest` or >= `0.104.0`.
  (**default:** `''`, i.e. defer to the default `OTEL_EXPORTER_OTLP_PROTOCOL`
  value for each activated SDK)

- `auto_instrumentation_metrics_exporter` (Linux only): Comma-separated list of
  exporters for collected metrics by all activated SDKs, for example
  `otlp,prometheus`. Set the value to `none` to disable collection and export
  of metrics. The value will be set to the `OTEL_METRICS_EXPORTER` environment
  variable. Only applicable if `auto_instrumentation_version` is `latest` or >=
  `0.104.0`. (**default:** `''`, i.e. defer to the default
  `OTEL_METRICS_EXPORTER` value for each activated SDK)

- `auto_instrumentation_logs_exporter` (Linux only): Set the exporter for
  collected logs by all activated SDKs, for example `otlp`. Set the value to
  `none` to disable collection and export of logs. The value will be set to the
  `OTEL_LOGS_EXPORTER` environment variable. Only applicable if
  `auto_instrumentation_version` is `latest` or >= `0.108.0`.
  (**default:** `''`, i.e. defer to the default `OTEL_LOGS_EXPORTER` value for
  each activated SDK)

### SignalFx Auto Instrumentation for .NET on Windows

The option to install the [SignalFx Auto Instrumentation for .NET](
https://docs.splunk.com/Observability/gdi/get-data-in/application/otel-dotnet/get-started.html)
`with_signalfx_dotnet_auto_instrumentation` is deprecated and
will have no effect after release `0.16.0`.
Install the `Splunk Distribution of OpenTelemetry .NET`.
