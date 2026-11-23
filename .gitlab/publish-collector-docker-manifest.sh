#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh

manifest="$1"
fips="$2"
version_tag="$3"
image_tag="$4"
artifact_name="${manifest}${fips}"
artifact_dir="dist/docker/${QUAY_COLLECTOR_REPO}"
records="${artifact_dir}/images_${artifact_name}.txt"

repository="${QUAY_COLLECTOR_REPO}${fips}"
if [[ "$manifest" == windows_multiarch ]]; then
    repository="${repository}-windows"
fi
if [[ -z "$version_tag" ]]; then
    repository="${repository}-dev"
fi

sources=()
linux_arches=()
: > "$records"
if [[ "$manifest" == multiarch ]]; then
    read -r linux_index < "${artifact_dir}/linux_index${fips}"
    sources+=("$linux_index")
    linux_arches=(amd64 arm64)
    if [[ -z "$fips" ]]; then
        linux_arches+=(ppc64le)
    fi
    for arch in "${linux_arches[@]}"; do
        cat "${artifact_dir}/tags_to_sign_${arch}${fips}" >> "$records"
    done
fi

windows_versions=()
for windows in 2019 2022; do
    image="${repository}:${image_tag}-${windows}"
    digest="$(awk -v image="$image" '$1 == image { print $2 }' "${artifact_dir}/tags_to_sign_${windows}${fips}")"
    if [[ ! "$digest" =~ ^sha256:[a-f0-9]{64}$ ]]; then
        echo "Missing Windows digest for $image" >&2
        exit 1
    fi
    reference="${repository}@${digest}"
    sources+=("$reference")
    os_version="$(docker buildx imagetools inspect --raw "$reference" | jq -er '.manifests[0].platform."os.version"')"
    windows_versions+=("$os_version")
    printf '%s %s\n' "$image" "$digest" >> "$records"
done

image="${repository}:${image_tag}"
if [[ "$image_tag" =~ ^[0-9]+\.[0-9]+\.[0-9]+-staging-[0-9]+$ ]]; then
    printf 'Publishing image %s\n' "$image"
fi
digest="$(push_collector_index "$image" "${sources[@]}")"
json="$(docker buildx imagetools inspect --raw "${repository}@${digest}")"
jq -e --argjson expected "$((${#linux_arches[@]} + ${#windows_versions[@]}))" \
    '.manifests | length == $expected' <<< "$json" > /dev/null
for arch in "${linux_arches[@]}"; do
    jq -e --arg arch "$arch" 'any(.manifests[]; .platform.os == "linux" and .platform.architecture == $arch)' \
        <<< "$json" > /dev/null
done
for os_version in "${windows_versions[@]}"; do
    jq -e --arg os_version "$os_version" \
        'any(.manifests[]; .platform.os == "windows" and .platform.architecture == "amd64" and .platform."os.version" == $os_version)' \
        <<< "$json" > /dev/null
done

printf '%s %s\n' "$image" "$digest" | tee "${artifact_dir}/tags_to_sign_${artifact_name}" >> "$records"
printf '[%s@%s]\n' "$repository" "$digest" | tee "${artifact_dir}/${artifact_name}_digest.txt"
