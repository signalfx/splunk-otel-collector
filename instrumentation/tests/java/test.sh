#!/bin/bash

set -euo pipefail

arch="${ARCH:-amd64}"
BASE="eclipse-temurin:11"
if [[ "$arch" = "arm64" ]]; then
  BASE="arm64v8/eclipse-temurin:11"
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libotelinject_${arch}.so" libotelinject.so
docker buildx build -q --platform linux/${arch} --build-arg BASE=$BASE -o type=image,name=zeroconfig-test-java,push=false .
OUTPUT=$(docker run --platform linux/${arch} --rm zeroconfig-test-java)
echo "========== OUTPUT =========="
echo "$OUTPUT"
echo "============================"
echo "Test presence of OTEL_SERVICE_NAME (set via ENV, injector must not override)"
echo "$OUTPUT" | grep "OTEL_SERVICE_NAME=iknowmyownservicename" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of SPLUNK_METRICS_ENABLED (set via ENV)"
echo "$OUTPUT" | grep "SPLUNK_METRICS_ENABLED=true" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of ANOTHER_VAR (set via ENV)"
echo "$OUTPUT" | grep "ANOTHER_VAR=foo" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of JAVA_TOOL_OPTIONS (injected by libotelinject)"
echo "$OUTPUT" | grep "JAVA_TOOL_OPTIONS=" > /dev/null && echo "Test passes"  || exit 1

# Check we didn't inject env vars into processes outside of java.
OUTPUT_BASH=$(docker run --platform linux/${arch} --rm zeroconfig-test-java /usr/bin/env)
echo "======= BASH OUTPUT ========"
echo "$OUTPUT_BASH"
echo "============================"
[[ ! "$OUTPUT_BASH" =~ JAVA_TOOL_OPTIONS ]] && echo "Test passes"  || exit 1
