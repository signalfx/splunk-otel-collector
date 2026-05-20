#!/bin/bash

set -euox pipefail

arch="${ARCH:-amd64}"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libotelinject_${arch}.so" libotelinject.so
docker buildx build -q --platform linux/${arch} --build-arg TARGETARCH=${arch} -o type=image,name=zeroconfig-test-dotnet,push=false .
OUTPUT=$(docker run --platform linux/${arch} --rm zeroconfig-test-dotnet)
echo "========== OUTPUT =========="
echo "$OUTPUT"
echo "============================"
echo "Test presence of OTEL_SERVICE_NAME (set via ENV, injector must not override)"
echo "$OUTPUT" | grep "OTEL_SERVICE_NAME=iknowmyownservicename" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of SPLUNK_METRICS_ENABLED (set via ENV)"
echo "$OUTPUT" | grep "SPLUNK_METRICS_ENABLED=true" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of ANOTHER_VAR (set via ENV)"
echo "$OUTPUT" | grep "ANOTHER_VAR=foo" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of CORECLR_ENABLE_PROFILING (injected by libotelinject)"
echo "$OUTPUT" | grep "CORECLR_ENABLE_PROFILING=1" > /dev/null && echo "Test passes"  || exit 1

# Check we didn't inject env vars into processes outside of dotnet.
OUTPUT_BASH=$(docker run --platform linux/${arch} --rm zeroconfig-test-dotnet /usr/bin/env)
echo "======= BASH OUTPUT ========"
echo "$OUTPUT_BASH"
echo "============================"
[[ ! "$OUTPUT_BASH" =~ CORECLR_ENABLE_PROFILING ]] && echo "Test passes"  || exit 1
