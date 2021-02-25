# Linux Installer Script

For non-containerized Linux environments, an installer script is available. The
script deploys and configures:

- Splunk OpenTelemetry Connector for Linux
- [TD Agent
(Fluentd)](https://www.fluentd.org/)

> IMPORTANT: systemd is required to use this script.

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 8, 9, 10
- Ubuntu: 16.04, 18.04, 20.04

## Getting Started

Run the below command on your host. Replace these variables:

- `SPLUNK_REALM`: Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN`: Access token to authenticate requests

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM -- SPLUNK_ACCESS_TOKEN
```

You can view the [source](../../internal/buildscripts/packaging/installer/install.sh)
for more details and available options.

## Advanced Configuration

### Additional Script Options

Additional configuration options supported by the script can be found by
running the script with the `-h` flag. One variable that may need to changed
is `SPLUNK_MEMORY_TOTAL_MIB` in order to configure the memory allocation:

> By default, this variable is set to `512`.

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --memory SPLUNK_TOTAL_MEMORY_MIB \
    -- SPLUNK_ACCESS_TOKEN
```

### Collector Configuration

The Collector comes with a default configuration which can be found at
`/etc/otel/collector/splunk_otel_linux.yaml`. This configuration can be
modified as needed. Possible configuration options can be found in the
`receivers`, `processors`, `exporters`, and `extensions` folders of either:

- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector)
- [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)

After modification, the Collector services needed to be restarted: `sudo
systemctl restart splunk-otel-collector`.

### Fluentd Configuration

By default, the fluentd service will be installed and configured to forward
log events with the `@SPLUNK` label to the collector (see the note below for
how to add fluentd log sources), and the collector will send these events to
the HEC ingest endpoint determined by the `--realm SPLUNK_REALM` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.

To configure the collector to send log events to a custom HEC endpoint URL, you
can specify the following parameters:

- `--hec-url URL`
- `--hec-token TOKEN`

**Note**: The installer script does not include any fluentd log sources. Custom
fluentd source config files can be added to the
`/etc/otel/collector/fluentd/conf.d` directory after installation. Config files
added to this directory should have a `.conf` extension, and the `td-agent`
service will need to be restarted to include/enable the new files, i.e.
`sudo systemctl restart td-agent`.  A sample config and instructions for
collecting journald log events is available at
`/etc/otel/collector/fluentd/conf.d/journald.conf.example`.
