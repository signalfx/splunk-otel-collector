# Container entrypoint for the splunkinputs-on-windows example.
#
# Starts the collector with the splunk_inputs receiver and splunk_outputs
# exporter. The complete Splunk home, including the TA and outputs.conf, is
# mounted at C:\var\splunk_home by run-example.ps1.

$ErrorActionPreference = 'Stop'

& C:\otelcol\otelcol.exe --config=C:\otelcol\otel-collector-config.yaml --feature-gates=+enableTARunner
exit $LASTEXITCODE
