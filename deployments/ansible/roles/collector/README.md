# Splunk OpenTelemetry Collector Ansible Role

Ansible role that installs Splunk OpenTelemetry Collector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/observability/admin/authentication/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://docs.splunk.com/observability/en/gdi/opentelemetry/exposed-endpoints.html)
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Usage

To use this role, simply include the 
`signalfx.splunk_otel_collector.collector` role invocation in your 
playbook. Note that this role requires root access. The following example shows 
how to use the role in a playbook with minimal required configuration:


```yaml
- name: Install Splunk OpenTelemetry Collector
  hosts: all
  become: yes
  # For Windows "become: yes" will raise error.
  # "The Powershell family is incompatible with the sudo become plugin". Remove "become: yes" tag to run on Windows
  tasks:
    - name: "Include splunk_otel_collector"
      include_role:
        name: "signalfx.splunk_otel_collector.collector"
      vars:
        splunk_access_token: YOUR_ACCESS_TOKEN
        splunk_hec_token: YOUR_HEC_TOKEN
        splunk_realm: SPLUNK_REALM
```

> **_NOTE:_**  Setting splunk_hec_token is optional.

You can disable starting the collector and fluentd services by setting 
the argument `start_service` to `false`:

```terminal
$> ansible-playbook playbook.yaml -e start_service=false
```

## Role Variables

### Collector

- `splunk_access_token` (**Required**): The Splunk access token to
  authenticate requests.

- `splunk_realm`: Which realm to send the data to. The `SPLUNK_REALM`
  environment variable will be set with this value for the Splunk OTel 
  Collector service. (**default:** `us0`)

- `splunk_ingest_url`: The Splunk ingest URL, e.g.
  `https://ingest.us0.signalfx.com`. The `SPLUNK_INGEST_URL` environment
  variable will be set with this value for the collector service. (**default:**
  `https://ingest.{{ splunk_realm }}.signalfx.com`)

- `splunk_api_url`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`.
  The `SPLUNK_API_URL` environment variable will be set with this value for the
  collector service. (**default:** `https://api.{{ splunk_realm }}.signalfx.com`)

- `splunk_hec_url`: The Splunk HEC endpoint URL, e.g.
  `https://ingest.us0.signalfx.com/v1/log`. The `SPLUNK_HEC_URL` environment
  variable will be set with this value for the collector service. (**default:**
  `{{ splunk_ingest_url }}/v1/log`)

- `splunk_otel_collector_version`: Version of the collector package to install, e.g.
  `0.25.0`. (**default:** `latest`)

- `splunk_otel_collector_config`: Splunk OTel Collector config YAML file. Can be set to 
  `/etc/otel/collector/gateway_config.yaml` to install the collector in gateway
  mode. (**default:** `/etc/otel/collector/agent_config.yaml` on Linux, 
  **default:** `%ProgramData%\Splunk\OpenTelemetry Collector\agent_config.yaml` on Windows)

- `splunk_config_override`: Custom Splunk OTel Collector config that will be merged into the default config.

- `splunk_config_override_list_merge`: This variable is used to configure `list_merge` option for merging lists in `splunk_config_override` with lists in default config. Allowed options are `replace`, `keep`, `append`, `prepend`, `append_rp` or `prepend_rp`. More details: https://docs.ansible.com/ansible/latest/user_guide/playbooks_filters.html#combining-hashes-dictionaries. (**default:** `replace`)

- `splunk_otel_collector_config_source`: Source path to a Splunk OTel Collector config YAML 
  file on your control host that will be uploaded and set in place of
  `splunk_otel_collector_config` in remote hosts. Can be used to submit a custom collector 
  config, e.g. `./custom_collector_config.yaml`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_otel_collector_config` will be used)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/pkg/extension/smartagentextension/README.md).
  The default path is provided by the collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service. (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle` on Linux, **default:** 
  `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent-bundle` on Windows)

- `splunk_collectd_dir`: The path to the collectd config directory for the
  Smart Agent bundle. The default path is provided by the collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node. The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the collector service.
  (**default:** `/usr/lib/splunk-otel-collector/agent-bundle` on Linux,
  **default:** `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent-bundle\run\collectd` on Windows)

