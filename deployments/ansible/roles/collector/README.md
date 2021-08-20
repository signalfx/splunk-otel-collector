# Splunk OpenTelemetry Connector Ansible Role

Ansible role that installs Splunk OpenTelemetry Connector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Linux
Currently, the following Linux distributions and versions are supported:

- Amazon Linux
- RedHat
- Debian: buster, stretch
- Ubuntu: 16.04, 18.04, 20.04

## Windows
Currently, the following Windows versions are supported:
Ansible requires PowerShell 3.0 or newer and atleast .NET4.0 to be installed on Windows host.
A WinRM listner should be created and activated. 
For setting up Windows Host Refer : [Ansible Docs](https://docs.ansible.com/ansible/latest/user_guide/windows_setup.html)

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit

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
- name: Install Splunk OpenTelemetry Connector
  hosts: all
  # For Windows "become: yes" will raise error.
  # "The Powershell family is incompatible with the sudo become plugin" Remove "become: yes" tag to run on Windows
  become: yes
  tasks:
    - name: "Include splunk_otel_collector"
      include_role:
        name: "signalfx.splunk_otel_collector.collector"
      vars:
        splunk_access_token: YOUR_ACCESS_TOKEN
        splunk_realm: SPLUNK_REALM
```

## Role Variables

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

- `splunk_otel_collector_version`: Version of the collector package to install, e.g.
  `0.25.0`. (**default:** `latest`)

- `splunk_otel_collector_config`: Splunk OTel Collector config YAML file. Can be set to 
  `/etc/otel/collector/gateway_config.yaml` to install the collector in gateway
  mode. (**default:** `/etc/otel/collector/agent_config.yaml`)

- `splunk_otel_collector_config_source`: Source path to a Splunk OTel Collector config YAML 
  file on your control host that will be uploaded and set in place of
  `splunk_otel_collector_config` in remote hosts. Can be used to submit a custom collector 
  config, e.g. `./custom_collector_config.yaml`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_otel_collector_config` will be used)

- `splunk_bundle_dir` & : The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
  The default path is provided by the collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service.  (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle`)

- `splunk_collectd_dir` : The path to the collectd config directory for the
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

- `splunk_ballast_size_mib`: Memory ballast size in MiB that will be set to the Splunk 
  OTel Collector. (**default:** 1/3 of `splunk_memory_total_mib`)

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
  installed (**default:** `3.3.0` for Debian jessie, `3.7.1` for Debian 
  stretch, and `4.1.1` for other distros`)

- `splunk_fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf`)

## Windows Specific Role Variables

- `collector_path_win`: Default path of Splunk-otel-collector in windows 
  (**default:** `C:\\Program Files\Splunk\OpenTelemetry Collector\`)

- `splunk_bundle_dir_win` & : The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
  The default path is provided by the collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service.  (**default:** `${collector_path_win}\agent-bundle`)

- `splunk_collectd_dir_win`: The path to the collectd config directory for the
  Smart Agent bundle. The default path is provided by the collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node. The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the collector service.
  (**default:** `${splunk_bundle_dir_win}\run\collectd`)

- `splunk_fluentd_config_source`: Source path to a fluentd config file on your 
  control host that will be uploaded and set in place of `splunk_fluentd_config` on
  remote hosts. Can be used to submit a custom fluentd config,
  e.g. `./custom_fluentd_config.conf`. (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_fluentd_config` will be used)

- `collector_config_source_win`: Source path to the collector config YAML file. This file will 
  be copied to the $collector_config_dest path on the node. See the source attribute of the file 
  resource for supported value types. The default source file is provided by the collector package.
  (**default:** `C:\\Program Files\Splunk\OpenTelemetry Collector\agent_config.yaml`)

- `collector_config_dest_win`: Destination path of the collector config file on the node. 
  The SPLUNK_CONFIG environment variable will be set with this value for the collector service.
  (**default:** `C:\\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`)

- `win_fluentd_config_source`: Source path to the fluentd config file. 
  (**default:** `${collector_path_win}\\fluentd\\td-agent.conf`)
  
- `win_fluentd_config_dest`: On Windows, the path will always be set to default
  (**default:** `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf`)

- `win_td-agent_version`: Version of td-agent (fluentd package) that will be 
  installed in Windows distro (`4.1.1`)

- `win_otel_version`: Version of splunk-otel-collector that will be installed in
  Windows distri(`0.31.0`) 

 
