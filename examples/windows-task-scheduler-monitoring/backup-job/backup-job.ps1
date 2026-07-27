# Stand-in for a pre-existing backup script, run periodically by the
# \BackupJob scheduled task. Its only telemetry is writing progress to the
# Windows Application event log and pushing its own duration/size metrics as
# InfluxDB line protocol, the same way it might already report to an existing
# InfluxDB/Telegraf setup; it knows nothing about the collector otherwise, and
# nothing about its own exit status (see run-scheduled-job-with-status.ps1 and
# its scheduled task action wiring for that).

$ErrorActionPreference = 'Stop'

$source = 'BackupJob'
$influxDbEndpoint = if ($env:INFLUXDB_ENDPOINT) {
    $env:INFLUXDB_ENDPOINT
} else {
    'http://127.0.0.1:8086/api/v2/write?org=example-org&bucket=example-bucket&precision=s'
}

Write-EventLog -LogName Application -Source $source -EntryType Information -EventId 1000 -Message 'backup-job: starting backup of C:\Windows\System32\drivers\etc'

$startTime = Get-Date
$timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$archive = "C:\Windows\Temp\etc-backup-$timestamp.zip"

Compress-Archive -Path 'C:\Windows\System32\drivers\etc\*' -DestinationPath $archive -Force

$durationSeconds = [Math]::Max(0, [int]((Get-Date) - $startTime).TotalSeconds)
$sizeBytes = (Get-Item $archive).Length

Write-EventLog -LogName Application -Source $source -EntryType Information -EventId 1001 -Message "backup-job: finished in ${durationSeconds}s, size ${sizeBytes} bytes"

$body = "backup_job duration_seconds=${durationSeconds},size_bytes=${sizeBytes}"
$response = Invoke-WebRequest -UseBasicParsing -Method Post -Uri $influxDbEndpoint -Body $body

Write-EventLog -LogName Application -Source $source -EntryType Information -EventId 1002 -Message "backup-job: pushed metrics to $influxDbEndpoint, HTTP $($response.StatusCode)"
