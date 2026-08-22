#!/bin/bash
#
# Stand-in for a pre-existing backup script, run periodically by the
# backup-job.timer systemd timer. Its only telemetry is pushing its own
# duration/size metrics as InfluxDB line protocol, the same way it likely
# already reports to an existing InfluxDB/Telegraf setup; it knows nothing
# about the collector otherwise, and nothing about its own exit status (see
# report-job-status.sh and its ExecStopPost= wiring in backup-job.service for
# that).

INFLUXDB_ENDPOINT="${INFLUXDB_ENDPOINT:-http://otelcollector:8086/api/v2/write?org=example-org&bucket=example-bucket&precision=s}"

echo "backup-job: starting backup of /etc"

start_ts=$(date +%s)
archive="/tmp/etc-backup-${start_ts}.tar.gz"
tar -czf "$archive" /etc >/tmp/backup-job.tar.log 2>&1
rc=$?
end_ts=$(date +%s)
duration=$((end_ts - start_ts))
size_bytes=$(stat -c%s "$archive" 2>/dev/null || echo 0)

echo "backup-job: finished in ${duration}s, size ${size_bytes} bytes, exit code ${rc}"

http_code=$(curl -s -o /dev/null -w "%{http_code}" -XPOST "$INFLUXDB_ENDPOINT" \
  --data-binary "backup_job duration_seconds=${duration},size_bytes=${size_bytes}")
echo "backup-job: pushed metrics to ${INFLUXDB_ENDPOINT}, HTTP ${http_code}"

exit "$rc"
