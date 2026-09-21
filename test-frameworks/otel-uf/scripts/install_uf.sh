#!/usr/bin/env bash
# Install Splunk Universal Forwarder on VM A and configure it to run a TA.
#
# What this script does:
#   1. Installs the UF under UF_SPLUNK_HOME
#   2. Deploys each TA from TA_SOURCE_DIR to $SPLUNK_HOME/etc/apps/
#   3. Writes outputs.conf pointing to VM C via S2S, routing to UF_INDEX
#   4. Starts/restarts the UF
#
# Usage:
#   ./scripts/install_uf.sh
#
# Requires: config.env with UF_HOST, SSH_KEY, SSH_USER, UF_VERSION,
#           UF_SPLUNK_HOME, SPLUNK_HOST, SPLUNK_S2S_PORT, UF_INDEX,
#           SPLUNK_ADMIN_PASSWORD, TA_LIST, TA_SOURCE_DIR
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$ROOT_DIR/config.env" ]; then
  # shellcheck source=/dev/null
  source "$ROOT_DIR/config.env"
fi

: "${UF_HOST:?UF_HOST is not set}"
: "${SSH_KEY:?SSH_KEY is not set}"
: "${SSH_USER:?SSH_USER is not set}"
: "${UF_VERSION:?UF_VERSION is not set}"
: "${UF_SPLUNK_HOME:?UF_SPLUNK_HOME is not set}"
: "${SPLUNK_HOST:?SPLUNK_HOST is not set}"
: "${SPLUNK_S2S_PORT:=9997}"
: "${UF_INDEX:?UF_INDEX is not set}"
: "${SPLUNK_ADMIN_PASSWORD:?SPLUNK_ADMIN_PASSWORD is not set}"
: "${TA_LIST:=Splunk_TA_nix}"
: "${TA_SOURCE_DIR:=$ROOT_DIR/tas}"

SSH_OPTS="-i ${SSH_KEY} -o StrictHostKeyChecking=no"
SSH="ssh ${SSH_OPTS} ${SSH_USER}@${UF_HOST}"
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

UF_BIN="$UF_SPLUNK_HOME/bin/splunk"
# Build hash is required for some Splunk versions (e.g. 9.4.0 = 6b4ebe426ca6).
# Set UF_BUILD_HASH in config.env if the plain URL returns 404.
# Falls back to SPLUNK_BUILD_HASH if UF_BUILD_HASH is not set separately.
_UF_HASH="${UF_BUILD_HASH:-${SPLUNK_BUILD_HASH:-}}"
if [ -n "$_UF_HASH" ]; then
  UF_PKG="splunkforwarder-${UF_VERSION}-${_UF_HASH}-linux-amd64.tgz"
else
  UF_PKG="splunkforwarder-${UF_VERSION}-linux-amd64.tgz"
fi
UF_URL="https://download.splunk.com/products/universalforwarder/releases/${UF_VERSION}/linux/${UF_PKG}"

echo "=== install_uf.sh ==="
echo "Host      : $UF_HOST"
echo "Version   : $UF_VERSION"
echo "SPLUNK_HOME: $UF_SPLUNK_HOME"
echo "Indexer   : $SPLUNK_HOST:$SPLUNK_S2S_PORT -> index=$UF_INDEX"
echo "TAs       : $TA_LIST"
echo ""

# ------------------------------------------------------------
# 1. Install Universal Forwarder
# ------------------------------------------------------------
echo "[1/4] Installing Universal Forwarder..."

$SSH bash -s << ENDSSH
set -euo pipefail

if [ -f "$UF_BIN" ]; then
  echo "  UF already installed at $UF_SPLUNK_HOME, skipping download."
  exit 0
fi

echo "  Downloading $UF_PKG..."
curl -fsSL -o "/tmp/$UF_PKG" "$UF_URL"

echo "  Extracting..."
sudo tar -xzf "/tmp/$UF_PKG" -C /opt
rm "/tmp/$UF_PKG"

