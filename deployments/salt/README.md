# Splunk OpenTelemetry Collector Salt Formula

This formula installs and configures Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Linux
Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 9, 10, 11
- SUSE: 12, 15 (**Note:** Only for collector versions v0.34.0 or higher. Log collection with fluentd not currently supported.)
- Ubuntu: 16.04, 18.04, 20.04

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
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
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
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
  (**default:** `True`)

- `td_agent_version`: Version of [td-agent](
  https://td-agent-package-browser.herokuapp.com/) (fluentd package) that will
  be installed (**default:** `3.7.1-0` for Debian 9, and `4.3.0` for other
  distros)

- `splunk_fluentd_config`: Path to the fluentd config file on the remote host.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf`)

- `splunk_fluentd_config_source`: Source path to a fluentd config file on your 
  control host that will be uploaded and set in place of `splunk_fluentd_config` on
  remote hosts. To use custom fluentd config add the config file into salt dir, 
  e.g. `salt://templates/td_agent.conf` (**default:** `""` meaning 
  that nothing will be copied and existing `splunk_fluentd_config` will be used)
