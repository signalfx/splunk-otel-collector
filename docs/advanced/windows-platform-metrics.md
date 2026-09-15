# Collect host metrics for Splunk Platform on Windows

> **Note:** For most users, sending metrics to [Splunk Observability Cloud](https://www.splunk.com/en_us/products/observability.html) via the standard installer is the recommended approach. This guide covers the alternative: forwarding host metrics directly to Splunk Enterprise or Splunk Cloud via HEC.

## Prerequisites

- A running Splunk Platform instance (Splunk Enterprise or Splunk Cloud) with a metrics index
- A HEC token with write permissions to the target metrics index. See <https://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector>.
- The Splunk Distribution of OpenTelemetry Collector installed on a Windows host

## Install the Collector with metrics collection enabled

Pass your HEC token, endpoint URL, and target metrics index to the installer:

```Powershell
& {
  Set-ExecutionPolicy Bypass -Scope Process -Force
  $script = (New-Object System.Net.WebClient).DownloadString('https://dl.observability.splunkcloud.com/splunk-otel-collector.ps1')
  $params = @{
    msi_public_properties = "SPLUNK_PLATFORM_URL=<URL> SPLUNK_PLATFORM_TOKEN=<TOKEN> SPLUNK_PLATFORM_METRICS_INDEX=<INDEX>"
  }
  & ([scriptblock]::Create($script)) @params
}
```

To also send metrics and traces to Splunk Observability Cloud at the same time, include your O11y access token and realm:

```Powershell
& {
  Set-ExecutionPolicy Bypass -Scope Process -Force
  $script = (New-Object System.Net.WebClient).DownloadString('https://dl.observability.splunkcloud.com/splunk-otel-collector.ps1')
  $params = @{
    access_token = "<ACCESS_TOKEN>"
    realm = "<SPLUNK_REALM>"
    msi_public_properties = "SPLUNK_PLATFORM_URL=<URL> SPLUNK_PLATFORM_TOKEN=<TOKEN> SPLUNK_PLATFORM_METRICS_INDEX=<METRICS_INDEX>"
  }
  & ([scriptblock]::Create($script)) @params
}
```

## Installer options

| **Option**                              | **Description**                                                                                     |
|-----------------------------------------|-----------------------------------------------------------------------------------------------------|
| `SPLUNK_PLATFORM_URL <token>`           | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`.              |
| `SPLUNK_PLATFORM_TOKEN <url>`           | Required. HEC token for authenticating to the HEC endpoint.                                         |
| `SPLUNK_PLATFORM_METRICS_INDEX <index>` | Required. Splunk metrics index to send host metrics to. Enables Splunk Platform metrics collection. |

## What gets collected

By default, the Collector uses the `windows_perf_counters` receiver to collect system metrics from the Windows host at `10s` or `30s` collection interval. Metrics collected by default come from following objects:

- Processor
- Processor Information
- LogicalDisk
- PhysicalDisk
- Memory
- Network Interface
- Process
- System
- DFS Replicated Folders
- NTDS
- DNS

## Enable or disable scrapers

Edit `C:\ProgramData\Splunk\OpenTelemetry Collector\splunk_metrics_config_windows.yaml` and comment/add counters:

```yaml
  windows_perf_counters/cpu:
    collection_interval: 10s
    perfcounters:
      - object: Processor
        instances: ["*"]
        counters:
          - name: "% Processor Time"
          - name: "% User Time"
          - name: "% Privileged Time"
          - name: "Interrupts/sec"
          - name: "% DPC Time"
          - name: "% Interrupt Time"
```

After editing, restart the service:

```Powershell
Restart-Service splunk-otel-collector
```

See the [Windows perfcounters receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowsperfcountersreceiver) for the full glob syntax and all available options.

## Verify metrics ingestion

```
| mpreview index="<your-index>"
```