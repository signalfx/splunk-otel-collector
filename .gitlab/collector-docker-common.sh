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

collector_release_refs() {
    git ls-remote --tags --sort=-version:refname \
        https://github.com/signalfx/splunk-otel-collector.git "$@"
}

collector_release_tag_commit() {
    local ref="refs/tags/$1" refs
    refs="$(collector_release_refs "$ref" "$ref^{}")" || return
    # Prefer the target of an annotated tag. A lightweight tag points directly at the commit.
    awk -v ref="$ref" '$2 == ref { tag = $1 } $2 == ref "^{}" { commit = $1 }
        END { if (tag) print (commit ? commit : tag) }' <<< "$refs"
}

latest_collector_release_tag() {
    local tags
    tags="$(collector_release_refs 'refs/tags/v*')" || return
    awk '$2 ~ /^refs\/tags\/v[0-9]+\.[0-9]+\.[0-9]+$/ {
        sub(/^refs\/tags\//, "", $2)
        print $2
        exit
    }' <<< "$tags"
}

collector_github_api() {
    local endpoint="$1"
    shift
    curl --silent --show-error --header 'Accept: application/vnd.github+json' \
        --header "Authorization: Bearer ${GITHUB_TOKEN:?}" \
        "$@" "https://api.github.com/repos/signalfx/splunk-otel-collector/$endpoint"
}

latest_collector_github_release() {
    local response
    response="$(collector_github_api releases/latest --write-out $'\n%{http_code}')" || return
    case "${response##*$'\n'}" in
        200)
            jq -ces '
                select(length == 1) | .[0]
                | select(.draft == false and .prerelease == false and (.tag_name | test("^v[0-9]+\\.[0-9]+\\.[0-9]+$")))
            ' <<< "${response%$'\n'*}"
            ;;
        404) return 0 ;;
        *)
            echo "Could not determine latest published GitHub release (HTTP ${response##*$'\n'})" >&2
            return 1
            ;;
    esac
}

latest_published_collector_release() {
    local release
    release="$(latest_collector_github_release)" || return
    jq -r '.tag_name // empty' <<< "$release"
}

collector_version_is_older() {
    [[ -n "$2" && "$1" != "$2" && "$(printf '%s\n' "$1" "$2" | sort -V | sed -n '1p')" == "$1" ]]
}

newest_collector_dev_pipeline() {
    local page response summary count latest current current_seen=false newest=0
    # The resource group holds the lock before GitLab starts this deployment.
    # A failed or canceled job can have written aliases if it started running.
    for ((page = 1; page <= 100; page++)); do
        response="$(curl --fail --silent --show-error --max-time 30 --get \
            --header "JOB-TOKEN: ${CI_JOB_TOKEN:?}" \
            --data-urlencode "environment=${CI_ENVIRONMENT_NAME:?}" \
            --data-urlencode "page=$page" --data-urlencode 'per_page=100' \
            --data-urlencode 'order_by=id' --data-urlencode 'sort=asc' \
            "${CI_API_V4_URL:?}/projects/${CI_PROJECT_ID:?}/deployments")" || return
        summary="$(jq -er --arg environment "$CI_ENVIRONMENT_NAME" --arg sha "${CI_COMMIT_SHA:?}" \
            --arg job "${CI_JOB_ID:?}" --arg pipeline "${CI_PIPELINE_ID:?}" '
            if type != "array" then error("Invalid development deployment list") else . end
            | map(
                if .environment.name != $environment or .sha != $sha
                then error("Unexpected development deployment") else . end
                | .status as $status
                | if ["created","blocked","running","success","failed","canceled","skipped"] | index($status) != null
                  then . else error("Invalid development deployment status") end
                | (.deployable.pipeline.id
                   | if type == "number" and . > 0 and . == floor
                     then . else error("Invalid development deployment pipeline ID") end) as $id
                | {status: $status, pipeline: $id, job: .deployable.id,
                   never_started: (.deployable | has("started_at") and .started_at == null)}
            )
            | [length,
               ([.[] | select(.status == "running" or .status == "success"
                              or ((.status == "failed" or .status == "canceled") and (.never_started | not)))
                      | .pipeline] | max // 0),
               any(.[]; (.job | tostring) == $job and (.pipeline | tostring) == $pipeline)]
            | @tsv
        ' <<< "$response")" || return
        read -r count latest current <<< "$summary"
        if (( latest > newest )); then
            newest="$latest"
        fi
        if [[ "$current" == true ]]; then
            current_seen=true
        fi
        if (( count < 100 )); then
            if [[ "$current_seen" != true ]]; then
                echo "GitLab deployment history did not include the current promotion job" >&2
                return 1
            fi
            printf '%s\n' "$newest"
            return 0
        fi
    done
    echo "Could not read the complete development deployment history" >&2
    return 1
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
    local source="$1" digest="$2" target="$3"
    docker buildx imagetools create --tag "$target" "${source}@${digest}" || return
    verify_collector_image "$digest" "$target"
}

verify_collector_image() {
    local digest="$1" target="$2" actual_digest
    actual_digest="$(docker buildx imagetools inspect --format '{{json .Manifest}}' "$target" | jq -er '.digest')" || return
    if [[ "$actual_digest" != "$digest" ]]; then
        echo "Digest mismatch for $target: expected $digest, got $actual_digest" >&2
        return 1
    fi
}
