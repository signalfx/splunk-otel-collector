# Splunk OpenTelemetry Connector Puppet Module

This is a Puppet module that will install and configure the Splunk
OpenTelemetry Connector.

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 8, 9, 10
- Ubuntu: 16.04, 18.04, 20.04

To use this module, simply include the class `splunk_otel_connector` in your
manifests.  For example, the simplest deployment definition with the
default parameters would be (replace `SPLUNK_ACCESS_TOKEN` with your Splunk
access token to authenticate requests):

```ruby
class { splunk_otel_connector:
  splunk_access_token => 'SPLUNK_ACCESS_TOKEN',
}
```

This class accepts the following parameters:

- `$splunk_access_token` (**Required**): The Splunk access token to
  authenticate requests.

- `$collector_config_source`: Source path to the collector config YAML file.
  This file will be copied to the `$collector_config_dest` path on the node.
  See the [source attribute](
  https://puppet.com/docs/puppet/latest/types/file.html#file-attribute-source)
  of the `file` resource for supported value types.  The default source file is
  provided by the collector package. (**default:**
  `file:///etc/otel/collector/agent_config.yaml`)

- `$collector_config_dest`: Destination path of the collector config file on
  the node.  (**default:** `/etc/otel/collector/agent_config.yaml`)

- `$collector_version`: Version of the collector package to install, e.g.
  `0.25.0`.  (**default:** `latest`)

- `splunk_realm`: Which realm to send the data to.  The `SPLUNK_REALM`
  environment variable will be set with this value for the collector service.
  (**default:** `us0`)

- `splunk_ingest_url`: The Splunk ingest URL, e.g.
  `https://ingest.us0.signalfx.com`.  The `SPLUNK_INGEST_URL` environment
  variable will be set with this value for the collector service. (**default:**
  `https://ingest.${splunk_realm}.signalfx.com`)

- `$splunk_api_url`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`.
  The `SPLUNK_API_URL` environment variable will be set with this value for the
  collector service.  (**default:** `https://api.${splunk_realm}.signalfx.com`)

- `$splunk_trace_url`: The Splunk trace endpoint URL, e.g.
  `https://ingest.us0.signalfx.com/v2/trace`.  The `SPLUNK_TRACE_URL`
  environment variable will be set with this value for the collector service.
  (**default:** `${splunk_ingest_url}/v2/trace`)

- `$splunk_hec_url`: The Splunk HEC endpoint URL, e.g.
  `https://ingest.us0.signalfx.com/v1/log`.  The `SPLUNK_HEC_URL` environment
  variable will be set with this value for the collector service.
  (**default:** `${splunk_ingest_url}/v1/log`)

- `$splunk_hec_token`: The Splunk HEC authentication token.  The
  `SPLUNK_HEC_TOKEN` environment variable will be set with this value for the
  collector service.  (**default:** `$splunk_access_token`)

- `$splunk_bundle_dir`: The path to the [Smart Agent bundle directory](
  https://github.com/signalfx/splunk-otel-collector/blob/main/internal/extension/smartagentextension/README.md).
  The default path is provided by the collector package.  If the specified path
  is changed from the default value, the path should be an existing directory
  on the node.  The `SPLUNK_BUNDLE_DIR` environment variable will be set to
  this value for the collector service.  (**default:**
  `/usr/lib/splunk-otel-collector/agent-bundle`)

- `$splunk_collectd_dir`: The path to the collectd config directory for the
  Smart Agent bundle.  The default path is provided by the collector package.
  If the specified path is changed from the default value, the path should be
  an existing directory on the node.  The `SPLUNK_COLLECTD_DIR` environment
  variable will be set to this value for the collector service.  (**default:**
  `${splunk_bundle_dir}/run/collectd`)

- `$service_user` and `$service_group` (Linux only): Set the user/group
  ownership for the collector service. The user/group will be created if they
  do not exist.  (**default:** `splunk-otel-collector`)

- `$with_fluentd`: Whether to install/manage fluentd and dependencies for log
  collection.  The dependencies include [capng_c](
  https://github.com/fluent-plugins-nursery/capng_c) for enabling
  [Linux capabilities](
  https://docs.fluentd.org/deployment/linux-capability),
  [fluent-plugin-systemd](
  https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd
  journal log collection, and the required libraries/development tools.
  (**default:** `true`)

- `$fluentd_config_source`: Source path to the fluentd config file.  This file
  will be copied to the `$fluentd_config_dest` path on the node.
  See the [source attribute](
  https://puppet.com/docs/puppet/latest/types/file.html#file-attribute-source)
  of the `file` resource for supported value types.  The default source file is
  provided by the collector package. (**default:**
  `file:///etc/otel/collector/fluentd/fluent.conf`)

- `$fluentd_config_dest`: Destination path to the fluentd config file on the
  node.  (**default:** `/etc/otel/collector/fluentd/fluent.conf`)

- `$manage_repo` (Linux only): In cases where the collector and fluentd apt/yum
  repositories are managed externally, set this to `false` to disable
  management of the repositories by this module.  **Note:** If set
  to `false`, the apt (`/etc/apt/sources.list.d/splunk-otel-collector.list` and
  `/etc/apt/sources.list.d/splunk-td-agent.list`) and yum
  (`/etc/yum.repos.d/splunk-otel-collector.repo` and
  `/etc/yum.repos.d/splunk-td-agent.repo`) repository definition files will be
  deleted if they exist in order to avoid any conflicts.  (**default:** `true`)

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

## Release Process
To release a new version of the module, run `./release` in this directory.  You
will need access to the SignalFx account on the Puppet Forge website, and the
release script will give you instructions for what to do there.

You should update the version in `metadata.json` to whatever is most appropriate
for semver and have that committed before running `./release`.

The release script will try to make and push an annotated tag of the form
`puppet-vX.Y.Z` where `X.Y.Z` is the version in the `./metadata.json` file.
