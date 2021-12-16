#!/bin/bash

set -euxo pipefail

SMART_AGENT_RELEASE="${SMART_AGENT_RELEASE:-}"
JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE:-}"
OUTPUT_DIR="${OUTPUT_DIR:-}"
VERSION="${VERSION:-}"

if [ $# -eq 0 ]; then
    if [ -z "$SMART_AGENT_RELEASE" ]; then
        echo "SMART_AGENT_RELEASE env var not set!" >&2
        exit 1
    fi

    if [ -z "$JMX_METRIC_GATHERER_RELEASE" ]; then
        echo "JMX_METRIC_GATHERER_RELEASE env var not set!" >&2
        exit 1
    fi

    if [ -z "$OUTPUT_DIR" ]; then
        echo "OUTPUT_DIR env var not set!" >&2
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo "VERSION env var not set!" >&2
        exit 1
    fi

    buildargs="--output /work/build/stage --smart-agent ${SMART_AGENT_RELEASE#v} --jmx-metric-gatherer ${JMX_METRIC_GATHERER_RELEASE} ${VERSION#v}"
    su wix -c "/project/internal/buildscripts/packaging/msi/msi-builder/build.sh $buildargs"
    mkdir -p $OUTPUT_DIR
    echo "Copying MSI to $OUTPUT_DIR"
    cp /work/build/stage/*.msi $OUTPUT_DIR/
else
    exec "$@"
fi
