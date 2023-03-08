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

OUTPUT_DIR="${OUTPUT_DIR:-}"
VERSION="${VERSION:-}"

if [ $# -eq 0 ]; then
    if [ -z "$OUTPUT_DIR" ]; then
        echo "OUTPUT_DIR env var not set!" >&2
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo "VERSION env var not set!" >&2
        exit 1
    fi

    buildargs="--output /work/build/stage ${VERSION#v}"
    /project/internal/buildscripts/packaging/msi/msi-builder/build.sh $buildargs
    mkdir -p $OUTPUT_DIR
    echo "Copying MSI to $OUTPUT_DIR"
    cp /work/build/stage/*.msi $OUTPUT_DIR/
else
    exec "$@"
fi
