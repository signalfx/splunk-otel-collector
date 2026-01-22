# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

$ErrorActionPreference = "Stop"

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$ASSETS_DIR = if ($env:ASSETS_DIR) { $(Resolve-Path $env:ASSETS_DIR).Path } else { Join-Path $SCRIPT_DIR "assets" }
$LOG_DIR = Join-Path $SCRIPT_DIR "local-test-logs\var\log\splunk"
$CONTAINER_NAME = if ($env:CONTAINER_NAME) { $env:CONTAINER_NAME } else { "splunk-ta-otel-test" }
$IMAGE_TAG = if ($env:IMAGE_TAG) { $env:IMAGE_TAG } else { "latest" }

# Check if assets directory exists
if (-not (Test-Path $ASSETS_DIR)) {
    Write-Host "Error: Assets directory not found at $ASSETS_DIR" -ForegroundColor Red
    Write-Host "Please run 'make build-otelcol' first to build the collector binaries." -ForegroundColor Red
    exit 1
}

# Clean up previous log directory if it exists
if (Test-Path $LOG_DIR) {
    Write-Host "Cleaning up previous log directory at $LOG_DIR"
    Remove-Item -Path $LOG_DIR -Recurse -Force
}

# Create log directory
New-Item -ItemType Directory -Path $LOG_DIR -Force | Out-Null

# Stop and remove existing container if it exists
$existingContainer = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $CONTAINER_NAME }
if ($existingContainer) {
    Write-Host "Stopping and removing existing container: $CONTAINER_NAME"
    docker rm -f $CONTAINER_NAME | Out-Null
}

Write-Host "Launching Splunk Universal Forwarder container..."
Write-Host "  Container name: $CONTAINER_NAME"
Write-Host "  Image tag: $IMAGE_TAG"
Write-Host "  Assets directory: $ASSETS_DIR"
Write-Host "  Log directory: $LOG_DIR"

# Generate a random password
$SPLUNK_PASSWORD = [System.Guid]::NewGuid().ToString("N")

# Launch Splunk Universal Forwarder container
docker run -d --name $CONTAINER_NAME `
    --user ContainerAdministrator `
    -v "${ASSETS_DIR}:C:/Program Files/SplunkUniversalForwarder/etc/apps/Splunk_TA_OTel_Collector" `
    -v "${LOG_DIR}:C:/Program Files/SplunkUniversalForwarder/var/log/splunk" `
    "splunk-uf-windows:${IMAGE_TAG}"

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to launch container" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 5

Write-Host ""
Write-Host "Container launched successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Initial container output:" -ForegroundColor Yellow
docker logs $CONTAINER_NAME > container-initial.log 2>&1
Get-Content -Path "container-initial.log"
Remove-Item -Path "container-initial.log"
Write-Host ""

$splunkdLog = Join-Path $LOG_DIR "splunkd.log"

# Wait for splunkd.log to be created
$timeout = 180
$elapsed = 0
Write-Host -NoNewline "Waiting for splunkd.log creation: "
while (-not (Test-Path $splunkdLog)) {
    if ($elapsed -ge $timeout) {
        Write-Host ""
        Write-Host "Timeout: splunkd.log was not created within $timeout seconds" -ForegroundColor Red
        docker logs $CONTAINER_NAME > container-timeout.log 2>&1
        Get-Content -Path "container-timeout.log"
        Remove-Item -Path "container-timeout.log"
        exit 1
    }
    Start-Sleep -Seconds 2
    $elapsed += 2
    Write-Host -NoNewline "."
}
Write-Host ""

# Wait for Splunk TA OTel Collector to be recorded on the log
$timeout = 180
$elapsed = 0
Write-Host -NoNewline "Waiting for Splunk_TA_OTel_Collector to be recorded on splunkd.log: "
while (-not (Select-String -Path $splunkdLog -Pattern "Splunk_TA_OTel_Collector" -Quiet)) {
    if ($elapsed -ge $timeout) {
        Write-Host ""
        Write-Host "Timeout: Splunk_TA_OTel_Collector was not recorded on splunkd.log within $timeout seconds" -ForegroundColor Red
        exit 1
    }
    Start-Sleep -Seconds 2
    $elapsed += 2
    Write-Host -NoNewline "."
}
Write-Host ""
Write-Host "Splunk_TA_OTel_Collector in splunkd.log:"
Select-String -Path $splunkdLog -Pattern "Splunk_TA_OTel_Collector" | ForEach-Object { $_.Line }

Write-Host ""
Write-Host ""
Write-Host "Useful commands:" -ForegroundColor Cyan
Write-Host "  View container logs: docker logs -f $CONTAINER_NAME"
Write-Host "  Stop container: docker stop $CONTAINER_NAME"
Write-Host "  Remove container: docker rm -f $CONTAINER_NAME"
Write-Host "  View splunkd logs: Get-Content -Path '$splunkdLog' -Wait"
Write-Host "  Grep Splunk_TA_OTel_Collector logs: Select-String -Path '$splunkdLog' -Pattern 'Splunk_TA_OTel_Collector'"
Write-Host "  Docker exec shell: docker exec -it $CONTAINER_NAME powershell"
Write-Host ""
