#!/bin/bash

set -euo pipefail

manifest="$1"
fips="$2"
latest_tag="$3"
artifact_name="${manifest}${fips}"
source_dir="dist/docker/${QUAY_COLLECTOR_REPO}"

read -r source_image < "${source_dir}/tags_to_sign_${artifact_name}"
read -r source_digest_entry < "${source_dir}/${artifact_name}_digest.txt"
source_repo="${source_image%:*}"
version="${source_image##*:}"
source_digest="${source_digest_entry##*@}"
source_digest="${source_digest%]}"
target_repo="${DOCKER_HUB_COLLECTOR_REPO}${source_repo#"$QUAY_COLLECTOR_REPO"}"

docker_hub_path="${DOCKER_HUB_COLLECTOR_REPO#docker.io/}"
# Docker Hub organization tokens use the destination namespace as the username.
printf '%s\n' "${DOCKER_HUB_TOKEN:?DOCKER_HUB_TOKEN is required}" |
    docker login --username "${DOCKER_HUB_USERNAME:-${docker_hub_path%%/*}}" --password-stdin

get_digest() {
    docker buildx imagetools inspect --format '{{json .Manifest}}' "$1" |
        jq -er '.digest | select(test("^sha256:[a-f0-9]{64}$"))'
}

copy_image() {
    local repository="$1" digest="$2" target="$3" actual_digest
    docker buildx imagetools create --prefer-index=false --tag "$target" "${repository}@${digest}"
    actual_digest="$(get_digest "$target")"
    if [[ "$actual_digest" != "$digest" ]]; then
        echo "Digest mismatch for $target: expected $digest, got $actual_digest" >&2
        return 1
    fi
}

# Publish the individual platform tags before the manifest that contains them.
platforms=(2019 2022)
if [[ "$manifest" == multiarch ]]; then
    platforms+=(amd64 arm64)
    if [[ -z "$fips" ]]; then
        platforms+=(ppc64le)
    fi
fi

for platform in "${platforms[@]}" ""; do
    suffix="${platform:+-${platform}}"
    digest="$source_digest"
    if [[ -n "$platform" ]]; then
        digest="$(get_digest "${source_repo}:${version}${suffix}")"
    fi
    target="${target_repo}:${version}${suffix}"
    copy_image "$source_repo" "$digest" "$target"
    artifact-ci sign docker "$target"
    if [[ -n "$latest_tag" ]]; then
        copy_image "$target_repo" "$digest" "${target_repo}:${latest_tag}${suffix}"
    fi
done

mkdir -p "dist/docker/${DOCKER_HUB_COLLECTOR_REPO}"
printf '[%s@%s]\n' "$target_repo" "$source_digest" |
    tee "dist/docker/${DOCKER_HUB_COLLECTOR_REPO}/${artifact_name}_digest.txt"
