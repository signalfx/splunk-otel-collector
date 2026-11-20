#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh

image_tag="$1"
latest_tag="$2"
source_dir="dist/docker/${DOCKER_HUB_COLLECTOR_REPO}"
images=()
digests=()
suffixes=()

# Assemble the complete promotion plan before changing any mutable tags.
for artifact in multiarch multiarch-fips windows_multiarch windows_multiarch-fips; do
    records="${source_dir}/images_${artifact}.txt"
    if [[ ! -s "$records" ]]; then
        echo "Missing image records for $artifact" >&2
        exit 1
    fi
    while read -r image digest; do
        tag="${image##*:}"
        if [[ "$tag" != "$image_tag" && "$tag" != "$image_tag"-* ]] || [[ ! "$digest" =~ ^sha256:[a-f0-9]{64}$ ]]; then
            echo "Unexpected image record for $image" >&2
            exit 1
        fi
        images+=("${image%:*}")
        digests+=("$digest")
        suffixes+=("${tag#"$image_tag"}")
    done < "$records"
done

login_to_quay
login_to_docker_hub
for i in "${!images[@]}"; do
    quay_repo="${images[i]}"
    hub_repo="${DOCKER_HUB_COLLECTOR_REPO}${quay_repo#"$QUAY_COLLECTOR_REPO"}"
    alias="${latest_tag}${suffixes[i]}"
    copy_collector_image "$quay_repo" "${digests[i]}" "${quay_repo}:${alias}"
    copy_collector_image "$hub_repo" "${digests[i]}" "${hub_repo}:${alias}"
done
