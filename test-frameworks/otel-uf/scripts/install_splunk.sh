#!/usr/bin/env bash
# Install Splunk Enterprise on VM C and configure it for the test environment:
#   - Two indexes (uf_<ta> and otel_<ta>)
#   - HEC enabled with a known token
#   - S2S receiving enabled on port 9997 via inputs.conf
#   - Splunk_TA_nix installed for search-time field extraction
#
# Usage:
#   ./scripts/install_splunk.sh
#
# Requires: config.env with SPLUNK_HOST, SSH_KEY, SSH_USER, SPLUNK_VERSION,
#           SPLUNK_ADMIN_PASSWORD, SPLUNK_HEC_TOKEN, UF_INDEX, OTEL_INDEX
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$ROOT_DIR/config.env" ]; then
  # shellcheck source=/dev/null
  source "$ROOT_DIR/config.env"
fi

: "${SPLUNK_HOST:?SPLUNK_HOST is not set}"
: "${SSH_KEY:?SSH_KEY is not set}"
: "${SSH_USER:?SSH_USER is not set}"
: "${SPLUNK_VERSION:?SPLUNK_VERSION is not set}"
: "${SPLUNK_ADMIN_PASSWORD:?SPLUNK_ADMIN_PASSWORD is not set}"
: "${SPLUNK_HEC_TOKEN:?SPLUNK_HEC_TOKEN is not set}"
: "${UF_INDEX:?UF_INDEX is not set}"
: "${OTEL_INDEX:?OTEL_INDEX is not set}"
: "${SPLUNK_S2S_PORT:=9997}"
: "${SPLUNK_HEC_PORT:=8088}"
: "${SPLUNK_MGMT_PORT:=8089}"
: "${TA_LIST:=Splunk_TA_nix}"
: "${TA_SOURCE_DIR:=$ROOT_DIR/tas}"

SSH_OPTS="-i ${SSH_KEY} -o StrictHostKeyChecking=no"
SSH="ssh ${SSH_OPTS} ${SSH_USER}@${SPLUNK_HOST}"
SCP="scp ${SSH_OPTS}"

# resolve_ta_dir <TA_NAME>
# Accepts a pre-extracted directory or a .tgz/.tar.gz tarball under TA_SOURCE_DIR.
# Sets global TA_TMPDIR if a tarball was extracted (caller must clean up).
TA_TMPDIR=""
resolve_ta_dir() {
  local ta="$1"
  TA_TMPDIR=""

  # 1. Pre-extracted directory
  if [ -d "$TA_SOURCE_DIR/$ta" ]; then
    echo "$TA_SOURCE_DIR/$ta"
    return 0
  fi

  # 2. Tarball named exactly <TA_NAME>.<ext>
  local tarball=""
  for ext in tgz tar.gz spl; do
    if [ -f "$TA_SOURCE_DIR/$ta.$ext" ]; then
      tarball="$TA_SOURCE_DIR/$ta.$ext"
      break
    fi
  done

  # 3. Any tarball in tas/ whose top-level directory matches the TA name
  if [ -z "$tarball" ]; then
    for f in "$TA_SOURCE_DIR"/*.tgz "$TA_SOURCE_DIR"/*.tar.gz "$TA_SOURCE_DIR"/*.spl; do
      [ -f "$f" ] || continue
      # peek at the first directory entry without extracting
      local top
      top="$(tar -tzf "$f" 2>/dev/null | head -1 | cut -d/ -f1)"
      if [ "$top" = "$ta" ]; then
        tarball="$f"
        break
      fi
    done
  fi

  if [ -z "$tarball" ]; then
    return 1
  fi

  TA_TMPDIR="$(mktemp -d)"
  tar -xzf "$tarball" -C "$TA_TMPDIR"

  local extracted=""
  for d in "$TA_TMPDIR"/*/; do
    [ -d "$d" ] && extracted="$d" && break
  done
  if [ -n "$extracted" ]; then
    echo "${extracted%/}"
  else
    echo "$TA_TMPDIR"
  fi
}

