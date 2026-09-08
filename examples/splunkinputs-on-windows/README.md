# Example of deployment with a Splunk Universal Forwarder add-on, on Windows

This is the Windows counterpart to the [`splunkinputs`](../splunkinputs/README.md)
example. It uses the [`splunk_inputs`](https://github.com/splunk/tarunner/tree/main/pkg/splunkinputsreceiver)
receiver to read a TA's `inputs.conf`, `transforms.conf`, and `props.conf`
directly, emulating its modular input(s) without running a real `splunkd`.
Like `splunkinputs`'s `/var/ta`, `C:\var\ta` is a generic mount point for any
TA; this example just happens to use the
[Splunk Add-on for Microsoft Windows](https://splunkbase.splunk.com/app/742)
as its example TA.

Docker Compose does not support Windows containers, so this example follows
a single-container pattern instead.

Because this example runs standalone, it has no local Splunk instance to send
data to like `splunkinputs` does. You must provide the HEC endpoint URL and
token of a Splunk instance (Splunk Cloud, Splunk Enterprise, or Splunk
Observability Cloud) that will receive the collected events; there is no way
for this example to know that in advance.

## Download a TA

`run-example.ps1` (below) takes the path to a TA package, extracts it to
`C:\var\ta` in the container, and enables its inputs for you. This example
uses the Splunk Add-on for Microsoft Windows, but any TA compatible with the
`splunk_inputs` receiver can be used instead. The receiver only reads a
single TA per configured path, so only one TA package can be used at a time.

By default every stanza in the TA's `inputs.conf` is enabled. To keep some
disabled, pass their stanza names (the part of the bracketed header before
`://`, or the whole bracketed string if there is no `://`) via
`-DisabledStanzaNames`, e.g. `-DisabledStanzaNames 'powershell','script'`.

Splunkbase requires a login, so the Windows add-on can't be downloaded
automatically. Download it yourself as a `.tgz`/`.spl` file from
[splunkbase.splunk.com/app/742](https://splunkbase.splunk.com/app/742).

## Get a Windows collector binary

Before building the image, place a Windows collector binary named
`otelcol.exe` in this directory.

The `splunk_inputs` receiver this example depends on is only registered from
`v0.158.0` onward (see the `enableTARunner` feature gate in
[`internal/components/components.go`](../../internal/components/components.go)),
so download the latest release from the
[project's GitHub releases page](https://github.com/signalfx/splunk-otel-collector/releases)
and fail fast if it predates that version:

```powershell
$minVersion = [version]'0.158.0'
$release = Invoke-RestMethod -Uri 'https://api.github.com/repos/signalfx/splunk-otel-collector/releases/latest'
$collectorVersion = $release.tag_name -replace '^v', ''
if ([version]$collectorVersion -lt $minVersion) {
    Write-Host "Error: latest release is v$collectorVersion, but the splunk_inputs receiver requires v$minVersion or newer." -ForegroundColor Red
    exit 1
}

$downloadUrl = "https://github.com/signalfx/splunk-otel-collector/releases/download/v${collectorVersion}/otelcol_windows_amd64.exe"
Invoke-WebRequest -Uri $downloadUrl -OutFile .\examples\splunkinputs-on-windows\otelcol.exe
```

Alternatively, build it locally from the repository root:

```powershell
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
go build -trimpath -o .\examples\splunkinputs-on-windows\otelcol.exe .\cmd\otelcol
```

Or, if you already built the Windows binary under `bin`:

```powershell
Copy-Item .\bin\otelcol_windows_amd64.exe .\examples\splunkinputs-on-windows\otelcol.exe
```

## Run the example

From this directory, run `run-example.ps1` with the path to the TA package
you downloaded and the HEC endpoint/token to send data to:

```powershell
.\run-example.ps1 `
    -TaPackagePath 'C:\Users\you\Downloads\splunk-add-on-for-microsoft-windows_1100.spl' `
    -SplunkHecUrl 'https://your-splunk-host:8088/services/collector/event' `
    -SplunkHecToken '00000000-0000-0000-0000-0000000000000'
```

This extracts the TA to a local `ta` folder, copies its `default/` directory
to `local/` and enables every input by rewriting `disabled = 1` to
`disabled = 0`, then builds and runs the container with that folder mounted at
`C:\var\ta` and the HEC endpoint/token passed in as environment variables.

The container waits for the mounted TA, then starts the collector with the
`splunk_inputs` receiver reading it and exporting to `splunk_hec`. See
[`otel-collector-config.yaml`](./otel-collector-config.yaml) for the full
configuration.

## Notes on the setup

- Unlike `splunkinputs`, which deploys a real `splunk/splunk` container as
  destination via Docker Compose, this example has no bundled destination.
  Point `-SplunkHecUrl`/`-SplunkHecToken` at any HEC endpoint reachable from
  the container, including a Splunk instance running on the host.
- Like `splunkinputs`'s `/var/ta`, `C:\var\ta` is a generic mount point for
  any TA compatible with the `splunk_inputs` receiver, not just the Splunk
  Add-on for Microsoft Windows used here. The receiver reads a single TA per
  configured path, so only one TA can be mounted there at a time.
- The `splunk_inputs` receiver parses the TA's configuration files directly;
  no Splunk Universal Forwarder process runs in this container. See
  [`packaging/ta-v2`](../../packaging/ta-v2/) for examples that install and
  run a real Universal Forwarder on Windows.
- The example uses `mcr.microsoft.com/windows/servercore:ltsc2022` as its base
  image and copies in a local `otelcol.exe`, avoiding a multi-container
  Compose setup, since Windows containers don't support Docker Compose.
