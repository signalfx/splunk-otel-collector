#!/bin/bash -eux
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

exit 0
