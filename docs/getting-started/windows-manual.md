# Windows Manual

The following deployment options are supported:

- [MSI](#msi-installation)
- [Docker](#docker)

## Getting Started

All installation methods offer [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which can be configured via environment variables. How these variables are
configured depends on the installation method leveraged.

### MSI Installation

A Windows MSI package (64-bit only) is available to download at
[https://github.com/signalfx/splunk-otel-collector/releases
](https://github.com/signalfx/splunk-otel-collector/releases).

The collector will be installed to
`\Program Files\Splunk\OpenTelemetry Collector`, and the
`splunk-otel-collector` service will be created but not started.

A default config file will be copied to
`\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml` if it does not
already exist.  This file is required to start the `splunk-otel-collector`
service.

### GUI

Double-click on the downloaded package and follow the wizard.

### PowerShell

In a PowerShell terminal:

```sh
PS> Start-Process -Wait msiexec "/i PATH_TO_MSI /qn"
```

Replace `PATH_TO_MSI` with the *full* path to the downloaded package, e.g.
`C:\your\download\folder\splunk-otel-collector-0.4.0-amd64.msi`.

### Configuration

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
PS> Start-Service splunk-otel-collector
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

After modifying the configuration file or registry key, apply the changes by
restarting the system or running the following PowerShell commands:

```powershell
Stop-Service splunk-otel-collector
Start-Service splunk-otel-collector
```

The collector logs and errors can be viewed in the Windows Event Viewer.

### Docker

Create build of otel collector on local system.
```bash
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector
$ docker build -t otelcol --build-arg SMART_AGENT_RELEASE=5.11.2 -f .\cmd\otelcol\Dockerfile.windows .\cmd\otelcol\
```

Deploy the latest Docker image:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0  `
	-p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060  `
	-p 8888:8888 -p 9080:9080 -p 9411:9411 -p 9943:9943 --name=otelcol quay.io/signalfx/splunk-otel-collector-windows:latest
```
### Custom Configuration

When we want to make changes in to the default configuration YAML file, for that create a
custom configuration YAML file. Then after we can use `SPLUNK_CONFIG` environment variable  or
command line argument `--config` to provide the path to this file.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 `
	-e SPLUNK_CONFIG=c:\splunk_config\gateway_config.yaml -p 13133:13133  `
	-p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060 -p 8888:8888 -p 9080:9080 `
	-p 9411:9411 -p 9943:9943 -v ${PWD}\splunk_config:c:\splunk_config:RO `
	--name otelcol quay.io/signalfx/splunk-otel-collector-windows:latest
```
> For mounting configuration files on a windows container, we have to specify a directory name in which the configuration file is present. because just like Linux containers we can not mount files to containers.
