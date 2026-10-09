#!/usr/bin/env bash
# Builds the PTP receiver integration test image and runs it against a real
# ptp4l "master" sidecar container on a shared Docker network. See README.md
# in this directory for background.
set -euo pipefail

cd "$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"

IMAGE="ptpreceiver-integration-test"
NETWORK="ptpreceiver-integration-test"
MASTER="ptpreceiver-integration-test-master"
DOCKERFILE="internal/receiver/ptpreceiver/testdata/integration/Dockerfile"

cleanup() {
  docker rm -f "$MASTER" >/dev/null 2>&1 || true
  docker network rm "$NETWORK" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker build -t "$IMAGE" -f "$DOCKERFILE" .

cleanup
docker network create "$NETWORK" >/dev/null

# Low-priority1 grandmaster. The test container (entrypoint.sh) starts the
# high-priority1 "slave" instance that synchronizes to this one.
docker run -d --name "$MASTER" --network "$NETWORK" --entrypoint ptp4l "$IMAGE" \
  -f /etc/ptp4l/master.conf -i eth0 -S -m >/dev/null

# SYS_TIME: the slave's clock servo needs to call clock_adjtime/adjtimex on
# its own (containerized) clock to leave LinuxPTP's UNCALIBRATED state.
docker run --rm --network "$NETWORK" --cap-add=SYS_TIME -v "$(pwd):/src" "$IMAGE" "$@"
