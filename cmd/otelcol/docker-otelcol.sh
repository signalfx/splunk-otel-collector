#!/bin/bash

set -exo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$( cd ${SCRIPT_DIR}/../../ && pwd )"

ARCH="${ARCH:-amd64}"
SKIP_COMPILE="${SKIP_COMPILE:-false}"
SKIP_BUNDLE="${SKIP_BUNDLE:-false}"
DOCKER_REPO="${DOCKER_REPO:-docker.io}"
JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE:-}"

if [ "$SKIP_COMPILE" != "true" ]; then
    for arch in $(echo $ARCH | tr "," " "); do
        make -C $REPO_DIR binaries-linux_${arch}
	done
fi

if [ "$SKIP_BUNDLE" != "true" ]; then
    for arch in $(echo $ARCH | tr "," " "); do
        if [[ "${arch}" = "amd64" || "$arch" = "arm64" ]]; then
            make -C ${REPO_DIR}/internal/signalfx-agent/bundle agent-bundle-linux ARCH=${arch} DOCKER_REPO=${DOCKER_REPO}
        fi
    done
fi

platforms=""
for arch in $(echo $ARCH | tr "," " "); do
    platforms="${platforms},linux/${arch}"
done

docker buildx create --name docker-otelcol --driver docker-container --use || true
docker buildx build --builder docker-otelcol --platform ${platforms#,} --load --tag otelcol:latest --build-arg DOCKER_REPO=${DOCKER_REPO} --build-arg JMX_METRIC_GATHERER_RELEASE=${JMX_METRIC_GATHERER_RELEASE} -f ${REPO_DIR}/cmd/otelcol/Dockerfile $REPO_DIR
