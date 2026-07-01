#!/usr/bin/env bash

set -euo pipefail

stage="${STAGE:-release}"
repo_url="https://splunk.jfrog.io/splunk/otel-collector-deb"
key_url="${SPLUNK_DEB_GPG_KEY_URL:-${repo_url}/splunk-B3CD4420.gpg}"
fingerprint="${SPLUNK_GPG_FINGERPRINT:-58C33310B7A354C1279DB6695EFA01EDB3CD4420}"
keyring="/usr/share/keyrings/splunk-otel-collector.gpg"
source_list="/etc/apt/sources.list.d/splunk-otel-collector.list"
tmpdir="$(mktemp -d)"

trap 'rm -rf "$tmpdir"' EXIT

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y ca-certificates curl gnupg

install -d -m 0755 "$(dirname "$keyring")"
curl -fsSL "$key_url" -o "$keyring"

export GNUPGHOME="${tmpdir}/gnupg"
mkdir -m 0700 "$GNUPGHOME"
gpg --import "$keyring" >/dev/null

curl -fsSL "${repo_url}/dists/${stage}/Release" -o "${tmpdir}/Release"
curl -fsSL "${repo_url}/dists/${stage}/Release.gpg" -o "${tmpdir}/Release.gpg"
gpg --status-fd=1 --verify "${tmpdir}/Release.gpg" "${tmpdir}/Release" \
  | grep -Eq "^\[GNUPG:\] VALIDSIG ${fingerprint}( |$)"

curl -fsSL "${repo_url}/dists/${stage}/InRelease" -o "${tmpdir}/InRelease"
gpg --status-fd=1 --output /dev/null --decrypt "${tmpdir}/InRelease" \
  | grep -Eq "^\[GNUPG:\] VALIDSIG ${fingerprint}( |$)"

echo "deb [signed-by=${keyring}] ${repo_url} ${stage} main" > "$source_list"

apt-get update
apt-get install -y splunk-otel-collector splunk-otel-auto-instrumentation

dpkg -s splunk-otel-collector >/dev/null
dpkg -s splunk-otel-auto-instrumentation >/dev/null
