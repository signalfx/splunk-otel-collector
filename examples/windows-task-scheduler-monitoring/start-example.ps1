# Container entrypoint for the Windows Task Scheduler monitoring example.
#
# Starts the collector, enables the Task Scheduler Operational event log
# channel, registers the demo \BackupJob scheduled task, runs it once
# immediately, and keeps the container alive by waiting on the collector.

$ErrorActionPreference = 'Stop'

$collector = Start-Process `
    -FilePath 'C:\otelcol\otelcol.exe' `
    -ArgumentList '--config=C:\otelcol\otel-collector-config.yaml' `
    -NoNewWindow `
    -PassThru

try {
    wevtutil.exe sl Microsoft-Windows-TaskScheduler/Operational /e:true

    $deadline = (Get-Date).AddSeconds(30)
    do {
        try {
            $client = [System.Net.Sockets.TcpClient]::new()
            $client.Connect('127.0.0.1', 4318)
            $client.Close()
            break
        } catch {
            if ((Get-Date) -ge $deadline) {
                throw 'Timed out waiting for the collector OTLP HTTP endpoint.'
            }
            Start-Sleep -Seconds 1
        }
    } while ($true)

    $taskAction = 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File C:\backup-job\run-scheduled-job-with-status.ps1 -TaskName \BackupJob -FilePath C:\backup-job\backup-job.ps1'
    schtasks.exe /Create /TN \BackupJob /SC MINUTE /MO 1 /TR $taskAction /RU SYSTEM /F
    schtasks.exe /Run /TN \BackupJob

    Write-Host 'BackupJob is scheduled to run every minute. Collector output follows.'
    Wait-Process -Id $collector.Id
} finally {
    if (-not $collector.HasExited) {
        Stop-Process -Id $collector.Id -Force
    }
}