SPLUNK_HOME="/opt/splunk"
SPLUNK_BIN="$SPLUNK_HOME/bin/splunk"
# Build hash is required for some Splunk versions (e.g. 9.4.0 = 6b4ebe426ca6).
# Set SPLUNK_BUILD_HASH in config.env if the plain URL returns 404.
if [ -n "${SPLUNK_BUILD_HASH:-}" ]; then
  SPLUNK_PKG="splunk-${SPLUNK_VERSION}-${SPLUNK_BUILD_HASH}-linux-amd64.tgz"
else
  SPLUNK_PKG="splunk-${SPLUNK_VERSION}-linux-amd64.tgz"
fi
SPLUNK_URL="https://download.splunk.com/products/splunk/releases/${SPLUNK_VERSION}/linux/${SPLUNK_PKG}"

echo "=== install_splunk.sh ==="
echo "Host    : $SPLUNK_HOST"
echo "Version : $SPLUNK_VERSION"
echo "Indexes : $UF_INDEX  $OTEL_INDEX"
echo ""

# ------------------------------------------------------------
# 1. Install Splunk Enterprise
# ------------------------------------------------------------
echo "[1/5] Installing Splunk Enterprise..."

$SSH bash -s << ENDSSH
set -euo pipefail

if [ ! -f "$SPLUNK_BIN" ]; then
  echo "  Downloading $SPLUNK_PKG..."
  curl -fsSL -o "/tmp/$SPLUNK_PKG" "$SPLUNK_URL"

  echo "  Extracting to /opt..."
  sudo tar -xzf "/tmp/$SPLUNK_PKG" -C /opt
  rm "/tmp/$SPLUNK_PKG"
else
  echo "  Binary already present at $SPLUNK_BIN, skipping download."
fi

# Stop any leftover process holding ports (safe to run even on first install)
sudo "$SPLUNK_BIN" stop 2>/dev/null || true
sleep 2

# Write user-seed.conf - sets admin password on first start.
# Idempotent: overwriting it before each start is harmless.
sudo mkdir -p "$SPLUNK_HOME/etc/system/local"
sudo tee "$SPLUNK_HOME/etc/system/local/user-seed.conf" > /dev/null << EOF
[user_info]
USERNAME = admin
PASSWORD = ${SPLUNK_ADMIN_PASSWORD}
EOF

echo "  Starting Splunk (accept license)..."
sudo "$SPLUNK_BIN" start --accept-license --answer-yes --no-prompt

echo "  Enabling boot start..."
sudo "$SPLUNK_BIN" enable boot-start --accept-license --answer-yes --no-prompt
ENDSSH

echo "  Done."

# ------------------------------------------------------------
# 2. Wait for Splunk to be ready
# ------------------------------------------------------------
echo ""
echo "[2/5] Waiting for Splunk to be ready..."

$SSH bash -s << ENDSSH
set -euo pipefail
timeout=300
elapsed=0
while true; do
  HTTP_CODE=\$(curl -sk -o /dev/null -w "%{http_code}" \
    "https://localhost:${SPLUNK_MGMT_PORT}/services" \
    -u "admin:${SPLUNK_ADMIN_PASSWORD}" 2>/dev/null || echo "000")
  if [ "\$HTTP_CODE" = "200" ] || [ "\$HTTP_CODE" = "401" ]; then
    break
  fi
  if [ \$elapsed -ge \$timeout ]; then
    echo ""
    echo "ERROR: Splunk did not become ready within \${timeout}s (last HTTP code: \$HTTP_CODE)"
    echo "--- splunkd.log (last 30 lines) ---"
    tail -30 /opt/splunk/var/log/splunk/splunkd.log 2>/dev/null || echo "(log not found)"
    echo "--- splunk status ---"
    sudo /opt/splunk/bin/splunk status 2>/dev/null || true
    exit 1
  fi
  sleep 5
  elapsed=\$((elapsed + 5))
  echo -n "."
done
echo ""
echo "  Splunk is ready (HTTP \$HTTP_CODE)."
ENDSSH

