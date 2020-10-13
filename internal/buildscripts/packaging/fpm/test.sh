#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/common.sh

PKG_PATH="${1:-}"


if [[ -z "$PKG_PATH" ]]; then
    echo "usage: ${BASH_SOURCE[0]} DEB_OR_RPM_PATH" >&2
    exit 1
fi

if [[ ! -f "$PKG_PATH" ]]; then
    echo "$PKG_PATH not found!" >&2
    exit 1
fi

pkg_base="$( basename "$PKG_PATH" )"
pkg_type="${pkg_base##*.}"
if [[ ! "$pkg_type" =~ ^(deb|rpm)$ ]]; then
    echo "$PKG_PATH not supported!" >&2
    exit 1
fi
image_name="splunk-otel-collector-$pkg_type-test"
container_name="$image_name"
docker_run_opts="--name $container_name -d -v /sys/fs/cgroup:/sys/fs/cgroup:ro --privileged"
docker_run="docker run $docker_run_opts $image_name"
docker_exec="docker exec $container_name"

trap "docker rm -fv $container_name >/dev/null 2>&1 || true" EXIT

docker build -t $image_name -f "$SCRIPT_DIR/$pkg_type/Dockerfile.test" "$SCRIPT_DIR"
docker rm -fv $container_name >/dev/null 2>&1 || true

# test install
echo
$docker_run
install_pkg $container_name "$PKG_PATH"

# ensure service has not started after initial install
sleep 5
echo "Checking $SERVICE_NAME service status ..."
if $docker_exec systemctl --no-pager status $SERVICE_NAME; then
    $docker_exec journalctl -u $SERVICE_USER --no-pager
    echo "$SERVICE_NAME service started after initial install but should not be"
    exit 1
fi

# start the service with the sample env file
$docker_exec cp /etc/otel/collector/splunk_env.example /etc/otel/collector/splunk_env
$docker_exec systemctl daemon-reload
$docker_exec systemctl start splunk-otel-collector.service

# ensure service has started and still running after 5 seconds
sleep 5
echo "Checking $SERVICE_NAME service status ..."
if ! $docker_exec systemctl --no-pager status $SERVICE_NAME; then
    $docker_exec journalctl -u $SERVICE_USER --no-pager
    echo "$SERVICE_NAME service not started"
    exit 1
fi

echo "Checking $PROCESS_NAME process is running as $SERVICE_USER user ..."
$docker_exec pgrep -a -u $SERVICE_USER $PROCESS_NAME

# test reinstall
install_pkg $container_name "$PKG_PATH"

# ensure service has started and still running after 5 seconds
sleep 5
echo "Checking $SERVICE_NAME service status ..."
if ! $docker_exec systemctl --no-pager status $SERVICE_NAME; then
    $docker_exec journalctl -u $SERVICE_USER --no-pager
    echo "$SERVICE_NAME service not started"
    exit 1
fi

echo "Checking $PROCESS_NAME process is running as $SERVICE_USER user ..."
$docker_exec pgrep -a -u $SERVICE_USER $PROCESS_NAME

# test uninstall
echo
uninstall_pkg $container_name $pkg_type

echo "Checking $SERVICE_NAME service status after uninstall ..."
if $docker_exec systemctl --no-pager status $SERVICE_NAME; then
    echo "$SERVICE_NAME service still running after uninstall" >&2
    exit 1
fi
echo "$SERVICE_NAME service successfully stopped after uninstall"

echo "Checking $SERVICE_NAME service existence after uninstall ..."
if $docker_exec systemctl list-unit-files --all | grep $SERVICE_NAME; then
    echo "$SERVICE_NAME service still exists after uninstall" >&2
    exit 1
fi
echo "$SERVICE_NAME service successfully removed after uninstall"

echo "Checking $PROCESS_NAME process after uninstall ..."
if $docker_exec pgrep $PROCESS_NAME; then
    echo "$PROCESS_NAME process still running after uninstall"
    exit 1
fi
echo "$PROCESS_NAME process successfully killed after uninstall"
