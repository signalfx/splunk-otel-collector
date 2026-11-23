#!/bin/bash

release_input_path() {
    case "$1" in
        assets) echo dist/assets ;;
        docker) echo dist/docker ;;
        tools) echo dist/github-release-tools ;;
        *) return 1 ;;
    esac
}

release_input_digest() {
    local unsupported
    unsupported="$(find "$1" ! -type d ! -type f -print -quit)" || return
    if [[ -n "$unsupported" ]]; then
        echo "Unexpected file type in release input $unsupported" >&2
        return 1
    fi
    (cd "$1" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 -r sha256sum --) |
        sha256sum | cut -d ' ' -f 1
}

release_input_record() {
    local part="$1" path job digest
    path="$(release_input_path "$part")" || return
    job="$(cat "dist/release-inputs/$part")" || return
    [[ "$job" =~ ^[1-9][0-9]*$ ]] || return 1
    digest="$(release_input_digest "$path")" || return
    jq -cn --arg job "$job" --arg sha256 "$digest" '{job:$job, sha256:$sha256}'
}

capture_release_snapshot() {
    local assets docker tools
    assets="$(release_input_record assets)" || return
    docker="$(release_input_record docker)" || return
    tools="$(release_input_record tools)" || return
    jq -cn --arg project "${CI_PROJECT_ID:?}" --arg pipeline "${CI_PIPELINE_ID:?}" \
        --arg quay "${QUAY_COLLECTOR_REPO:?}" --arg hub "${DOCKER_HUB_COLLECTOR_REPO:?}" \
        --argjson assets "$assets" --argjson docker "$docker" --argjson tools "$tools" \
        '{project:$project, pipeline:$pipeline, registries:{quay:$quay, hub:$hub}, assets:$assets, docker:$docker, tools:$tools}'
}

read_release_snapshot() {
    local version="$1" tag_ref="$2" mode="${3:-}" tag_sha ref tag snapshot
    read -r tag_sha ref <<< "$tag_ref"
    if [[ ! "$tag_sha" =~ ^[0-9a-f]{40}$ || "$ref" != "refs/tags/$version" ]]; then
        echo "Could not resolve the published tag for $version" >&2
        return 1
    fi
    tag="$(collector_github_api "git/tags/$tag_sha" --fail)" || return
    # GitHub verifies that a GPG tag's email belongs to the signing account.
    if ! jq -e '.verification.verified == true and .tagger.email == "ssg-srv-gh-o11y-gdi@splunk.com"' <<< "$tag" > /dev/null; then
        echo "$version is missing a verified signature from the collector release bot" >&2
        return 1
    fi
    if ! snapshot="$(jq -ce --arg version "$version" '
        select(.tag == $version and .object.type == "commit" and (.object.sha | test("^[0-9a-f]{40}$")))
        | .object.sha as $commit
        | .message | split("\n") | map(select(startswith("Collector-Release-Snapshot: ")))
        | select(length == 1) | .[0] | ltrimstr("Collector-Release-Snapshot: ") | fromjson
        | . + {commit:$commit}
    ' <<< "$tag")"; then
        echo "$version does not identify a release snapshot" >&2
        return 1
    fi
    if ! jq -e --arg project "${CI_PROJECT_ID:?}" '.project == $project and (.pipeline | test("^[1-9][0-9]*$"))' <<< "$snapshot" > /dev/null; then
        echo "$version does not identify a release pipeline in this GitLab project" >&2
        return 1
    fi
    if [[ "$mode" != --recover-from-tag ]] && ! jq -e --arg pipeline "${CI_PIPELINE_ID:?}" \
        --arg commit "${CI_COMMIT_SHA:?}" '.pipeline == $pipeline and .commit == $commit' <<< "$snapshot" > /dev/null; then
        echo "Retry $version in its original release pipeline or start a recovery-only pipeline on main" >&2
        return 1
    fi
    printf '%s\n' "$snapshot"
}

export_release_snapshot_registries() {
    QUAY_COLLECTOR_REPO="$(jq -er '.registries.quay | strings | select(length > 0)' <<< "$1")" || return
    DOCKER_HUB_COLLECTOR_REPO="$(jq -er '.registries.hub | strings | select(length > 0)' <<< "$1")" || return
    export QUAY_COLLECTOR_REPO DOCKER_HUB_COLLECTOR_REPO
}

release_snapshot_staging_tag() {
    jq -er --arg version "${1#v}" '$version + "-staging-" + .pipeline' <<< "$2"
}

restore_release_input() (
    local snapshot="$1" part="$2" path job expected current digest workdir
    path="$(release_input_path "$part")"
    job="$(jq -er --arg part "$part" '.[$part].job | select(test("^[1-9][0-9]*$"))' <<< "$snapshot")"
    expected="$(jq -er --arg part "$part" '.[$part].sha256 | select(test("^[0-9a-f]{64}$"))' <<< "$snapshot")"
    current="$(cat "dist/release-inputs/$part" 2>/dev/null || true)"

    if [[ "$current" == "$job" ]]; then
        digest="$(release_input_digest "$path")"
        if [[ "$digest" != "$expected" ]]; then
            echo "Checksum mismatch for release input $part from job $job" >&2
            return 1
        fi
        return 0
    fi

    # GitLab needs can supply artifacts from a newer attempt of the same job.
    echo "Restoring release input $part from its original job $job"
    if ! command -v unzip > /dev/null; then
        apt-get update
        apt-get install -y unzip
    fi
    mkdir -p dist/release-inputs
    workdir="$(mktemp -d "$PWD/dist/.release-inputs.XXXXXX")"
    trap 'rm -rf "$workdir"' EXIT
    curl --fail --silent --show-error --location --output "$workdir/artifacts.zip" \
        "${CI_API_V4_URL:?}/projects/${CI_PROJECT_ID:?}/jobs/$job/artifacts?job_token=${CI_JOB_TOKEN:?}"
    unzip -q "$workdir/artifacts.zip" "$path/*" "dist/release-inputs/$part" -d "$workdir"
    digest="$(release_input_digest "$workdir/$path")"
    if [[ "$digest" != "$expected" || "$(cat "$workdir/dist/release-inputs/$part")" != "$job" ]]; then
        echo "Checksum or job ID mismatch for restored release input $part" >&2
        return 1
    fi
    rm -rf -- "$path"
    mv -- "$workdir/$path" "$path"
    mv -- "$workdir/dist/release-inputs/$part" "dist/release-inputs/$part"
)
