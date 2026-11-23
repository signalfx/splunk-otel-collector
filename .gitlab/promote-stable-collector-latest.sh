#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh
source .gitlab/collector-release-snapshot.sh

version="$1"
staging_tag="$2"
mode="${3:-}"
if [[ -n "$mode" && "$mode" != --recover-from-tag ]]; then
    echo "Usage: $0 <version> <staging-tag> [--recover-from-tag]" >&2
    exit 1
fi
latest_published="$(latest_published_collector_release)"
if collector_version_is_older "$version" "$latest_published"; then
    echo "Skipping promotion of $version because $latest_published is already published"
    exit 0
fi
if [[ "$version" != "$latest_published" ]]; then
    echo "Cannot promote $version; latest published GitHub release is ${latest_published:-missing}" >&2
    exit 1
fi

published_ref="$(collector_release_refs "refs/tags/${version}")"
snapshot="$(read_release_snapshot "$version" "$published_ref" "$mode")"
if [[ "$mode" == --recover-from-tag ]]; then
    export_release_snapshot_registries "$snapshot"
    staging_tag="$(release_snapshot_staging_tag "$version" "$snapshot")"
fi
restore_release_input "$snapshot" docker
bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}" verify
bash .gitlab/promote-collector-docker-images.sh "$staging_tag" latest
