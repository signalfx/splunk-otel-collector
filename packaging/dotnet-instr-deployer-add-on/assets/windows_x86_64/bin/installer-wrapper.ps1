<#
.SYNOPSIS
This script drives the installation of the Splunk Distribution of OpenTelemetry .NET.

.DESCRIPTION
This script is intended to be called by the splunk_otel_dotnet_deployer.exe as
the later is launched by splunkd with proper environment variables and configuration.

.PARAMETER uninstall
If present, the script will uninstall the Splunk Distribution of OpenTelemetry .NET.

#>

param (
    [switch]$uninstall
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$scriptPath = $MyInvocation.MyCommand.Path
$scriptDir = Split-Path $scriptPath

$modulePath = Join-Path $scriptDir "Splunk.OTel.DotNet.psm1"
Import-Module $modulePath

$otelSDKInstallVersion = Get-OpenTelemetryInstallVersion

if ($uninstall) {
    if ($otelSDKInstallVersion) {
        Write-Host "Uninstalling Splunk Distribution of OpenTelemetry .NET ..."
        Uninstall-OpenTelemetryCore

        $w3svc = Get-Service -name "W3SVC" -ErrorAction SilentlyContinue
        $was = Get-Service -name "WAS" -ErrorAction SilentlyContinue
        if ($w3svc -And $was) {
            Write-Host "Unregistering OpenTelemetry for IIS ..."
            Unregister-OpenTelemetryForIIS
        }

        Write-Host "Splunk Distribution of OpenTelemetry .NET uninstalled successfully"
    } else {
        Write-Host "Nothing to do since Splunk Distribution of OpenTelemetry .NET is not installed"
    }
# Install the SDK if it is not already installed
} elseif ($otelSDKInstallVersion) {
    Write-Host "Nothing to do since Splunk Distribution of OpenTelemetry .NET is already installed. OpenTelemetry .NET SDK version: $otelSDKInstallVersion"
} else {
    Write-Host "Installing Splunk Distribution of OpenTelemetry .NET ..."

    # Avoid issues with NGEN assemblies by forcing SingleDomain mode.
    Set-ItemProperty -Path "HKLM:\\SOFTWARE\\Microsoft\\.NETFramework" -Name "LoaderOptimization" -Value 1 -Type DWord

    $zipPath = Join-Path $scriptDir "splunk-opentelemetry-dotnet-windows.zip"
    Install-OpenTelemetryCore -LocalPath $zipPath

    $w3svc = Get-Service -name "W3SVC" -ErrorAction SilentlyContinue
    $was = Get-Service -name "WAS" -ErrorAction SilentlyContinue
    if ($w3svc -And $was) {
        Write-Host "Registering OpenTelemetry for IIS ..."
        Register-OpenTelemetryForIIS
    }

    Write-Host "Splunk Distribution of OpenTelemetry .NET installed successfully"
}
