#!/bin/bash

# Docker Compose Integration Script for Testing Delayed OTLP Receiver
# This script demonstrates how to use docker-compose with the startup probe script

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="${ROOT_DIR}/tests/repro"

# Configuration
COMPOSE_FILE="${COMPOSE_FILE:-${SCRIPT_DIR}/docker-compose.yml}"
STARTUP_PROBE_SCRIPT="${TESTS_DIR}/otl4172_startup_probe.sh"
EXPECTED_DELAY_SECONDS="${EXPECTED_DELAY_SECONDS:-10}"
COLLECTOR_CONTAINER="${COLLECTOR_CONTAINER:-otelcollector}"
HAPROXY_CONTAINER="${HAPROXY_CONTAINER:-haproxy}"
TELEMETRYGEN_CONTAINER="${TELEMETRYGEN_CONTAINER:-telemetrygen}"
TELEMETRYGEN_HTTP_CONTAINER="${TELEMETRYGEN_HTTP_CONTAINER:-telemetrygen-http}"
OTLP_DIRECT_HTTP_URL="${OTLP_DIRECT_HTTP_URL:-http://localhost:14318/v1/traces}"
OTLP_PROXY_HTTP_URL="${OTLP_PROXY_HTTP_URL:-http://localhost:4318/v1/traces}"
HAPROXY_STATS_URL="${HAPROXY_STATS_URL:-http://localhost:8404/stats;csv}"
PROXY_LOAD_TIMEOUT_SECONDS="${PROXY_LOAD_TIMEOUT_SECONDS:-90}"
STARTUP_ERROR_PATTERN='authentication handshake failed|broken pipe|server preface|code = Unavailable|exporter export timeout|addrConn.createTransport failed|TRANSIENT_FAILURE'
POST_READY_ERROR_PATTERN='authentication handshake failed|broken pipe|server preface|addrConn.createTransport failed|TRANSIENT_FAILURE'
BAD_REQUEST_PATTERN='400[[:space:]]+bad request|bad request|HTTP/[0-9.]*"[[:space:]]+400([[:space:]]|$)|status[=:][[:space:]]*400|code[=:][[:space:]]*400|http=400'
LAST_OTLP_HTTP_CODE="000"

compose() {
    if docker compose version >/dev/null 2>&1; then
        docker compose "$@"
    else
        docker-compose "$@"
    fi
}

