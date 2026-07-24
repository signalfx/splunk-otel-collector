# Collect host metrics for Splunk Platform on Linux

> **Note:** For most users, sending metrics to [Splunk Observability Cloud](https://www.splunk.com/en_us/products/observability.html) via the standard installer is the recommended approach. This guide covers the alternative: forwarding host metrics directly to Splunk Enterprise or Splunk Cloud via HEC.

## Prerequisites

- A running Splunk Platform instance (Splunk Enterprise or Splunk Cloud) with a metrics index
- A HEC token with write permissions to the target metrics index. See <https://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector>.
- The Splunk Distribution of OpenTelemetry Collector installed on a Linux host

## Install the Collector with metrics collection enabled

Pass your HEC token, endpoint URL, and target metrics index to the installer:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-metrics-index "<your-metrics-index>"
```

To also send metrics and traces to Splunk Observability Cloud at the same time, include your O11y access token and realm:

```sh
curl -sSL https://dl.observability.splunkcloud.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm us0 \
  --splunk-platform-token "<your-hec-token>" \
  --splunk-platform-url   "https://<your-splunk-host>:8088/services/collector" \
  --splunk-platform-metrics-index "<your-metrics-index>" \
  -- YOUR_O11Y_ACCESS_TOKEN
```

## Installer options

| **Option**                                | **Description**                                                                                                                                                              |
|-------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--splunk-platform-token <token>`         | Required. HEC token for authenticating to the HEC endpoint.                                                                                                                  |
| `--splunk-platform-url <url>`             | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`.                                                                                       |
| `--splunk-platform-metrics-index <index>` | Required. Splunk metrics index to send host metrics to. Enables Splunk Platform metrics collection.                                                                          |

## What gets collected

By default, the Collector uses the `host_metrics` receiver to collect system metrics from the Linux host at a `10s` collection interval. The default scrapers collect CPU, disk, filesystem, memory, network, load, paging, and process count metrics.

| **Scraper**  | **Description**                                                                                                                                                                                                                                                  |
|--------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `cpu`        | [CPU utilisation and time metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/documentation.md)                                                                       |
| `disk`       | [Disk I/O and operation metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/diskscraper/documentation.md)                                                                        |
| `filesystem` | [Filesystem usage and capacity metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper/documentation.md)                                                           |
| `memory`     | [System memory usage metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/memoryscraper/documentation.md)                                                                         |
| `network`    | [Network interface traffic and error metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/networkscraper/documentation.md)                                                        |
| `load`       | [System load average metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/loadscraper/documentation.md) See [Load (computing)](https://en.wikipedia.org/wiki/Load_(computing)).   |
| `paging`     | [Paging, swap space utilisation, and I/O metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/pagingscraper/documentation.md)                                                     |
| `processes`  | [Aggregated system process count metrics.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processesscraper/documentation.md)                                                          |
| `process`    | [Per-process metrics. Disabled by default.](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/hostmetricsreceiver/internal/scraper/processscraper/documentation.md)                                                           |

## Enable or disable scrapers

Edit `/etc/otel/collector/splunk_metrics_config_linux.yaml` and comment or uncomment scrapers:

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

```
| mpreview index="<your-metrics-index>" | search source=otel
```