# ------------------------------------------------------------
# 3. Create indexes
# ------------------------------------------------------------
echo ""
echo "[3/5] Creating indexes..."

$SSH bash -s << ENDSSH
set -euo pipefail
AUTH="admin:${SPLUNK_ADMIN_PASSWORD}"
MGMT="https://localhost:${SPLUNK_MGMT_PORT}"

# Verify credentials before proceeding
HTTP_CODE=\$(curl -sk -o /dev/null -w "%{http_code}" -u "\$AUTH" "\$MGMT/services")
if [ "\$HTTP_CODE" != "200" ]; then
  echo "ERROR: Splunk REST API returned HTTP \$HTTP_CODE - check SPLUNK_ADMIN_PASSWORD in config.env"
  exit 1
fi

for INDEX in "${UF_INDEX}" "${OTEL_INDEX}"; do
  HTTP=\$(curl -sk -o /dev/null -w "%{http_code}" -u "\$AUTH" "\$MGMT/services/data/indexes/\$INDEX")
  if [ "\$HTTP" = "200" ]; then
    echo "  Index \$INDEX already exists, skipping."
  else
    curl -sf -k -u "\$AUTH" "\$MGMT/services/data/indexes" \
      -d name="\$INDEX" -d datatype=event > /dev/null
    echo "  Created index: \$INDEX"
  fi
done
ENDSSH

# ------------------------------------------------------------
# 4. Enable HEC
# ------------------------------------------------------------
echo ""
echo "[4/5] Enabling HEC..."

$SSH bash -s << ENDSSH
set -euo pipefail
AUTH="admin:${SPLUNK_ADMIN_PASSWORD}"
MGMT="https://localhost:${SPLUNK_MGMT_PORT}"

# Enable HEC globally
OUT=\$(curl -sk -w "\n%{http_code}" -u "\$AUTH" "\$MGMT/services/data/inputs/http/http" \
  -d disabled=0 -d port="${SPLUNK_HEC_PORT}" -d enableSSL=0)
HTTP=\$(echo "\$OUT" | tail -1)
[ "\$HTTP" = "200" ] || { echo "ERROR: HEC global enable returned HTTP \$HTTP"; echo "\$OUT"; exit 1; }
echo "  HEC enabled on port ${SPLUNK_HEC_PORT} (SSL disabled)."

# Create or update the HEC token using the value from config.env.
TOKEN_NAME="otel-uf-tests"

EXISTS=\$(curl -sk -o /dev/null -w "%{http_code}" -u "\$AUTH" "\$MGMT/services/data/inputs/http/\$TOKEN_NAME")
if [ "\$EXISTS" = "200" ]; then
  OUT=\$(curl -sk -w "\n%{http_code}" -u "\$AUTH" "\$MGMT/services/data/inputs/http/\$TOKEN_NAME" \
    -d token="${SPLUNK_HEC_TOKEN}" -d disabled=0 \
    -d index="${OTEL_INDEX}" -d index="${UF_INDEX}" \
    -d defaultIndex="${OTEL_INDEX}")
  HTTP=\$(echo "\$OUT" | tail -1)
  [ "\$HTTP" = "200" ] || { echo "ERROR: HEC token update returned HTTP \$HTTP"; echo "\$OUT"; exit 1; }
  echo "  HEC token updated: \$TOKEN_NAME (defaultIndex=${OTEL_INDEX})"
else
  OUT=\$(curl -sk -w "\n%{http_code}" -u "\$AUTH" "\$MGMT/services/data/inputs/http" \
    -d name="\$TOKEN_NAME" -d token="${SPLUNK_HEC_TOKEN}" -d disabled=0 \
    -d index="${OTEL_INDEX}" -d index="${UF_INDEX}" \
    -d defaultIndex="${OTEL_INDEX}")
  HTTP=\$(echo "\$OUT" | tail -1)
  [ "\$HTTP" = "201" ] || { echo "ERROR: HEC token create returned HTTP \$HTTP"; echo "\$OUT"; exit 1; }
  echo "  HEC token created: \$TOKEN_NAME (defaultIndex=${OTEL_INDEX})"
