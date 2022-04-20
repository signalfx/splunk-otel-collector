# Splunk OpenTelemetry Collector Cookbook

This cookbook installs and configures the Splunk OpenTelemetry Collector to
collect metrics, traces and logs from Linux and Windows machines and sends
data to [Splunk Observability Cloud](
https://www.splunk.com/en_us/observability.html).

## Prerequisites

- [Splunk Access Token](https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html)
- [Splunk Realm](https://dev.splunk.com/observability/docs/realms_in_endpoints/)
- [Double-check exposed ports](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/security.md#exposed-endpoints) 
  to make sure your environment doesn't have conflicts. Ports can be changed in the collector's configuration.

## Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 9, 10, 11
- SUSE: 12, 15 (**Note:** Only for Collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 18.04, 20.04, 22.04 (**Note:** Log collection with Fluentd [not currently supported on Ubuntu 22.04](https://www.fluentd.org/blog/td-agent-v4.3.1-has-been-released).)

## Windows

Currently, the following Windows versions are supported:

- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

## Usage

To install the Collector and Fluentd, include the
`splunk-otel-collector::default` recipe in the `run_list`, and set the
attributes on the node's `run_state`. Below is an example to configure the
required `splunk_access_token` attribute and some optional attributes:
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

# This cookbook accepts the following attributes

- `splunk_access_token` (**Required**): The [Splunk access token](
  https://docs.splunk.com/Observability/admin/authentication-tokens/org-tokens.html)
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

- `splunk_memory_total_mib`: Amount of memory in MiB allocated to the
  Collector. The `SPLUNK_MEMORY_TOTAL_MIB` environment variable will be set
  with this value for the Collector service. (**default:** `512`)

- `splunk_ballast_size_mib`: Explicitly set the ballast size for the Collector
  instead of the value calculated from the `splunk_memory_total_mib` attribute.
  This should be set to 1/3 to 1/2 of configured memory. The
  `SPLUNK_BALLAST_SIZE_MIB` environment variable will be set with this value
  for the Collector service. (**default:** `''`)

- `splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
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

- `package_stage`: The Collector package repository stage to use.  Can be
  `release`, `beta`, or `test`. (**default:** `release`)

- `splunk_service_user` and `splunk_service_group` (Linux only): Set the
  user/group ownership for the Collector service. The user/group will be
  created if they do not exist. (**default:** `splunk-otel-collector`)

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
  `3.7.1` for Debian stretch, and `4.3.0` for all other Linux distros and
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

- `node['splunk-otel-collector']['collector_config']`: The Collector
  configuration object. Everything underneath this object gets directly
  converted to YAML and becomes the Collector config file. (**default:** `{}`)
