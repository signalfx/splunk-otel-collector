#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

HEALTH_HOST="${HEALTH_HOST:-127.0.0.1}"
HEALTH_PORT="${HEALTH_PORT:-13133}"
OTLP_HOST="${OTLP_HOST:-127.0.0.1}"
OTLP_HTTP_PORT="${OTLP_HTTP_PORT:-4318}"
OTLP_GRPC_PORT="${OTLP_GRPC_PORT:-4317}"
OTLP_BACKEND_HTTP_PORT="${OTLP_BACKEND_HTTP_PORT:-14318}"
COLLECTOR_BIND_HOST="${COLLECTOR_BIND_HOST:-127.0.0.1}"

INTERVAL_SECONDS="${INTERVAL_SECONDS:-1}"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-120}"
CURL_MAX_TIME_SECONDS="${CURL_MAX_TIME_SECONDS:-2}"
DELAY_OTLP_START_SECONDS="${DELAY_OTLP_START_SECONDS:-0}"
BUILD_IF_MISSING="${BUILD_IF_MISSING:-true}"
KEEP_TMP="${KEEP_TMP:-true}"

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/otl4172-startup.XXXXXX")"
CONFIG_FILE="${TMP_DIR}/collector.yaml"
COLLECTOR_LOG="${TMP_DIR}/collector.log"
OTLP_RESPONSE_BODY="${TMP_DIR}/otlp-response.out"
HEALTH_CURL_ERR="${TMP_DIR}/health-curl.err"
OTLP_CURL_ERR="${TMP_DIR}/otlp-curl.err"

COLLECTOR_PID=""
PROXY_PID=""

timestamp() {
  date "+%Y-%m-%dT%H:%M:%S%z"
}

cleanup() {
  if [[ -n "${PROXY_PID}" ]] && kill -0 "${PROXY_PID}" 2>/dev/null; then
    kill "${PROXY_PID}" 2>/dev/null || true
    wait "${PROXY_PID}" 2>/dev/null || true
  fi

  if [[ -n "${COLLECTOR_PID}" ]] && kill -0 "${COLLECTOR_PID}" 2>/dev/null; then
    kill "${COLLECTOR_PID}" 2>/dev/null || true
    wait "${COLLECTOR_PID}" 2>/dev/null || true
  fi

  if [[ "${KEEP_TMP}" != "true" ]]; then
    rm -rf "${TMP_DIR}"
  else
    echo "Work dir kept: ${TMP_DIR}"
  fi
}

trap cleanup EXIT

die() {
  echo "error: $*" >&2
  exit 1
}

normalize_goarch() {
  case "$1" in
    x86_64) echo "amd64" ;;
    aarch64) echo "arm64" ;;
    *) echo "$1" ;;
  esac
}

find_collector_bin() {
  if [[ -n "${COLLECTOR_BIN:-}" ]]; then
    [[ -x "${COLLECTOR_BIN}" ]] || die "COLLECTOR_BIN is not executable: ${COLLECTOR_BIN}"
    echo "${COLLECTOR_BIN}"
    return
  fi

  local goos goarch
  goos="$(go env GOOS 2>/dev/null || true)"
  goarch="$(go env GOARCH 2>/dev/null || true)"
  if [[ -z "${goos}" ]]; then
    goos="$(uname -s | tr '[:upper:]' '[:lower:]')"
  fi
  if [[ -z "${goarch}" ]]; then
    goarch="$(normalize_goarch "$(uname -m)")"
  fi

  local candidates=(
    "${ROOT_DIR}/bin/otelcol_${goos}_${goarch}"
    "${ROOT_DIR}/bin/otelcol"
  )

  local candidate
  for candidate in "${candidates[@]}"; do
    if [[ -x "${candidate}" ]]; then
      echo "${candidate}"
      return
    fi
  done

  if [[ "${BUILD_IF_MISSING}" == "true" ]]; then
    echo "Collector binary not found; running: make otelcol" >&2
    make -C "${ROOT_DIR}" otelcol
    for candidate in "${candidates[@]}"; do
      if [[ -x "${candidate}" ]]; then
        echo "${candidate}"
        return
      fi
    done
  fi

  die "collector binary not found. Run 'make otelcol' or set COLLECTOR_BIN=/path/to/otelcol"
}

