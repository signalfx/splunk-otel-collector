#!/bin/bash -eux
set -o pipefail

[[ -z "$ORCA_CLOUD" ]] && echo "ORCA_CLOUD not set" && exit 1
[[ -z "$UF_VERSION" ]] && echo "UF_VERSION not set" && exit 1
[[ -z "$OLLY_ACCESS_TOKEN" ]] && echo "OLLY_ACCESS_TOKEN not set" && exit 1

exit 0
