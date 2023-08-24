> The official Splunk documentation for this page is [Install on Windows](https://docs.splunk.com/Observability/gdi/opentelemetry/install-windows.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

# Windows Manual

The following deployment options are supported:

- [Collector MSI](#collector-msi-installation)
- [Fluentd MSI](#fluentd-msi-installation) for log collection
- [Chocolatey](#chocolatey-installation)
- [Docker](#docker)

## Getting Started

All installation methods offer [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which can be configured via environment variables. How these variables are
configured depends on the installation method leveraged.

### Collector MSI Installation

A Windows MSI package (64-bit only) is available to download at
[https://github.com/signalfx/splunk-otel-collector/releases
](https://github.com/signalfx/splunk-otel-collector/releases).

The Collector will be installed to
`\Program Files\Splunk\OpenTelemetry Collector`, and the
`splunk-otel-collector` service will be created but not started.

A default config file will be copied to
`\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml` if it does not
already exist.  This file is required to start the `splunk-otel-collector`
service.

#### GUI

Double-click on the downloaded package and follow the wizard.

#### PowerShell

In a PowerShell terminal:

```sh
Start-Process -Wait msiexec "/i PATH_TO_MSI /qn"
```

Replace `PATH_TO_MSI` with the *full* path to the downloaded package, e.g.
`C:\your\download\folder\splunk-otel-collector-0.4.0-amd64.msi`.

#### Configuration

Before starting the `splunk-otel-collector` service, the following variables
in the default config file need to be replaced by the appropriate values for
your environment:

- `${SPLUNK_ACCESS_TOKEN}`: The Splunk access token to authenticate requests
- `${SPLUNK_API_URL}`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`
- `${SPLUNK_HEC_TOKEN}`: The Splunk HEC authentication token
- `${SPLUNK_HEC_URL}`: The Splunk HEC endpoint URL, e.g. `https://ingest.us0.signalfx.com/v1/log`
- `${SPLUNK_INGEST_URL}`: The Splunk ingest URL, e.g. `https://ingest.us0.signalfx.com`
- `${SPLUNK_TRACE_URL}`: The Splunk trace endpoint URL, e.g. `https://ingest.us0.signalfx.com/v2/trace`
- `${SPLUNK_BUNDLE_DIR}`: The location of your Smart Agent bundle for monitor functionality, e.g. `C:\Program Files\Splunk\OpenTelemetry Collector\agent-bundle`

After updating all variables in the config file, start the
`splunk-otel-collector` service by rebooting the system or running the
following command in a PowerShell terminal:

```sh
Start-Service splunk-otel-collector
```

To modify the default path to the configuration file for the
`splunk-otel-collector` service, run `regdit` and modify the `SPLUNK_CONFIG`
value in the
`HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
registry key, or run the following PowerShell command (replace `PATH` with the
full path to the new configuration file):

```powershell
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" -name "SPLUNK_CONFIG" -value "PATH"
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

#### Service Logging

The Collector logs and errors can be viewed in the Windows Event Viewer when run as a service. The service logs are
displayed in `Event Viewer` > `Windows Logs` > `Application` and are visible by source "collector".

The Collector will have a log level of info by default.  To change this in versions 0.36.0 and later
you can add a "logs" entry in the [service telemetry definition](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/troubleshooting.md#logs) in the currently used config before restarting the service:

```yaml
service:
  telemetry:
    logs:
      level: debug
```

For older versions of the Collector you can alter the service `ImagePath` before restarting:

```sh
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -name "ImagePath" -value "C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe --set=service.telemetry.logs.level=debug"
Restart-Service splunk-otel-collector

# Reverting after observing logs:
Set-ItemProperty -path "HKLM:\SYSTEM\CurrentControlSet\Services\splunk-otel-collector" -name "ImagePath" -value "C:\Program Files\Splunk\OpenTelemetry Collector\otelcol.exe"
Restart-Service splunk-otel-collector
```

### Fluentd MSI Installation

If log collection is required, perform the following steps to install Fluentd
and forward collected log events to the Collector (requires `Administrator`
privileges):

1. Install, configure, and start the Collector as described in the previous
   section.  The Collector's default configuration file
   (`\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`) listens
   for log events on `127.0.0.1:8006` and sends them to the Splunk
   Observability Cloud.
1. Check [https://docs.fluentd.org/installation/install-by-msi#td-agent-v4](
   https://docs.fluentd.org/installation/install-by-msi#td-agent-v4) to install
   the Fluentd MSI.  Requires version 4.0 or newer.
1. Configure Fluentd to collect log events and forward them to the Collector:
   - Option 1: Update the default config file provided by the Fluentd MSI at
     `\opt\td-agent\etc\td-agent\td-agent.conf` to collect the desired log
     events and [forward](https://docs.fluentd.org/output/forward) them to
     `127.0.0.1:8006`.
   - Option 2: The installed Collector package provides a custom Fluentd config
     file
     (`\Program Files\Splunk\OpenTelemetry Collector\fluentd\td-agent.conf`) to
     collect log events from the Windows Event Log
     (`\Program Files\Splunk\OpenTelemetry Collector\fluentd\conf.d\eventlog.conf`)
     and forwards them to `127.0.0.1:8006`.  To utilize these files, backup any
     files as necessary in the `\opt\td-agent\etc\td-agent\` directory, and
     copy the contents from the
     `\Program Files\Splunk\OpenTelemetry Collector\fluentd\` directory to the
     `\opt\td-agent\etc\td-agent\` directory.
1. Apply the changes by restarting the system or by running the following
   Powershell commands to restart the Fluentd service:
   ```sh
   Stop-Service fluentdwinsvc
   Start-Service fluentdwinsvc
   ```
   **Note**: The `fluentdwinsvc` service must be restarted in order for any
   changes made to the Fluentd config files to take effect.
1. The Fluentd service logs and errors can be viewed in
   `\opt\td-agent\td-agent.log`.
1. See [https://docs.fluentd.org/configuration](
   https://docs.fluentd.org/configuration) for general Fluentd configuration
   details.

### Chocolatey Installation

A [Chocolatey package](
https://community.chocolatey.org/packages/splunk-otel-collector) is available
to download, install, and configure the Collector with the following PowerShell
command (replace `MY_SPLUNK_ACCESS_TOKEN` and `MY_SPLUNK_REALM`):

```powershell
choco install splunk-otel-collector --params="'/SPLUNK_ACCESS_TOKEN:MY_SPLUNK_ACCESS_TOKEN /SPLUNK_REALM:MY_SPLUNK_REALM'"
```

#### Package Parameters

The following package parameters are available:

 * `/SPLUNK_ACCESS_TOKEN`: The Splunk access token (org token) used to send data to Splunk Observability Cloud. This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_ACCESS_TOKEN` registry value.
 * `/SPLUNK_REALM`: The Splunk realm to send the data to. This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_REALM` registry value. Default value is `us0`.
 * `/SPLUNK_INGEST_URL:`: URL of the Splunk ingest endpoint (e.g. `https://ingest.us1.signalfx.com`). This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_INGEST_URL` registry value. Default value is `https://ingest.$SPLUNK_REALM.signalfx.com`.
 * `/SPLUNK_API_URL`: URL of the Splunk API endpoint (e.g. `https://api.us1.signalfx.com`). This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_API_URL` registry value. Default value is `https://api.$SPLUNK_REALM.signalfx.com`.
 * `/SPLUNK_HEC_TOKEN`: The Splunk HEC authentication token. This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_HEC_TOKEN` registry value. Default value is the same as `SPLUNK_ACCESS_TOKEN`.
 * `/SPLUNK_HEC_URL`: URL of the Splunk HEC endpoint (e.g. `https://ingest.us1.signalfx.com/v1/log`). This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_HEC_URL` registry value. Default value is `https://ingest.$SPLUNK_REALM.signalfx.com/v1/log`
 * `/SPLUNK_TRACE_URL`: URL of the Splunk trace endpoint (e.g. `https://ingest.us1.signalfx.com/v2/trace`). This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_TRACE_URL` registry value. Default value is `https://ingest.$SPLUNK_REALM.signalfx.com/v2/trace`
 * `/SPLUNK_BUNDLE_DIR`: The path to the Smart Agent bundle directory for the `smartagent` receiver and extension. The default path is provided by the Collector package. If the specified path is changed from the default value, the path should be an existing directory on the system. This parameter is saved to the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_BUNDLE_DIR` registry value. Default value is `\Program Files\Splunk\OpenTelemetry Collector\agent-bundle`.
 * `/MODE`: This parameter is used for setting the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_CONFIG` registry value to `\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml` or `\ProgramData\Splunk\OpenTelemetry Collector\gateway_config.yaml`. Possible values are `agent` and `gateway`. Default value is `agent`.
 * `/WITH_FLUENTD`: Whether to download, install, and configure Fluentd to collect and forward log events to the Collector. Possible values are `true` and `false`. If set to `true`, the Fluentd MSI package will be downloaded from `https://packages.treasuredata.com`. Default value is `false`.

To pass parameters, use `--params "''"` (e.g. `choco install splunk-otel-collector --params="'/SPLUNK_ACCESS_TOKEN:MY_SPLUNK_ACCESS_TOKEN /SPLUNK_REALM:MY_SPLUNK_REALM'"`).

To have choco remember parameters on upgrade, be sure to set `choco feature enable -n=useRememberedArgumentsForUpgrades` before installation.

#### Notes

 * If the `SPLUNK_ACCESS_TOKEN` parameter is not specified on initial installation, the Collector service will not be started. In order to start the Collector service, manually create/set the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_ACCESS_TOKEN` registry value to the Splunk access token and run the `Start-Service splunk-otel-collector` PowerShell command.
 * If the Collector configuration file or any of the `HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment\SPLUNK_*` registry values are modified after installation, restart the Collector service by restarting the system or by running the `Restart-Service splunk-otel-collector` PowerShell command.
 * If the `WITH_FLUENTD` parameter is set to `true` and the `\opt\td-agent\etc\td-agent\td-agent.conf` Fluentd configuration file does not exist, this file will be created and customized to collect events from the Windows Event Log and forward the collected events to the Collector. If this file is modified after installation, restart the Fluentd service by restarting the system or by running the `Restart-Service fluentdwinsvc` PowerShell command.
 * See [Collector Configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/windows-installer.md#collector-configuration) for additional configuration details.

### Docker

Deploy the latest Docker image:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0  `
	-p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060  `
	-p 8888:8888 -p 9080:9080 -p 9411:9411 -p 9943:9943 `
	--name=otelcol quay.io/signalfx/splunk-otel-collector-windows:latest
```
### Custom Configuration

If using a custom configuration file, you will need to mount the directory containing the file and either use the `SPLUNK_CONFIG=<path>` environment variable or the `--config=<path>` command line argument (replace `<path>` with the path to the custom file within the container).

Example with `SPLUNK_CONFIG`:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 `
	-e SPLUNK_CONFIG=c:\splunk_config\gateway_config.yaml -p 13133:13133  `
	-p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060 -p 8888:8888 -p 9080:9080 `
	-p 9411:9411 -p 9943:9943 -v ${PWD}\splunk_config:c:\splunk_config:RO `
	--name otelcol quay.io/signalfx/splunk-otel-collector-windows:latest
```

Example with `--config`:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 `
    -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060 `
    -p 8888:8888 -p 9080:9080 -p 9411:9411 -p 9943:9943 `
    -v ${PWD}\splunk_config:c:\splunk_config:RO `
    --name otelcol quay.io/signalfx/splunk-otel-collector-windows:latest `
    --config c:\splunk_config\gateway_config.yaml 
```

> For mounting configuration files on a windows container, we have to specify a directory name in which the configuration file is present. because just like Linux containers we can not mount files to containers.
