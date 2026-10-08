#!/usr/bin/env bash

# Copyright Splunk, Inc.
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

# Build the collector from the checkout that contains this test framework.
# The collector already contains the splunk_inputs and splunk_outputs modules,
# so no external tarunner checkout or go.mod replacement is required.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COLLECTOR_ROOT="$(cd "$TEST_ROOT/../.." && pwd)"

if [ -f "$TEST_ROOT/config.env" ]; then
  # shellcheck source=/dev/null
  source "$TEST_ROOT/config.env"
fi

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
OUTPUT_DIR="${OUTPUT_DIR:-$TEST_ROOT/bin}"
OUTPUT="$OUTPUT_DIR/otelcol_${GOOS}_${GOARCH}"

echo "=== build_otel.sh ==="
echo "Collector repo : $COLLECTOR_ROOT"
echo "Target         : ${GOOS}/${GOARCH}"

mkdir -p "$OUTPUT_DIR"

(
  cd "$COLLECTOR_ROOT"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -mod=readonly -trimpath \
    -o "$OUTPUT" \
    ./cmd/otelcol
)

if [ ! -x "$OUTPUT" ]; then
  echo "ERROR: expected executable not found at $OUTPUT" >&2
  exit 1
fi

echo "Done. Binary: $OUTPUT"
file "$OUTPUT"
