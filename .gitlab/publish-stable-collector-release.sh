#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh
source .gitlab/collector-release-snapshot.sh

version="$1"
staging_tag="$2"
mode="${3:-}"
if [[ -n "$mode" && "$mode" != --recover-only && "$mode" != --recover-from-tag ]]; then
    echo "Usage: $0 <version> <staging-tag> [--recover-only|--recover-from-tag]" >&2
    exit 1
fi

# ghr can upload to a draft, but does not expose GitHub's make_latest flag.
publish_superseded_draft() {
    local page=1 releases release id
    while :; do
        releases="$(collector_github_api "releases?per_page=100&page=$page" --fail)"
        release="$(jq -c --arg version "$version" '[.[] | select(.tag_name == $version)] | sort_by(.draft) | first // empty' <<< "$releases")"
        if [[ -n "$release" ]]; then
            if [[ "$(jq -r '.draft' <<< "$release")" == true ]]; then
                id="$(jq -er '.id | numbers' <<< "$release")"
                collector_github_api "releases/$id" --fail --request PATCH \
                    --header 'Content-Type: application/json' \
                    --data '{"draft":false,"make_latest":"false"}' > /dev/null
            fi
            return
        fi
        if [[ "$(jq 'length' <<< "$releases")" -lt 100 ]]; then
            echo "Could not find GitHub draft release for $version" >&2
            return 1
        fi
        ((page += 1))
    done
}

: "${GITHUB_TOKEN:?}" "${CI_COMMIT_SHA:?}"

# The GitHub release job holds collector-docker-latest for this entire script.
published_ref="$(collector_release_refs "refs/tags/${version}")"
if [[ -n "$mode" && -z "$published_ref" ]]; then
    echo "Recovery requires an existing signed release tag for $version" >&2
    exit 1
fi
latest_published="$(latest_published_collector_release)"
if [[ -n "$published_ref" ]]; then
    snapshot="$(read_release_snapshot "$version" "$published_ref" "$mode")"
    if [[ -n "$mode" ]]; then
        release_status="$(collector_github_api "releases/tags/$version" --output /dev/null --write-out '%{http_code}')"
        case "$release_status" in
            200)
                echo "$version is already published on GitHub"
                exit 0
                ;;
            404) ;; # No published release exists for this tag yet.
            *)
                echo "Could not check publication of $version (HTTP $release_status)" >&2
                exit 1
                ;;
        esac
    fi
    if [[ "$mode" == --recover-from-tag ]]; then
        export_release_snapshot_registries "$snapshot"
        staging_tag="$(release_snapshot_staging_tag "$version" "$snapshot")"
    fi
    for part in assets docker tools; do
        restore_release_input "$snapshot" "$part"
    done
else
    newest_version="$(latest_collector_release_tag)"
    if collector_version_is_older "$newest_version" "$latest_published"; then
        newest_version="$latest_published"
    fi
    if collector_version_is_older "$version" "$newest_version"; then
        echo "Cannot publish a new stable version $version after $newest_version" >&2
        exit 1
    fi
    (cd dist/assets && sha256sum --quiet -c checksums.txt)
    snapshot="$(capture_release_snapshot)"
fi

if [[ "$mode" == --recover-from-tag ]]; then
    release_commit="$(jq -er '.commit' <<< "$snapshot")"
    if ! git cat-file -e "$release_commit:CHANGELOG.md"; then
        git fetch --no-tags https://github.com/signalfx/splunk-otel-collector.git "refs/tags/$version"
    fi
    release_notes="$(git show "$release_commit:CHANGELOG.md" | ./packaging/release/gh-release-notes.sh "$version" /dev/stdin)"
else
    release_notes="$(./packaging/release/gh-release-notes.sh "$version")"
fi
test -x dist/github-release-tools/ghr
export PATH="$PWD/dist/github-release-tools:$PATH"
if [[ -n "$published_ref" ]]; then
    bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}" verify
else
    test -x dist/github-release-tools/multimod
    bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}"
    ./.gitlab/create-github-tag.sh "$version" "$CI_COMMIT_SHA" "$snapshot"
fi
# Preserve a published release, while allowing a retry to finish a draft.
ghr_args=(-t "$GITHUB_TOKEN" -u signalfx -r splunk-otel-collector -n "$version" -b "$release_notes" --soft --replace "$version" dist/assets/)
if collector_version_is_older "$version" "$latest_published"; then
    ghr --draft "${ghr_args[@]}"
    publish_superseded_draft
else
    ghr "${ghr_args[@]}"
fi
