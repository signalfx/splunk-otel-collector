#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp "${SCRIPT_DIR}/../../dist/libsplunk_amd64.so" libsplunk.so
docker build -q -t zeroconfig-test-nodejs .
OUTPUT=$(docker run -it zeroconfig-test-nodejs)
echo "$OUTPUT"
echo "Test presence of OTEL_RESOURCE_ATTRIBUTES"
echo $OUTPUT | grep "OTEL_RESOURCE_ATTRIBUTES=foo=bar,this=that" || exit 1
echo "Test presence of OTEL_SERVICE_NAME"
echo $OUTPUT | grep "OTEL_SERVICE_NAME=myservice" || exit 1
echo "Test absence of FOO"
echo $OUTPUT | grep -v "FOO" || exit 1
echo "Test absence of OTEL_EXPORTER_OTLP_ENDPOINT"
echo $OUTPUT | grep -v "OTEL_EXPORTER_OTLP_ENDPOINT" || exit 1