write_config() {
  local otlp_http_bind_port="$1"

  cat > "${CONFIG_FILE}" <<EOF
extensions:
  health_check:
    endpoint: "${COLLECTOR_BIND_HOST}:${HEALTH_PORT}"

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "${COLLECTOR_BIND_HOST}:${OTLP_GRPC_PORT}"
      http:
        endpoint: "${COLLECTOR_BIND_HOST}:${otlp_http_bind_port}"

processors:
  batch:

exporters:
  debug:
    verbosity: basic

service:
  extensions: [health_check]
  telemetry:
    logs:
      level: debug
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
EOF
}

curl_status() {
  local url="$1"
  shift

  local status
  status="$(curl \
    --silent \
    --show-error \
    --output /dev/null \
    --write-out "HTTP_CODE:%{http_code}\nHTTP_CONNECT:%{http_connect}\nTIME_TOTAL:%{time_total}s\nTIME_CONNECT:%{time_connect}s\nTIME_STARTTRANSFER:%{time_starttransfer}s\nTIME_NAMELOOKUP:%{time_namelookup}s\nRETRY_AFTER:%{retry-after}\nCONTENT_TYPE:%{content_type}" \
    --connect-timeout "${CURL_MAX_TIME_SECONDS}" \
    --max-time "${CURL_MAX_TIME_SECONDS}" \
    "$@" \
    "${url}" 2>/dev/null || true)"

  if [[ -z "${status}" ]]; then
    status="000"
  fi

  # Print complete curl status to command line
  echo "=== Complete curl status ===" >&2
  echo "${status}" >&2
  echo "============================" >&2

  # Extract and return just the HTTP code for compatibility
  echo "${status}" | grep "^HTTP_CODE:" | cut -d: -f2
}

health_status() {
  curl_status "http://${HEALTH_HOST}:${HEALTH_PORT}/" 2>"${HEALTH_CURL_ERR}"
}

otlp_curl_one() {
  local signal="$1"
  local url="$2"
  local payload="$3"
  local body_file="${TMP_DIR}/otlp-${signal}-body.txt"
  local headers_file="${TMP_DIR}/otlp-${signal}-headers.txt"

  local summary
  summary="$(curl -sS \
    -o "${body_file}" \
    -D "${headers_file}" \
    -w "http=%{http_code} exit=%{exitcode} error=%{errormsg}" \
    --connect-timeout "${CURL_MAX_TIME_SECONDS}" \
    --max-time "${CURL_MAX_TIME_SECONDS}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "${url}" 2>"${OTLP_CURL_ERR}" || true)"

  if [[ -z "${summary}" ]]; then
    summary="http=000 exit=999 error=no response"
  fi

  # Print complete info to console
  echo "=== OTLP ${signal} ===" >&2
  echo "  summary: ${summary}" >&2
  if [[ -s "${headers_file}" ]]; then
    echo "  headers:" >&2
    sed 's/^/    /' "${headers_file}" >&2
  fi
  if [[ -s "${body_file}" ]]; then
    echo "  body:" >&2
    sed 's/^/    /' "${body_file}" >&2
  fi
  echo "================================" >&2

  # Extract and return just the http code
  echo "${summary}" | grep -o 'http=[^ ]*' | cut -d= -f2
}

otlp_status() {
  local traces_code metrics_code logs_code

  traces_code="$(otlp_curl_one traces \
    "http://${OTLP_HOST}:${OTLP_HTTP_PORT}/v1/traces" \
    '{"resourceSpans":[]}')"

  metrics_code="$(otlp_curl_one metrics \
    "http://${OTLP_HOST}:${OTLP_HTTP_PORT}/v1/metrics" \
    '{"resourceMetrics":[]}')"

  logs_code="$(otlp_curl_one logs \
    "http://${OTLP_HOST}:${OTLP_HTTP_PORT}/v1/logs" \
    '{"resourceLogs":[]}')"

  # Return 200 only if all endpoints return 200
  if [[ "${traces_code}" == "200" && "${metrics_code}" == "200" && "${logs_code}" == "200" ]]; then
    echo "200"
  else
    [[ "${traces_code}" != "200" ]] && echo "${traces_code}" && return
    [[ "${metrics_code}" != "200" ]] && echo "${metrics_code}" && return
    echo "${logs_code}"
  fi
}

