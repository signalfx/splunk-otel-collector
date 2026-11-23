#!/bin/bash

github_release_pages() {
    local endpoint="$1" page response count
    for ((page = 1; page <= 100; page++)); do
        response="$(collector_github_api "$endpoint?per_page=100&page=$page" --fail)" || return
        count="$(jq -ers '
            if length == 1 and (.[0] | type) == "array" and (.[0] | length) <= 100
            then .[0] | length else error("Invalid GitHub response") end
        ' <<< "$response")" || return
        printf '%s\n' "$response"
        if (( count < 100 )); then
            return 0
        fi
    done
    echo "Could not read all GitHub pages for $endpoint" >&2
    return 1
}

github_release_id() {
    jq -ers --arg version "$1" '
        select(length == 1) | .[0]
        | select(.tag_name == $version and .prerelease == false and (.draft | type) == "boolean")
        | .id | select(type == "number" and . > 0 and . == floor)
    ' <<< "$2"
}

local_release_assets() (
    local path name checksum size
    shopt -s dotglob
    for path in dist/assets/*; do
        name="${path##*/}"
        if [[ ! -f "$path" || ! "$name" =~ ^[A-Za-z0-9_-][A-Za-z0-9_.-]*$ || "$name" == *. ]]; then
            echo "Invalid GitHub release asset path $path" >&2
            return 1
        fi
        checksum="$(sha256sum -- "$path")" || return
        size="$(wc -c < "$path")" || return
        jq -cn --arg name "$name" --arg digest "sha256:${checksum%% *}" \
            --argjson size "${size//[[:space:]]/}" '{name:$name, digest:$digest, size:$size}' || return
    done
)

github_asset_upload_plan() {
    local id="$1" expected="$2"
    github_release_pages "releases/$id/assets" | jq -cs --argjson expected "$expected" '
        add as $assets
        | if any($assets[]; (.id | type) != "number" or .id <= 0 or .id != (.id | floor)
                    or (.name | type) != "string")
              or ($assets | length) != ($assets | unique_by(.name) | length)
              or ($assets | length) != ($assets | unique_by(.id) | length)
          then error("Invalid GitHub release asset inventory") else . end
        | if any($assets[]; .name as $name | ($expected | any(.name == $name) | not))
          then error("GitHub release contains assets outside the signed snapshot") else . end
        | [$expected[] as $wanted | ($assets | map(select(.name == $wanted.name)) | first) as $asset
            | if $asset == null then {name:$wanted.name}
              elif $asset.state == "starter" and $asset.size == 0 then {name:$wanted.name, starter:$asset.id}
              elif $asset.state == "uploaded" and $asset.size == $wanted.size and $asset.digest == $wanted.digest then empty
              else error("GitHub release asset does not match the signed snapshot: " + $wanted.name) end]
    '
}