help() {
    cat <<EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
  start      Start docker-compose with delayed OTLP receiver
  stop       Stop and cleanup docker containers
  test-delay Test that OTLP receiver actually delays
  test-proxy Test load through HAProxy during delayed OTLP startup
  test-shutdown Test shutdown: health check remains OK while OTLP returns 503
  test-probe Run startup probe against the collector
  logs       Show collector logs
  clean      Remove all containers and volumes
  help       Show this help message

Environment Variables:
  EXPECTED_DELAY_SECONDS   Expected receiver delay in seconds (default: 10)
  OTLP_DIRECT_HTTP_URL     Direct collector OTLP HTTP URL (default: http://localhost:14318/v1/traces)
  PROXY_LOAD_TIMEOUT_SECONDS Timeout for telemetrygen completion (default: 90)
  COMPOSE_FILE             Path to docker-compose file
  KEEP_LOGS                Keep logs on cleanup (default: false)

Examples:
  # Test that receiver delays
  $0 test-delay

  # Test proxy load while the collector starts delayed OTLP
  $0 test-proxy

  # Test shutdown behavior with load-balancer health checks
  $0 test-shutdown

  # Run startup probe
  $0 test-probe

  # View logs
  $0 logs

  # Full cleanup
  $0 clean
EOF
}

start() {
    echo "Starting full docker-compose stack with delayed OTLP receiver config..."
    compose -f "${COMPOSE_FILE}" up -d

    echo ""
    echo "Waiting for collector container to be running..."
    for ((i = 1; i <= 20; i++)); do
        if [[ "$(container_status "${COLLECTOR_CONTAINER}" 2>/dev/null || true)" == "running" ]]; then
            echo "✓ Collector container is running"
            break
        fi

        if [[ "${i}" == "20" ]]; then
            echo "✗ Collector container is not running yet"
            break
        fi

        sleep 1
    done

    echo ""
    echo "Collector started. OTLP receiver is expected after ${EXPECTED_DELAY_SECONDS}s."
    echo "Logs available with: $0 logs"
}

container_exit_code() {
    docker inspect "${1}" --format '{{.State.ExitCode}}'
}

container_status() {
    docker inspect "${1}" --format '{{.State.Status}}'
}

wait_for_container_exit() {
    local container="${1}"
    local timeout_seconds="${2}"
    local deadline=$((SECONDS + timeout_seconds))
    local status

    while (( SECONDS < deadline )); do
        status="$(container_status "${container}" 2>/dev/null || true)"
        if [[ "${status}" == "exited" ]]; then
            return 0
        fi
        if [[ "${status}" == "dead" ]]; then
            return 1
        fi
        sleep 1
    done

    return 1
}

start_collector_only() {
    echo "Starting collector only for direct OTLP delay validation..."
    compose -f "${COMPOSE_FILE}" up -d otelcollector

    for ((i = 1; i <= 10; i++)); do
        if docker ps --filter "name=${COLLECTOR_CONTAINER}" --filter "status=running" -q | grep -q .; then
            echo "✓ Collector container is running"
            return
        fi
        sleep 1
    done

    echo "✗ Collector container did not start"
    docker logs "${COLLECTOR_CONTAINER}" 2>&1 | tail -50 || true
    return 1
}

stop() {
    echo "Stopping docker-compose..."
    compose -f "${COMPOSE_FILE}" down || true
    echo "Stopped."
}

wait_for_otlp_http() {
    local timeout_seconds="${1}"
    local deadline=$((SECONDS + timeout_seconds))
    local code

    while (( SECONDS < deadline )); do
        code="$(curl -sS -o /dev/null -w "%{http_code}" \
            -X POST "${OTLP_DIRECT_HTTP_URL}" \
            -H "Content-Type: application/json" \
            -d '{"resourceSpans":[]}' \
            --max-time 2 2>/dev/null || true)"
        if [[ "${code}" == "200" ]]; then
            return 0
        fi
        sleep 1
    done

    LAST_OTLP_HTTP_CODE="${code:-000}"
    return 1
}

wait_for_delay_log() {
    local timeout_seconds="${1}"
    local deadline=$((SECONDS + timeout_seconds))

    while (( SECONDS < deadline )); do
        if docker logs "${COLLECTOR_CONTAINER}" 2>&1 | grep -q 'Starting OTLP receiver after configured delay'; then
            return 0
        fi
        sleep 1
    done

    return 1
}

wait_for_collector_ready() {
    local timeout_seconds="${1}"
    local deadline=$((SECONDS + timeout_seconds))

    while (( SECONDS < deadline )); do
        if docker logs "${COLLECTOR_CONTAINER}" 2>&1 | grep -q 'Everything is ready'; then
            return 0
        fi
        sleep 1
    done

    return 1
}

haproxy_http_backend_stats() {
    curl -fsS "${HAPROXY_STATS_URL}" 2>/dev/null |
        awk -F, '$1 == "collector_otlp_http" && $2 == "otelcollector" {print}'
}

wait_for_haproxy_http_health_ok() {
    local timeout_seconds="${1}"
    local deadline=$((SECONDS + timeout_seconds))

    while (( SECONDS < deadline )); do
        if haproxy_http_backend_stats |
            awk -F, '$18 ~ /^UP/ && $37 == "L7OK" && $38 == "200" {found = 1} END {exit(found ? 0 : 1)}'; then
            return 0
        fi
        sleep 1
    done

    return 1
}

test_delay() {
    echo "Testing OTLP receiver delay..."
    echo ""

    echo "1. Attempting to send direct OTLP HTTP immediately (should not return 200)..."
    before_code="$(curl -sS -o /dev/null -w "%{http_code}" \
        -X POST "${OTLP_DIRECT_HTTP_URL}" \
        -H "Content-Type: application/json" \
        -d '{"resourceSpans":[]}' \
        --max-time 2 2>/dev/null || true)"
    if [[ "${before_code}" != "200" ]]; then
        echo "   ✓ Direct OTLP HTTP unavailable during delay (http=${before_code:-000})"
    else
        echo "   ✗ Direct OTLP HTTP returned 200 before the receiver delay elapsed"
        return 1
    fi

    echo ""
    echo ""
    echo "2. Waiting ${EXPECTED_DELAY_SECONDS} seconds for receiver to become available..."
    for ((i = EXPECTED_DELAY_SECONDS; i > 0; i--)); do
        echo -ne "\r   Waiting... ${i}s remaining"
        sleep 1
    done
    echo ""

    echo "3. Checking collector logged the configured receiver delay..."
    if wait_for_delay_log 20; then
        echo "   ✓ Collector logged delayed OTLP receiver startup"
    else
        echo "   ✗ Collector delay log not found"
        docker logs "${COLLECTOR_CONTAINER}" 2>&1 | tail -50
        return 1
    fi

    echo ""
    echo "4. Attempting direct OTLP HTTP after delay (should succeed)..."
    if wait_for_otlp_http 20; then
        echo "   ✓ OTLP HTTP successful after delay"
    else
        echo "   ✗ OTLP HTTP still failing (last http=${LAST_OTLP_HTTP_CODE})"
        return 1
    fi
}

assert_no_bad_requests() {
    local logs_file
    logs_file="$(mktemp "${TMPDIR:-/tmp}/otl4172-compose-logs.XXXXXX")"
    compose -f "${COMPOSE_FILE}" logs --no-color > "${logs_file}" 2>&1 || true

    if grep -Ein "${BAD_REQUEST_PATTERN}" "${logs_file}"; then
        echo "   ✗ Found 400 Bad Request in compose logs"
        echo "   Logs captured at: ${logs_file}"
        return 1
    fi

    rm -f "${logs_file}"
    echo "   ✓ No 400 Bad Request entries found in compose logs"
}

first_log_timestamp() {
    local container="${1}"
    local pattern="${2}"

    docker logs --timestamps "${container}" 2>&1 | awk -v pat="${pattern}" 'found == "" && $0 ~ pat {found = $1} END {if (found != "") print found}'
}

assert_telemetrygen_started_before_receiver_ready() {
    local telemetrygen_start_ts
    local receiver_ready_ts

    telemetrygen_start_ts="$(first_log_timestamp "${TELEMETRYGEN_CONTAINER}" 'starting gRPC exporter')"
    receiver_ready_ts="$(first_log_timestamp "${COLLECTOR_CONTAINER}" 'Everything is ready|Starting GRPC server')"

    if [[ -z "${telemetrygen_start_ts}" || -z "${receiver_ready_ts}" ]]; then
        echo "   ✗ Could not determine telemetrygen start or receiver ready timestamp"
        docker logs "${TELEMETRYGEN_CONTAINER}" 2>&1 | tail -40 || true
        docker logs "${COLLECTOR_CONTAINER}" 2>&1 | tail -40 || true
        return 1
    fi

    if [[ "${telemetrygen_start_ts}" < "${receiver_ready_ts}" ]]; then
        echo "   ✓ telemetrygen started before the OTLP receiver was ready"
        echo "     telemetrygen=${telemetrygen_start_ts}, receiver_ready=${receiver_ready_ts}"
        return 0
    fi

    echo "   ✗ telemetrygen waited until after the OTLP receiver was ready"
    echo "     telemetrygen=${telemetrygen_start_ts}, receiver_ready=${receiver_ready_ts}"
    return 1
}

assert_startup_errors_stop_after_ready() {
    local telemetrygen_logs
    local ready_ts
    local startup_errors
    local post_ready_errors

    telemetrygen_logs="$(mktemp "${TMPDIR:-/tmp}/otl4172-telemetrygen-logs.XXXXXX")"
    docker logs --timestamps "${TELEMETRYGEN_CONTAINER}" > "${telemetrygen_logs}" 2>&1 || true

    ready_ts="$(awk '/Channel Connectivity change to READY/ {print $1; exit}' "${telemetrygen_logs}")"
    if [[ -z "${ready_ts}" ]]; then
        echo "   ✗ telemetrygen never reached gRPC READY state"
        echo "   Logs captured at: ${telemetrygen_logs}"
        return 1
    fi

    startup_errors="$(awk -v cutoff="${ready_ts}" -v pat="${STARTUP_ERROR_PATTERN}" '$1 < cutoff && $0 ~ pat {print}' "${telemetrygen_logs}")"
    if [[ -z "${startup_errors}" ]]; then
        echo "   ✗ Did not observe any startup-phase gRPC error before READY"
        echo "   Logs captured at: ${telemetrygen_logs}"
        return 1
    fi

    post_ready_errors="$(awk -v cutoff="${ready_ts}" -v pat="${POST_READY_ERROR_PATTERN}" '$1 > cutoff && $0 !~ /traces export: exporter export timeout/ && $0 ~ pat {print}' "${telemetrygen_logs}")"
    if [[ -n "${post_ready_errors}" ]]; then
        echo "   ✗ Found gRPC connection errors after telemetrygen reached READY"
        echo "${post_ready_errors}"
        echo "   Logs captured at: ${telemetrygen_logs}"
        return 1
    fi

    echo "   ✓ Startup gRPC connection errors were observed before READY and stopped after READY"
    echo "     first_ready=${ready_ts}"
    echo "     startup error sample:"
    awk 'NR <= 3 {print "       " $0}' <<< "${startup_errors}"
    rm -f "${telemetrygen_logs}"
}

assert_no_bad_requests_after_ready() {
    local logs_file
    local ready_ts
    local post_ready_bad_requests

    logs_file="$(mktemp "${TMPDIR:-/tmp}/otl4172-container-logs.XXXXXX")"
    ready_ts="$(first_log_timestamp "${TELEMETRYGEN_CONTAINER}" 'Channel Connectivity change to READY')"

    if [[ -z "${ready_ts}" ]]; then
        echo "   ✗ Cannot check post-ready 400s because telemetrygen never reached READY"
        return 1
    fi

    docker logs --timestamps "${COLLECTOR_CONTAINER}" >> "${logs_file}" 2>&1 || true
    docker logs --timestamps "${HAPROXY_CONTAINER}" >> "${logs_file}" 2>&1 || true
    docker logs --timestamps "${TELEMETRYGEN_CONTAINER}" >> "${logs_file}" 2>&1 || true

    post_ready_bad_requests="$(awk -v cutoff="${ready_ts}" -v pat="${BAD_REQUEST_PATTERN}" '$1 > cutoff && $0 ~ pat {print}' "${logs_file}")"
    if [[ -n "${post_ready_bad_requests}" ]]; then
        echo "   ✗ Found 400 Bad Request after telemetrygen reached READY"
        echo "${post_ready_bad_requests}"
        echo "   Logs captured at: ${logs_file}"
        return 1
    fi

    if grep -Ein "${BAD_REQUEST_PATTERN}" "${logs_file}" >/dev/null; then
        echo "   ✓ 400 Bad Request did not appear after telemetrygen reached READY"
    else
        echo "   ✓ No 400 Bad Request entries found; startup failure was gRPC transport unavailability"
    fi

    rm -f "${logs_file}"
}

test_proxy_load() {
    echo "Testing telemetrygen load through HAProxy during delayed collector startup..."
    echo ""

    echo "1. Waiting for collector delayed receiver startup log..."
    if wait_for_delay_log 30; then
        echo "   ✓ Collector logged delayed OTLP receiver startup"
    else
        echo "   ✗ Collector delay log not found"
        docker logs "${COLLECTOR_CONTAINER}" 2>&1 | tail -80
        return 1
    fi

    echo ""
    echo "2. Validating telemetrygen did not wait for the OTLP receiver..."
    assert_telemetrygen_started_before_receiver_ready

    echo ""
    echo "3. Waiting for telemetrygen to finish sending through HAProxy..."
    if ! wait_for_container_exit "${TELEMETRYGEN_CONTAINER}" "${PROXY_LOAD_TIMEOUT_SECONDS}"; then
        echo "   ✗ telemetrygen did not exit within ${PROXY_LOAD_TIMEOUT_SECONDS}s"
        docker logs "${TELEMETRYGEN_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi

    telemetrygen_exit_code="$(container_exit_code "${TELEMETRYGEN_CONTAINER}")"
    if [[ "${telemetrygen_exit_code}" == "0" ]]; then
        echo "   ✓ telemetrygen completed successfully"
    else
        echo "   ✗ telemetrygen exited with code ${telemetrygen_exit_code}"
        docker logs "${TELEMETRYGEN_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi

    echo ""
    echo "4. Validating startup gRPC errors stop once the proxy path is ready..."
    assert_startup_errors_stop_after_ready

    echo ""
    echo "5. Validating no 400 Bad Request appears after the receiver is ready..."
    assert_no_bad_requests_after_ready
}

assert_shutdown_healthcheck_ok_and_otlp_503() {
    local log_count_before
    local stats_after_shutdown
    local otlp_code
    local shutdown_503

    echo "1. Waiting for collector and HAProxy health check to be ready..."
    if ! wait_for_collector_ready 40; then
        echo "   ✗ Collector did not become ready"
        docker logs "${COLLECTOR_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi

    if ! wait_for_haproxy_http_health_ok 40; then
        echo "   ✗ HAProxy did not record health_check 200 for OTLP HTTP backend"
        curl -fsS "${HAPROXY_STATS_URL}" 2>/dev/null | grep -E '(^#|collector_otlp_http)' || true
        return 1
    fi
    echo "   ✓ HAProxy OTLP HTTP backend health check is L7OK/200"

    if [[ "$(container_status "${TELEMETRYGEN_HTTP_CONTAINER}" 2>/dev/null || true)" != "running" ]]; then
        echo "   ✗ ${TELEMETRYGEN_HTTP_CONTAINER} is not running before collector shutdown"
        docker logs "${TELEMETRYGEN_HTTP_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi
    echo "   ✓ ${TELEMETRYGEN_HTTP_CONTAINER} is running"

    log_count_before="$(docker logs "${HAPROXY_CONTAINER}" 2>&1 | wc -l | tr -d ' ')"

    echo ""
    echo "2. Sending SIGTERM to collector while telemetrygen-http is still running..."
    docker kill --signal=TERM "${COLLECTOR_CONTAINER}" >/dev/null
    sleep 1

    stats_after_shutdown="$(haproxy_http_backend_stats || true)"
    if ! awk -F, '$18 ~ /^UP/ && $37 == "L7OK" && $38 == "200" {found = 1} END {exit(found ? 0 : 1)}' <<< "${stats_after_shutdown}"; then
        echo "   ✗ HAProxy no longer has health_check 200 after collector shutdown"
        echo "${stats_after_shutdown}"
        return 1
    fi
    echo "   ✓ HAProxy still has health_check 200 after collector shutdown"
    echo "     ${stats_after_shutdown}"

    echo ""
    echo "3. Validating OTLP HTTP through HAProxy returns 503..."
    otlp_code="$(curl -sS -o /dev/null -w "%{http_code}" \
        -X POST "${OTLP_PROXY_HTTP_URL}" \
        -H "Content-Type: application/json" \
        -d '{"resourceSpans":[]}' \
        --max-time 2 2>/dev/null || true)"
    if [[ "${otlp_code}" != "503" ]]; then
        echo "   ✗ OTLP HTTP returned ${otlp_code:-000}; expected 503"
        return 1
    fi
    echo "   ✓ Direct OTLP HTTP request through HAProxy returned 503"

    echo ""
    echo "4. Validating telemetrygen-http receives 503 while still running..."
    for ((i = 1; i <= 20; i++)); do
        shutdown_503="$(docker logs --timestamps "${HAPROXY_CONTAINER}" 2>&1 |
            awk -v start="${log_count_before}" '
                NR > start && /otlp_http_in/ && /172\.20\./ && / 503 / && /POST \/v1\/traces/ {
                    print
                    found = 1
                }
                END {exit(found ? 0 : 1)}
            ' || true)"
        if [[ -n "${shutdown_503}" ]]; then
            break
        fi
        sleep 1
    done

    if [[ -z "${shutdown_503}" ]]; then
        echo "   ✗ Did not find telemetrygen-http POST /v1/traces with 503 after shutdown"
        docker logs "${HAPROXY_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi

    if [[ "$(container_status "${TELEMETRYGEN_HTTP_CONTAINER}" 2>/dev/null || true)" != "running" ]]; then
        echo "   ✗ ${TELEMETRYGEN_HTTP_CONTAINER} was not running when shutdown 503 was observed"
        docker logs "${TELEMETRYGEN_HTTP_CONTAINER}" 2>&1 | tail -80 || true
        return 1
    fi

    echo "   ✓ telemetrygen-http POST /v1/traces received 503 after collector shutdown"
    awk 'NR <= 3 {print "     " $0}' <<< "${shutdown_503}"
}

test_shutdown() {
    echo "Testing collector shutdown with health-check-backed load balancer..."
    echo ""
    assert_shutdown_healthcheck_ok_and_otlp_503
}

test_probe() {
    echo "Running startup probe script..."
    echo ""

    if [[ ! -f "${STARTUP_PROBE_SCRIPT}" ]]; then
        echo "Error: Startup probe script not found at: ${STARTUP_PROBE_SCRIPT}"
        exit 1
    fi

    # Adjust OTLP_HOST to match docker network
    export OTLP_HOST="127.0.0.1"
    export OTLP_HTTP_PORT="4318"
    export HEALTH_HOST="127.0.0.1"
    export HEALTH_PORT="13133"
    export DELAY_OTLP_START_SECONDS="0"

    echo "Running: ${STARTUP_PROBE_SCRIPT}"
    "${STARTUP_PROBE_SCRIPT}" "$@"
}

show_logs() {
    echo "Compose logs (last 50 lines per service):"
    echo "=========================================="
    compose -f "${COMPOSE_FILE}" logs -f --tail 50 2>/dev/null || echo "Stack not running"
}

clean() {
    echo "Removing all containers and volumes..."
    compose -f "${COMPOSE_FILE}" down -v

    # Remove dangling containers
    docker container prune -f 2>/dev/null || true

    echo "Cleanup complete."
}

# Parse arguments
COMMAND="${1:-help}"
shift || true

case "${COMMAND}" in
    start)
        start "$@"
        ;;
    stop)
        stop
        ;;
    test-delay)
        trap stop EXIT
        start
        test_delay
        trap - EXIT
        stop
        ;;
  test-proxy)
        trap stop EXIT
        start
        test_proxy_load
        trap - EXIT
        stop
        ;;
    test-shutdown)
        trap stop EXIT
        start
        test_shutdown
        trap - EXIT
        stop
        ;;
    test-probe)
        test_probe "$@"
        ;;
    logs)
        show_logs
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        help
        ;;
    *)
        echo "Unknown command: ${COMMAND}"
        help
        exit 1
        ;;
esac
