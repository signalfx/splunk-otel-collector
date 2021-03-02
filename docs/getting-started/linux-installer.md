# Linux Installer Script

For non-containerized Linux environments, an installer script is available. The
script deploys and configures:

- Splunk OpenTelemetry Connector for Linux
- [Fluentd (via the TD Agent)](https://www.fluentd.org/)

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
running the script with the `-h` flag.

```sh
$ sh /tmp/splunk-otel-collector.sh -h
```

One additional parameter that may need to changed is `--memory` in order to
configure the memory allocation.

> By default, this variable is set to `512`. If you have allocated more memory
> to the Collector then you must increase this setting.

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --memory SPLUNK_MEMORY_TOTAL_MIB \
    -- SPLUNK_ACCESS_TOKEN
```

### Collector Configuration

The Collector comes with a default configuration which can be found at
`/etc/otel/collector/splunk_otel_linux.yaml`. This configuration can be
modified as needed. Possible configuration options can be found in the
`receivers`, `processors`, `exporters`, and `extensions` folders of either:

- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector)
- [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)

After modification, the Collector service needs to be restarted:

```sh
sudo systemctl restart splunk-otel-collector
```

### Fluentd Configuration

By default, the fluentd service will be installed and configured to forward log
events with the `@SPLUNK` label to the collector (see below for how to add
custom fluentd log sources), and the collector will send these events to the
HEC ingest endpoint determined by the `--realm SPLUNK_REALM` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.

To configure the collector to send log events to a custom HEC endpoint URL, you
can specify the following parameters for the installer script:

- `--hec-url URL`
- `--hec-token TOKEN`

The main fluentd configuration file will be installed to
`/etc/otel/collector/fluentd/fluent.conf`. Custom fluentd source config files
can be added to the `/etc/otel/collector/fluentd/conf.d` directory after 
installation. Please note:

- All files in this directory ending `.conf` extension will automatically be
 included by Fluentd.
- The "td-agent" user must have permissions to access the files specified in
  the config files and the paths defined within.
- By default, Fluentd will be configured to collect systemd journal log events
from `/var/log/journal`.

After any configuration modification, the td-agent service needs to be restarted:

```sh
sudo systemctl restart td-agent
```

### Uninstall

If you wish to uninstall you can run:

```sh
$ sudo sh /tmp/splunk-otel-collector.sh --uninstall
```

> Note that configuration files may be left on the filesystem.
