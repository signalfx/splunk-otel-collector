# Ingest Linux data to Splunk Platform

> **Scope:** This guide covers Splunk Platform integration via the [Linux installer script](linux-installer.md) only. For manual configuration or other installation methods, see the [Splunk Platform log and metrics collection reference](../splunk-platform-log-collection.md).

## Prerequisites

- A running Splunk Platform instance (Splunk Enterprise or Splunk Cloud)
- A HEC token with write permissions to the target index. See <https://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector>.
- The Splunk Distribution of OpenTelemetry Collector installed on a Linux host

## Install the Collector with log collection enabled

Pass your Splunk Platform HEC token, endpoint URL, and target index to the installer. Log collection is enabled automatically when `--splunk-platform-url` is provided:

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
  --splunk-platform-logs-index "<your-index>" -- YOUR_O11Y_ACCESS_TOKEN
```

## Installer options

| **Option**                             | **Description**                                                                                                                                                                                       |
|----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--splunk-platform-token <token>`      | Required. HEC token for authenticating to Splunk Platform.                                                                                                                                            |
| `--splunk-platform-url <url>`          | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`.                                                                                                                |
| `--splunk-platform-logs-index <index>` | Recommended. Splunk index to send logs to. If omitted, logs are routed to the default index configured on the HEC token. If no default index is configured on the token, events are silently dropped. |

> **Note:** If you omit `--splunk-platform-logs-index`, the `SPLUNK_PLATFORM_LOGS_INDEX` environment variable is not set in the collector's env file. The collector then resolves the index to an empty string and omits it from the HEC payload, leaving Splunk to route events to the HEC token's default index. If the token has no default index configured, events are rejected silently with no error visible in the collector logs. To avoid data loss, always specify `--splunk-platform-logs-index` or ensure the HEC token has a default index set in Splunk.

## What gets collected

By default, the Collector tails files under `/var/log` matching common log patterns (`*.log`, `*log`, `*auth*`, `*secure*`, `*messages*`, and others). This mirrors the behavior of the Splunk Add-on for Unix and Linux.

For the full list of included paths and all available receivers, see the [default configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/splunk\_logs\_config\_linux.yaml).

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

Search Splunk for events using the sourcetype set by the default receiver:

```
index="<your-index>" sourcetype="linux:varlog"
```

To see all sourcetypes currently being ingested:

```
index="<your-index>" sourcetype="linux:*" | stats count by sourcetype
```

## Install the Collector with metrics collection enabled

Pass your Splunk Platform HEC token, endpoint URL, and target index to the installer. Metrics collection is enabled automatically when `--splunk-platform-url` is provided:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-metrics-index "<your-index>"
```

To also send metrics and traces to Splunk Observability Cloud at the same time, include your O11y access token and realm:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm us0 \
  --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-metrics-index "<your-index>" -- YOUR_O11Y_ACCESS_TOKEN
```

## Installer options

| **Option**                                | **Description**                                                                                                                                                              |
|-------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--splunk-platform-token <token>`         | Required. HEC token for authenticating to Splunk Platform.                                                                                                                   |
| `--splunk-platform-url <url>`             | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`.                                                                                       |
| `--splunk-platform-metrics-index <index>` | Required. Set the Splunk index to send metrics to. This option enables Splunk Platform metrics collection and must be specified when configuring metrics via this installer. |

## What gets collected

By default, the Collector uses the `host_metrics` receiver to collect system metrics from the Linux host at a `10s` collection interval. The default scrapers collect CPU, disk, filesystem, memory, network, load, paging, and process count metrics.

For host metrics, the receiver is configured as follows:

| **Setting or scraper** | **Description**                                                                                                                                                                                                                                                  |
|------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `cpu`                  | [CPU utilisation and time metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/documentation.md)                                                                       |
| `disk`                 | [Disk I/O and operation metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/diskscraper/documentation.md)                                                                        |
| `filesystem`           | [Filesystem usage and capacity metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper/documentation.md)                                                           |
| `memory`               | [System memory usage metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/memoryscraper/documentation.md)                                                                         |
| `network`              | [Network interface traffic and error metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/networkscraper/documentation.md)                                                        |
| `load`                 | [System load average metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/loadscraper/documentation.md) See [Load (computing)](<https://en.wikipedia.org/wiki/Load_(computing)>). |
| `paging`               | [Paging, swap space utilisation, and I/O metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/pagingscraper/documentation.md)                                                     |
| `processes`            | [Aggregated system process count metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processesscraper/documentation.md)                                                          |
| `process`              | [System process metrics. Disabled by default.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processscraper/documentation.md)                                                        |

## Enable or disable scrapers

Edit `/etc/otel/collector/splunk_metrics_config_linux.yaml` and uncomment/comment scrapers:

```yaml
host_metrics:
  collection_interval: 10s
  scrapers:
    cpu:
    disk:
    filesystem:
    memory:
    network:
    # System load average metrics https://en.wikipedia.org/wiki/Load_(computing)
    load:
    # Paging/Swap space utilization and I/O metrics
    paging:
    # Aggregated system process count metrics
    processes:
    # System processes metrics, disabled by default
    #process:
```

To add [other scrapers available in the host_metrics receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) (such as `nfs` or `system`), add them to the `scrapers` block:

```yaml
host_metrics:
  collection_interval: 10s
  scrapers:
    cpu:
    disk:
    filesystem:
    memory:
    network:
    load:
    paging:
    processes:
    nfs:
    system:
```

After editing, restart the service:

```sh
sudo systemctl restart splunk-otel-collector
```

### Enable or disable individual metrics

Individual metrics within a scraper group can be toggled using the metric name as a key. For example, to enable `system.memory.linux.hugepages.reserved` and `system.memory.linux.hugepages.page_size`, which are [disabled by default](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/memoryscraper/documentation.md):

```yaml
host_metrics:
  collection_interval: 10s
  scrapers:
    memory:
      system.memory.linux.hugepages.reserved:
        enabled: true
      system.memory.linux.hugepages.page_size:
        enabled: true
```

To disable `system.cpu.load_average.15m`, which is [enabled by default](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/loadscraper/documentation.md):

```yaml
host_metrics:
  collection_interval: 10s
  scrapers:
    load:
      system.cpu.load_average.15m:
        enabled: false
```

## Verify metrics ingestion

Search Splunk for events using the sourcetype set by the default receiver:

```
| mpreview index="<your-index>" | search source=otel
```
