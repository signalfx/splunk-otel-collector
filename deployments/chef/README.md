# Splunk OpenTelemetry Collector Cookbook

This cookbook installs and configures the Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux and Windows machines and sends
data to [Splunk Observability Cloud](
https://www.splunk.com/en_us/observability.html).

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 9, 10, 11
- SUSE: 12, 15 (**Note:** Only for Collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 18.04, 20.04, 22.04

## Windows

Currently, the following Windows versions are supported:

- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

## Usage

This cookbook can be downloaded and installed from [Chef Supermarket](https://supermarket.chef.io/cookbooks/splunk_otel_collector).

To install the Collector and Fluentd, include the
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

- `splunk_ballast_size_mib`: Explicitly set the ballast size for the Collector
  instead of the value calculated from the `splunk_memory_total_mib` attribute.
  This should be set to 1/3 to 1/2 of configured memory. The
  `SPLUNK_BALLAST_SIZE_MIB` environment variable will be set with this value
  for the Collector service. (**default:** `''`)

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
  On Windows, the variables/values will be added to the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
  registry key.

### Fluentd

- `with_fluentd`: Whether to install/manage Fluentd and dependencies for log
  collection. On Linux, the dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `true`)

- `fluentd_version`: Version of the [td-agent](
  https://www.fluentd.org/download) (Fluentd) package to install (**default:**
  `3.7.1` for Debian stretch, and `4.3.1` for all other Linux distros and
  Windows)

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

### Auto Instrumentation

**Note:** The Java application(s) on the node need to be restarted separately
after installation/configuration in order for any change to take effect.

- `with_auto_instrumentation`: Whether to install/manage [Splunk OpenTelemetry
  Auto Instrumentation for Java](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation).
  When set to `true`, the `splunk-otel-auto-instrumentation` deb/rpm package
  will be downloaded and installed from the Collector repository.
  (**default:** `false`)

- `auto_instrumentation_version`: Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`. The
  minimum supported version is `0.48.0`. (**default:** `latest`)

- `auto_instrumentation_ld_so_preload`: By default, the `/etc/ld.so.preload`
  file on the node will be configured for the
  `/usr/lib/splunk-instrumentation/libsplunk.so` [shared object library](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#operation)
  provided by the `splunk-otel-auto-instrumentation` package and is required
  for auto instrumentation. Configure this variable to include additional
  library paths, e.g. `/path/to/my.library.so`. (**default:** `''`)

- `auto_instrumentation_java_agent_path`: Path to the [Splunk OpenTelemetry
  Java agent](https://github.com/signalfx/splunk-otel-java). The default path
  is provided by the `splunk-otel-auto-instrumentation` package. If the path is
  changed from the default value, the path should be an existing file on the
  node. The specified path will be added to the
  `/usr/lib/splunk-instrumentation/instrumentation.conf` config file on the
  node. (**default:**
  `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`)

- `auto_instrumentation_resource_attributes`: Configure the OpenTelemetry
  instrumentation [resource attributes](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#configuration-file),
  e.g. `deployment.environment=prod`. The specified resource attribute(s) will
  be added to the `/usr/lib/splunk-instrumentation/instrumentation.conf` config
  file on the node. (**default:** `''`)

- `auto_instrumentation_service_name`: Explicitly set the [service name](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#configuration-file)
  for the instrumented Java application, e.g. `my.service`. By default, the
  service name is automatically derived from the arguments of the Java
  executable on the node. However, if this variable is set to a non-empty
  value, the value will override the derived service name and be added to the
  `/usr/lib/splunk-instrumentation/instrumentation.conf` config file on the
  node. (**default:** `''`)

- `auto_instrumentation_generate_service_name`: Set this option to `false` to
  prevent the preloader from setting the `OTEL_SERVICE_NAME` environment
  variable. (**default:** `true`)

- `auto_instrumentation_disable_telemetry` (Linux only): Enable or disable the
  preloader from sending the `splunk.linux-autoinstr.executions` metric to the
  local collector. (**default:** `false`)

- `auto_instrumentation_enable_profiler` (Linux only): Enable or disable
  AlwaysOn CPU Profiling. (**default**: `false`)

- `auto_instrumentation_enable_profiler_memory` (Linux only): Enable or disable
  AlwaysOn Memory Profiling. (**default:** `false`)

- `auto_instrumentation_enable_metrics` (Linux only): Enable or disable
  exporting Micrometer metrics. (**default**: `false`)
