#!/bin/bash

set -eo pipefail

export DOCKER_BUILDKIT=1

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$( cd ${SCRIPT_DIR}/../ && pwd )"
OTELCOL_DIR="${REPO_DIR}/cmd/otelcol"
DIST_DIR="${OTELCOL_DIR}/dist"

FIPS="${FIPS:-false}"
PUSH="${PUSH:-false}"
ARCH="${ARCH:-amd64}"
if [ "$FIPS" != "true" ]; then
    IMAGE_NAME="${IMAGE_NAME:-otelcol}"
else
    IMAGE_NAME="${IMAGE_NAME:-otelcol-fips}"
fi
IMAGE_TAG="${IMAGE_TAG:-latest}"
SKIP_COMPILE="${SKIP_COMPILE:-false}"
SKIP_BUNDLE="${SKIP_BUNDLE:-false}"
DOCKER_REPO="${DOCKER_REPO:-docker.io}"
JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE:-}"
MULTIARCH_OTELCOL_BUILDER="${MULTIARCH_OTELCOL_BUILDER:-}"
if [ "$FIPS" != "true" ]; then
    DOCKER_OPTS="--provenance=false --build-arg DOCKER_REPO=${DOCKER_REPO} --build-arg JMX_METRIC_GATHERER_RELEASE=${JMX_METRIC_GATHERER_RELEASE} $OTELCOL_DIR"
else
    OTELCOL_DIR="${OTELCOL_DIR}/fips"
    DIST_DIR="${OTELCOL_DIR}/dist"
    DOCKER_OPTS="--provenance=false --build-arg DOCKER_REPO=${DOCKER_REPO} $OTELCOL_DIR"
fi

CONTAINERD_ENABLED="false"
if docker info -f '{{ .DriverStatus }}' | grep -q "io.containerd.snapshotter"; then
    # containerd image store is required to save a multiarch image locally
    # https://docs.docker.com/storage/containerd/
    CONTAINERD_ENABLED="true"
fi

archs=$(echo $ARCH | tr "," " ")
platforms=""

for arch in $archs; do
    if [ "$FIPS" != "true" ]; then
        if [[ ! "$arch" =~ ^amd64|arm64|ppc64le$ ]]; then
            echo "$arch not supported!" >&2
            exit 1
        fi
    elif [[ ! "$arch" =~ ^amd64|arm64$ ]]; then
        echo "$arch not supported!" >&2
        exit 1
    fi
done

if [ -d "$DIST_DIR" ]; then
    rm -rf "$DIST_DIR"
fi
mkdir -p "$DIST_DIR"

if [ "$FIPS" != "true" ]; then
    cp "${SCRIPT_DIR}/collect-libs.sh" "$DIST_DIR"
fi

for arch in $archs; do
    if [ "$SKIP_COMPILE" != "true" ]; then
        if [ "$FIPS" != "true" ]; then
            make -C "$REPO_DIR" binaries-linux_${arch}
        else
            GOOS=linux GOARCH=$arch make -C "$REPO_DIR" otelcol-fips
        fi
    fi
    if [ "$FIPS" != "true" ]; then
        bins="otelcol_linux_${arch}"
    else
        bins="otelcol-fips_linux_${arch}"
    fi
    for bin in $bins; do
        if [ ! -f "${REPO_DIR}/bin/${bin}" ]; then
            echo "${REPO_DIR}/bin/${bin} not found!" >&2
            exit 1
        fi
        cp "${REPO_DIR}/bin/${bin}" "$DIST_DIR"
    done
    if [ "$FIPS" != "true" ]; then
        if [[ "$arch" =~ ^amd64|arm64$ ]]; then
            if [ "$SKIP_BUNDLE" != "true" ]; then
                make -C "${REPO_DIR}/packaging/bundle" agent-bundle-linux ARCH=${arch} DOCKER_REPO=${DOCKER_REPO}
            fi
        else
            # create a dummy file to copy for the docker build
            touch "${REPO_DIR}/dist/agent-bundle_linux_${arch}.tar.gz"
        fi
        if [ ! -f "${REPO_DIR}/dist/agent-bundle_linux_${arch}.tar.gz" ]; then
            echo "${REPO_DIR}/dist/agent-bundle_linux_${arch}.tar.gz not found!" >&2
            exit 1
        fi
        cp "${REPO_DIR}/dist/agent-bundle_linux_${arch}.tar.gz" "$DIST_DIR"
    fi
    if [[ "$PUSH" = "true" || "$CONTAINERD_ENABLED" = "true" ]]; then
        platforms="${platforms},linux/${arch}"
    else
        # multiarch images must be built and tagged individually when not pushing or not using the containerd image store
        # https://github.com/docker/buildx/issues/59
        docker buildx build --platform linux/${arch} --tag ${IMAGE_NAME}:${arch} --load $DOCKER_OPTS
        docker tag ${IMAGE_NAME}:${arch} ${IMAGE_NAME}:${IMAGE_TAG}
    fi
done

if [ -n "$platforms" ]; then
    if [ -z "$MULTIARCH_OTELCOL_BUILDER" ]; then
        # multiarch builds require a builder with the docker-container driver; create one if not specified
        MULTIARCH_OTELCOL_BUILDER="docker-otelcol"
        if ! docker buildx inspect --builder $MULTIARCH_OTELCOL_BUILDER >/dev/null 2>&1; then
            docker buildx create --name $MULTIARCH_OTELCOL_BUILDER --driver docker-container --bootstrap
        fi
    fi
    if [ "$PUSH" = "true" ]; then
        DOCKER_OPTS="--push $DOCKER_OPTS"
    else
        DOCKER_OPTS="--load $DOCKER_OPTS"
    fi
    docker buildx build --builder $MULTIARCH_OTELCOL_BUILDER --platform ${platforms#,} --tag ${IMAGE_NAME}:${IMAGE_TAG} $DOCKER_OPTS
fi
