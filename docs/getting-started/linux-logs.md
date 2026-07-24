# Collect logs on Linux

> **Scope:** This guide covers log collection via the [Linux installer script](linux-installer.md) only.

## Prerequisites

- A HEC token with write permissions to the target index. See <https://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector>.
- The Splunk Distribution of OpenTelemetry Collector installed on a Linux host

## Install the Collector with log collection enabled

Pass your HEC token, endpoint URL, and target index to the installer. Log collection is enabled automatically when `--splunk-platform-url` is provided:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-logs-index "<your-index>"
```

To also send metrics and traces to Splunk Observability Cloud at the same time, include your O11y access token and realm:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm us0 \
  --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-logs-index "<your-index>" \
  -- YOUR_O11Y_ACCESS_TOKEN
```

## Installer options

| **Option**                             | **Description**                                                                                                                                                                                       |
|----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--splunk-platform-token <token>`      | Required. HEC token for authenticating to the HEC endpoint.                                                                                                                                           |
| `--splunk-platform-url <url>`          | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`.                                                                                                                |
| `--splunk-platform-logs-index <index>` | Recommended. Index to send logs to. If omitted, logs are routed to the default index configured on the HEC token. If no default index is configured on the token, events are silently dropped.        |

> **Note:** If you omit `--splunk-platform-logs-index`, the `SPLUNK_PLATFORM_LOGS_INDEX` environment variable is not set in the collector's env file. The collector then resolves the index to an empty string and omits it from the HEC payload, leaving Splunk to route events to the HEC token's default index. If the token has no default index configured, events are rejected silently with no error visible in the collector logs. To avoid data loss, always specify `--splunk-platform-logs-index` or ensure the HEC token has a default index set in Splunk.

## What gets collected

By default, the Collector tails files under `/var/log` matching common log patterns (`*.log`, `*log`, `*auth*`, `*secure*`, `*messages*`, and others). This mirrors the behavior of the Splunk Add-on for Unix and Linux.

For the full list of included paths and all available receivers, see the [default configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/splunk_logs_config_linux.yaml).

Additional receivers are defined in the config but disabled by default:

| **Receiver**                                                | **Description**                                                      |
|-------------------------------------------------------------|----------------------------------------------------------------------|
| `file_log/etc`                                              | Config files under `/etc` — one-time read on each collector start    |
| `file_log/bash_history`                                     | Bash history files                                                   |
| `journald`                                                  | Systemd journal — requires group membership (see below)              |
| `file_log/nginx-access`, `file_log/mysql-error`, and others | Application-specific log files — **experimental**, subject to change |

## Enable or disable receivers

Edit `/etc/otel/collector/splunk_logs_config_linux.yaml` and uncomment receivers in the `service.pipelines.logs/hec.receivers` section:

```yaml
service:
  pipelines:
    logs/hec:
      receivers:
        - file_log/varlog       # enabled by default
        #- file_log/etc         # uncomment to enable
        #- file_log/bash_history # uncomment to enable
        #- journald             # uncomment to enable
        #- file_log/nginx-access
        #- file_log/mysql-error
        # ... (see full list in the config file)
```

After editing, restart the service:

```sh
sudo systemctl restart splunk-otel-collector
```

### Enable journald

The `journald` receiver requires the `splunk-otel-collector` user to be in the `systemd-journal` group. Without this, the receiver starts without error but produces no data:

```sh
sudo usermod -aG systemd-journal splunk-otel-collector
sudo systemctl restart splunk-otel-collector
```

## Add additional log files

Every `file_log` receiver has an `include` list of glob patterns. To collect logs from a custom application, add its path to the `file_log/varlog` receiver:

```yaml
file_log/varlog:
  include:
    - /var/log/*.log
    - /var/log/*log
    # ... existing patterns ...
    - /opt/myapp/logs/*.log   # add your path here
```

To exclude specific files within a matched pattern:

```yaml
file_log/varlog:
  include:
    - /var/log/*.log
  exclude:
    - /var/log/noisy-app.log
```

See the [filelog receiver documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) for the full glob syntax and all available options.

## Verify log ingestion

Search for events using the sourcetype set by the default receiver:

```
index="<your-index>" sourcetype="linux:varlog"
```

To see all sourcetypes currently being ingested:

```
index="<your-index>" sourcetype="linux:*" | stats count by sourcetype
```