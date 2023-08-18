#!/bin/bash

set -euo pipefail

arch="${ARCH:-amd64}"
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libsplunk_${arch}.so" libsplunk.so
docker build -q -t zeroconfig-test-java .
OUTPUT=$(docker run --rm zeroconfig-test-java)
echo "========== OUTPUT =========="
echo "$OUTPUT"
echo "============================"
echo "Test presence of OTEL_RESOURCE_ATTRIBUTES"
echo $OUTPUT | grep "OTEL_RESOURCE_ATTRIBUTES=foo=bar,this=that" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of OTEL_SERVICE_NAME"
echo $OUTPUT | grep "OTEL_SERVICE_NAME=myservice" > /dev/null && echo "Test passes"  || exit 1
echo "Test presence of SPLUNK_METRICS_ENABLED"
echo $OUTPUT | grep "SPLUNK_METRICS_ENABLED=abcd" > /dev/null && echo "Test passes"  || exit 1
echo "Test absence of FOO"
echo $OUTPUT | grep -v "FOO" > /dev/null && echo "Test passes"  || exit 1
echo "Test absence of OTEL_EXPORTER_OTLP_ENDPOINT"
echo $OUTPUT | grep -v "OTEL_EXPORTER_OTLP_ENDPOINT" > /dev/null && echo "Test passes"  || exit 1
