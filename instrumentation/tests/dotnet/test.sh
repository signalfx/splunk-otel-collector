#!/bin/bash

set -euox pipefail

arch="${ARCH:-amd64}"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libsplunk_${arch}.so" libsplunk.so
docker buildx build -q --platform linux/${arch} -o type=image,name=zeroconfig-test-dotnet,push=false .
OUTPUT=$(docker run --platform linux/${arch} --rm zeroconfig-test-dotnet)
echo "========== OUTPUT =========="
echo "$OUTPUT"
echo "============================"
echo "Test presence of OTEL_RESOURCE_ATTRIBUTES"
echo "$OUTPUT" | grep "OTEL_RESOURCE_ATTRIBUTES=foo=bar,this=that" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of OTEL_SERVICE_NAME"
echo "$OUTPUT" | grep "OTEL_SERVICE_NAME=iknowmyownservicename" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of SPLUNK_METRICS_ENABLED"
echo "$OUTPUT" | grep "SPLUNK_METRICS_ENABLED=abcd" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of ANOTHER_VAR"
echo "$OUTPUT" | grep "ANOTHER_VAR=foo" > /dev/null && echo "Test passes"  || exit 1
echo "Test absence of FOO"
echo "$OUTPUT" | grep -v "FOO" > /dev/null && echo "Test passes"  || exit 1
echo "Test absence of OTEL_EXPORTER_OTLP_ENDPOINT"
echo "$OUTPUT" | grep -v "OTEL_EXPORTER_OTLP_ENDPOINT" > /dev/null && echo "Test passes"  || exit 1

# Check we didn't inject env vars into processes outside of dotnet.
OUTPUT_BASH=$(docker run --platform linux/${arch} --rm zeroconfig-test-dotnet /usr/bin/env)
echo "======= BASH OUTPUT ========"
echo "$OUTPUT_BASH"
echo "============================"
echo "$OUTPUT_BASH" | grep -v "OTEL_RESOURCE_ATTRIBUTES" | grep -v "OTEL_SERVICE_NAME" | grep -v "SPLUNK_METRICS_ENABLED" > /dev/null && echo "Test passes"  || exit 1