- `splunk_service_user` and `splunk_service_group` (Linux only): Set the user/group
  ownership for the collector service. The user/group will be created if they
  do not exist. (**default:** `splunk-otel-collector`)

- `splunk_otel_collector_proxy_http` and `splunk_otel_collector_proxy_https`
  (Linux only): Set the proxy address, respectively for `http_proxy` and
  `https_proxy` environment variables, to be used by the collector service
  **if** at least one of them is not empty. It must be a full URL like
  `http://user:pass@10.0.0.42`. Notice this proxy is not used by ansible
  itself during deployment. (**default:** ``)

- `splunk_otel_collector_no_proxy` (Linux only): Set the ip and/or hosts that
  will not use `splunk_otel_collector_proxy_http` or
  `splunk_otel_collector_proxy_https`. This variable is only used if
  `splunk_otel_collector_proxy_http` or `splunk_otel_collector_proxy_https` is
  defined. (**default:** `localhost,127.0.0.1,::1`)

- `splunk_otel_collector_command_line_args`: Command-line arguments to pass to the
  Splunk OpenTelemetry Collector. These will be added as arguments to the service
  command line.
  (**default:** `""`)
  
  Example:
  ```yaml
  splunk_otel_collector_command_line_args: "--discovery --set=processors.batch.timeout=10s"
  ```

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the Splunk OTel 
  Collector. (**default:** `512`)

- `gomemlimit`: The `GOMEMLIMIT` environment variable is introduced for the Splunk Otel Collector version >=0.97.0, allowing the limitation of memory usage in the GO runtime. This feature can help enhance GC (Garbage Collection) related performance and prevent GC related Out of Memory (OOM) situations.

- `splunk_listen_interface`: The network interface the collector receivers will listen on.
  (**default** `0.0.0.0`).

- `splunk_skip_repo` (Linux only): If installing the collector from a custom or self-hosted
  apt/yum repo, set to `true` to skip the installation of the default repo
  (**default:** `false`)

- `start_service`: Whether to restart the services installed by the playbook. (**default:** true)

- `splunk_otel_collector_additional_env_vars`: Dictionary of additional environment variables
  from the collector configuration file for the collector service (**default:** `{}`).
  For example, if the collector configuration file includes references to `${MY_CUSTOM_VAR1}`
  and `${MY_CUSTOM_VAR2}`, specify the following to allow the collector service to expand these
  variables:
  ```yaml
  splunk_otel_collector_additional_env_vars:
    MY_CUSTOM_VAR1: value1
    MY_CUSTOM_VAR2: value2
  ```
  On Linux, the variables/values will be added to the
  `/etc/otel/collector/splunk-otel-collector.conf` systemd environment file.
  On Windows, the variables/values will be added to the `Environment` value under the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\splunk-otel-collector`
  registry key.

#### Windows Proxy

The collector and fluentd installation on Windows relies on [win_get_url](https://docs.ansible.com/ansible/latest/collections/ansible/windows/win_get_url_module.html),
which allows setting up a proxy to download the collector binaries.

- `win_proxy_url` (Windows only): An explicit proxy to use for the request. By default, the request will use the IE defined proxy unless `win_use_proxy` is set to `no`. (**default:** ``)
- `win_use_proxy` (Windows only): If set to `no`, it will not use the proxy defined in IE for the current user. (**default:** `no`)
- `win_proxy_username` (Windows only): The username to use for proxy authentication. (**default:** ``)
- `win_proxy_password` (Windows only): The password for `win_proxy_username`. (**default:** ``)

### Fluentd

> **_NOTE:_**  Fluentd support has been deprecated and will be removed in a future release.
> Please refer to [deprecation documentation](../../../../docs/deprecations/fluentd-support.md) for more information.

- `install_fluentd`: Whether to install/manage fluentd and dependencies for log
  collection. The dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](
  https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `false`)

- `td_agent_version`: Version of td-agent (fluentd package) that will be 
  installed (**default:** `4.3.2`)

- `splunk_fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf` on Linux, 
  **default:** `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf` on Windows)

- `splunk_fluentd_config_source`: Source path to a fluentd config file on your 
  control host that will be uploaded and set in place of `splunk_fluentd_config` on
  remote hosts. Can be used to submit a custom fluentd config,
  e.g. `./custom_fluentd_config.conf`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_fluentd_config` will be used)

