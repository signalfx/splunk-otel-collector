#!/bin/bash

set -euo pipefail

arch="${ARCH:-amd64}"
BASE="node:16"
if [[ "$arch" = "arm64" ]]; then
  BASE="arm64v8/node:16"
fi
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libotelinject_${arch}.so" libotelinject.so
docker buildx build -q --platform linux/${arch} --build-arg BASE=$BASE -o type=image,name=zeroconfig-test-nodejs,push=false .
OUTPUT=$(docker run --platform linux/${arch} --rm zeroconfig-test-nodejs)
echo "========== OUTPUT =========="
echo "$OUTPUT"
echo "============================"
echo "Test presence of SPLUNK_METRICS_ENABLED (set via ENV)"
echo "$OUTPUT" | grep "SPLUNK_METRICS_ENABLED=true" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of NODE_OPTIONS (injected by libotelinject)"
echo "$OUTPUT" | grep "NODE_OPTIONS=" > /dev/null && echo "Test passes"  || exit 1
