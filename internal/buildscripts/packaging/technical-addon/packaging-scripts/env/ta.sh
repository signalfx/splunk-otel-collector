#!/bin/bash -eux
set -o pipefail

[[ -z "$PLATFORM" ]] && echo "PLATFORM not set" && exit 1
[[ -z "$ARCH" ]] && echo "ARCH not set" && exit 1
[[ -z "$OTEL_COLLECTOR_VERSION" ]] && echo "OTEL_COLLECTOR_VERSION not set" && exit 1

exit 0
