#!/bin/bash

set -euo pipefail

export DEBIAN_FRONTEND="${DEBIAN_FRONTEND:-noninteractive}"

ARCH="${ARCH:-}"
STAGE="${STAGE:-test}"
DEB_REPO_BASE="${DEB_REPO_BASE:-https://splunk.jfrog.io/splunk/otel-collector-deb}"
DEB_REPO_FILE="${DEB_REPO_FILE:-/etc/apt/sources.list.d/splunk-otel-collector.list}"
DEB_KEYRING="${DEB_KEYRING:-/usr/share/keyrings/splunk-otel-collector.gpg}"

if [[ "$#" -gt 0 ]]; then
    deb_paths=("$@")
else
    ARCH="${ARCH:-amd64}"
    shopt -s nullglob
    deb_paths=(dist/signed/*_"${ARCH}".deb)
    shopt -u nullglob
fi

if [[ "${#deb_paths[@]}" -lt 2 ]]; then
    echo "Expected collector and instrumentation DEBs for ${ARCH}, found ${#deb_paths[@]}" >&2
    exit 1
fi

collector_version=""
instrumentation_version=""
instrumentation_deb=""
for path in "${deb_paths[@]}"; do
    if [[ ! -f "$path" ]]; then
        echo "$path not found!" >&2
        exit 1
    fi

    package="$(dpkg-deb -f "$path" Package)"
    version="$(dpkg-deb -f "$path" Version)"
    deb_arch="$(dpkg-deb -f "$path" Architecture)"

    if [[ -z "$ARCH" ]]; then
        ARCH="$deb_arch"
    elif [[ "$deb_arch" != "$ARCH" ]]; then
        echo "DEB arch mismatch: expected ${ARCH}, got ${deb_arch} for ${path}" >&2
        exit 1
    fi

    case "$package" in
        splunk-otel-collector)
            collector_version="$version"
            ;;
        splunk-otel-auto-instrumentation)
            instrumentation_version="$version"
            instrumentation_deb="$(basename "$path")"
            ;;
    esac
done

test -n "$collector_version"
test -n "$instrumentation_version"
test -n "$instrumentation_deb"

apt-get -y update
apt-get -y install apt-transport-https ca-certificates curl gnupg

install -d -m 0755 "$(dirname "$DEB_REPO_FILE")" "$(dirname "$DEB_KEYRING")"
curl -fsSL "${DEB_REPO_BASE}/splunk-B3CD4420.gpg" -o "$DEB_KEYRING"
chmod 0644 "$DEB_KEYRING"

cat > "$DEB_REPO_FILE" <<EOF
deb [arch=${ARCH} signed-by=${DEB_KEYRING}] ${DEB_REPO_BASE} ${STAGE} main
EOF

apt-get -y update
apt-get -y install \
    "splunk-otel-collector=${collector_version}" \
    "splunk-otel-auto-instrumentation=${instrumentation_version}"

dpkg-query -W -f='${Package} ${Version} ${Architecture}\n' \
    splunk-otel-collector splunk-otel-auto-instrumentation

test "$(dpkg-query -W -f='${Version}' splunk-otel-collector)" = "$collector_version"
test "$(dpkg-query -W -f='${Version}' splunk-otel-auto-instrumentation)" = "$instrumentation_version"
test "$(dpkg-query -W -f='${Architecture}' splunk-otel-collector)" = "$ARCH"
test "$(dpkg-query -W -f='${Architecture}' splunk-otel-auto-instrumentation)" = "$ARCH"

# Direct .deb URL installs bypass repo metadata signature checks, but some users install this way.
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
instrumentation_url="${DEB_REPO_BASE}/pool/${STAGE}/${ARCH}/${instrumentation_deb}"
apt-get -y remove splunk-otel-auto-instrumentation
curl -fsSL "$instrumentation_url" -o "${tmpdir}/${instrumentation_deb}"
apt-get -y install "${tmpdir}/${instrumentation_deb}"

dpkg-query -W -f='${Package} ${Version} ${Architecture}\n' splunk-otel-auto-instrumentation
test "$(dpkg-query -W -f='${Version}' splunk-otel-auto-instrumentation)" = "$instrumentation_version"
test "$(dpkg-query -W -f='${Architecture}' splunk-otel-auto-instrumentation)" = "$ARCH"
