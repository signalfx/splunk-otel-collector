#!/bin/bash

# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

if [[ "$#" -ne 2 ]]; then
    echo "Usage: $0 <candidate|previous> <x86_64|aarch64>" >&2
    exit 1
fi

test_case="$1"
arch="$2"
repo_stage="${RPM_REPO_STAGE:-test}"
repo_base_url="${RPM_REPO_BASE_URL:-https://splunk.jfrog.io/splunk/otel-collector-rpm}"
gpg_key_url="${RPM_GPG_KEY_URL:-https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub}"
gpgcheck="${RPM_REPO_GPGCHECK:-1}"
repo_id="splunk-otel-collector-$repo_stage"
packages=(splunk-otel-collector splunk-otel-auto-instrumentation)

if [[ "$test_case" != "candidate" && "$test_case" != "previous" ]]; then
    echo "Unsupported test case: $test_case" >&2
    exit 1
fi
if [[ "$arch" != "x86_64" && "$arch" != "aarch64" ]]; then
    echo "Unsupported RPM architecture: $arch" >&2
    exit 1
fi
if [[ "$repo_stage" != "test" && "$repo_stage" != "beta" ]]; then
    echo "Refusing to run repository installation checks outside the test or beta stages." >&2
    exit 1
fi
if [[ "$gpgcheck" != "0" && "$gpgcheck" != "1" ]]; then
    echo "RPM_REPO_GPGCHECK must be 0 or 1." >&2
    exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "$script_dir/../.." && pwd)"

candidate_evr="${CANDIDATE_RPM_EVR:-}"
if [[ -z "$candidate_evr" ]]; then
    # Match the version selection and RPM normalization performed by both builders.
    candidate_version="${RELEASE_VERSION:-${CI_COMMIT_TAG:-}}"
    if [[ -z "$candidate_version" ]]; then
        commit_tag="$(git -C "$repo_dir" describe --abbrev=0 --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true)"
        if [[ -n "$commit_tag" ]]; then
            candidate_version="$commit_tag"
        else
            latest_tag="$(git -C "$repo_dir" describe --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)"
            candidate_version="${latest_tag:-0.0.1}-post"
        fi
    fi
    candidate_version="${candidate_version/'-'/'_'}"
    candidate_version="${candidate_version#v}"
    candidate_evr="${candidate_version}-1"
fi

cat > "/etc/yum.repos.d/splunk-otel-collector-${repo_stage}.repo" <<EOF
[$repo_id]
name=Splunk OpenTelemetry Collector ${repo_stage^} Repository
baseurl=${repo_base_url}/${repo_stage}/\$basearch
gpgcheck=$gpgcheck
repo_gpgcheck=$gpgcheck
gpgkey=$gpg_key_url
enabled=1
EOF

dnf -y --refresh --repo "$repo_id" makecache

query_versions() {
    dnf -q repoquery --repo "$repo_id" --arch "$arch" \
        --qf '%{version}-%{release}' "$1" | sort -Vru
}

target_evr="$candidate_evr"
if [[ "$test_case" == "previous" ]]; then
    mapfile -t collector_versions < <(query_versions "${packages[0]}")
    mapfile -t instrumentation_versions < <(query_versions "${packages[1]}")

    target_evr=""
    for collector_version in "${collector_versions[@]}"; do
        if [[ "$collector_version" == "$candidate_evr" ]]; then
            continue
        fi
        for instrumentation_version in "${instrumentation_versions[@]}"; do
            if [[ "$collector_version" == "$instrumentation_version" ]]; then
                target_evr="$collector_version"
                break 2
            fi
        done
    done

    if [[ -z "$target_evr" ]]; then
        echo "No previous version shared by both RPM packages was found in the test repository." >&2
        exit 1
    fi
fi

echo "Installing $test_case RPM packages at version-release $target_evr for $arch"
dnf -y --repo "$repo_id" install \
    "${packages[0]}-${target_evr}.${arch}" \
    "${packages[1]}-${target_evr}.${arch}"

for package in "${packages[@]}"; do
    installed_evr="$(rpm -q --queryformat '%{VERSION}-%{RELEASE}' "$package")"
    file_digest_algo="$(rpm -q --queryformat '%{FILEDIGESTALGO}' "$package")"
    echo "$package: version-release=$installed_evr FILEDIGESTALGO=$file_digest_algo"

    if [[ "$installed_evr" != "$target_evr" ]]; then
        echo "Unexpected installed version for $package: $installed_evr" >&2
        exit 1
    fi
    if [[ "$test_case" == "candidate" && "$file_digest_algo" != "8" ]]; then
        echo "Candidate package does not use SHA-256 per-file digests: $package" >&2
        exit 1
    fi
done

if [[ "$test_case" == "candidate" ]]; then
    verify_candidate_file() {
        local package="$1"
        local path="$2"
        local owner=""
        local digest=""

        if [[ ! -f "$path" ]]; then
            echo "Expected candidate package file not found: $path" >&2
            exit 1
        fi

        owner="$(rpm -qf --queryformat '%{NAME}' "$path")"
        if [[ "$owner" != "$package" ]]; then
            echo "Unexpected package owner for $path: $owner" >&2
            exit 1
        fi

        digest="$(rpm -q --dump "$package" | awk -v expected="$path" '$1 == expected { print $4; exit }')"
        if [[ ! "$digest" =~ ^[[:xdigit:]]{64}$ ]]; then
            echo "Expected a SHA-256 file digest for $path, got: ${digest:-none}" >&2
            exit 1
        fi

        echo "$package owns $path with SHA-256 digest $digest"
    }

    for binary in /usr/bin/otelcol /usr/bin/otelcollauncher /usr/bin/opampsupervisor; do
        verify_candidate_file "${packages[0]}" "$binary"
        if [[ ! -x "$binary" ]]; then
            echo "Expected candidate package executable not found: $binary" >&2
            exit 1
        fi
    done

    instrumentation_files=(
        /usr/lib/splunk-instrumentation/libsplunk.so
        /usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar
        /usr/lib/splunk-instrumentation/splunk-otel-js.tgz
    )
    if [[ "$arch" == "x86_64" ]]; then
        instrumentation_files+=(
            /usr/lib/splunk-instrumentation/splunk-otel-dotnet/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so
        )
    fi
    for instrumentation_file in "${instrumentation_files[@]}"; do
        verify_candidate_file "${packages[1]}" "$instrumentation_file"
    done
fi
