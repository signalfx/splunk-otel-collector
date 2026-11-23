#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh
source .gitlab/collector-release-snapshot.sh
source .gitlab/collector-release-assets.sh

version="$1"
staging_tag="$2"
mode="${3:-}"
if [[ -n "$mode" && "$mode" != --recover-from-tag ]]; then
    echo "Usage: $0 <version> <staging-tag> [--recover-from-tag]" >&2
    exit 1
fi
latest_release="$(latest_collector_github_release)"
latest_published="$(jq -r '.tag_name // empty' <<< "$latest_release")"
if collector_version_is_older "$version" "$latest_published"; then
    echo "Skipping promotion of $version because $latest_published is already published"
    exit 0
fi
if [[ "$version" != "$latest_published" ]]; then
    echo "Cannot promote $version; latest published GitHub release is ${latest_published:-missing}" >&2
    exit 1
fi
release_id="$(github_release_id "$version" "$latest_release")"

published_ref="$(collector_release_refs "refs/tags/${version}")"
snapshot="$(read_release_snapshot "$version" "$published_ref" "$mode")"
if [[ "$mode" == --recover-from-tag ]]; then
    export_release_snapshot_registries "$snapshot"
    staging_tag="$(release_snapshot_staging_tag "$version" "$snapshot")"
fi
restore_release_input "$snapshot" docker
restore_release_input "$snapshot" assets
expected_assets="$(local_release_assets | jq -cs '.')"
bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}" verify

plan="$(github_asset_upload_plan "$release_id" "$expected_assets")"
if [[ "$plan" != '[]' ]]; then
    echo "Cannot promote $version; GitHub release assets are incomplete" >&2
    exit 1
fi

# Artifact restoration and verification can take time. Confirm the same GitHub
# release is still current before writing either registry.
latest_release="$(latest_collector_github_release)"
latest_published="$(jq -r '.tag_name // empty' <<< "$latest_release")"
if collector_version_is_older "$version" "$latest_published"; then
    echo "Skipping promotion of $version because $latest_published is already published"
    exit 0
fi
if [[ "$(github_release_id "$version" "$latest_release")" != "$release_id" ]]; then
    echo "Cannot promote $version; the latest published GitHub release changed during verification" >&2
    exit 1
fi
bash .gitlab/promote-collector-docker-images.sh "$staging_tag" latest
