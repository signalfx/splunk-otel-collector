#!/bin/sh
# Creates a Splunk HEC token via the local management REST API and writes its
# value to SECRETS_DIR/hec-token. Intended to run inside the Splunk container
# via a Docker Compose post_start lifecycle hook.
#
# If the token already exists the existing value is retrieved instead.
#
# Usage: sh generate-hec-token.sh [<secrets-dir>]   (default: /secrets)
#
# Environment variables:
#   SPLUNK_URL       Management API base URL  (default: https://splunk:8089)
#   SPLUNK_USER      Admin username           (default: admin)
#   SPLUNK_PASSWORD  Admin password           (required, no default)

set -eu

SECRETS_DIR="${1:-/secrets}"
SPLUNK_URL="${SPLUNK_URL:-https://splunk:8089}"
SPLUNK_USER="${SPLUNK_USER:-admin}"
TOKEN_NAME="otelcollector"

mkdir -p "$SECRETS_DIR"

echo "==> Waiting for Splunk management API at ${SPLUNK_URL} ..."
until curl -sk -o /dev/null -w "%{http_code}" \
        -u "${SPLUNK_USER}:${SPLUNK_PASSWORD}" \
        "${SPLUNK_URL}" | grep -q "200"; do
    sleep 3
done
echo " Splunk is ready."

echo "==> Enabling HEC globally..."
curl -sk -u "${SPLUNK_USER}:${SPLUNK_PASSWORD}" \
    -d 'disabled=0' \
    "${SPLUNK_URL}/services/data/inputs/http/http" > /dev/null

# Try to create the token; fall back to GET if it already exists (HTTP 409).
echo "==> Creating HEC token '${TOKEN_NAME}'..."
RESPONSE=$(curl -sk -u "${SPLUNK_USER}:${SPLUNK_PASSWORD}" \
    -d "name=${TOKEN_NAME}&index=main&disabled=0" \
    "${SPLUNK_URL}/services/data/inputs/http?output_mode=json")
echo "==> Got API response"

TOKEN=$(printf '%s' "$RESPONSE" \
    | grep -oE '"token":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"' \
    | head -1 \
    | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "===> Extracted token"

if [ -z "$TOKEN" ]; then
    echo "==> Token already exists, retrieving it..."
    RESPONSE=$(curl -sk -u "${SPLUNK_USER}:${SPLUNK_PASSWORD}" \
        "${SPLUNK_URL}/services/data/inputs/http/${TOKEN_NAME}?output_mode=json")

    TOKEN=$(printf '%s' "$RESPONSE" \
        | grep -oE '"token":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"' \
        | head -1 \
        | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
fi

if [ -z "$TOKEN" ]; then
    echo "ERROR: failed to obtain HEC token. Last API response:"
    printf '%s\n' "$RESPONSE"
    exit 1
fi

# Write without a trailing newline so ${file:/secrets/hec-token} expands cleanly.
echo "==> Writing token to ${SECRETS_DIR}/hec-token"
printf '%s' "$TOKEN" > "${SECRETS_DIR}/hec-token"
# Mode 644: the named volume is the access control boundary; any process that
# can mount this volume (otelcollector) may read the token.
chmod 644 "${SECRETS_DIR}/hec-token"

echo "==> Done. Waiting 30s for Splunk to fully stabilize before otelcollector starts..."
sleep 30
