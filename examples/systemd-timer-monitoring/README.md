# Monitoring a systemd timer with the Splunk OpenTelemetry Collector

This example shows how to monitor a script that runs on a recurring schedule
via a [systemd timer](https://www.freedesktop.org/software/systemd/man/latest/systemd.timer.html),
using the Splunk OpenTelemetry Collector.

`backup-job.timer` triggers `backup-job.service` every 30 seconds (a short
interval chosen for this demo; a real timer would typically use `OnCalendar`
for something like a nightly cron-style schedule). The service runs
[`backup-job.sh`](./backup-job/backup-job.sh), which stands in for a
pre-existing job script someone is moving onto a systemd timer: it logs
progress to stdout, and reports its own `duration_seconds`/`size_bytes` as
[InfluxDB line protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/)
after each run — the same way it might already report to an existing
InfluxDB/Telegraf setup. It has no other telemetry code, and in particular
knows nothing about its own exit status.

Exit-status monitoring is added on top without touching that script, using
[`report-job-status.sh`](./backup-job/report-job-status.sh), a generic
script wired up as `backup-job.service`'s `ExecStopPost=`
(see [`backup-job.service`](./backup-job/backup-job.service)). systemd runs
`ExecStopPost=` after the main process exits, success or failure, and passes
it `$SERVICE_RESULT`/`$EXIT_CODE`/`$EXIT_STATUS` describing how it ended;
`report-job-status.sh` turns those into a metric and pushes it to the
collector. Because it only depends on systemd's own exec/service variables,
the same script can be reused as the `ExecStopPost=` for any other job's
`.service` unit.

The collector is configured with three receivers:

- [`journald`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/journaldreceiver/README.md),
  filtered to the `backup-job.service` and `backup-job.timer` units. This
  captures every run of the job (start, the script's own log lines, and
  systemd's own success/failure messages) without any code changes to the
  script itself.
- [`otlp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/receiver/otlpreceiver/README.md),
  to receive the metrics `report-job-status.sh` pushes directly.
- [`influxdb`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/influxdbreceiver/README.md),
  to receive the `backup_job duration_seconds=...,size_bytes=...` line
  protocol metric `backup-job.sh` pushes directly.

See [`otel-collector-config.yaml`](./otel-collector-config.yaml) for the full,
minimal configuration. Both pipelines send to the `debug` exporter so you can
see the results directly in the collector's logs.

## Running the example

```bash
$> docker compose up --build
```

This starts two containers:

- `backup-job`: runs systemd as PID 1 (following the well-known
  ["systemd in Docker"](https://developers.redhat.com/blog/2016/09/13/running-systemd-in-a-non-privileged-container)
  pattern) with the timer/service installed and enabled.
- `otelcollector`: the Splunk OpenTelemetry Collector, reading the journal
  shared from `backup-job` via a Docker volume and listening for OTLP on port
  4318 and InfluxDB line protocol on port 8086.

Wait for the timer to fire (up to 30s) and check the collector's output:

```bash
$> docker logs otelcollector
```

You should see a `LogRecord` for each of the script's messages (tagged
`SYSLOG_IDENTIFIER: backup-job.sh`), as well as `systemd_job.successful_run`
and `systemd_job.failed_run` metrics (from `report-job-status.sh`) and
`backup_job_duration_seconds`/`backup_job_size_bytes` metrics (from
`backup-job.sh`) appearing after each run.

## Notes on the setup

- The journal is shared between the two containers via the `journal` Docker
  volume, mounted at `/run/log/journal` (systemd's volatile/in-memory journal
  location) in both containers. This is the standard location in containerized
  environments; see the
  [journald receiver's container guidance](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/journaldreceiver/README.md#journal-location-in-container-environments)
  for more detail, including how to use a persistent (`/var/log/journal`)
  location on a real host.
- The collector image used here (`quay.io/signalfx/splunk-otel-collector`)
  bundles a `journalctl` binary, so no extra installation is needed. It must
  run as root (`user: "0"` in the compose file) to read the journal.
- You may see a transient `No journal boot entry found for the specified
  boot` error in the collector's logs right after startup. This happens if
  the collector starts reading before `backup-job`'s journald has written its
  first entries; the receiver retries automatically and recovers within a
  few seconds.
