Start-Service msiserver
msiexec.exe /i C:\splunk-uf.msi /qn /norestart /l*v C:\splunk-uf-install.log AGREETOLICENSE=Yes
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Splunk Universal Forwarder installation failed with exit code $LASTEXITCODE" -ForegroundColor Red
    Get-Content C:/splunk-uf-install.log
    exit 1
}

$splunkdLog = "C:/Program Files/SplunkUniversalForwarder/var/log/splunk/splunkd.log"

Get-Content C:/splunk-uf-install.log
Get-WinEvent -LogName Application
Get-WinEvent -LogName System
Get-Service SplunkForwarder

Write-Host "Waiting for splunkd.log ..."
$timeout = 60
$elapsed = 0
while (-not (Test-Path -Path $splunkdLog)) {
    $elapsed += 2
    if ($elapsed -ge $timeout) {
        Write-Host -NoNewline "."
        Write-Host "Timeout: splunkd.log was not created within $timeout seconds" -ForegroundColor Red
        exit 1
    }
    Start-Sleep -Seconds 2
}

Get-Service SplunkForwarder

Get-Content -Path $splunkdLog -Wait