# The tarball extracts to splunkforwarder/ - rename to match UF_SPLUNK_HOME if needed
EXTRACTED="/opt/splunkforwarder"
if [ "\$EXTRACTED" != "$UF_SPLUNK_HOME" ]; then
  sudo mv "\$EXTRACTED" "$UF_SPLUNK_HOME"
fi

echo "  Starting UF (first run, accept license)..."
sudo "$UF_BIN" start --accept-license --answer-yes --no-prompt \
  --seed-passwd "${SPLUNK_ADMIN_PASSWORD}"

echo "  Enabling boot start..."
sudo "$UF_BIN" stop --accept-license --answer-yes --no-prompt
sudo "$UF_BIN" enable boot-start -user root --accept-license --answer-yes --no-prompt
sudo "$UF_BIN" start --accept-license --answer-yes --no-prompt
ENDSSH

echo "  Done."

# ------------------------------------------------------------
# 2. Write outputs.conf (S2S -> VM C)
# ------------------------------------------------------------
echo ""
echo "[2/4] Writing outputs.conf..."

$SSH bash -s << ENDSSH
set -euo pipefail
SYSTEM_LOCAL="$UF_SPLUNK_HOME/etc/system/local"
sudo mkdir -p "\$SYSTEM_LOCAL"

sudo tee "\$SYSTEM_LOCAL/outputs.conf" > /dev/null << 'EOF'
[tcpout]
defaultGroup = splunk_indexer

[tcpout:splunk_indexer]
server = ${SPLUNK_HOST}:${SPLUNK_S2S_PORT}

[tcpout-server://${SPLUNK_HOST}:${SPLUNK_S2S_PORT}]
EOF

echo "  outputs.conf written."
ENDSSH

# ------------------------------------------------------------
# 3. Deploy TAs
# ------------------------------------------------------------
echo ""
echo "[3/4] Deploying TAs..."

for TA in $TA_LIST; do
  TA_DIR="$(resolve_ta_dir "$TA")" || {
    echo "  WARNING: TA source not found for $TA (checked directory and .tgz/.tar.gz), skipping."
    continue
  }

  echo "  Deploying $TA (from $TA_DIR)..."

  # Create app dir on remote
  $SSH "sudo mkdir -p $UF_SPLUNK_HOME/etc/apps/$TA"

  # Upload TA contents
  $SCP -r "$TA_DIR/." "${SSH_USER}@${UF_HOST}:/tmp/${TA}_upload"

  # Move into place
  $SSH "sudo cp -r /tmp/${TA}_upload/. $UF_SPLUNK_HOME/etc/apps/$TA/ && rm -rf /tmp/${TA}_upload"

  [ -n "$TA_TMPDIR" ] && rm -rf "$TA_TMPDIR" && TA_TMPDIR=""

  # Apply inputs.conf override from tas/<TA>.inputs.conf if it exists
  OVERLAY_FILE="$TA_SOURCE_DIR/$TA.inputs.conf"
  if [ -f "$OVERLAY_FILE" ]; then
    echo "  Applying inputs.conf override from $OVERLAY_FILE..."
    $SSH "sudo mkdir -p $UF_SPLUNK_HOME/etc/apps/$TA/local"
    $SCP "$OVERLAY_FILE" "${SSH_USER}@${UF_HOST}:/tmp/${TA}_inputs.conf"
    $SSH "sudo mv /tmp/${TA}_inputs.conf $UF_SPLUNK_HOME/etc/apps/$TA/local/inputs.conf"
  fi

  # Replace __INDEX__ in default and local inputs.conf. The deterministic
  # test TA intentionally keeps its single stanza in default/inputs.conf.
  $SSH bash -s << INNERSH
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

  echo "  $TA deployed."
done

# ------------------------------------------------------------
# 4. Restart UF
# ------------------------------------------------------------
echo ""
echo "[4/4] Restarting UF..."

$SSH "sudo $UF_BIN restart --accept-license --answer-yes --no-prompt"

echo ""
echo "=== install_uf.sh complete ==="
echo "  UF installed : $UF_SPLUNK_HOME"
echo "  Forwarding to: $SPLUNK_HOST:$SPLUNK_S2S_PORT"
echo "  Index        : $UF_INDEX"
