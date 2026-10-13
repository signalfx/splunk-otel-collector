#!/usr/bin/env bash
# Builds the PTP receiver integration test image and runs the full suite of
# integration scenarios against real ptp4l instances. See README.md in this
# directory for background on each scenario.
#
# With no arguments, runs every scenario (this is what CI does). With
# arguments, runs exactly that command once against the standard
# master+slave topology, for manual debugging -- e.g. to run a single test
# with -v output:
#   run.sh go test -tags=ptp_integration -run TestPTPReceiverIntegration -v ./internal/receiver/ptpreceiver/...
set -euo pipefail

cd "$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"

IMAGE="ptpreceiver-integration-test"
NETWORK="ptpreceiver-integration-test"
DOWNSTREAM_NETWORK="ptpreceiver-integration-test-downstream"
MASTER="ptpreceiver-integration-test-master"
BOUNDARY="ptpreceiver-integration-test-boundary"
LOSS_RECOVERY="ptpreceiver-integration-test-loss-recovery"
DOCKERFILE="internal/receiver/ptpreceiver/testdata/integration/Dockerfile"
PKG="./internal/receiver/ptpreceiver/..."
SIGNAL_DIR="/tmp/ptp-integration-signals"

cleanup() {
  docker rm -f "$MASTER" "$BOUNDARY" "$LOSS_RECOVERY" >/dev/null 2>&1 || true
  docker network rm "$NETWORK" "$DOWNSTREAM_NETWORK" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker build -t "$IMAGE" -f "$DOCKERFILE" .
cleanup

# start_master (re)starts the low-priority1 grandmaster sidecar that the
# other containers' high-priority1 "slave" instances synchronize to. Its MAC
# is pinned: ptp4l derives clockIdentity from the interface's MAC (EUI-64),
# and Docker randomizes an unpinned MAC on every `docker start`, which would
# otherwise make the grandmaster-loss-recovery test's "same grandmaster
# identity after restart" assertion flaky.
start_master() {
  docker run -d --name "$MASTER" --network "$NETWORK" --mac-address 02:42:ac:11:00:10 --entrypoint ptp4l "$IMAGE" \
    -f /etc/ptp4l/master.conf -i eth0 -S -m >/dev/null
}

# wait_for_container_file polls, via `docker exec`, for a file to appear
# inside a running container -- used to receive a "ready" signal from the Go
# test running inside it, since run.sh (on the host) and that test don't
# otherwise share a process or a mounted volume.
wait_for_container_file() {
  local container="$1" name="$2" timeout="${3:-60}"
  for _ in $(seq 1 "$timeout"); do
    if docker exec "$container" test -f "$SIGNAL_DIR/$name" 2>/dev/null; then
      return 0
    fi
    sleep 1
  done
  echo "timed out waiting for $container to signal $name" >&2
  return 1
}

# send_container_file is the reverse: it tells the Go test running inside a
# container that run.sh finished a host-level action.
send_container_file() {
  local container="$1" name="$2"
  docker exec "$container" mkdir -p "$SIGNAL_DIR"
  docker exec "$container" touch "$SIGNAL_DIR/$name"
}

if [ "$#" -gt 0 ]; then
  docker network create "$NETWORK" >/dev/null
  start_master
  # SYS_TIME: the slave's clock servo needs to call clock_adjtime/adjtimex on
  # its own (containerized) clock to leave LinuxPTP's UNCALIBRATED state.
  docker run --rm --network "$NETWORK" --cap-add=SYS_TIME -v "$(pwd):/src" "$IMAGE" "$@"
  exit 0
fi

echo "=== no remote grandmaster / socket not present (negative paths, no network dependency) ==="
docker run --rm --cap-add=SYS_TIME -v "$(pwd):/src" "$IMAGE" \
  go test -tags=ptp_integration -run 'TestPTPReceiverIntegration_NoRemoteGrandmaster|TestPTPReceiverIntegration_SocketNotPresent' -v "$PKG"

docker network create "$NETWORK" >/dev/null
start_master

echo "=== successful path: real metric contract and startup states ==="
docker run --rm --network "$NETWORK" --cap-add=SYS_TIME -v "$(pwd):/src" "$IMAGE" \
  go test -tags=ptp_integration -run '^TestPTPReceiverIntegration$' -v "$PKG"

echo "=== boundary clock (multiple ports) ==="
docker network create "$DOWNSTREAM_NETWORK" >/dev/null
# Created, not run: both networks must already be attached before ptp4l
# starts, since LinuxPTP checks for eth1 at startup and never retries.
# --network at create time becomes eth0 (and, importantly, suppresses the
# default bridge network Docker would otherwise attach as a 3rd interface);
# the second network, connected before start, becomes eth1.
docker create --name "$BOUNDARY" --network "$NETWORK" --cap-add=SYS_TIME -e PTP4L_IFACE_ARGS="-i eth0 -i eth1" -v "$(pwd):/src" "$IMAGE" \
  go test -tags=ptp_integration -run TestPTPReceiverIntegration_BoundaryClock -v "$PKG" >/dev/null
docker network connect "$DOWNSTREAM_NETWORK" "$BOUNDARY"
docker start -a "$BOUNDARY"
docker rm -f "$BOUNDARY" >/dev/null
docker network rm "$DOWNSTREAM_NETWORK" >/dev/null

echo "=== grandmaster loss and recovery ==="
docker run -d --name "$LOSS_RECOVERY" --network "$NETWORK" --cap-add=SYS_TIME -v "$(pwd):/src" "$IMAGE" \
  go test -tags=ptp_integration -run TestPTPReceiverIntegration_GrandmasterLossRecovery -v "$PKG" >/dev/null

wait_for_container_file "$LOSS_RECOVERY" ready-for-master-stop 150
docker stop "$MASTER" >/dev/null
send_container_file "$LOSS_RECOVERY" master-stopped

wait_for_container_file "$LOSS_RECOVERY" ready-for-master-restart 60
docker start "$MASTER" >/dev/null
send_container_file "$LOSS_RECOVERY" master-restarted

# `docker wait` blocks until the container stops and prints its exit code to
# stdout; its own command exit status does not reflect the container's.
exit_code=$(docker wait "$LOSS_RECOVERY")
docker logs "$LOSS_RECOVERY"
if [ "$exit_code" != "0" ]; then
  echo "grandmaster loss/recovery test failed (exit $exit_code)" >&2
  exit 1
fi
