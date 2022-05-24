# Splunk OpenTelemetry Collector Puppet Module

This is a Puppet module that will install and configure the Splunk
OpenTelemetry Collector.

## Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 9, 10, 11
- SUSE: 12, 15 (**Note:** Only for collector versions v0.34.0 or higher. Log collection with Fluentd not currently supported.)
- Ubuntu: 16.04, 18.04, 20.04, 22.04 (**Note:** Log collection with Fluentd [not currently supported on Ubuntu 22.04](https://www.fluentd.org/blog/td-agent-v4.3.1-has-been-released).)

> Note: `systemd` is required to be installed on the host for service
> management.

## Windows

Currently, the following Windows versions are supported and requires PowerShell
3.0 or newer:

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit

To use this module, include the `splunk_otel_collector` class in your
manifests with the supported parameters (see the table below for descriptions
of the available parameters).  For example, the simplest deployment definition
with the default parameters would be (replace `VERSION` with the desired
collector version, `SPLUNK_ACCESS_TOKEN` with your Splunk access token to
authenticate requests, and `SPLUNK_REALM` for the realm to send the data to):

```ruby
class { splunk_otel_collector:
  collector_version => 'VERSION'
  splunk_access_token => 'SPLUNK_ACCESS_TOKEN',
  splunk_realm => 'SPLUNK_REALM',
}
```

This class accepts the following parameters:

| Name | Description | Default value |
| :--- | :---------- | :------------ |
| `collector_version` | **Required on Windows**: Version of the collector package to install, e.g., `0.25.0`.  The version should correspond to [Github Releases](https://github.com/signalfx/splunk-otel-collector/releases) _without_ the preceding `v`.  **Note**: On Linux, the latest collector version will be installed if this parameter is not specified. | None |
| `splunk_access_token` | **Required**: The Splunk access token to authenticate requests. | None |
| `splunk_realm` | **Required**: Which realm to send the data to, e.g. `us0`.  The Splunk ingest and API URLs will be inferred by this value.  The `SPLUNK_REALM` environment variable will be set with this value for the collector service. | None |
| `splunk_ingest_url` | Set the Splunk ingest URL explicitly instead of the URL inferred by the `$splunk_realm` parameter.  The `SPLUNK_INGEST_URL` environment variable will be set with this value for the collector service. | `https://ingest.${splunk_realm}.signalfx.com` |
| `splunk_api_url` | Set the Splunk API URL explicitly instead of the URL inferred by the `$splunk_realm` parameter.  The `SPLUNK_API_URL` environment variable will be set with this value for the collector service. | `https://api.${splunk_realm}.signalfx.com` |
| `splunk_trace_url` | Set the Splunk trace endpoint URL explicitly instead of the URL inferred by the `$splunk_ingest_url` parameter.  The `SPLUNK_TRACE_URL` environment variable will be set with this value for the collector service. | `${splunk_ingest_url}/v2/trace` |
| `splunk_hec_url` | Set the Splunk HEC endpoint URL explicitly instead of the URL inferred by the `$splunk_ingest_url` parameter.  The `SPLUNK_HEC_URL` environment variable will be set with this value for the collector service. | `${splunk_ingest_url}/v1/log` |
| `splunk_hec_token` | Set the Splunk HEC authentication token if different than `$splunk_access_token`.  The `SPLUNK_HEC_TOKEN` environment variable will be set with this value for the collector service. | `$splunk_access_token` |
| `splunk_bundle_dir` | The path to the [Smart Agent bundle directory](https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).  The default path is provided by the collector package.  If the specified path is changed from the default value, the path should be an existing directory on the node.  The `SPLUNK_BUNDLE_DIR` environment variable will be set to this value for the collector service. | Linux: `/usr/lib/splunk-otel-collector/agent-bundle`<br>Windows: `%PROGRAMFILES%\Splunk\OpenTelemetry Collector\agent-bundle` |
| `splunk_collectd_dir` | The path to the collectd config directory for the Smart Agent bundle.  The default path is provided by the collector package.  If the specified path is changed from the default value, the path should be an existing directory on the node.  The `SPLUNK_COLLECTD_DIR` environment variable will be set to this value for the collector service. | Linux: `${splunk_bundle_dir}/run/collectd`<br>Windows: `${splunk_bundle_dir}\run\collectd` |
| `splunk_memory_total_mib` | Total memory in MIB to allocate to the collector; automatically calculates the ballast size.  The `SPLUNK_MEMORY_TOTAL_MIB` environment variable will be set with this value for the collector service. | `512` |
| `splunk_ballast_size_mib` | Set the ballast size for the collector explicitly instead of the value calculated from the `$splunk_memory_total_mib` parameter.  This should be set to 1/3 to 1/2 of configured memory.  The `SPLUNK_BALLAST_SIZE_MIB` environment variable will be set with this value for the collector service. | None |
| `collector_config_source` | Source path to the collector config YAML file. This file will be copied to the `$collector_config_dest` path on the node.  See the [source attribute](https://puppet.com/docs/puppet/latest/types/file.html#file-attribute-source) of the `file` resource for supported value types.  The default source file is provided by the collector package. | Linux: `/etc/otel/collector/agent_config.yaml`<br>Windows: `%PROGRAMFILES\Splunk\OpenTelemetry Collector\agent_config.yaml` |
| `collector_config_dest` | Destination path of the collector config file on the node.  The `SPLUNK_CONFIG` environment variable will be set with this value for the collector service. | Linux: `/etc/otel/collector/agent_config.yaml`<br>Windows: `%PROGRAMDATA%\Splunk\OpenTelemetry Collector\agent_config.yaml` |
| `service_user` and `$service_group` | **Linux only**: Set the user/group ownership for the collector service. The user/group will be created if they do not exist. | `splunk-otel-collector` |
| `with_fluentd` | Whether to install/manage fluentd and dependencies for log collection.  **Note**: On Linux, the dependencies include [capng_c](https://github.com/fluent-plugins-nursery/capng_c) for enabling [Linux capabilities](https://docs.fluentd.org/deployment/linux-capability), [fluent-plugin-systemd](https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd journal log collection, and the [required libraries/development tools](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md#fluentd-configuration). | `true` |
| `fluentd_config_source` | Source path to the fluentd config file.  This file will be copied to the `$fluentd_config_dest` path on the node.  See the [source attribute](https://puppet.com/docs/puppet/latest/types/file.html#file-attribute-source) of the `file` resource for supported value types.  The default source file is provided by the collector package.  Only applicable if `$with_fluentd` is set to `true`. | Linux: `/etc/otel/collector/fluentd/fluent.conf`<br>Windows: `%PROGRAMFILES\Splunk\OpenTelemetry Collector\fluentd\td-agent.conf` |
| `fluentd_config_dest` | **Linux only**: Destination path to the fluentd config file on the node.  Only applicable if `$with_fluentd` is set to `true`.  **Note**: On Windows, the path will always be set to `%SYSTEMDRIVE%\opt\td-agent\etc\td-agent\td-agent.conf`. | `/etc/otel/collector/fluentd/fluent.conf` |
| `manage_repo` | **Linux only**: In cases where the collector and fluentd apt/yum repositories are managed externally, set this to `false` to disable management of the repositories by this module.  **Note:** If set to `false`, the externally managed repositories should provide the `splunk-otel-collector` and `td-agent` packages.  Also, the apt (`/etc/apt/sources.list.d/splunk-otel-collector.list`, `/etc/apt/sources.list.d/splunk-td-agent.list`) and yum (`/etc/yum.repos.d/splunk-otel-collector.repo`, `/etc/yum.repos.d/splunk-td-agent.repo`) repository definition files will be deleted if they exist in order to avoid any conflicts. | `true` |

## Dependencies

On Linux-based systems, the
[puppetlabs/stdlib](https://forge.puppet.com/puppetlabs/stdlib) module is
required.

On Debian-based systems, the
[puppetlabs/apt](https://forge.puppet.com/puppetlabs/apt) module is required to
manage the collector and fluentd apt repositories.

On RPM-based systems, the
[puppet/yum](https://forge.puppet.com/puppet/yum) module is required to
install the "Development Tools" package group as a dependency for fluentd.

On Windows systems, the
[puppetlabs/registry](https://forge.puppet.com/modules/puppetlabs/registry)
module is required to set the registry key/values, and the
[puppetlabs/powershell](https://forge.puppet.com/modules/puppetlabs/powershell)
module is required to execute Powershell commands.
