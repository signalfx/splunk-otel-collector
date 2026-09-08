# Monitoring a Windows Scheduled Task with the Splunk OpenTelemetry Collector

This example shows how to monitor a script that runs on a recurring schedule
via the [Windows Task Scheduler](https://learn.microsoft.com/windows/win32/taskschd/task-scheduler-start-page),
using the Splunk OpenTelemetry Collector.

The `\BackupJob` scheduled task runs every minute (a short interval chosen for
this demo; a real task would typically use a daily, weekly, or event-based
trigger). The task runs [`backup-job.ps1`](./backup-job/backup-job.ps1), which
stands in for a pre-existing job script: it logs progress to the Windows
Application event log and reports its own `duration_seconds`/`size_bytes` as
[InfluxDB line protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/)
after each run. It has no other telemetry code, and in particular does not
report whether the scheduled task succeeded or failed.

Exit-status monitoring is added on top without touching that script, using
[`run-scheduled-job-with-status.ps1`](./backup-job/run-scheduled-job-with-status.ps1),
a generic wrapper configured as the scheduled task action. The wrapper runs the
job script, preserves its exit code for Task Scheduler, and pushes a
success/failure metric to the collector. The same wrapper can be reused for any
other PowerShell script, cmd/bat script, or executable run by Task Scheduler.

The collector is configured with four receivers:

- [`windows_event_log`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/windowseventlogreceiver/README.md),
  reading `Microsoft-Windows-TaskScheduler/Operational`. This captures task
  lifecycle events such as task start, action launch, action completion with
  return code, and task finish.
- `windows_event_log`, reading `Application`. This captures the demo script's
  own log entries from the `BackupJob` event source.
- [`otlp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/receiver/otlpreceiver/README.md),
  to receive the metrics `run-scheduled-job-with-status.ps1` pushes directly.
- [`influxdb`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/influxdbreceiver/README.md),
  to receive the `backup_job duration_seconds=...,size_bytes=...` line
  protocol metric `backup-job.ps1` pushes directly.

See [`otel-collector-config.yaml`](./otel-collector-config.yaml) for the full,
minimal configuration. Both pipelines send to the `debug` exporter so you can
see the results directly in the collector's logs.

## Running the example

This example runs a single Windows container instead of using Docker Compose.
Before building the image, place a Windows collector binary named
`otelcol.exe` in this directory.

To download it from the
[project's GitHub releases page](https://github.com/signalfx/splunk-otel-collector/releases),
replace the version below with the release you want to use, then run from the
repository root:

```powershell
$collectorVersion = '0.149.0'
$asset = "agent-bundle_${collectorVersion}_windows_amd64.zip"
$downloadUrl = "https://github.com/signalfx/splunk-otel-collector/releases/download/v${collectorVersion}/${asset}"
$downloadPath = ".\examples\windows-task-scheduler-monitoring\$asset"
$extractPath = ".\examples\windows-task-scheduler-monitoring\release"

Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadPath
Expand-Archive -Path $downloadPath -DestinationPath $extractPath -Force
Copy-Item (Get-ChildItem $extractPath -Recurse -Filter otelcol.exe | Select-Object -First 1).FullName .\examples\windows-task-scheduler-monitoring\otelcol.exe
```

Alternatively, build it locally from the repository root:

```powershell
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
go build -trimpath -o .\examples\windows-task-scheduler-monitoring\otelcol.exe .\cmd\otelcol
```

Or, if you already built the Windows binary under `bin`:

```powershell
Copy-Item .\bin\otelcol_windows_amd64.exe .\examples\windows-task-scheduler-monitoring\otelcol.exe
```

Then build and run the container:

```powershell
docker build -t windows-task-scheduler-monitoring .\examples\windows-task-scheduler-monitoring
docker run --rm --name windows-task-scheduler-monitoring windows-task-scheduler-monitoring
```

The container starts the collector, enables the Task Scheduler Operational
event log channel, registers `\BackupJob`, runs it once immediately, then lets
Task Scheduler run it every minute.

You should see `LogRecord` output for Task Scheduler events from
`Microsoft-Windows-TaskScheduler/Operational`, `LogRecord` output for the
script's `BackupJob` Application events, `windows_task_scheduler.job.successful_run`
and `windows_task_scheduler.job.failed_run` metrics from the wrapper, and
`backup_job_duration_seconds`/`backup_job_size_bytes` metrics from
`backup-job.ps1`.

## Notes on the setup

- Windows Task Scheduler does not have a direct equivalent to systemd's
  `ExecStopPost=`. This example uses a reusable wrapper as the scheduled task
  action so exit-status metrics can be added without editing the existing job
  script.
- Task Scheduler's event log records action return codes but not a task
  process's stdout/stderr. Scripts that need searchable execution logs should
  write to an event log, a file collected by the `filelog` receiver, or another
  existing logging destination.
- The example uses `mcr.microsoft.com/windows/servercore:ltsc2022` as its base
  image and copies in a local `otelcol.exe`, avoiding a multi-container Compose
  setup for Windows containers.
