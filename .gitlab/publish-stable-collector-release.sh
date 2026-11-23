#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh
source .gitlab/collector-release-snapshot.sh
source .gitlab/collector-release-assets.sh

version="$1"
staging_tag="$2"
mode="${3:-}"
if [[ -n "$mode" && "$mode" != --recover-only && "$mode" != --recover-from-tag ]]; then
    echo "Usage: $0 <version> <staging-tag> [--recover-only|--recover-from-tag]" >&2
    exit 1
fi

github_release_for_tag() {
    local response status
    response="$(collector_github_api "releases/tags/$version" --write-out $'\n%{http_code}')" || return
    status="${response##*$'\n'}"
    case "$status" in
        200)
            response="${response%$'\n'*}"
            github_release_id "$version" "$response" > /dev/null || return
            printf '%s\n' "$response"
            ;;
        404)
            # The tag endpoint only returns published releases. Drafts are visible in the authenticated list.
            github_release_pages releases | jq -cs --arg version "$version" '
                [.[][] | select(.tag_name == $version)]
                | if length <= 1 then .[] else error("Several GitHub releases match the tag") end
            '
            ;;
        *)
            echo "Could not check publication of $version (HTTP $status)" >&2
            return 1
            ;;
    esac
}

finish_github_assets() {
    local id="$1" expected="$2" plan uploads name starter encoded_name
    plan="$(github_asset_upload_plan "$id" "$expected")" || return
    if [[ "$plan" == '[]' ]]; then
        return 0
    fi
    uploads="$(jq -r '.[] | [.name, .starter] | @tsv' <<< "$plan")" || return
    while IFS=$'\t' read -r name starter; do
        if [[ -n "$starter" ]]; then
            # GitHub can leave an empty placeholder after an upload fails upstream.
            collector_github_api "releases/assets/$starter" --fail --request DELETE > /dev/null || return
        fi
        encoded_name="$(jq -rn --arg name "$name" '$name | @uri')" || return
        curl --silent --show-error --fail --header "Authorization: Bearer $GITHUB_TOKEN" \
            --header 'Accept: application/vnd.github+json' --header 'Content-Type: application/octet-stream' \
            --data-binary "@dist/assets/$name" \
            "https://uploads.github.com/repos/signalfx/splunk-otel-collector/releases/$id/assets?name=$encoded_name" \
            > /dev/null || return
    done <<< "$uploads"
    plan="$(github_asset_upload_plan "$id" "$expected")" || return
    if [[ "$plan" != '[]' ]]; then
        echo "GitHub release $version still has incomplete asset uploads" >&2
        return 1
    fi
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
    release="$(github_release_for_tag)"
    if [[ "$mode" == --recover-from-tag ]]; then
        export_release_snapshot_registries "$snapshot"
        staging_tag="$(release_snapshot_staging_tag "$version" "$snapshot")"
    fi
    for part in assets docker; do
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
    release_notes="$(git show "$release_commit:CHANGELOG.md" | ./packaging/release/gh-release-notes.sh "$version" "$staging_tag" /dev/stdin)"
else
    release_notes="$(./packaging/release/gh-release-notes.sh "$version" "$staging_tag")"
fi
expected_assets="$(local_release_assets | jq -cs '.')"
if [[ -n "$published_ref" ]]; then
    bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}" verify
else
    test -x dist/github-release-tools/multimod
    export PATH="$PWD/dist/github-release-tools:$PATH"
    bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}"
    ./.gitlab/create-github-tag.sh "$version" "$CI_COMMIT_SHA" "$snapshot"
    release="$(github_release_for_tag)"
fi

if [[ -z "$release" ]]; then
    payload="$(jq -cn --arg tag "$version" --arg notes "$release_notes" \
        '{tag_name:$tag,name:$tag,body:$notes,draft:true,prerelease:false}')"
    release="$(collector_github_api releases --fail --request POST \
        --header 'Content-Type: application/json' --data "$payload")"
    jq -e '.draft == true' <<< "$release" > /dev/null
fi
release_id="$(github_release_id "$version" "$release")"
finish_github_assets "$release_id" "$expected_assets"

if [[ "$(jq -r '.draft' <<< "$release")" == true ]]; then
    latest_published="$(latest_published_collector_release)"
    make_latest=true
    if collector_version_is_older "$version" "$latest_published"; then
        make_latest=false
    fi
    payload="$(jq -cn --arg latest "$make_latest" '{draft:false,make_latest:$latest}')"
    published="$(collector_github_api "releases/$release_id" --fail --request PATCH \
        --header 'Content-Type: application/json' --data "$payload")"
    [[ "$(github_release_id "$version" "$published")" == "$release_id" ]]
    jq -e '.draft == false' <<< "$published" > /dev/null
fi
