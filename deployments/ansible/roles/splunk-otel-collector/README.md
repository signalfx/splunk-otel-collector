# Role Name

Ansible role that installs Splunk OpenTelemetry Connector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Usage

To install the role, clone the agent repo to your control host and add the 
`splunk-otel-collector` role directory to its `roles_path` in `ansible.cfg`, 
or use this document's directory as your working directory.

To use this role, simply include the `splunk-otel-collector` role invocation 
in your playbook. The following example shows how to use the role in a
playbook with minimal required configuration:

```yaml
- name: Install Splunk OpenTelemetry Connector
  hosts: all
  become: yes
  vars:
    splunk_access_token: YOUR_ACCESS_TOKEN
    splunk_realm: SPLUNK_REALM
  tasks:
    - name: "Include splunk-otel-collector"
      include_role:
        name: "splunk-otel-collector"
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

- `collector_version`: Version of the collector package to install, e.g.
  `0.26.0`. (**default:** `latest`)

- `collector_config`: Splunk OTel Collector config YAML file. Can be set to 
  /etc/otel/collector/gateway_config.yaml to install the collector in gateway
  mode. (**default:** `/etc/otel/collector/agent_config.yaml`)

- `collector_config_source`: Source path to a Splunk OTel Collector config YAML 
  file on a control node that will be uploaded and set in place of
  `collector_config` in remote nodes. Can be used to submit a custom collector 
  config, e.g. `./custom_collector_config.yaml`. (**default:** `""` meaning 
  that nothing will be copied and existing `collector_config` will be used)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
  The default path is provided by the collector package. If the specified path
  is changed from the default value, the path should be an existing directory
  on the node. The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service.  (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle`)

- `splunk_collectd_dir`: The path to the collectd config directory for the
  Smart Agent bundle. The default path is provided by the collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node. The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the collector service.  (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle`)

- `service_user` and `service_group` (Linux only): Set the user/group
  ownership for the collector service. The user/group will be created if they
  do not exist. (**default:** `splunk-otel-collector`)

- `memory_total_mib`: Amount of memory in MiB allocated to the Splunk OTel 
  Collector. (**default:** `512`)

- `ballast_size_mib`: Memory ballast size in MiB that will be set to the Splunk 
  OTel Collector. (**default:** 1/3 of `memory_total_mib`)

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

- `fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf`)

- `fluentd_config_source`: Source path to a fluentd config file on a 
  control node that will be uploaded and set in place of `fluentd_config` on
  remote nodes. Can be used to submit a custom fluentd config,
  e.g. `./custom_fluentd_config.conf`. (**default:** `""` meaning 
  that nothing will be copied and existing `fluentd_config` will be used)

## Contributing

Check [Contributing guidelines](./CONTRIBUTING.md) if you see something that 
needs to be improved in this Ansible role.