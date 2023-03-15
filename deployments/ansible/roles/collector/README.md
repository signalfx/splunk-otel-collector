# Splunk OpenTelemetry Collector Ansible Role

Ansible role that installs Splunk OpenTelemetry Collector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
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
        splunk_realm: SPLUNK_REALM
```

You can disable starting the collector and fluentd services by setting 
the argument `start_service` to `false`:

```
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

- `splunk_trace_url`: The Splunk trace endpoint URL, e.g.
  `https://ingest.us0.signalfx.com/v2/trace`. The `SPLUNK_TRACE_URL` environment
  variable will be set with this value for the collector service. (**default:**
  `{{ splunk_ingest_url }}/v2/trace`)

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

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the Splunk OTel 
  Collector. (**default:** `512`)

- `splunk_ballast_size_mib`: Memory ballast size in MiB that will be set to the Splunk 
  OTel Collector. (**default:** 1/3 of `splunk_memory_total_mib`)

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
  On Windows, the variables/values will be added to the
  `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
  registry key.

#### Windows Proxy

The collector and fluentd installation on Windows relies on [win_get_url](https://docs.ansible.com/ansible/latest/collections/ansible/windows/win_get_url_module.html),
which allows setting up a proxy to download the collector binaries.

- `win_proxy_url` (Windows only): An explicit proxy to use for the request. By default, the request will use the IE defined proxy unless `win_use_proxy` is set to `no`. (**default:** ``)
- `win_use_proxy` (Windows only): If set to `no`, it will not use the proxy defined in IE for the current user. (**default:** `no`)
- `win_proxy_username` (Windows only): The username to use for proxy authentication. (**default:** ``)
- `win_proxy_password` (Windows only): The password for `win_proxy_username`. (**default:** ``)

### Fluentd

- `install_fluentd`: Whether to install/manage fluentd and dependencies for log
  collection. The dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](
  https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `true`)

- `td_agent_version`: Version of td-agent (fluentd package) that will be 
  installed (**default:** `3.7.1` for Debian stretch and `4.3.2` for other
  distros)

- `splunk_fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf` on Linux, 
  **default:** `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf` on Windows)

- `splunk_fluentd_config_source`: Source path to a fluentd config file on your 
  control host that will be uploaded and set in place of `splunk_fluentd_config` on
  remote hosts. Can be used to submit a custom fluentd config,
  e.g. `./custom_fluentd_config.conf`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_fluentd_config` will be used)

### Auto Instrumentation

**Note:** The Java application(s) on the node need to be restarted separately
after installation/configuration in order for any change to take effect.

- `install_splunk_otel_auto_instrumentation` (Linux only): Whether to
  install/manage [Splunk OpenTelemetry Auto Instrumentation for Java](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation).
  When set to `true`, the `splunk-otel-auto-instrumentation` deb/rpm package
  will be downloaded and installed from the Collector repository. (**default:**
  `false`)

- `splunk_otel_auto_instrumentation_version` (Linux only): Version of the
  `splunk-otel-auto-instrumentation` package to install, e.g. `0.50.0`.
  The minimum supported version is `0.48.0`. (**default:** `latest`)

- `splunk_otel_auto_instrumentation_ld_so_preload` (Linux only): By default,
  the `/etc/ld.so.preload` file on the node will be configured for the
  `/usr/lib/splunk-instrumentation/libsplunk.so` [shared object library](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#operation)
  provided by the `splunk-otel-auto-instrumentation` package and is required
  for auto instrumentation. Configure this variable to include additional
  library paths, e.g. `/path/to/my.library.so`. (**default:** ``)

- `splunk_otel_auto_instrumentation_java_agent_jar` (Linux only): Path to the
  [Splunk OpenTelemetry Java agent](
  https://github.com/signalfx/splunk-otel-java). The default path is provided
  by the `splunk-otel-auto-instrumentation` package. If the path is changed
  from the default value, the path should be an existing file on the node. The
  specified path will be added to the
  `/usr/lib/splunk-instrumentation/instrumentation.conf` config file on the
  node. (**default:**
  `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`)

- `splunk_otel_auto_instrumentation_resource_attributes` (Linux only):
  Configure the OpenTelemetry instrumentation [resource attributes](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#configuration-file),
  e.g. `deployment.environment=prod`. The specified resource attribute(s) will
  be added to the `/usr/lib/splunk-instrumentation/instrumentation.conf` config
  file on the node. (**default:** ``)

- `splunk_otel_auto_instrumentation_service_name` (Linux only): Explicitly set
  the [service name](
  https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#configuration-file),
  for the instrumented Java application, e.g. `my.service`. By default, the
  service name is automatically derived from the arguments of the Java
  executable on the node. The specified service name will be added to the
  `/usr/lib/splunk-instrumentation/instrumentation.conf` config file on the
  node, overriding any generated service name. (**default:** ``)

- `splunk_otel_auto_instrumentation_generate_service_name` (Linux only): Set
  this option to `false` to prevent the preloader from setting the
  `OTEL_SERVICE_NAME` environment variable. (**default:** `true`)

- `splunk_otel_auto_instrumentation_disable_telemetry` (Linux only): Enable or
  disable the preloader from sending the `splunk.linux-autoinstr.executions`
  metric to the local collector. (**default:** `false`)

- `splunk_otel_auto_instrumentation_enable_profiler` (Linux only): Enable or
  disable AlwaysOn CPU Profiling. (**default**: `false`)

- `splunk_otel_auto_instrumentation_enable_profiler_memory` (Linux only):
  Enable or disable AlwaysOn Memory Profiling. (**default:** `false`)

- `splunk_otel_auto_instrumentation_enable_metrics` (Linux only): Enable or
  disable exporting Micrometer metrics. (**default**: `false`)
