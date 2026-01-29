$ufRoot = Join-Path -Path $Env:ProgramFiles -ChildPath "SplunkUniversalForwarder"
$logDir = Join-Path -Path $ufRoot -ChildPath "var\log\splunk"
$splunkdLog = Join-Path -Path $logDir -ChildPath "splunkd.log"
Get-Service SplunkForwarder

Write-Host "Waiting for splunkd.log ..."
$timeout = 60
$elapsed = 0
while (-not (Test-Path -Path $splunkdLog)) {
    if ($elapsed -ge $timeout) {
        Write-Host "Timeout: splunkd.log was not created within $timeout seconds" -ForegroundColor Red
        exit 1
    }
    Start-Sleep -Seconds 2
    $elapsed += 2
    Write-Host -NoNewline "."
}

Get-Service SplunkForwarder

Get-Content -Path $splunkdLog -Wait
