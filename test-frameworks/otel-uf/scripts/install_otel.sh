#!/usr/bin/env bash
# Install the Splunk OTel Collector on VM B and configure it to run TAs via tarunner.
#
# What this script does:
#   1. Installs TA runtime dependencies (tools required by Splunk_TA_nix scripts)
#   2. Copies the otelcol binary to VM B
#   3. Creates the $SPLUNK_HOME layout (etc/apps/, etc/system/local/)
#   4. Deploys each TA from TA_SOURCE_DIR
#   5. Writes outputs.conf pointing to VM C via HEC, routing to OTEL_INDEX
#   6. Writes the collector config (splunk_inputs + splunk_outputs)
#   7. Installs and starts a systemd service
#
# Usage:
#   ./scripts/install_otel.sh
#
# Requires: config.env with OTEL_HOST, SSH_KEY, SSH_USER, OTEL_SPLUNK_HOME,
#           OTEL_BIN_DIR, SPLUNK_HOST, SPLUNK_HEC_PORT, SPLUNK_HEC_TOKEN,
#           OTEL_INDEX, TA_LIST, TA_SOURCE_DIR
#           bin/otelcol_linux_amd64 (produced by build_otel.sh)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$ROOT_DIR/config.env" ]; then
  # shellcheck source=/dev/null
  source "$ROOT_DIR/config.env"
fi

: "${OTEL_HOST:?OTEL_HOST is not set}"
: "${SSH_KEY:?SSH_KEY is not set}"
: "${SSH_USER:?SSH_USER is not set}"
: "${OTEL_SPLUNK_HOME:?OTEL_SPLUNK_HOME is not set}"
: "${OTEL_BIN_DIR:=/usr/local/bin}"
: "${SPLUNK_HOST:?SPLUNK_HOST is not set}"
: "${SPLUNK_HEC_PORT:=8088}"
: "${SPLUNK_HEC_TOKEN:?SPLUNK_HEC_TOKEN is not set}"
: "${OTEL_INDEX:?OTEL_INDEX is not set}"
: "${TA_LIST:=Splunk_TA_nix}"
: "${TA_SOURCE_DIR:=$ROOT_DIR/tas}"

SSH_OPTS="-i ${SSH_KEY} -o StrictHostKeyChecking=no"
SSH="ssh ${SSH_OPTS} ${SSH_USER}@${OTEL_HOST}"
SCP="scp ${SSH_OPTS}"

# resolve_ta_dir <TA_NAME>
# Prints the local directory containing the TA's files.
# Accepts either:
#   $TA_SOURCE_DIR/<TA_NAME>/          (pre-extracted directory)
#   $TA_SOURCE_DIR/<TA_NAME>.tgz       (tarball - extracted to a temp dir)
#   $TA_SOURCE_DIR/<TA_NAME>.tar.gz    (tarball - extracted to a temp dir)
# Sets global TA_TMPDIR to the temp dir (if any) so the caller can clean up.
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

LOCAL_BINARY="$ROOT_DIR/bin/otelcol_linux_amd64"
SERVICE_NAME="otelcol"
CONFIG_FILE="$OTEL_SPLUNK_HOME/etc/otel_collector.yaml"

if [ ! -f "$LOCAL_BINARY" ]; then
  echo "ERROR: binary not found at $LOCAL_BINARY - run 'make build' first."
  exit 1
fi

echo "=== install_otel.sh ==="
echo "Host        : $OTEL_HOST"
echo "SPLUNK_HOME : $OTEL_SPLUNK_HOME"
echo "HEC target  : http://$SPLUNK_HOST:$SPLUNK_HEC_PORT -> index=$OTEL_INDEX"
echo "TAs         : $TA_LIST"
echo ""

# ------------------------------------------------------------
# 1. Install TA runtime dependencies
# ------------------------------------------------------------
echo "[1/7] Installing TA runtime dependencies..."

# Required by Splunk_TA_nix scripted inputs (net-tools, lsof, sysstat, auditd,
# ntpdate, lastlog2). Harmless on hosts that don't use those inputs.
# lastlog2 is the newer package name; fall back to the older variant if absent.
$SSH bash -s << 'ENDSSH'
set -euo pipefail
sudo apt-get update -qq
sudo apt-get install -y --no-install-recommends \
  net-tools lsof sysstat auditd ntpsec-ntpdate lastlog2 2>/dev/null || \
sudo apt-get install -y --no-install-recommends \
  net-tools lsof sysstat auditd ntpdate
echo "  Dependencies installed."
ENDSSH

# ------------------------------------------------------------
# 2. Copy otelcol binary
# ------------------------------------------------------------
echo ""
echo "[2/7] Copying otelcol binary..."

$SCP "$LOCAL_BINARY" "${SSH_USER}@${OTEL_HOST}:/tmp/otelcol"
$SSH "sudo install -m 755 /tmp/otelcol $OTEL_BIN_DIR/otelcol && rm /tmp/otelcol"
echo "  Binary installed at $OTEL_BIN_DIR/otelcol."

# ------------------------------------------------------------
# 3. Create SPLUNK_HOME layout
# ------------------------------------------------------------
echo ""
echo "[3/7] Creating SPLUNK_HOME layout at $OTEL_SPLUNK_HOME..."

