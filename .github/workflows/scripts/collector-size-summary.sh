#!/bin/bash

set -euo pipefail

ARCHES=("amd64" "arm64" "ppc64le")

format_bytes() {
    local bytes=$1
    if [[ -z "$bytes" || "$bytes" == "null" ]]; then
        echo "not available"
        return
    fi
    awk -v bytes="$bytes" 'BEGIN {
        mib = bytes / 1048576
        if (mib >= 1024) {
            printf "%.2f GiB", mib / 1024
        } else {
            printf "%.1f MiB", mib
        }
    }'
}

format_delta_cell() {
    local base=$1
    local current=$2
    if [[ -z "$current" || "$current" == "null" ]]; then
        echo "not available"
        return
    fi

    local current_fmt
    current_fmt=$(format_bytes "$current")
    if [[ -z "$base" || "$base" == "null" ]]; then
        echo "$current_fmt"
        return
    fi

    local delta=$((current - base))
    if [[ "$delta" -eq 0 ]]; then
        echo "$current_fmt (no change)"
        return
    fi

    local sign="+"
    if [[ "$delta" -lt 0 ]]; then
        sign="-"
    fi

    local delta_abs=${delta#-}
    local delta_fmt percent
    delta_fmt=$(format_bytes "$delta_abs")
    percent=$(awk -v delta="$delta" -v base="$base" 'BEGIN {
        if (base == 0) {
            printf "n/a"
        } else {
            printf "%+.2f%%", delta * 100 / base
        }
    }')
    echo "$current_fmt (${sign}${delta_fmt}, ${percent})"
}

format_commit() {
    local sha=$1
    if [[ -z "$sha" || "$sha" == "null" || "$sha" == "not available" ]]; then
        echo "not available"
        return
    fi
    echo "\`${sha:0:12}\`"
}

load_image() {
    local image_tar=$1
    local tag=$2

    docker load -i "$image_tar"
    docker tag otelcol:latest "otelcol:${tag}"
    docker rmi otelcol:latest >/dev/null 2>&1 || true
}

extract_binary_size() {
    local tag=$1
    local arch=$2
    local output_dir=$3
    local cid size

    cid=$(docker create --platform "linux/${arch}" "otelcol:${tag}")
    if ! docker cp "${cid}:/otelcol" "${output_dir}/otelcol-${tag}-${arch}"; then
        docker rm "$cid" >/dev/null
        return 1
    fi
    size=$(stat -c%s "${output_dir}/otelcol-${tag}-${arch}")
    docker rm "$cid" >/dev/null
    echo "$size"
}

write_summary() {
    local summary_file=$1
    local base_json=$2
    local pr_json=$3

    local base_sha pr_sha base_archive pr_archive
    base_sha=$(jq -r '.sha // "not available"' "$base_json")
    pr_sha=$(jq -r '.sha // "not available"' "$pr_json")
    base_archive=$(jq -r '.image_archive_bytes // null' "$base_json")
    pr_archive=$(jq -r '.image_archive_bytes // null' "$pr_json")

    {
        echo "### Collector Size Summary"
        echo
        echo "| Metric | Base | PR |"
        echo "|---|---:|---:|"
        echo "| Commit | $(format_commit "$base_sha") | $(format_commit "$pr_sha") |"
        echo "| Multi-arch image archive | $(format_bytes "$base_archive") | $(format_delta_cell "$base_archive" "$pr_archive") |"

        for arch in "${ARCHES[@]}"; do
            local base_binary pr_binary
            base_binary=$(jq -r --arg arch "$arch" '.binaries[$arch] // null' "$base_json")
            pr_binary=$(jq -r --arg arch "$arch" '.binaries[$arch] // null' "$pr_json")
            echo "| linux/${arch} /otelcol | $(format_bytes "$base_binary") | $(format_delta_cell "$base_binary" "$pr_binary") |"
        done

        echo
        echo "Source: existing \`linux-package-test / docker-otelcol\` image artifact."
    } >>"$summary_file"
}

collect_image_sizes() {
    local tag=$1
    local image_tar=$2
    local sha=$3
    local output_json=$4
    local work_dir=$5

    local binaries_json="{}"
    for arch in "${ARCHES[@]}"; do
        local binary_size
        binary_size=$(extract_binary_size "$tag" "$arch" "$work_dir")
        binaries_json=$(jq -c --arg arch "$arch" --argjson size "$binary_size" '. + {($arch): $size}' <<<"$binaries_json")
    done

    jq -n \
        --arg sha "$sha" \
        --argjson imageArchiveBytes "$(stat -c%s "$image_tar")" \
        --argjson binaries "$binaries_json" \
        '{
            sha: $sha,
            image_archive_bytes: $imageArchiveBytes,
            binaries: $binaries
        }' >"$output_json"
}

main() {
    local pr_image_tar=${PR_IMAGE_TAR:?PR_IMAGE_TAR is required}
    local pr_sha=${PR_SHA:?PR_SHA is required}
    local base_image_tar=${BASE_IMAGE_TAR:-}
    local base_sha=${BASE_SHA:-not available}
    local summary_file=${GITHUB_STEP_SUMMARY:-/dev/stdout}
    local output_json=${OUTPUT_JSON:-collector-size-summary.json}
    local tmp_dir

    tmp_dir=$(mktemp -d)
    trap 'rm -rf "${tmp_dir:-}"' EXIT

    local base_json="$tmp_dir/base.json"
    jq -n --arg sha "$base_sha" '{sha: $sha, image_archive_bytes: null, binaries: {}}' >"$base_json"

    if [[ -n "$base_image_tar" && -f "$base_image_tar" ]]; then
        load_image "$base_image_tar" base
        collect_image_sizes base "$base_image_tar" "$base_sha" "$base_json" "$tmp_dir"
    fi

    load_image "$pr_image_tar" pr
    local pr_json="$tmp_dir/pr.json"
    collect_image_sizes pr "$pr_image_tar" "$pr_sha" "$pr_json" "$tmp_dir"

    jq -n \
        --slurpfile base "$base_json" \
        --slurpfile pr "$pr_json" \
        '{base: $base[0], pr: $pr[0]}' >"$output_json"

    write_summary "$summary_file" "$base_json" "$pr_json"
}

main "$@"
