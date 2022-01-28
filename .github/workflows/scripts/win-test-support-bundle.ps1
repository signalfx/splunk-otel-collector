param (
    [string]$mode = "agent",
    [string]$with_fluentd = "true"
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

if ( "$with_fluentd" -eq "true" ) {
    Test-Path -Path "\tmp\splunk-support-bundle\logs\td-agent.log"
    Test-Path -Path "\tmp\splunk-support-bundle\logs\td-agent.txt"
    Test-Path -Path "\tmp\splunk-support-bundle\config\td-agent\td-agent.conf"
}
