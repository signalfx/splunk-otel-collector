#!/bin/bash

login_to_quay() {
    local username="${CIRCLECI_QUAY_USERNAME:-}" password="${CIRCLECI_QUAY_PASSWORD:-}"
    if [[ -n "${DRY_RUN:-}" ]]; then
        username="${QUAY_DEV_USERNAME:?QUAY_DEV_USERNAME is required for a Docker dry run}"
        password="${QUAY_DEV_PASSWORD:?QUAY_DEV_PASSWORD is required for a Docker dry run}"
    fi
    printf '%s\n' "$password" | docker login --username "$username" --password-stdin quay.io
}

login_to_docker_hub() {
    local path="${DOCKER_HUB_COLLECTOR_REPO#docker.io/}"
    local username="${DOCKER_HUB_USERNAME:-${path%%/*}}" token
    if [[ -n "${DRY_RUN:-}" ]]; then
        username="${DOCKER_HUB_DEV_USERNAME:-${path%%/*}}"
        token="${DOCKER_HUB_DEV_TOKEN:?DOCKER_HUB_DEV_TOKEN is required for a Docker dry run}"
    else
        token="${DOCKER_HUB_TOKEN:?DOCKER_HUB_TOKEN is required}"
    fi
    printf '%s\n' "$token" | docker login --username "$username" --password-stdin
}

latest_collector_release_tag() {
    local tags
    tags="$(git ls-remote --tags --refs --sort=-version:refname \
        https://github.com/signalfx/splunk-otel-collector.git 'refs/tags/v*')" || return
    awk '$2 ~ /^refs\/tags\/v[0-9]+\.[0-9]+\.[0-9]+$/ {
        sub(/^refs\/tags\//, "", $2)
        print $2
        exit
    }' <<< "$tags"
}

read_build_digest() {
    jq -er '."containerimage.digest" | select(test("^sha256:[a-f0-9]{64}$"))' "$1"
}

# These indexes are assembled from digest references or descriptors. Hash the
# exact JSON Buildx will publish and confirm the result is retrievable by digest.
push_collector_index() {
    local tag="$1" json checksum digest
    shift
    json="$(docker buildx imagetools create --dry-run --tag "$tag" "$@")" || return
    checksum="$(printf '%s' "$json" | sha256sum)" || return
    digest="sha256:${checksum%% *}"
    docker buildx imagetools create --tag "$tag" "$@" >&2 || return
    docker buildx imagetools inspect --raw "${tag%:*}@${digest}" > /dev/null || return
    printf '%s\n' "$digest"
}

copy_collector_image() {
    local source="$1" digest="$2" target="$3" actual_digest
    docker buildx imagetools create --tag "$target" "${source}@${digest}" || return
    actual_digest="$(docker buildx imagetools inspect --format '{{json .Manifest}}' "$target" | jq -er '.digest')" || return
    if [[ "$actual_digest" != "$digest" ]]; then
        echo "Digest mismatch for $target: expected $digest, got $actual_digest" >&2
        return 1
    fi
}
