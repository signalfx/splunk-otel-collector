#!/usr/bin/env bash

set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 <compose-profile> [<compose-profile> ...]" >&2
  exit 2
fi

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
compose_file="${script_dir}/docker-compose.yml"
compose_args=()

for profile in "$@"; do
  if [[ -z "${profile}" ]]; then
    echo "Compose profiles cannot be empty" >&2
    exit 2
  fi
  compose_args+=(--profile "${profile}")
done

docker compose --file "${compose_file}" "${compose_args[@]}" build
