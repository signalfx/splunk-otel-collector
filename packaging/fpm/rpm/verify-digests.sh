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

if [[ "$#" -eq 0 ]]; then
    echo "Usage: $0 <rpm> [<rpm> ...]" >&2
    exit 1
fi

status=0
rpm_query_tags="$(rpm --querytags)"
if grep -Fxq "PAYLOADDIGESTALGO" <<< "$rpm_query_tags"; then
    payload_digest_tag="PAYLOADDIGESTALGO"
elif grep -Fxq "PAYLOADSHA256ALGO" <<< "$rpm_query_tags"; then
    # RPM 6 renamed the query tag while preserving the SHA-256 algorithm ID.
    payload_digest_tag="PAYLOADSHA256ALGO"
else
    echo "RPM does not expose a supported payload digest algorithm tag." >&2
    exit 1
fi

for rpm_path in "$@"; do
    if [[ ! -f "$rpm_path" ]]; then
        echo "RPM not found: $rpm_path" >&2
        status=1
        continue
    fi

    metadata="$(rpm -qp --queryformat "%{FILEDIGESTALGO} %{$payload_digest_tag}" "$rpm_path")"
    read -r file_digest_algo payload_digest_algo <<< "$metadata"
    echo "$rpm_path: FILEDIGESTALGO=$file_digest_algo $payload_digest_tag=$payload_digest_algo"

    if [[ "$file_digest_algo" != "8" || "$payload_digest_algo" != "8" ]]; then
        echo "Expected SHA-256 digest algorithm ID 8 for RPM files and payload: $rpm_path" >&2
        status=1
    fi
done

exit "$status"
