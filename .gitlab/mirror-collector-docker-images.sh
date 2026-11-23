#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh

artifact_name="$1"
source_dir="dist/docker/${QUAY_COLLECTOR_REPO}"
target_dir="dist/docker/${DOCKER_HUB_COLLECTOR_REPO}"
records="${source_dir}/images_${artifact_name}.txt"

if [[ ! -s "$records" ]]; then
    echo "Missing image records for $artifact_name" >&2
    exit 1
fi

login_to_docker_hub
while read -r image digest; do
    target="${DOCKER_HUB_COLLECTOR_REPO}${image#"$QUAY_COLLECTOR_REPO"}"
    if [[ "$target" =~ :[0-9]+\.[0-9]+\.[0-9]+-staging-[0-9]+($|-) ]]; then
        printf 'Publishing image %s\n' "$target"
    fi
    copy_collector_image "${image%:*}" "$digest" "$target"
    artifact-ci sign docker "${target%:*}@${digest}"
done < "$records"

read -r image digest < "${source_dir}/tags_to_sign_${artifact_name}"
repository="${DOCKER_HUB_COLLECTOR_REPO}${image#"$QUAY_COLLECTOR_REPO"}"
mkdir -p "$target_dir"
cp "$records" "${target_dir}/images_${artifact_name}.txt"
printf '[%s@%s]\n' "${repository%:*}" "$digest" | tee "${target_dir}/${artifact_name}_digest.txt"
