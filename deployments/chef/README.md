# Splunk OpenTelemetry Collector Cookbook

This cookbook installs and configures the Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux and Windows machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html).

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html#admin-org-tokens)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Linux
Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 9, 10
- SUSE: 12, 15 (**Note:** Only for collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 16.04, 18.04, 20.04

## Windows
Currently, the following Windows versions are supported:

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit

## Usage

To install the collector, simply include the `splunk-otel-collector::default` recipe in run_list,
and set attribute on the node's run_state. Below is example to configure attributes
```yaml 
{
    "splunk-otel-collector": {
        "splunk_access_token": "testing123",
        "splunk_realm": "test",
        "splunk_ingest_url": "https://ingest.test.signalfx.com",
        "splunk_api_url": "https://api.test.signalfx.com",
        "splunk_service_user": "splunk-otel-collector",
        "splunk_service_group": "splunk-otel-collector",
        "with_fluentd": true
    }
}
```

# This Cookbook accepts the following attributes

- `splunk_access_token` (**Required**): The Splunk access token to
  authenticate requests.

- `splunk_realm` (**Required:**): Which realm to send the data to. The `SPLUNK_REALM`
  environment variable will be set with this value for the Splunk OTel 
  Collector service.

- `splunk_ingest_url`: The Splunk ingest URL, e.g.
  `https://ingest.us0.signalfx.com`. The `SPLUNK_INGEST_URL` environment
  variable will be set with this value for the collector service. (**default:**
  `https://ingest.{{ splunk_realm }}.signalfx.com`)

- `splunk_api_url`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`.
  The `SPLUNK_API_URL` environment variable will be set with this value for the
  collector service. (**default:** `https://api.{{ splunk_realm }}.signalfx.com`)

- `splunk_otel_collector_version`: Version of the collector package to install, e.g.
  `0.25.0`. (**default:** `nil` on Linux, **default:** `latest` on Windows)

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the Splunk OTel 
  Collector. (**default:** `512`)

- `splunk_ballast_size_mib`: Memory ballast size in MiB that will be set to the Splunk 
  OTel Collector. (**default:** 1/3 of `splunk_memory_total_mib`)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
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

- `collector_config_source`: Source path to the collector config YAML file. This file will be copied to the `collector_config_dest` path on the node. See the [source attribute](https://docs.chef.io/resources/remote_file/) of the file resource for supported value types. The default source file is provided by the collector package. (**default:** `/etc/otel/collector/agent_config.yaml` on Linux, 
  **default:** `%ProgramFiles%\Splunk\OpenTelemetry Collector\agent_config.yaml` on Windows)

- `package_stage`: The package repository to use.  Can
be `release` (default, for main releases), `beta` (for beta releases), or `test`
(for unsigned test releases).

- `splunk_service_user` and `splunk_service_group` (Linux only): Set the user/group
  ownership for the collector service. The user/group will be created if they
  do not exist. (**default:** `splunk-otel-collector`)

- `install_fluentd`: Whether to install/manage fluentd and dependencies for log
  collection. The dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](
  https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `true`)

- `with_fluentd`: Version of td-agent (fluentd package) that will be 
  installed (**default:** `3.3.0` for Debian jessie, `3.7.1` for Debian 
  stretch, and `4.1.1` for other distros`)

- `fluentd_config_source`: Source path to the fluentd config file. This file will be copied to the `fluentd_config_dest` path on the node. See the [source attribute](https://docs.chef.io/resources/remote_file/) of the file resource for supported value types. The default source file is provided by the collector package. Only applicable if `with_fluentd` is set to true.
  (**default:** `/etc/otel/collector/fluentd/fluent.conf` on Linux, 
  **default:** `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf` on Windows)

- `node['splunk-otel-collector']['collector_config']`: Collector configuration object.  Everything
underneath this object gets directly converted to YAML and becomes the collector
config file.
