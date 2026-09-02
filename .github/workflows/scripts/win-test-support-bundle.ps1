param (
    [string]$mode = "agent",
    [switch]$with_supervisor
)

$ErrorActionPreference = 'Stop'
Set-PSDebug -Trace 1

# test support bundle script
Set-Location -Path "$env:ProgramFiles\Splunk\OpenTelemetry Collector"
Test-Path -Path ".\splunk-support-bundle.ps1"
powershell.exe -File "$env:ProgramFiles\Splunk\OpenTelemetry Collector\splunk-support-bundle.ps1" -t \tmp\splunk-support-bundle
Test-Path -Path ".\splunk-support-bundle.zip"
Test-Path -Path "\tmp\splunk-support-bundle\logs\splunk-otel-collector.log"
Test-Path -Path "\tmp\splunk-support-bundle\logs\splunk-otel-collector.txt"
Test-Path -Path "\tmp\splunk-support-bundle\metrics\collector-metrics.txt"
Test-Path -Path "\tmp\splunk-support-bundle\metrics\df.txt"
Test-Path -Path "\tmp\splunk-support-bundle\metrics\free.txt"
Test-Path -Path "\tmp\splunk-support-bundle\metrics\top.txt"
Test-Path -Path "\tmp\splunk-support-bundle\zpages\tracez.html"
Test-Path -Path "\tmp\splunk-support-bundle\config\${mode}_config.yaml"
Test-Path -Path "\tmp\splunk-support-bundle\config\service_environment.txt"
if ($with_supervisor) {
    $supervisor_config = "\tmp\splunk-support-bundle\config\supervisor\supervisor_config.yaml"
    $supervisor_runtime_config = "\tmp\splunk-support-bundle\config\supervisor\supervisor_runtime_config.yaml"
    if (!(Test-Path -Path $supervisor_config) -or !(Test-Path -Path $supervisor_runtime_config)) {
        throw "Supervisor configuration was not included in the support bundle."
    }
}
