#!/bin/sh

set -eu

COLLECTOR_BIN="${COLLECTOR_BIN:-/otelcol}"
COLLECTOR_CONFIG="${COLLECTOR_CONFIG:-/etc/otel-collector-delayed-config.yaml}"

if [ ! -x "${COLLECTOR_BIN}" ]; then
  echo "collector binary not found or not executable: ${COLLECTOR_BIN}" >&2
  exit 1
fi

if [ ! -f "${COLLECTOR_CONFIG}" ]; then
  echo "collector config not found: ${COLLECTOR_CONFIG}" >&2
  exit 1
fi

echo "Starting collector with delayed OTLP receiver config: ${COLLECTOR_CONFIG}"
exec "${COLLECTOR_BIN}" --config="${COLLECTOR_CONFIG}" "$@"