start_delayed_otlp_proxy_if_requested() {
  if [[ "${DELAY_OTLP_START_SECONDS}" == "0" ]]; then
    return
  fi

  echo "Delaying public OTLP HTTP listener by ${DELAY_OTLP_START_SECONDS}s:"
  echo "  public:  ${OTLP_HOST}:${OTLP_HTTP_PORT}"
  echo "  backend: ${COLLECTOR_BIND_HOST}:${OTLP_BACKEND_HTTP_PORT}"

  (
    sleep "${DELAY_OTLP_START_SECONDS}"
    exec socat \
      "TCP-LISTEN:${OTLP_HTTP_PORT},fork,reuseaddr,bind=${OTLP_HOST}" \
      "TCP:${COLLECTOR_BIND_HOST}:${OTLP_BACKEND_HTTP_PORT}"
  ) &
  PROXY_PID="$!"
}

main() {
  local collector_bin
  collector_bin="$(find_collector_bin)"

  if [[ "${DELAY_OTLP_START_SECONDS}" != "0" ]] && ! command -v socat >/dev/null 2>&1; then
    echo "DELAY_OTLP_START_SECONDS=${DELAY_OTLP_START_SECONDS} requested, but socat is not installed." >&2
    echo "Continuing without delayed OTLP exposure; the collector OTLP receiver itself cannot be delayed by config." >&2
    DELAY_OTLP_START_SECONDS=0
  fi

  local otlp_http_bind_port="${OTLP_HTTP_PORT}"
  if [[ "${DELAY_OTLP_START_SECONDS}" != "0" ]]; then
    otlp_http_bind_port="${OTLP_BACKEND_HTTP_PORT}"
  fi

  write_config "${otlp_http_bind_port}"

  echo "Collector binary: ${collector_bin}"
  echo "Collector config: ${CONFIG_FILE}"
  echo "Collector log:    ${COLLECTOR_LOG}"
  echo "Health URL:       http://${HEALTH_HOST}:${HEALTH_PORT}/"
  echo "OTLP HTTP URL:    http://${OTLP_HOST}:${OTLP_HTTP_PORT}/v1/traces"

  "${collector_bin}" --config "${CONFIG_FILE}" >"${COLLECTOR_LOG}" 2>&1 &
  COLLECTOR_PID="$!"

  start_delayed_otlp_proxy_if_requested

  local health="000"
  local otlp="000"
  local start_seconds="${SECONDS}"

  echo "Polling until health=200 and otlp_http=200..."
  while [[ "${health}" != "200" || "${otlp}" != "200" ]]; do
    if ! kill -0 "${COLLECTOR_PID}" 2>/dev/null; then
      echo "Collector exited before probe succeeded. Last collector log lines:" >&2
      tail -n 80 "${COLLECTOR_LOG}" >&2 || true
      exit 1
    fi

    health="$(health_status)"
    otlp="$(otlp_status)"
    printf "\n"
    printf "%s health=%s otlp_http=%s\n" "$(timestamp)" "${health}" "${otlp}"
    printf "\n"
    if [[ "${health}" == "200" && "${otlp}" == "200" ]]; then
      break
    fi

    if (( SECONDS - start_seconds >= TIMEOUT_SECONDS )); then
      echo "Timed out after ${TIMEOUT_SECONDS}s waiting for health=200 and otlp_http=200." >&2
      echo "Last collector log lines:" >&2
      tail -n 80 "${COLLECTOR_LOG}" >&2 || true
      exit 1
    fi

    sleep "${INTERVAL_SECONDS}"
  done

  echo "Ready: health=200 and otlp_http=200"
}

main "$@"
