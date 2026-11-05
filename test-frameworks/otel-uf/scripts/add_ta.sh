#!/usr/bin/env bash
# Deploy a TA to both VM A (UF) and VM B (OTel) in sync, then restart both agents.
#
# What this script does:
#   1. Validates TA source directory exists
#   2. Deploys TA to VM A (UF) under $UF_SPLUNK_HOME/etc/apps/<TA>/
#   3. Sets index in TA's default/local inputs.conf on VM A (UF_INDEX)
#   4. Deploys TA to VM B (OTel) under $OTEL_SPLUNK_HOME/etc/apps/<TA>/
#   5. Sets index in TA's default/local inputs.conf on VM B (OTEL_INDEX)
#   6. Restarts UF on VM A
#   7. Restarts otelcol on VM B
#
# Usage:
#   ./scripts/add_ta.sh <TA_NAME>
#   make add-ta TA=Splunk_TA_nix
#
# Requires: config.env with UF_HOST, OTEL_HOST, SSH_KEY, SSH_USER,
#           UF_SPLUNK_HOME, OTEL_SPLUNK_HOME, UF_INDEX, OTEL_INDEX, TA_SOURCE_DIR
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$ROOT_DIR/config.env" ]; then
  # shellcheck source=/dev/null
  source "$ROOT_DIR/config.env"
fi

TA="${1:?Usage: $0 <TA_NAME>}"

: "${UF_HOST:?UF_HOST is not set}"
: "${OTEL_HOST:?OTEL_HOST is not set}"
: "${SSH_KEY:?SSH_KEY is not set}"
: "${SSH_USER:?SSH_USER is not set}"
: "${UF_SPLUNK_HOME:?UF_SPLUNK_HOME is not set}"
: "${OTEL_SPLUNK_HOME:?OTEL_SPLUNK_HOME is not set}"
: "${UF_INDEX:?UF_INDEX is not set}"
: "${OTEL_INDEX:?OTEL_INDEX is not set}"
: "${TA_SOURCE_DIR:=$ROOT_DIR/tas}"

SSH_OPTS="-i ${SSH_KEY} -o StrictHostKeyChecking=no"
SSH_UF="ssh ${SSH_OPTS} ${SSH_USER}@${UF_HOST}"
SSH_OTEL="ssh ${SSH_OPTS} ${SSH_USER}@${OTEL_HOST}"
SCP="scp ${SSH_OPTS}"

UF_BIN="$UF_SPLUNK_HOME/bin/splunk"
SERVICE_NAME="otelcol"

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

# ------------------------------------------------------------
# 1. Resolve TA source
# ------------------------------------------------------------
TA_DIR="$(resolve_ta_dir "$TA")" || {
  echo "ERROR: TA source not found for $TA (checked directory and .tgz/.tar.gz in $TA_SOURCE_DIR)"
  exit 1
}

echo "=== add_ta.sh ==="
echo "TA          : $TA"
echo "Source      : $TA_DIR"
echo "VM A (UF)   : $UF_HOST -> $UF_SPLUNK_HOME/etc/apps/$TA  (index=$UF_INDEX)"
echo "VM B (OTel) : $OTEL_HOST -> $OTEL_SPLUNK_HOME/etc/apps/$TA  (index=$OTEL_INDEX)"
echo ""

# ------------------------------------------------------------
# 2. Deploy TA to VM A (UF)
# ------------------------------------------------------------
echo "[1/6] Deploying $TA to VM A (UF)..."

$SSH_UF "sudo mkdir -p $UF_SPLUNK_HOME/etc/apps/$TA"
$SCP -r "$TA_DIR/." "${SSH_USER}@${UF_HOST}:/tmp/${TA}_upload"
$SSH_UF "sudo cp -r /tmp/${TA}_upload/. $UF_SPLUNK_HOME/etc/apps/$TA/ && rm -rf /tmp/${TA}_upload"
echo "  Files copied."

# Apply inputs.conf override from tas/<TA>.inputs.conf if it exists
OVERLAY_FILE="$TA_SOURCE_DIR/$TA.inputs.conf"
if [ -f "$OVERLAY_FILE" ]; then
  echo "  Applying inputs.conf override from $OVERLAY_FILE..."
  $SSH_UF "sudo mkdir -p $UF_SPLUNK_HOME/etc/apps/$TA/local"
  $SCP "$OVERLAY_FILE" "${SSH_USER}@${UF_HOST}:/tmp/${TA}_inputs.conf"
  $SSH_UF "sudo mv /tmp/${TA}_inputs.conf $UF_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"
fi

# ------------------------------------------------------------
# 3. Set index on VM A
# ------------------------------------------------------------
echo "[2/6] Setting index on VM A..."

$SSH_UF bash -s << INNERSH
set -euo pipefail
for CONF in \
  "$UF_SPLUNK_HOME/etc/apps/$TA/default/inputs.conf" \
  "$UF_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"; do
  if [ -f "\$CONF" ]; then
    sudo sed -i "s/__INDEX__/${UF_INDEX}/g" "\$CONF"
    echo "  Substituted index=${UF_INDEX} in \$CONF."
  fi
done
INNERSH

# ------------------------------------------------------------
# 4. Deploy TA to VM B (OTel)
# ------------------------------------------------------------
echo "[3/6] Deploying $TA to VM B (OTel)..."

$SSH_OTEL "mkdir -p $OTEL_SPLUNK_HOME/etc/apps/$TA"
$SCP -r "$TA_DIR/." "${SSH_USER}@${OTEL_HOST}:$OTEL_SPLUNK_HOME/etc/apps/$TA/"
[ -n "$TA_TMPDIR" ] && rm -rf "$TA_TMPDIR" && TA_TMPDIR=""
echo "  Files copied."

if [ -f "$OVERLAY_FILE" ]; then
  echo "  Applying inputs.conf override from $OVERLAY_FILE..."
  $SSH_OTEL "mkdir -p $OTEL_SPLUNK_HOME/etc/apps/$TA/local"
  $SCP "$OVERLAY_FILE" "${SSH_USER}@${OTEL_HOST}:/tmp/${TA}_inputs.conf"
  $SSH_OTEL "mv /tmp/${TA}_inputs.conf $OTEL_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"
fi

# ------------------------------------------------------------
# 5. Set index on VM B
# ------------------------------------------------------------
echo "[4/6] Setting index on VM B..."

$SSH_OTEL bash -s << INNERSH
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

# ------------------------------------------------------------
# 6. Restart UF on VM A
# ------------------------------------------------------------
echo "[5/6] Restarting UF on VM A..."

$SSH_UF "sudo $UF_BIN restart --accept-license --answer-yes --no-prompt"
echo "  UF restarted."

# ------------------------------------------------------------
# 7. Restart otelcol on VM B
# ------------------------------------------------------------
echo "[6/6] Restarting otelcol on VM B..."

$SSH_OTEL "sudo systemctl restart ${SERVICE_NAME}"
sleep 3
if $SSH_OTEL "sudo systemctl is-active --quiet ${SERVICE_NAME}"; then
  echo "  otelcol restarted and running."
else
  echo "  ERROR: otelcol failed to start after restart. Last 20 log lines:"
  $SSH_OTEL "sudo journalctl -u ${SERVICE_NAME} --no-pager -n 20"
  exit 1
fi

echo ""
echo "=== add_ta.sh complete ==="
echo "  TA $TA deployed to both VMs."
echo "  UF   (VM A): index=$UF_INDEX"
echo "  OTel (VM B): index=$OTEL_INDEX"
