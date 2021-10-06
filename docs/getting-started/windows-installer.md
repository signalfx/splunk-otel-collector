> The official Splunk documentation for this page is [Install on Windows](https://docs.splunk.com/Observability/gdi/opentelemetry/install-windows.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

# Windows Installer Script

For Windows 64-bit environments, an installer script is available. The
script deploys and configures:

- Splunk OpenTelemetry Connector for Windows
- [Fluentd (via the TD Agent)](https://www.fluentd.org/)

Currently, the following Windows versions are supported and requires PowerShell
3.0 or newer:

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit

## Getting Started

Run the below PowerShell command on your host. Replace these variables:

- `SPLUNK_REALM`: Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN`: Access token to authenticate requests

```powershell
& {Set-ExecutionPolicy Bypass -Scope Process -Force; $script = ((New-Object System.Net.WebClient).DownloadString('https://dl.signalfx.com/splunk-otel-collector.ps1')); $params = @{access_token = "SPLUNK_ACCESS_TOKEN"; realm = "SPLUNK_REALM"}; Invoke-Command -ScriptBlock ([scriptblock]::Create(". {$script} $(&{$args} @params)"))}
```

You can view the [source](../../internal/buildscripts/packaging/installer/install.ps1)
for more details and available options.

## Advanced Configuration

### Additional Script Options

One additional parameter that may need to changed is `-memory` in order to
configure the memory allocation.

> By default, this variable is set to `512`. If you have allocated more memory
> to the Collector then you must increase this setting.

Replace the `SPLUNK_MEMORY_TOTAL_MIB` variable with the desired value.

```powershell
& {Set-ExecutionPolicy Bypass -Scope Process -Force; $script = ((New-Object System.Net.WebClient).DownloadString('https://dl.signalfx.com/splunk-otel-collector.ps1')); $params = @{access_token = "SPLUNK_ACCESS_TOKEN"; realm = "SPLUNK_REALM"; memory = "SPLUNK_MEMORY_TOTAL_MIB"}; Invoke-Command -ScriptBlock ([scriptblock]::Create(". {$script} $(&{$args} @params)"))}
```

#### Custom MSI URLs

By default, the Collector MSI is downloaded from `https://dl.signalfx.com` and
the Fluentd MSI is downloaded from `https://packages.treasuredata.com`.  To
specify custom URLs for these downloads, use the `collector_msi_url` and
`fluentd_msi_url` options.

Replace `COLLECTOR_MSI_URL` and `FLUENTD_MSI_URL` with the URLs to the
desired MSI packages to install, e.g.
`https://my.host/splunk-otel-collector-1.2.3-amd64.msi` and
`https://my.host/td-agent-4.1.0-x64.msi`.

```powershell
& {Set-ExecutionPolicy Bypass -Scope Process -Force; $script = ((New-Object System.Net.WebClient).DownloadString('https://dl.signalfx.com/splunk-otel-collector.ps1')); $params = @{access_token = "SPLUNK_ACCESS_TOKEN"; realm = "SPLUNK_REALM"; collector_msi_url = "COLLECTOR_MSI_URL"; fluentd_msi_url = "FLUENTD_MSI_URL"}; Invoke-Command -ScriptBlock ([scriptblock]::Create(". {$script} $(&{$args} @params)"))}
```

### Collector Configuration

The Collector comes with a default configuration which can be found at
`\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`. This configuration
can be modified as needed. Possible configuration options can be found in the
`receivers`, `processors`, `exporters`, and `extensions` folders of either:

- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector)
- [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)

Based on the specified installation parameters, the following environment
variables will be saved to the
`HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment` registry
key and passed to the Collector service:

- `SPLUNK_ACCESS_TOKEN`: The Splunk access token to authenticate requests
- `SPLUNK_API_URL`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`
- `SPLUNK_BUNDLE_DIR`: The location of your Smart Agent bundle for monitor functionality, e.g. `C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle`
- `SPLUNK_CONFIG`: The path to the collector config file, e.g. `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`
- `SPLUNK_HEC_TOKEN`: The Splunk HEC authentication token
- `SPLUNK_HEC_URL`: The Splunk HEC endpoint URL, e.g. `https://ingest.us0.signalfx.com/v1/log`
- `SPLUNK_INGEST_URL`: The Splunk ingest URL, e.g. `https://ingest.us0.signalfx.com`
- `SPLUNK_MEMORY_TOTAL_MIB`: Total memory in MiB allocated to the collector, e.g. `512`
- `SPLUNK_REALM`: The Splunk realm to send the data to, e.g. `us0`
- `SPLUNK_TRACE_URL`: The Splunk trace endpoint URL, e.g. `https://ingest.us0.signalfx.com/v2/trace`

To modify these values, run `regdit` and browse to the path, or run the
following PowerShell command (replace `ENV_VAR` and `VALUE` for the desired
environment variable and value):

```powershell
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "ENV_VAR" -value "VALUE"
```

After modifying the configuration file or registry key, apply the changes by
restarting the system or running the following PowerShell commands:

```powershell
Stop-Service splunk-otel-collector
Start-Service splunk-otel-collector
```

### Fluentd Configuration

By default, the fluentd service will be installed and configured to forward log
events with the `@SPLUNK` label to the collector (see below for how to add
custom fluentd log sources), and the collector will send these events to the
HEC ingest endpoint determined by the `-realm SPLUNK_REALM` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.

To configure the collector to send log events to a custom HEC endpoint URL, you
can specify the following parameters for the installer script:

- `-hec_url URL`
- `-hec_token TOKEN`

The main fluentd configuration file will be installed to
`\opt\td-agent\etc\td-agent\td-agent.conf`. Custom fluentd source config files
can be added to the `\opt\td-agent\etc\td-agent\conf.d` directory after 
installation. Please note:

- All files in this directory ending `.conf` extension will automatically be
  included by Fluentd.
- By default, Fluentd will be configured to collect from the Windows Event Log.
  See `\opt\td-agent\etc\td-agent\conf.d\eventlog.conf` for the default
  configuration.

After any configuration modification, apply the changes by restarting the
system or running the following PowerShell commands:

```powershell
Stop-Service fluentdwinsvc
Start-Service fluentdwinsvc
```
