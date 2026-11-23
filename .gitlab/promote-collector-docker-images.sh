#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh

promote_tag() {
    local repository="$1" digest="$2" target="$3" attempt
    for attempt in 1 2 3; do
        if copy_collector_image "$repository" "$digest" "$target"; then
            return 0
        fi
        if (( attempt < 3 )); then
            echo "Could not promote $target on attempt $attempt. Retrying." >&2
            sleep "$((attempt * 2))"
        fi
    done
    echo "Could not promote $target after 3 attempts." >&2
    return 1
}

image_tag="$1"
target_tag="$2"
action="${3:-publish}"
if [[ -z "$image_tag" || -z "$target_tag" || ( "$action" != publish && "$action" != verify ) ]]; then
    echo "Usage: $0 <source-tag> <target-tag> [publish|verify]" >&2
    exit 1
fi

source_dir="dist/docker/${DOCKER_HUB_COLLECTOR_REPO}"
images=()
digests=()
suffixes=()
repository_suffix=""
if [[ "$image_tag" =~ ^[0-9a-f]{40}-staging-[0-9]+$ ]]; then
    repository_suffix="-dev"
fi

# Assemble the complete promotion plan before changing any mutable tags.
for artifact in multiarch multiarch-fips windows_multiarch windows_multiarch-fips; do
    expected_suffixes=('' -2019 -2022)
    case "$artifact" in
        multiarch)
            variant=""
            expected_suffixes+=(-amd64 -arm64 -ppc64le)
            ;;
        multiarch-fips)
            variant="-fips"
            expected_suffixes+=(-amd64 -arm64)
            ;;
        windows_multiarch) variant="-windows" ;;
        windows_multiarch-fips) variant="-fips-windows" ;;
    esac
    repository="${QUAY_COLLECTOR_REPO}${variant}${repository_suffix}"
    pending_refs=()
    for suffix in "${expected_suffixes[@]}"; do
        pending_refs+=("${repository}:${image_tag}${suffix}")
    done
    records="${source_dir}/images_${artifact}.txt"
    if [[ ! -s "$records" ]]; then
        echo "Missing image records for $artifact" >&2
        exit 1
    fi
    while read -r image digest || [[ -n "$image$digest" ]]; do
        if [[ ! "$digest" =~ ^sha256:[a-f0-9]{64}$ ]]; then
            echo "Invalid image digest for $image in $records" >&2
            exit 1
        fi
        for i in "${!pending_refs[@]}"; do
            if [[ "$image" == "${pending_refs[i]}" ]]; then
                unset 'pending_refs[i]'
                images+=("$repository")
                digests+=("$digest")
                suffixes+=("${image#"${repository}:${image_tag}"}")
                continue 2
            fi
        done
        echo "Unexpected or duplicate image record for $image in $records" >&2
        exit 1
    done < "$records"
    if [[ -n "${pending_refs[*]:-}" ]]; then
        echo "Missing image records in $records for: ${pending_refs[*]}" >&2
        exit 1
    fi
done

login_to_quay
login_to_docker_hub
for i in "${!images[@]}"; do
    quay_repo="${images[i]}"
    hub_repo="${DOCKER_HUB_COLLECTOR_REPO}${quay_repo#"$QUAY_COLLECTOR_REPO"}"
    tag="${target_tag}${suffixes[i]}"
    if [[ "$action" == verify ]]; then
        verify_collector_image "${digests[i]}" "${quay_repo}:${tag}"
        verify_collector_image "${digests[i]}" "${hub_repo}:${tag}"
    else
        promote_tag "$quay_repo" "${digests[i]}" "${quay_repo}:${tag}"
        promote_tag "$hub_repo" "${digests[i]}" "${hub_repo}:${tag}"
    fi
done
