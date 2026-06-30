#!/bin/bash

set -euo pipefail

ARCH="${ARCH:-}"
STAGE="${STAGE:-test}"
RPM_REPO_FILE="${RPM_REPO_FILE:-/etc/yum.repos.d/splunk-otel-collector.repo}"

if [[ "$#" -gt 0 ]]; then
    rpm_paths=("$@")
else
    ARCH="${ARCH:-x86_64}"
    shopt -s nullglob
    rpm_paths=(dist/signed/*"${ARCH}".rpm)
    shopt -u nullglob
fi

if [[ "${#rpm_paths[@]}" -lt 2 ]]; then
    echo "Expected collector and instrumentation RPMs for ${ARCH}, found ${#rpm_paths[@]}" >&2
    exit 1
fi

collector_nevra=""
instrumentation_nevra=""
instrumentation_rpm=""
for path in "${rpm_paths[@]}"; do
    if [[ ! -f "$path" ]]; then
        echo "$path not found!" >&2
        exit 1
    fi

    rpm_arch="$(rpm -qp --queryformat '%{ARCH}' "$path")"
    if [[ -z "$ARCH" ]]; then
        ARCH="$rpm_arch"
    elif [[ "$rpm_arch" != "$ARCH" ]]; then
        echo "RPM arch mismatch: expected ${ARCH}, got ${rpm_arch} for ${path}" >&2
        exit 1
    fi

    nevra="$(rpm -qp --queryformat '%{NAME}-%{VERSION}-%{RELEASE}.%{ARCH}' "$path")"
    case "$nevra" in
        splunk-otel-collector-*) collector_nevra="$nevra" ;;
        splunk-otel-auto-instrumentation-*)
            instrumentation_nevra="$nevra"
            instrumentation_rpm="$(basename "$path")"
            ;;
    esac
done

test -n "$collector_nevra"
test -n "$instrumentation_nevra"
test -n "$instrumentation_rpm"

cat > "$RPM_REPO_FILE" <<EOF
[splunk-otel-collector]
name=Splunk OpenTelemetry Collector Repository
baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/${STAGE}/\$basearch
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
enabled=1
EOF

dnf clean all
dnf -y --disablerepo='*' --enablerepo=splunk-otel-collector makecache
dnf -y install "$collector_nevra" "$instrumentation_nevra"
rpm -q "$collector_nevra" "$instrumentation_nevra"

# Direct rpm URL installs bypass repo_gpgcheck, but some users install this way.
instrumentation_url="https://splunk.jfrog.io/splunk/otel-collector-rpm/${STAGE}/${ARCH}/${instrumentation_rpm}"
rpm -e splunk-otel-auto-instrumentation
rpm -ivh "$instrumentation_url"
rpm -q "$instrumentation_nevra"
