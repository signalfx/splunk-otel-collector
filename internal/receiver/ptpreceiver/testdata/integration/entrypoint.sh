#!/bin/bash
# Starts this container's local "slave" ptp4l instance (high priority1, so it
# loses BMCA) over eth0, using software timestamping (-S) so no PTP-capable
# NIC is required. A sibling container on the same Docker network must
# already be running the low-priority1 "master" (see run.sh), so the two can
# exchange PTP multicast announce/sync messages and actually synchronize --
# two ptp4l instances on the same host/interface derive identical
# clockIdentity values and silently ignore each other's announcements, which
# is why this needs two containers rather than two processes in one.
#
# Once this instance's management socket is up, runs the command passed as
# arguments (normally `go test -tags=ptp_integration`).
set -euo pipefail

CONF_DIR="/etc/ptp4l"
LOG_DIR="/var/log/ptp-integration"
mkdir -p "$LOG_DIR"

ptp4l -f "$CONF_DIR/slave.conf" -i eth0 -S -m >"$LOG_DIR/slave.log" 2>&1 &
slave_pid=$!

cleanup() {
  kill "$slave_pid" 2>/dev/null || true
  wait "$slave_pid" 2>/dev/null || true
}
trap cleanup EXIT

echo "waiting for the local ptp4l management socket..."
for _ in $(seq 1 30); do
  if [ -S /var/run/ptp4l-slave.ro ]; then
    break
  fi
  sleep 1
done
if [ ! -S /var/run/ptp4l-slave.ro ]; then
  echo "slave management socket never appeared" >&2
  cat "$LOG_DIR/slave.log" >&2
  exit 1
fi

exec "$@"