$SSH bash -s << ENDSSH
set -euo pipefail
sudo mkdir -p "$OTEL_SPLUNK_HOME/etc/apps"
sudo mkdir -p "$OTEL_SPLUNK_HOME/etc/system/default"
sudo mkdir -p "$OTEL_SPLUNK_HOME/etc/system/local"
sudo chown -R "${SSH_USER}:${SSH_USER}" "$OTEL_SPLUNK_HOME"
echo "  Layout created."
ENDSSH

# ------------------------------------------------------------
# 4. Deploy TAs
# ------------------------------------------------------------
echo ""
echo "[4/7] Deploying TAs..."

for TA in $TA_LIST; do
  TA_DIR="$(resolve_ta_dir "$TA")" || {
    echo "  WARNING: TA source not found for $TA (checked directory and .tgz/.tar.gz), skipping."
    continue
  }

  echo "  Deploying $TA (from $TA_DIR)..."

  $SSH "mkdir -p $OTEL_SPLUNK_HOME/etc/apps/$TA"
  $SCP -r "$TA_DIR/." "${SSH_USER}@${OTEL_HOST}:$OTEL_SPLUNK_HOME/etc/apps/$TA/"
  [ -n "$TA_TMPDIR" ] && rm -rf "$TA_TMPDIR" && TA_TMPDIR=""

  # Apply inputs.conf override from tas/<TA>.inputs.conf if it exists
  OVERLAY_FILE="$TA_SOURCE_DIR/$TA.inputs.conf"
  if [ -f "$OVERLAY_FILE" ]; then
    echo "  Applying inputs.conf override from $OVERLAY_FILE..."
    $SSH "mkdir -p $OTEL_SPLUNK_HOME/etc/apps/$TA/local"
    $SCP "$OVERLAY_FILE" "${SSH_USER}@${OTEL_HOST}:$OTEL_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"
  fi

  # Replace __INDEX__ in default and local inputs.conf. The deterministic
  # test TA intentionally keeps its single stanza in default/inputs.conf.
  $SSH bash -s << INNERSH
set -euo pipefail
for CONF in \
  "$OTEL_SPLUNK_HOME/etc/apps/$TA/default/inputs.conf" \
  "$OTEL_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"; do
  if [ -f "\$CONF" ]; then
    sed -i "s/__INDEX__/${OTEL_INDEX}/g" "\$CONF"
    echo "  Substituted index=${OTEL_INDEX} in \$CONF."
  fi
done
INNERSH

  echo "  $TA deployed."
done

# ------------------------------------------------------------
# 5. Write outputs.conf (HEC -> VM C)
# ------------------------------------------------------------
echo ""
echo "[5/7] Writing outputs.conf..."

$SSH bash -s << ENDSSH
set -euo pipefail
cat > "$OTEL_SPLUNK_HOME/etc/system/local/outputs.conf" << EOF
[httpout]
uri = http://${SPLUNK_HOST}:${SPLUNK_HEC_PORT}/services/collector/event
httpEventCollectorToken = ${SPLUNK_HEC_TOKEN}
EOF
echo "  outputs.conf written."
ENDSSH

# ------------------------------------------------------------
# 6. Write collector config
# ------------------------------------------------------------
echo ""
echo "[6/7] Writing collector config..."

$SSH bash -s << ENDSSH
set -euo pipefail
cat > "$CONFIG_FILE" << EOF
receivers:
  splunk_inputs:
    base_dir: ${OTEL_SPLUNK_HOME}

exporters:
  splunk_outputs:
    base_dir: ${OTEL_SPLUNK_HOME}

service:
  pipelines:
    logs:
      receivers: [splunk_inputs]
      exporters: [splunk_outputs]
EOF
echo "  Collector config written to $CONFIG_FILE."
ENDSSH

# ------------------------------------------------------------
# 7. Install and start systemd service
# ------------------------------------------------------------
echo ""
echo "[7/7] Installing systemd service..."

$SSH bash -s << ENDSSH
set -euo pipefail

sudo tee /etc/systemd/system/${SERVICE_NAME}.service > /dev/null << EOF
[Unit]
Description=Splunk OTel Collector (tarunner)
After=network.target

[Service]
Type=simple
User=${SSH_USER}
Environment=SPLUNK_HOME=${OTEL_SPLUNK_HOME}
ExecStart=${OTEL_BIN_DIR}/otelcol \\
  --config ${CONFIG_FILE} \\
  --feature-gates=+enableTARunner
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable ${SERVICE_NAME}
sudo systemctl restart ${SERVICE_NAME}

sleep 3
if sudo systemctl is-active --quiet ${SERVICE_NAME}; then
  echo "  Service ${SERVICE_NAME} is running."
else
  echo "  ERROR: service ${SERVICE_NAME} failed to start. Last 30 log lines:"
  sudo journalctl -u ${SERVICE_NAME} --no-pager -n 30
  exit 1
fi
ENDSSH

echo ""
echo "=== install_otel.sh complete ==="
echo "  Binary   : $OTEL_BIN_DIR/otelcol"
echo "  Config   : $CONFIG_FILE"
echo "  Service  : $SERVICE_NAME (systemd)"
echo "  HEC      : http://$SPLUNK_HOST:$SPLUNK_HEC_PORT -> index=$OTEL_INDEX"
echo ""
echo "  Logs: ssh -i $SSH_KEY $SSH_USER@$OTEL_HOST 'sudo journalctl -u $SERVICE_NAME -f'"