fi
ENDSSH

# ------------------------------------------------------------
# 5. Enable S2S receiving
# ------------------------------------------------------------
echo ""
echo "[5/6] Enabling S2S receiving on port ${SPLUNK_S2S_PORT}..."

$SSH bash -s << ENDSSH
set -euo pipefail
INPUTS="$SPLUNK_HOME/etc/system/local/inputs.conf"
sudo mkdir -p "$(dirname "\$INPUTS")"
if ! sudo grep -q "^\[splunktcp://${SPLUNK_S2S_PORT}\]" "\$INPUTS" 2>/dev/null; then
  sudo tee -a "\$INPUTS" > /dev/null << EOF

[splunktcp://${SPLUNK_S2S_PORT}]
disabled = 0
EOF
  echo "  S2S receiving enabled on port ${SPLUNK_S2S_PORT}."
else
  echo "  S2S port ${SPLUNK_S2S_PORT} already in inputs.conf, skipping."
fi
ENDSSH

# ------------------------------------------------------------
# 6. Install TAs for search-time field extraction
# ------------------------------------------------------------
echo ""
echo "[6/6] Installing TAs on search head for search-time extraction..."

for TA in $TA_LIST; do
  TA_DIR="$(resolve_ta_dir "$TA")" || {
    echo "  WARNING: TA source not found for $TA (checked directory and .tgz/.tar.gz), skipping."
    continue
  }

  echo "  Deploying $TA (from $TA_DIR)..."
  $SSH "sudo mkdir -p $SPLUNK_HOME/etc/apps/$TA"
  $SCP -r "$TA_DIR/." "${SSH_USER}@${SPLUNK_HOST}:/tmp/${TA}_upload"
  $SSH "sudo cp -r /tmp/${TA}_upload/. $SPLUNK_HOME/etc/apps/$TA/ && rm -rf /tmp/${TA}_upload"
  [ -n "$TA_TMPDIR" ] && rm -rf "$TA_TMPDIR" && TA_TMPDIR=""

  # Apply inputs.conf override from tas/<TA>.inputs.conf if it exists
  OVERLAY_FILE="$TA_SOURCE_DIR/$TA.inputs.conf"
  if [ -f "$OVERLAY_FILE" ]; then
    echo "  Applying inputs.conf override from $OVERLAY_FILE..."
    $SSH "sudo mkdir -p $SPLUNK_HOME/etc/apps/$TA/local"
    $SCP "$OVERLAY_FILE" "${SSH_USER}@${SPLUNK_HOST}:/tmp/${TA}_inputs.conf"
    $SSH "sudo mv /tmp/${TA}_inputs.conf $SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"
  fi

  # The search head also receives the TA for props.conf/transforms.conf. Keep
  # any templated input config valid there, even though collection is driven
  # by VM A and VM B.
  $SSH bash -s << INNERSH
set -euo pipefail
for CONF in \
  "$SPLUNK_HOME/etc/apps/$TA/default/inputs.conf" \
  "$SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"; do
  if [ -f "\$CONF" ]; then
    sudo sed -i "s/__INDEX__/${OTEL_INDEX}/g" "\$CONF"
  fi
done
INNERSH
  echo "  $TA deployed."
done

echo ""
echo "  Restarting Splunk to apply changes..."
$SSH "sudo $SPLUNK_BIN restart --accept-license --answer-yes --no-prompt"

echo ""
echo "=== install_splunk.sh complete ==="
echo "  Splunk UI  : http://${SPLUNK_HOST}:8000  (admin / ${SPLUNK_ADMIN_PASSWORD})"
echo "  HEC        : http://${SPLUNK_HOST}:${SPLUNK_HEC_PORT}/services/collector"
echo "  S2S port   : ${SPLUNK_S2S_PORT}"
echo "  Indexes    : ${UF_INDEX}  ${OTEL_INDEX}"
