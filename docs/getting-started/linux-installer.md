# Linux Installer Script

For non-containerized Linux environments, a convenience script is available for
installing the Collector package and [TD Agent
(Fluentd)](https://www.fluentd.org/).

## Getting Started

Run the following command on your host. Replace `SPLUNK_REALM`,
`SPLUNK_TOTAL_MEMORY_MIB`, and `SPLUNK_ACCESS_TOKEN` for your
environment:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --memory SPLUNK_TOTAL_MEMORY_MIB \
    -- SPLUNK_ACCESS_TOKEN
```

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 8, 9, 10
- Ubuntu: 16.04, 18.04, 20.04

You can view the [source](../../internal/buildscripts/packaging/installer/install.sh)
for more details and available options.

## Advanced Configuration

By default, the fluentd service will be installed and configured to forward
log events with the `@SPLUNK` label to the collector (see the note below for
how to add fluentd log sources), and the collector will send these events to
the HEC ingest endpoint determined by the `--realm SPLUNK_REALM` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.  To configure the collector
to send log events to a custom HEC endpoint URL, specify the `--hec-url URL`
and `--hec-token TOKEN` options to the command above.

**Note**: The installer script does not include any fluentd log sources. Custom
fluentd source config files can be added to the
`/etc/otel/collector/fluentd/conf.d` directory after installation. Config files
added to this directory should have a `.conf` extension, and the `td-agent`
service will need to be restarted to include/enable the new files, i.e.
`sudo systemctl restart td-agent`.  A sample config and instructions for
collecting journald log events is available at
`/etc/otel/collector/fluentd/conf.d/journald.conf.example`.
