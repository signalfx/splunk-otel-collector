<#
.SYNOPSIS
    The script validates if the Splunk OpenTelemetry .NET Deployer TA can be
    installed and uninstalled successfully by Splunk.

.DESCRIPTION
    The script extracts the Splunk OpenTelemetry .NET Deployer TA from a .tgz file
    and installs it. It then validates the installation by checking if the expected
    files and directories are present. The script also uninstalls the Splunk OpenTelemetry
    .NET instrumentation, by changing the local configuration of the Deployer TA
    and validates the uninstallation by once again checking if the instrumentation
    files and folders were removed.

.PARAMETER tgzFilePath
    Path to the .tgz file containing the Splunk OpenTelemetry .NET Deployer TA.

.PARAMETER splunkInstallPath
    Path to the Splunk installation directory.

.PARAMETER splunkAppName
    Actual name on the disk of the folder with the Splunk OpenTelemetry .NET
    Deployer TA app.
#>
param (
    [string]$tgzFilePath,
    [string]$splunkInstallPath,
    [string]$splunkAppName = "splunk_otel_dotnet_deployer"
)

# Function to extract .tgz file into the given destination path
function Expand-Tgz {
    param (
        [string]$tgzFilePath,
        [string]$destinationPath
    )

    Write-Host "Expanding '$tgzFilePath' into '$destinationPath'"
    tar -xzf $tgzFilePath -C $destinationPath
}

# Validate parameters
if (-not (Test-Path $tgzFilePath)) {
    Write-Error "The specified .tgz file does not exist: $tgzFilePath"
    exit 1
}
if (-not (Test-Path $splunkInstallPath)) {
    Write-Error "The specified Splunk installation path does not exist: $splunkInstallPath"
    exit 1
}

$splunkAppsPath = Join-Path $splunkInstallPath "etc/apps"
Expand-Tgz -tgzFilePath $tgzFilePath -destinationPath $splunkAppsPath

# Check if the TA is installed
$taPath = Join-Path $splunkAppsPath $splunkAppName
if (Test-Path $taPath) {
    Write-Host "Splunk TA $splunkAppName is installed at $taPath"
} else {
    Write-Error "Splunk TA $splunkAppName is not installed"
    exit 1
}

# Restart Splunk
Write-Host "Restarting Splunk"
& "$splunkInstallPath/bin/splunk.exe" restart

# Check if the Splunk OpenTelemetry .NET instrumentation is installed
$splunkOtelDotNetPath = Join-Path ${env:ProgramFiles} "Splunk OpenTelemetry .NET"
if (Test-Path $splunkOtelDotNetPath) {
    Write-Host "Splunk OpenTelemetry .NET instrumentation is installed at $splunkOtelDotNetPath"
} else {
    Write-Error "Splunk OpenTelemetry .NET instrumentation is not installed"
    exit 1
}

# Create a local inputs.conf for the TA and configure it to uninstall the Splunk OpenTelemety .NET instrumentation
$localInputsConfPath = Join-Path $taPath "local/inputs.conf"
# Ensure the local directory exists
$localDir = Join-Path $taPath "local"
if (-not (Test-Path $localDir)) {
    New-Item -ItemType Directory -Path $localDir | Out-Null
}

# Copy the default inputs.conf to the local directory
$defaultInputsConfPath = Join-Path $taPath "default/inputs.conf"
if (Test-Path $defaultInputsConfPath) {
    Copy-Item -Path $defaultInputsConfPath -Destination $inputsConfPath
    Write-Host "Copied default inputs.conf to local directory"
} else {
    Write-Error "Default inputs.conf does not exist at $defaultInputsConfPath"
    exit 1
}

# Modify the local inputs.conf to uninstall the Splunk OpenTelemetry .NET instrumentation
$localInputsConfContent = (Get-Content $localInputsConfPath).Replace("uninstall = false", "uninstall = true")
Set-Content -Path $localInputsConfPath -Value $localInputsConfContent

# Restart Splunk
Write-Host "Restarting Splunk"
& "$splunkInstallPath/bin/splunk.exe" restart

# Check if the Splunk OpenTelemetry .NET instrumentation is uninstalled
if (-not (Test-Path $splunkOtelDotNetPath)) {
    Write-Host "Splunk OpenTelemetry .NET instrumentation is uninstalled"
} else {
    Write-Error "Splunk OpenTelemetry .NET instrumentation is not uninstalled"
    exit 1
}
