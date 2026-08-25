# Collect logs on Windows

> **Scope:** This guide covers log collection via the [Windows installer script](windows-installer.md) only.

## Prerequisites

- A HEC token with write permissions to the target index. See <https://docs.splunk.com/Documentation/Splunk/latest/Data/UsetheHTTPEventCollector>.
- The Splunk Distribution of OpenTelemetry Collector installed on a Windows host

## Install the Collector with log collection enabled

Pass your HEC token, endpoint URL, and target index to the installer. Log collection is enabled automatically when `SPLUNK_PLATFORM_URL` and `SPLUNK_PLATFORM_TOKEN` is provided:

```Powershell
& {
  Set-ExecutionPolicy Bypass -Scope Process -Force
  $script = (New-Object System.Net.WebClient).DownloadString('https://dl.observability.splunkcloud.com/splunk-otel-collector.ps1')

  $params = @{
    msi_public_properties = "SPLUNK_PLATFORM_URL=<URL> SPLUNK_PLATFORM_TOKEN=<TOKEN> SPLUNK_PLATFORM_LOGS_INDEX=<INDEX>"
  }
  & ([scriptblock]::Create($script)) @params
}
```

To also send metrics and traces to Splunk Observability Cloud at the same time, include your O11y access token and realm:

```Powershell
& {
  Set-ExecutionPolicy Bypass -Scope Process -Force
  $script = (New-Object System.Net.WebClient)
    .DownloadString('https://dl.observability.splunkcloud.com/splunk-otel-collector.ps1')
  $params = @{
    access_token = "<ACCESS_TOKEN>"
    realm = "<SPLUNK_REALM>"
    msi_public_properties = "SPLUNK_PLATFORM_URL=<URL> SPLUNK_PLATFORM_TOKEN=<TOKEN> SPLUNK_PLATFORM_LOGS_INDEX=<LOGS_INDEX>"
  }
  & ([scriptblock]::Create($script)) @params
}
```

## Installer options

| **Option**                           | **Description**                                                                        |
|--------------------------------------|----------------------------------------------------------------------------------------|
| `SPLUNK_PLATFORM_URL <url>`          | Required. HEC endpoint URL, e.g. `https://splunk.example.com:8088/services/collector`. |
| `SPLUNK_PLATFORM_TOKEN <token>`      | Required. HEC token for authenticating to the HEC endpoint.                            |
| `SPLUNK_PLATFORM_LOGS_INDEX <index>` | Recommended. Index to send logs to.                                                    |

> **Note:** If you omit `SPLUNK_PLATFORM_LOGS_INDEX`, the `SPLUNK_PLATFORM_LOGS_INDEX` variable is not set in the collector's ENVIRONMENT registry entry. The collector then resolves the index to an empty string and omits it from the HEC payload, leaving Splunk to route events to the HEC token's default index. If the token has no default index configured, events are rejected silently with no error visible in the collector logs. To avoid data loss, always specify `SPLUNK_PLATFORM_LOGS_INDEX` or ensure the HEC token has a default index set in Splunk.

## What gets collected

By default, the collector tails files from following location patterns

- C:\Windows\System32\DHCP\DhcpSrvLog*
- C:\Windows\WindowsUpdate.log
- C:\Windows\debug\netlogon.log
- C:\Windows\System32\LogFiles\Firewall\*

and monitors following Windows Event Log channels:

- Application
- Security
- System
- Microsoft-Windows-PrintService/Operational
- ForwardedEvents
- DFS Replication
- Directory Service
- File Replication Service
- DNS Server
- Key Management Service
- Microsoft-Windows-Windows Defender/Operational

This mirrors the behavior of the Splunk Add-on for Windows.

For the full list of included paths and all available receivers, see the [default configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/splunk_logs_config_windows.yaml).


## Enable or disable receivers

Edit `C:\ProgramData\Splunk\OpenTelemetry Collector\splunk_logs_config_windows.yaml` and uncomment receivers in the `service.pipelines.logs` section:

```yaml
service:
  pipelines:
    logs/application:
      receivers: [windows_event_log]
      processors: [transform/application, resource_detection]
      exporters: [splunk_hec/logs]
    logs/security:
      receivers: [windows_event_log/security]
      processors: [transform/security, resource_detection]
      exporters: [splunk_hec/logs]
```

After editing, restart the service:

```Powershell
Restart-Service splunk-otel-collector
```

## Add additional log files

Every `file_log` receiver has an `include` list of glob patterns. To collect logs from a custom application, add its path to the `file_log/dhcp` receiver:

```yaml
file_log/dhcp:
  include:
    - '${env:SystemRoot}\System32\DHCP\DhcpSrvLog*'
    - '${env:SystemRoot}\SomeOtherDirectory\DHCP*.log'
```

To exclude specific files within a matched pattern:

```yaml
file_log/firewall:
  include:
    - '${env:SystemRoot}\System32\LogFiles\Firewall\*'
  exclude:
    - '${env:SystemRoot}\System32\LogFiles\Firewall\sensitive_data.log'
```

See the [filelog receiver documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) for the full glob syntax and all available options.

## Verify log ingestion

Search for events using the sourcetype set by the default receiver:

```
index="<your-index>" sourcetype="XmlWinEventLog"
```

To see all sourcetypes currently being ingested:

```
index="<your-index>" | stats count by sourcetype
```