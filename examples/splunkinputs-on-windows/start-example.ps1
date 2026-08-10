# Container entrypoint for the splunkinputs-on-windows example.
#
# Waits for a TA to be mounted at C:\var\ta (see run-example.ps1), then starts
# the collector with the splunk_inputs receiver, which reads that TA's
# inputs.conf, transforms.conf, and props.conf directly, emulating its
# modular input(s) without running a real splunkd.

$ErrorActionPreference = 'Stop'

if (-not $env:SPLUNK_HEC_URL) {
    Write-Host 'Error: SPLUNK_HEC_URL environment variable is required.' -ForegroundColor Red
    exit 1
}
if (-not $env:SPLUNK_HEC_TOKEN) {
    Write-Host 'Error: SPLUNK_HEC_TOKEN environment variable is required.' -ForegroundColor Red
    exit 1
}

& C:\otelcol\otelcol.exe --config=C:\otelcol\otel-collector-config.yaml --feature-gates=+enableTARunner
exit $LASTEXITCODE