### Auto Instrumentation on Linux

**Note:** The application(s) to be instrumented on the node need to be
restarted separately after installation/configuration in order for any change
to take effect.

- `install_splunk_otel_auto_instrumentation` (Linux only): Whether to
  install/manage [Splunk OpenTelemetry Auto Instrumentation for Java](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation).
  When set to `true`, the `splunk-otel-auto-instrumentation` deb/rpm package
  will be downloaded and installed from the Collector repository. (**default:**
  `false`)

- `splunk_otel_auto_instrumentation_sdks`: List of Splunk OpenTelemetry Auto
  Instrumentation SDKs to install, configure, and activate. (**default:**
  `['java', 'nodejs', 'dotnet']`)

  Currently, the following values are supported:
  - `java`: [Splunk OpenTelemetry for Java](https://github.com/signalfx/splunk-otel-java)
  - `nodejs`: [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js)
  - `dotnet` [Splunk OpenTelemetry for .NET](https://github.com/signalfx/splunk-otel-dotnet) (x86_64/amd64 only)

  **Note:** This role does not manage the installation/configuration of
  Node.js, `npm`, or Node.js applications. If `nodejs` is included in this
  option, Node.js and `npm` are required to be pre-installed on the node in
  order to install and activate the Node.js SDK.

- `splunk_otel_auto_instrumentation_version` (Linux only): Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`.
  The minimum supported version is `0.48.0`. The minimum supported version for
  Node.js auto instrumentation is `0.87.0`. The minimum supported version for
  .NET auto instrumentation is `0.99.0` (**default:** `latest`)

- `splunk_otel_auto_instrumentation_systemd` (Linux only): By default, the
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

- `splunk_otel_auto_instrumentation_ld_so_preload` (Linux only): Configure this
  variable to include additional library paths, e.g. `/path/to/my.library.so`,
  to `/etc/ld.so.preload`. (**default:** ``)

- `splunk_otel_auto_instrumentation_java_agent_jar` (Linux only): Path to the
  [Splunk OpenTelemetry Java agent](
  https://github.com/signalfx/splunk-otel-java). The default path is provided
  by the `splunk-otel-auto-instrumentation` package. If the path is changed
  from the default value, the path should be an existing file on the node.
  (**default:** `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`)

- `splunk_otel_auto_instrumentation_npm_path`: If the
  `splunk_otel_auto_instrumentation_sdks` option includes `nodejs`, the Splunk
  OpenTelemetry for Node.js SDK will be installed only if the `npm --version`
  shell command is successful. Use this option to specify a custom path on the
  node for `npm`, for example `/my/custom/path/to/npm`. (**default:** `npm`)

  **Note:** This role does not manage the installation/configuration of
  Node.js or `npm`.

- `splunk_otel_auto_instrumentation_resource_attributes`:
  Configure the OpenTelemetry instrumentation resource attributes,
  e.g. `deployment.environment=prod,my.key=value` (comma-separated
  `key=value` pairs. (**default:** ``)

- `splunk_otel_auto_instrumentation_service_name` (Linux only): Explicitly set
  the service name for ***all*** instrumented applications on the node, e.g.
  `my.service`. By default, the service name is automatically generated for
  each instrumented application. (**default:** ``)

- **DEPRECATED** `splunk_otel_auto_instrumentation_generate_service_name`
  (Linux only): Set this option to `false` to prevent the preloader from
  setting the `OTEL_SERVICE_NAME` environment variable. Only applicable if
  `splunk_otel_auto_instrumentation_version` is < `0.87.0`.
  (**default:** `true`)

- **DEPRECATED** `splunk_otel_auto_instrumentation_disable_telemetry`
  (Linux only): Enable or disable the preloader from sending the
  `splunk.linux-autoinstr.executions` metric to the local collector. Only
  applicable if `splunk_otel_auto_instrumentation_version` is < `0.87.0`.
  (**default:** `false`)

- `splunk_otel_auto_instrumentation_enable_profiler` (Linux only): Enable or
  disable AlwaysOn CPU Profiling. (**default**: `false`)

- `splunk_otel_auto_instrumentation_enable_profiler_memory` (Linux only):
  Enable or disable AlwaysOn Memory Profiling. (**default:** `false`)

- `splunk_otel_auto_instrumentation_enable_metrics` (Linux only): Enable or
  disable exporting instrumentation metrics. (**default**: `false`)

- `splunk_otel_auto_instrumentation_otlp_endpoint` (Linux only): Set the OTLP
  endpoint for captured traces. The value will be set to the
  `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable. Only applicable if
  `splunk_otel_auto_instrumentation_version` is `latest` or >= `0.87.0`.
  (**default:** ``, i.e. defer to the default `OTEL_EXPORTER_OTLP_ENDPOINT`
  value for each activated SDK)

- `splunk_otel_auto_instrumentation_otlp_endpoint_protocol` (Linux only): Set
  the protocol for the OTLP endpoint, for example `grpc` or `http/protobuf`.
  The value will be set to the `OTEL_EXPORTER_OTLP_PROTOCOL` environment
  variable. Only applicable if `splunk_otel_auto_instrumentation_version` is
  `latest` or >= `0.104.0`.
  (**default:** ``, i.e. defer to the default `OTEL_EXPORTER_OTLP_PROTOCOL`
  value for each activated SDK)

- `splunk_otel_auto_instrumentation_metrics_exporter` (Linux only):
  Comma-separated list of exporters for collected metrics by all activated
  SDKs, for example `otlp,prometheus`. Set the value to `none` to disable
  collection and export of metrics. The value will be set to the
  `OTEL_METRICS_EXPORTER` environment variable. Only applicable if
  `splunk_otel_auto_instrumentation_version` is `latest` or >= `0.104.0`.
  (**default:** ```, i.e. defer to the default `OTEL_METRICS_EXPORTER` value
  for each activated SDK)

- `splunk_otel_auto_instrumentation_logs_exporter` (Linux only):
  Set the exporter for collected logs by all activated SDKs, for example
  `otlp`. Set the value to `none` to disable collection and export of logs.
  The value will be set to the `OTEL_LOGS_EXPORTER` environment variable. Only
  applicable if `splunk_otel_auto_instrumentation_version` is `latest` or >=
  `0.108.0`.
  (**default:** ```, i.e. defer to the default `OTEL_LOGS_EXPORTER` value
  for each activated SDK)

### Auto Instrumentation for .NET on Windows

***Warning:*** The `Environment` property in the
`HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC` registry key will
be overwritten by the options specified below to enable/configure auto
instrumentation for IIS. Use the
`splunk_dotnet_auto_instrumentation_additional_options` option (see below for
details) to include any other environment variables required for IIS.

**Note:** By default, IIS will be restarted with the `iisreset` command (if it
exists) after installation/configuration. Applications ***not*** running within
IIS need to be restarted/managed separately in order for any changes to take
effect.

For proxy options, see the [Windows Proxy](#windows-proxy) section.

- `install_splunk_dotnet_auto_instrumentation` (Windows only): Whether to
  install/manage [Splunk Distribution of OpenTelemetry .NET](
  https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/get-started.html). (**default:** `false`)

- `splunk_dotnet_auto_instrumentation_version` (Windows only): Version of the
  `splunk-otel-dotnet` project to download and install from
  [GitHub Releases](https://github.com/signalfx/splunk-otel-dotnet/releases).
  By default, a request will be made to
  `https://api.github.com/repos/signalfx/splunk-otel-dotnet/releases/latest`
  to determine the latest release. If a version is specified, for example
  `1.0.0`, the API request will be skipped and files will be
  downloaded from
  `https://github.com/signalfx/splunk-otel-dotnet/releases/download/v{{ splunk_dotnet_auto_instrumentation_version }}`.
  (**default:** `latest`)

- `splunk_dotnet_auto_instrumentation_url` (Windows only): Specify the
  URL to download the `splunk-otel-dotnet` files to skip the GitHub API
  request, for example
  `https://github.com/signalfx/splunk-otel-dotnet/releases/download/v1.8.0`,
  or to download the files from a custom host, for example
  `https://my.host/`. If specified, the
  `splunk_dotnet_auto_instrumentation_version` option is ignored.
  (**default:** ``)

- `splunk_dotnet_auto_instrumentation_github_token` (Windows only): Specify
  a token to authenticate with the GitHub API when making requests to get the
  latest `splunk-otel-dotnet` release. A token is recommended when
  `splunk_dotnet_auto_instrumentation_version` is `latest` or when not using
  `splunk_dotnet_auto_instrumentation_url` since unauthenticated requests
  are [rate-limited](https://docs.github.com/en/rest/rate-limit) by GitHub.
  (**default:** ``)

- `splunk_dotnet_auto_instrumentation_iisreset` (Windows only): By default,
  the `iisreset.exe` command (if it exists) will be executed after
  installation/configuration in order for any changes to take effect for IIS
  applications. Set this option to `false` to skip this step if IIS is managed
  separately or is not applicable. (**default:** `true`)

- `splunk_dotnet_auto_instrumentation_system_wide` (Windows only): By
  default, the `Environment` property in the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\W3SVC` registry key
  will be configured for the following environment variables and any from the
  `splunk_dotnet_auto_instrumentation_additional_options` option to
  enable/configure auto instrumentation for ***only*** IIS applications:
  ```yaml
  COR_ENABLE_PROFILING: "1"  # Required
  COR_PROFILER: "{918728DD-259F-4A6A-AC2B-B85E1B658318}"  # Required
  CORECLR_ENABLE_PROFILING: "1"  # Required
  CORECLR_PROFILER: "{918728DD-259F-4A6A-AC2B-B85E1B658318}"  # Required
  OTEL_RESOURCE_ATTRIBUTES: "deployment.environment={{ splunk_dotnet_auto_instrumentation_environment }},{{ splunk_otel_auto_instrumentation_resource_attributes }},splunk.zc.method=splunk-otel-dotnet-1.8.0"
  OTEL_SERVICE_NAME: "{{ splunk_dotnet_auto_instrumentation_service_name }}"
  SPLUNK_PROFILER_ENABLED: "{{ splunk_dotnet_auto_instrumentation_enable_profiler }}"
  SPLUNK_PROFILER_MEMORY_ENABLED: "{{ splunk_dotnet_auto_instrumentation_enable_profiler_memory }}"
  ```
  Set this option to `true` to also add these environment variables and any
  from the `splunk_dotnet_auto_instrumentation_additional_options` option to
  the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
  registry key to enable/configure auto instrumentation for ***all*** .NET
  applications on the node. (**default:** `false`)

- `splunk_dotnet_auto_instrumentation_environment` (Windows only): Configure
  this option to set the "Environment" value to be reported to Splunk APM, for
  example `production`. The value is assigned to the `OTEL_RESOURCE_ATTRIBUTES` environment
  variable in the Windows registry (**default:** ``, i.e. the "Environment"
  will appear as `unknown` in Splunk APM for the instrumented
  service/application) using the `deployment.environment` attribute key.

- `splunk_dotnet_auto_instrumentation_service_name` (Windows only): Configure
  this variable to override the [auto-generated service name](
  https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/configuration/advanced-dotnet-configuration.html#changing-the-default-service-name)
  for the instrumented service/application, for example `my-service-name`. The
  value is assigned to the `OTEL_SERVICE_NAME` environment variable in the
  Windows registry. (**default:** ``)

- `splunk_dotnet_auto_instrumentation_enable_profiler` (Windows only): Set
  this option to `true` to enable AlwaysOn Profiling. The value will be
  assigned to the `SPLUNK_PROFILER_ENABLED` environment variable in the
  Windows registry. (**default:** `false`)

- `splunk_dotnet_auto_instrumentation_enable_profiler_memory` (Windows only):
  Set this option to `true` to enable AlwaysOn Memory Profiling. The value will
  be assigned to the `SPLUNK_PROFILER_MEMORY_ENABLED` environment variable in
  the Windows registry. (**default:** `false`)

- `splunk_dotnet_auto_instrumentation_additional_options` (Windows only):
  Dictionary of environment variables to be added to the Windows registry
  ***in addition*** to the options above. (**default:** `{}`)

  For example:
  ```yaml
  splunk_dotnet_auto_instrumentation_additional_options:
    SOME_ENV_VAR_00: "1.2.3"
    SOME_ENV_VAR_01: "false"
  ```
  Check the [configuration options](
  https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/configuration/advanced-dotnet-configuration.html#configure-the-splunk-distribution-of-opentelemetry-net)
  for more details about the options above and other supported options.
