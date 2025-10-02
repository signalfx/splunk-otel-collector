#!/bin/bash

# Copyright 2020 Splunk, Inc.
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

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE:-}"
OUTPUT_DIR="${OUTPUT_DIR:-}"
REPO_DIR="${REPO_DIR:-}"
WORK_DIR="${WORK_DIR:-}"
VERSION="${VERSION:-}"

if [ -z "$JMX_METRIC_GATHERER_RELEASE" ]; then
    echo "JMX_METRIC_GATHERER_RELEASE env var not set!" >&2
    exit 1
fi

if [ -z "$OUTPUT_DIR" ]; then
    echo "OUTPUT_DIR env var not set!" >&2
    exit 1
fi

if [ -z "$REPO_DIR" ]; then
    echo "REPO_DIR env var not set!" >&2
    exit 1
fi

if [ -z "$WORK_DIR" ]; then
    echo "WORK_DIR env var not set!" >&2
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "VERSION env var not set!" >&2
    exit 1
fi

"$SCRIPT_DIR/build.sh" --output "$WORK_DIR/build/stage" --jmx-metric-gatherer "$JMX_METRIC_GATHERER_RELEASE" "${VERSION#v}"
mkdir -p "$OUTPUT_DIR"
echo "Copying MSI to $OUTPUT_DIR"
cp "$WORK_DIR/build/stage/*.msi" "$OUTPUT_DIR"
