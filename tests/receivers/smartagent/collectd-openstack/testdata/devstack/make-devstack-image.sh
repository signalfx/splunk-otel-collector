#!/bin/bash -ex
# Builds a devstack Docker image with OpenStack pre-installed.
# This script must be run once before running the integration tests.
# The resulting image is tagged as "devstack:latest".
#
# Requirements:
#   - Docker with privileged container support
#   - At least 20GB of free disk space
#   - A reliable internet connection (downloads ~2GB of packages)
#
# Usage:
#   cd tests/receivers/smartagent/collectd-openstack/testdata/devstack
#   ./make-devstack-image.sh

docker build -t devstack:builder .
docker run -d --privileged \
    --name devstack \
    -v /lib/modules:/lib/modules:ro \
    -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
    -e container=docker \
    devstack:builder

docker exec devstack systemctl start devstack
docker exec devstack systemctl stop devstack

docker export -o /tmp/devstack.tar devstack
# entrypoint doesn't seem to persist for all exports
docker import -c 'ENTRYPOINT ["/lib/systemd/systemd"]' /tmp/devstack.tar devstack:latest

docker rm -fv devstack
docker rmi devstack:builder
rm -f /tmp/devstack.tar
