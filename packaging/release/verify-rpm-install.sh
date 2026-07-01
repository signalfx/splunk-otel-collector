#!/usr/bin/env bash

set -euo pipefail

stage="${STAGE:-release}"
repo_file="/etc/yum.repos.d/splunk-otel-collector.repo"
repo_url="https://splunk.jfrog.io/splunk/otel-collector-rpm"
key_url="${SPLUNK_RPM_GPG_KEY_URL:-${repo_url}/splunk-B3CD4420.pub}"
fingerprint="${SPLUNK_GPG_FINGERPRINT:-58C33310B7A354C1279DB6695EFA01EDB3CD4420}"
tmpdir="$(mktemp -d)"

trap 'rm -rf "$tmpdir"' EXIT

dnf install -y curl-minimal gnupg2

export GNUPGHOME="${tmpdir}/gnupg"
mkdir -m 0700 "$GNUPGHOME"
curl -fsSL "$key_url" -o "${tmpdir}/splunk.pub"
gpg --import "${tmpdir}/splunk.pub" >/dev/null

for arch in x86_64 aarch64; do
  metadata_url="${repo_url}/${stage}/${arch}/repodata/repomd.xml"
  signature_url="${metadata_url}.asc"
  curl -fsSL "$metadata_url" -o "${tmpdir}/repomd-${arch}.xml"
  curl -fsSL "$signature_url" -o "${tmpdir}/repomd-${arch}.xml.asc"
  gpg --status-fd=1 --verify "${tmpdir}/repomd-${arch}.xml.asc" "${tmpdir}/repomd-${arch}.xml" \
    | grep -Eq "^\[GNUPG:\] VALIDSIG ${fingerprint}( |$)"
done

cat > "$repo_file" <<EOF
[splunk-otel-collector]
name=Splunk OpenTelemetry Collector Repository
baseurl=${repo_url}/${stage}/\$basearch
gpgcheck=1
repo_gpgcheck=1
gpgkey=${key_url}
enabled=1
EOF

dnf -y makecache --repo splunk-otel-collector
dnf install -y splunk-otel-collector splunk-otel-auto-instrumentation

rpm -q splunk-otel-collector >/dev/null
rpm -q splunk-otel-auto-instrumentation >/dev/null
