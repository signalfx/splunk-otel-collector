#!/bin/bash
#
# Generic systemd job status reporter.
#
# Wire this up as the ExecStopPost= of any systemd service (typically one
# triggered by a .timer) to push a success/failure metric for its last run,
# without changing the job's own script at all:
#
#   [Service]
#   ExecStart=/path/to/your-existing-job.sh
#   ExecStopPost=/usr/local/bin/report-job-status.sh %n
#
# systemd runs ExecStopPost= after the main process exits (even if it failed
# or timed out) and sets $SERVICE_RESULT/$EXIT_CODE/$EXIT_STATUS to describe
# how it ended; %n expands to the full unit name. See systemd.exec(5) and
# systemd.service(5) for details on these variables.

unit="${1:?usage: report-job-status.sh <unit-name>, pass systemd %n}"

OTLP_ENDPOINT="${OTLP_ENDPOINT:-http://otelcollector:4318/v1/metrics}"

successful_run=0
failed_run=1
if [ "$SERVICE_RESULT" = "success" ]; then
  successful_run=1
  failed_run=0
fi

service_result=${SERVICE_RESULT:-not_set}
exit_code=${EXIT_CODE:-not_set}
exit_status=${EXIT_STATUS:-not_set}

echo "report-job-status: ${unit} service_result=${service_result} exit_code=${exit_code} exit_status=${exit_status}"

now_ns="$(date +%s)000000000"
payload=$(cat <<JSON
{
  "resourceMetrics": [{
    "resource": {"attributes": [
      {"key": "systemd.unit", "value": {"stringValue": "${unit}"}}
    ]},
    "scopeMetrics": [{
      "scope": {"name": "report-job-status.sh"},
      "metrics": [
        {"name": "systemd_job.successful_run", "gauge": {"dataPoints": [{"timeUnixNano": "${now_ns}", "asInt": "${successful_run}"}]}},
        {"name": "systemd_job.failed_run", "gauge": {"dataPoints": [{"timeUnixNano": "${now_ns}", "asInt": "${failed_run}"}]}}
      ]
    }]
  }]
}
JSON
)

http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$OTLP_ENDPOINT" -H "Content-Type: application/json" --data-binary "$payload")
echo "report-job-status: pushed metrics for ${unit} to ${OTLP_ENDPOINT}, HTTP ${http_code}"
