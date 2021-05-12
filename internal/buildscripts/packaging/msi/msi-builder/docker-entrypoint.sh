#!/bin/bash

set -euo pipefail

SMART_AGENT_RELEASE="${SMART_AGENT_RELEASE#v}"
OUTPUT_DIR="${OUTPUT_DIR:-}"

if [ -z "$SMART_AGENT_RELEASE" ]; then
    echo "SMART_AGENT_RELEASE env var not set!" >&2
    exit 1
fi

if [ -z "$OUTPUT_DIR" ]; then
    echo "OUTPUT_DIR env var not set!" >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    echo "Required version argument not specified!" >&2
    exit 1
fi

buildargs="--output /work/build/stage --smart-agent $SMART_AGENT_RELEASE $@"
su wix -c "/work/build.sh $buildargs"

cp /work/build/stage/*.msi $OUTPUT_DIR/
