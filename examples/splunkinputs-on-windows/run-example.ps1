# Prepares a TA (by default, the Splunk Add-on for Microsoft Windows) and
# runs this example in a single Windows container. See README.md for the
# manual download step this script depends on.

param(
    [Parameter(Mandatory = $true)]
    [string]$TaPackagePath,

    [Parameter(Mandatory = $true)]
    [string]$SplunkHecUrl,

    [Parameter(Mandatory = $true)]
    [string]$SplunkHecToken,

    # Stanza names to keep disabled instead of enabling. A stanza name is the
    # part of the bracketed header before "://" (e.g. "script", "monitor"),
    # or the whole bracketed string if it has no "://".
    [string[]]$DisabledStanzaNames = @()
)

$ErrorActionPreference = 'Stop'

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$CONTAINER_NAME = if ($env:CONTAINER_NAME) { $env:CONTAINER_NAME } else { "splunkinputs-on-windows" }
$IMAGE_TAG = if ($env:IMAGE_TAG) { $env:IMAGE_TAG } else { "latest" }
# Local staging directory for the extracted TA; mounted into the container at
# C:\var\ta, the generic TA location the receiver reads.
$TA_DIR = Join-Path $SCRIPT_DIR "ta"

$resolvedTaPackagePath = Resolve-Path -ErrorAction SilentlyContinue $TaPackagePath
if (-not $resolvedTaPackagePath) {
    Write-Host "Error: TaPackagePath '$TaPackagePath' could not be resolved." -ForegroundColor Red
    exit 1
}

$otelcolPath = Join-Path $SCRIPT_DIR "otelcol.exe"
if (-not (Test-Path $otelcolPath)) {
    Write-Host "Error: otelcol.exe not found at $otelcolPath" -ForegroundColor Red
    Write-Host "Place a Windows collector binary there first. See README.md for download/build instructions." -ForegroundColor Red
    exit 1
}

# Extract the TA and enable its inputs
if (Test-Path $TA_DIR) {
    Write-Host "Removing previous extracted TA at $TA_DIR"
    Remove-Item -Path $TA_DIR -Recurse -Force
}
New-Item -ItemType Directory -Path $TA_DIR -Force | Out-Null

Write-Host "Extracting $($resolvedTaPackagePath.Path) to $TA_DIR"
tar -xzf $resolvedTaPackagePath.Path -C $TA_DIR --strip-components=1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to extract $($resolvedTaPackagePath.Path)" -ForegroundColor Red
    exit 1
}

$defaultDir = Join-Path $TA_DIR "default"
$localDir = Join-Path $TA_DIR "local"
Write-Host "Copying default/ to local/ and enabling inputs"
Copy-Item -Path $defaultDir -Destination $localDir -Recurse -Force

$localInputsConf = Join-Path $localDir "inputs.conf"
$currentStanzaDisabled = $false
$updatedLines = foreach ($line in Get-Content -Path $localInputsConf) {
    if ($line -match '^\s*\[(.+)\]\s*$') {
        $stanzaHeader = $Matches[1]
        $stanzaName = ($stanzaHeader -split '://', 2)[0]
        $currentStanzaDisabled = $DisabledStanzaNames -contains $stanzaName
        $line
    } elseif ($line -match '^disabled\s*=\s*\d\s*$') {
        if ($currentStanzaDisabled) { 'disabled = 1' } else { 'disabled = 0' }
    } else {
        $line
    }
}
$updatedLines | Set-Content -Path $localInputsConf

# Stop and remove existing container if it exists
$existingContainer = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $CONTAINER_NAME }
if ($existingContainer) {
    Write-Host "Stopping and removing existing container: $CONTAINER_NAME"
    docker rm -f $CONTAINER_NAME | Out-Null
}

Write-Host "Building image splunkinputs-on-windows:$IMAGE_TAG"
docker build -t "splunkinputs-on-windows:$IMAGE_TAG" $SCRIPT_DIR
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to build image" -ForegroundColor Red
    exit 1
}

Write-Host "Launching container: $CONTAINER_NAME"
docker run -d --name $CONTAINER_NAME `
    -e SPLUNK_HEC_URL=$SplunkHecUrl `
    -e SPLUNK_HEC_TOKEN=$SplunkHecToken `
    -v "${TA_DIR}:C:\var\ta" `
    "splunkinputs-on-windows:$IMAGE_TAG"
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to launch container" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Container launched successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Useful commands:" -ForegroundColor Cyan
Write-Host "  View collector logs: docker logs -f $CONTAINER_NAME"
Write-Host "  Stop container: docker stop $CONTAINER_NAME"
Write-Host "  Remove container: docker rm -f $CONTAINER_NAME"
Write-Host "  Docker exec shell: docker exec -it $CONTAINER_NAME powershell"
Write-Host ""
