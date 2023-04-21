> The official Splunk documentation for this page is [Install on Windows](https://docs.splunk.com/Observability/gdi/opentelemetry/install-windows.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

# Windows Installer Script

For Windows 64-bit environments, an installer script is available. The
script deploys and configures:

- Splunk OpenTelemetry Collector for Windows
- [Fluentd (via the TD Agent)](https://www.fluentd.org/)

Currently, the following Windows versions are supported and requires PowerShell
3.0 or newer:

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

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

For general advanced configuration and usage see the [manual installation documentation](./windows-manual.md). Advanced installer script usage is detailed in this section.

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
- `SPLUNK_CONFIG`: The path to the Collector config file, e.g. `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`
- `SPLUNK_HEC_TOKEN`: The Splunk HEC authentication token
- `SPLUNK_HEC_URL`: The Splunk HEC endpoint URL, e.g. `https://ingest.us0.signalfx.com/v1/log`
- `SPLUNK_INGEST_URL`: The Splunk ingest URL, e.g. `https://ingest.us0.signalfx.com`
- `SPLUNK_MEMORY_TOTAL_MIB`: Total memory in MiB allocated to the Collector, e.g. `512`
- `SPLUNK_REALM`: The Splunk realm to send the data to, e.g. `us0`
- `SPLUNK_TRACE_URL`: The Splunk trace endpoint URL, e.g. `https://ingest.us0.signalfx.com/v2/trace`

To modify these values, run `regdit` and browse to the path, or run the
following PowerShell command (replace `ENV_VAR` and `VALUE` for the desired
environment variable and value):

```powershell
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "ENV_VAR" -value "VALUE"
```

To add or remove command line options for the `splunk-otel-collector` service,
run `regedit` and modify the `ImagePath` value in the
`HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector` registry key,
or run the following PowerShell command (replace `OPTIONS` with the desired
command line options):

```powershell
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -name "ImagePath" -value "C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe OPTIONS"
```

For example, to change the default exposed metrics address of the Collector to
`0.0.0.0:9090`, run the following PowerShell command:

```powershell
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -name "ImagePath" -value "C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe --metrics-addr 0.0.0.0:9090"
```

To see all available command line options, run the following PowerShell
command:

```powershell
& 'C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe' --help
```

After modifying the configuration file or registry key, apply the changes by
restarting the system or running the following PowerShell commands:

```powershell
Stop-Service splunk-otel-collector
Start-Service splunk-otel-collector
```

### Command Line Operation

After successful installation with the installer script, the Collector is
automatically started as a Windows service.  However, if you need to start the
Collector manually via the command line, run the following PowerShell commands
as `Administrator`:

1. Stop the Collector service to avoid conflicts:
   ```powershell
   Stop-Service splunk-otel-collector
   ```
1. Check the available command line options:
   ```powershell
   & 'C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe' --help
   ```
1. Environment variables are automatically defined within the Windows registry
   by the installer script based on the specified and default installation
   parameter values. These environment variables are utilized by the Collector
   and configuration files. To check these environment variables, run:
   ```powershell
   dir env: | Select-Object -Property Name,Value | Where-object -Property Name -Like "SPLUNK_*"
   ```
   To temporarily change any of these environment variables within the current
   PowerShell session, run the following command for the desired variable and
   value.  For example, the `SPLUNK_CONFIG` environment variable can
   be changed to a new configuration file path with:
   ```powershell
   $env:SPLUNK_CONFIG = "C:\my\custom\config.yaml"
   ```
   To permanently change these environment variables in the Windows registry,
   see [Collector Configuration](#collector-configuration) and restart the
   system for the changes to take effect.
1. Start the Collector:
   - With the configuration file defined by the `SPLUNK_CONFIG` environment
     variable:
     ```powershell
     & 'C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe'
     ```
   - With a custom configuration file via command line:
     ```powershell
     & 'C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe' --config 'C:\my\custom\config.yaml'
     ```
1. Press `Ctrl-C` to stop the Collector.

### Fluentd Configuration

By default, the Fluentd service will be installed and configured to forward log
events with the `@SPLUNK` label to the Collector (see below for how to add
custom Fluentd log sources), and the Collector will send these events to the
HEC ingest endpoint determined by the `realm = <SPLUNK_REALM>` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.

To configure the Collector to send log events to a custom HEC endpoint URL, you
can specify the following parameters for the installer script:

- `hec_url = <SPLUNK_HEC_URL>`
- `hec_token = <SPLUNK_HEC_TOKEN>`

For example (replace the `<SPLUNK...>` values for your configuration):
```powershell
& {Set-ExecutionPolicy Bypass -Scope Process -Force; $script = ((New-Object System.Net.WebClient).DownloadString('https://dl.signalfx.com/splunk-otel-collector.ps1')); $params = @{access_token = "<SPLUNK_ACCESS_TOKEN>"; realm = "<SPLUNK_REALM>"; hec_url = "<SPLUNK_HEC_URL>"; hec_token = "<SPLUNK_HEC_TOKEN>"}; Invoke-Command -ScriptBlock ([scriptblock]::Create(". {$script} $(&{$args} @params)"))}
```

The main Fluentd configuration file will be installed to
`\opt\td-agent\etc\td-agent\td-agent.conf`. Custom Fluentd source config files
can be added to the `\opt\td-agent\etc\td-agent\conf.d` directory after 
installation. Please note:

- By default, Fluentd will be configured to collect from the Windows Event Log.
  See `\opt\td-agent\etc\td-agent\conf.d\eventlog.conf` for the default
  configuration.
- Any new source added to this directory should have a `.conf` extension and
  have the `@SPLUNK` label to automatically forward log events to the
  Collector.
- All files with a `.conf` extension in this directory will automatically be
  included when the Fluentd service starts/restarts.
- After any configuration modification, apply the changes by restarting the
  system or running the following PowerShell commands:
  ```powershell
  Stop-Service fluentdwinsvc
  Start-Service fluentdwinsvc
  ```
- The Fluentd service logs and errors can be viewed in
  `\opt\td-agent\td-agent.log`.
- See [https://docs.fluentd.org/configuration](
  https://docs.fluentd.org/configuration) for general Fluentd configuration
  details.
