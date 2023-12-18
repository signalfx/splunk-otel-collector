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
- CentOS / Red Hat / Oracle: 7, 8, 9
- Debian: 9, 10, 11
- SUSE: 12, 15 (**Note:** Only for Collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 18.04, 20.04, 22.04

## Windows

Currently, the following Windows versions are supported:

- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

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

- `splunk_ballast_size_mib`: Explicitly set the ballast size for the Collector
  instead of the value calculated from the `splunk_memory_total_mib` attribute.
  This should be set to 1/3 to 1/2 of configured memory. The
  `SPLUNK_BALLAST_SIZE_MIB` environment variable will be set with this value
  for the Collector service. (**default:** `''`)

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
  (**default:** `false`)

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

### Auto Instrumentation on Linux

**Note:** The application(s) on the node need to be restarted separately
after installation/configuration in order for any change to take effect.

- `with_auto_instrumentation`: Whether to install/manage Splunk OpenTelemetry
  Auto Instrumentation. When set to `true`, the
  `splunk-otel-auto-instrumentation` deb/rpm package will be downloaded and
  installed from the Collector repository. (**default:** `false`)

- `with_auto_instrumentation_sdks`: List of Splunk OpenTelemetry Auto
  Instrumentation SDKs to install, configure, and activate. (**default:**
  `%w(java nodejs)`)

  Currently, the following values are supported:
  - `java`: [Splunk OpenTelemetry for Java](https://github.com/signalfx/splunk-otel-java)
  - `nodejs`: [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js)

  **Note:** This recipe does not manage the installation/configuration of
  Node.js, `npm`, or Node.js applications. If `nodejs` is included in this
  option, Node.js and `npm` are required to be pre-installed on the node in
  order to install and activate the Node.js SDK.

- `auto_instrumentation_version`: Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`. The
  minimum supported version is `0.48.0`. The minimum supported version for
  Node.js auto instrumentation is `0.87.0`. (**default:** `latest`)

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

- `auto_instrumentation_otlp_endpoint` (Linux only): Set the OTLP gRPC endpoint
  for captured traces. Only applicable if `auto_instrumentation_version` is
  `latest` or >= `0.87.0`. (**default:** `http://127.0.0.1:4317`)

### Auto Instrumentation for .NET on Windows

***Warning:*** The `Environment` property in the
`HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC` registry key will
be overwritten by the options specified below to enable/configure auto
instrumentation for IIS. Use the
`signalfx_dotnet_auto_instrumentation_additional_options` option (see below for
details) to include any other environment variables required for IIS.

**Note:** By default, IIS will be restarted with the `iisreset` command (if it
exists) after installation/configuration. Applications ***not*** running within
IIS need to be restarted/managed separately in order for any changes to take
effect.

- `with_signalfx_dotnet_auto_instrumentation` (Windows only): Whether to
  install/manage [SignalFx Auto Instrumentation for .NET](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/dotnet/get-started.html).
  When set to `true`, the `signalfx-dotnet-tracing` MSI package will be
  downloaded and installed, and the Windows registry will be updated based on
  the options below. (**default:** `false`)

- `signalfx_dotnet_auto_instrumentation_version` (Windows only): Version of the
  `signalfx-dotnet-tracing` MSI package to download and install from
  [GitHub Releases](https://github.com/signalfx/signalfx-dotnet-tracing/releases).
  (**default:** `1.1.0`)

- `signalfx_dotnet_auto_instrumentation_msi_url` (Windows only): Specify the
  URL to download the MSI from a custom host, for example
  `https://my.host/signalfx-dotnet-tracing-1.0.0-x64.msi`. If specified, the
  `signalfx_dotnet_auto_instrumentation_version` option is ignored.
  (**default:** `https://github.com/signalfx/signalfx-dotnet-tracing/releases/download/v{{ signalfx_dotnet_auto_instrumentation_version }}/signalfx-dotnet-tracing-{{ signalfx_dotnet_auto_instrumentation_version }}-x64.msi`)

- `signalfx_dotnet_auto_instrumentation_iisreset` (Windows only): By default,
  the `iisreset.exe` command (if it exists) will be executed after
  installation/configuration in order for any changes to take effect for IIS
  applications. Set this option to `false` to skip this step if IIS is managed
  separately or is not applicable. (**default:** `true`)

- `signalfx_dotnet_auto_instrumentation_system_wide` (Windows only): By
  default, the `Environment` property in the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC` registry key
  will be configured for the following environment variables and any from the
  `signalfx_dotnet_auto_instrumentation_additional_options` option to
  enable/configure auto instrumentation for ***only*** IIS applications:
  ```yaml
  COR_ENABLE_PROFILING: true  # Required
  COR_PROFILER: "{B4C89B0F-9908-4F73-9F59-0D77C5A06874}"  # Required
  CORECLR_ENABLE_PROFILING: true  # Required
  CORECLR_PROFILER: "{B4C89B0F-9908-4F73-9F59-0D77C5A06874}"  # Required
  SIGNALFX_ENV: "{{ signalfx_dotnet_auto_instrumentation_environment }}"
  SIGNALFX_PROFILER_ENABLED: "{{ signalfx_dotnet_auto_instrumentation_enable_profiler }}"
  SIGNALFX_PROFILER_MEMORY_ENABLED: "{{ signalfx_dotnet_auto_instrumentation_enable_profiler_memory }}"
  SIGNALFX_SERVICE_NAME: "{{ signalfx_dotnet_auto_instrumentation_service_name }}"
  ```
  Set this option to `true` to also add these environment variables and any
  from the `signalfx_dotnet_auto_instrumentation_additional_options` option to
  the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
  registry key to enable/configure auto instrumentation for ***all*** .NET
  applications on the node. (**default:** `false`)

- `signalfx_dotnet_auto_instrumentation_environment` (Windows only): Configure
  this option to set the "Environment" value to be reported to Splunk APM, for
  example `production`. The value is assigned to the `SIGNALFX_ENV` environment
  variable in the Windows registry (**default:** `''`, i.e. the "Environment"
  will appear as `unknown` in Splunk APM for the instrumented
  service/application)

- `signalfx_dotnet_auto_instrumentation_service_name` (Windows only): Configure
  this variable to override the [auto-generated service name](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/dotnet/configuration/advanced-dotnet-configuration.html#changing-the-default-service-name)
  for the instrumented service/application, for example `my-service-name`. The
  value is assigned to the `SIGNALFX_SERVICE_NAME` environment variable in the
  Windows registry. (**default:** `''`)

- `signalfx_dotnet_auto_instrumentation_enable_profiler` (Windows only): Set
  this option to `true` to enable AlwaysOn Profiling. The value will be
  assigned to the `SIGNALFX_PROFILER_ENABLED` environment variable in the
  Windows registry. (**default:** `false`)

- `signalfx_dotnet_auto_instrumentation_enable_profiler_memory` (Windows only):
  Set this option to `true` to enable AlwaysOn Memory Profiling. The value will
  be assigned to the `SIGNALFX_PROFILER_MEMORY_ENABLED` environment variable in
  the Windows registry. (**default:** `false`)

- `signalfx_dotnet_auto_instrumentation_additional_options` (Windows only):
  Hash of additional options to be added to the Windows registry
  ***in addition*** to the options above. (**default:** `{}`)

  For example:
  ```yaml
  signalfx_dotnet_auto_instrumentation_additional_options: {
    'SIGNALFX_VERSION': '1.2.3',
    'SIGNALFX_FILE_LOG_ENABLED': false,
    # Hint: If the signalfx_dotnet_auto_instrumentation_system_wide option is
    # set to true, all .NET applications on the node will be instrumented. Use
    # the following options to include/exclude processes from auto
    # instrumentation.
    'SIGNALFX_PROFILER_PROCESSES': 'MyApp.exe;dotnet.exe',
    'SIGNALFX_PROFILER_EXCLUDE_PROCESSES': 'ReservedProcess.exe;powershell.exe',
  }
  ```
  Check the [Advanced Configuration Guide](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/dotnet/configuration/advanced-dotnet-configuration.html)
  for more details about the options above and other supported options.

To uninstall the `signalfx-dotnet-tracing` MSI and disable auto
instrumentation, include the following in your recipe and restart all
applicable services:
```
windows_package 'SignalFx .NET Tracing 64-bit' do
  action :remove
end

# If the "signalfx_dotnet_auto_instrumentation_system_wide" option was set to
# "true", include the following to remove the values from the
# "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment"
# registry key:

registry_key 'HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment' do
  values [
    { name: 'COR_ENABLE_PROFILING', type: :string, data: '' },
    { name: 'COR_PROFILER', type: :string, data: '' },
    { name: 'CORECLR_ENABLE_PROFILING', type: :string, data: '' },
    { name: 'CORECLR_PROFILER', type: :string, data: '' },
    { name: 'SIGNALFX_ENV', type: :string, data: '' },
    { name: 'SIGNALFX_PROFILER_ENABLED', type: :string, data: '' },
    { name: 'SIGNALFX_PROFILER_MEMORY_ENABLED', type: :string, data: '' },
    { name: 'SIGNALFX_SERVICE_NAME', type: :string, data: '' },
  ]
  action :delete
end
```
