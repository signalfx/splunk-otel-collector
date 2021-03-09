# Windows Standalone

A Windows MSI package (64-bit only) is available to download at
[https://github.com/signalfx/splunk-otel-collector/releases
](https://github.com/signalfx/splunk-otel-collector/releases).

## Installation

The collector will be installed to
`\Program Files\Splunk\OpenTelemetry Collector`, and the
`splunk-otel-collector` service will be created but not started.

A default config file will be copied to
`\ProgramData\Splunk\OpenTelemetry Collector\config.yaml` if it does not
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

## Configuration

Before starting the `splunk-otel-collector` service, the following variables
in the default config file need to be replaced by the appropriate values for
your environment:

- `${SPLUNK_ACCESS_TOKEN}`: The Splunk access token to authenticate requests
- `${SPLUNK_API_URL}`: The Splunk API URL, e.g. `https://api.us0.signalfx.com`
- `${SPLUNK_HEC_TOKEN}`: The Splunk HEC authentication token
- `${SPLUNK_HEC_URL}`: The Splunk HEC endpoint URL, e.g. `https://ingest.us0.signalfx.com/v1/log`
- `${SPLUNK_INGEST_URL}`: The Splunk ingest URL, e.g. `https://ingest.us0.signalfx.com`
- `${SPLUNK_TRACE_URL}`: The Splunk trace endpoint URL, e.g. `https://ingest.us0.signalfx.com/v2/trace`

After updating all variables in the config file, start the
`splunk-otel-collector` service by rebooting the system or running the
following command in a PowerShell terminal:

```sh
PS> Start-Service splunk-otel-collector
```

The collector logs and errors can be viewed in the Windows Event Viewer.
