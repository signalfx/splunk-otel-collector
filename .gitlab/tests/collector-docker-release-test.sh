#!/bin/bash

set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

mkdir -p .gitlab packaging/release
cp "$repo"/.gitlab/{collector-docker-common,collector-release-assets,collector-release-snapshot,promote-collector-docker-images,promote-stable-collector-latest,publish-prerelease-collector-release,publish-stable-collector-release}.sh .gitlab/
cp "$repo/packaging/release/gh-release-notes.sh" packaging/release/
printf '## v1.2.3\nTest release\n' > CHANGELOG.md

export QUAY_COLLECTOR_REPO=quay.io/splunk/collector DOCKER_HUB_COLLECTOR_REPO=docker.io/splunk/collector
export CIRCLECI_QUAY_PASSWORD=test DOCKER_HUB_TOKEN=test GITHUB_TOKEN=test
export CI_PIPELINE_ID=100 CI_PROJECT_ID=200 CI_JOB_TOKEN=fixture CI_API_V4_URL=https://gitlab.invalid/api/v4
CI_COMMIT_SHA="$(printf '%040d' 100)"
export CI_COMMIT_SHA
unset DRY_RUN

# Exercise the shell blocks GitLab uses for tag selection and the Linux build.
(
    mkdir linux-build
    cd linux-build
    pipeline="$repo/.gitlab/collector-pipeline.yml"
    tag_setup="$(awk '/^\.set-docker-tags:/ { print_block=1; next } print_block && /^[^ ]/ { exit } print_block' "$pipeline")"
    linux_build="$(awk '
        /^  script: &build-push-linux-docker$/ { found_job=1 }
        found_job && /^    - \|$/ { print_block=1; next }
        print_block && /^  artifacts:/ { exit }
        print_block
    ' "$pipeline")"
    FIXTURE_DIGEST="sha256:$(printf '%064d' 1)"
    export FIXTURE_DIGEST
    source "$repo/.gitlab/collector-docker-common.sh"
    # The child Bash shell uses these stubs to execute the build block offline.
    # shellcheck disable=SC2329
    make() {
        printf '%s\n' "$@"
        printf '{"containerimage.digest":"%s"}\n' "$FIXTURE_DIGEST" > build_metadata.json
    }
    # shellcheck disable=SC2329
    docker() {
        [[ "$1 $2 $3 $4" == 'buildx imagetools inspect --raw' ]]
        printf '{"manifests":[{"platform":{"architecture":"amd64","os":"linux"}}]}\n'
    }
    # shellcheck disable=SC2329
    push_collector_index() {
        printf '%s\n' "$1" > platform-tag
        printf '%s\n' "$FIXTURE_DIGEST"
    }
    export -f make docker push_collector_index read_build_digest
    verify_linux_build_tag() {
        local version="$1" dry_run="$2" repository_suffix="$3" tag="$4" args
        args="$(RELEASE_VERSION="$version" DRY_RUN="$dry_run" CI_COMMIT_BRANCH=main CI_COMMIT_TAG='' \
            DOCKER_FIPS_SUFFIX='' DOCKER_ARCHES=amd64 DOCKER_HUB_REPO=docker.io \
            bash -euc "$tag_setup"$'\n'"$linux_build")"
        if ! grep -Fxq "IMAGE_NAME=${QUAY_COLLECTOR_REPO}${repository_suffix}" <<< "$args" ||
           ! grep -Fxq "IMAGE_TAG=${tag}-linux" <<< "$args"; then
            printf 'Expected Linux build tag %s-linux. Build arguments:\n%s\n' "$tag" "$args" >&2
            return 1
        fi
        grep -Fxq "${QUAY_COLLECTOR_REPO}${repository_suffix}:$tag-amd64" platform-tag
    }
    verify_linux_build_tag v1.2.3 "" "" 1.2.3-staging-100
    verify_linux_build_tag v1.2.3-beta.1 "" "" 1.2.3-beta.1-staging-100
    verify_linux_build_tag v1.2.3-rc.1 "" "" 1.2.3-rc.1-staging-100
    verify_linux_build_tag "" "" -dev "$CI_COMMIT_SHA-staging-$CI_PIPELINE_ID"
    verify_linux_build_tag "" true -dev "$CI_COMMIT_SHA-dryrun-$CI_PIPELINE_ID"
)

quay_dir="dist/docker/$QUAY_COLLECTOR_REPO"
hub_dir="dist/docker/$DOCKER_HUB_COLLECTOR_REPO"
mkdir -p "$quay_dir" "$hub_dir"
digest="sha256:$(printf '%064d' 1)"
other_digest="sha256:$(printf '%064d' 2)"
printf '[%s@%s]\n' "$QUAY_COLLECTOR_REPO" "$digest" > "$quay_dir/multiarch_digest.txt"
printf '[%s@%s]\n' "$DOCKER_HUB_COLLECTOR_REPO" "$digest" > "$hub_dir/multiarch_digest.txt"

write_image_records() {
    local stage="$1" suffix="${2:-}" fips tag_suffix
    local linux_suffixes
    for fips in '' -fips; do
        linux_suffixes=('' -amd64 -arm64 -2019 -2022)
        if [[ -z "$fips" ]]; then
            linux_suffixes+=(-ppc64le)
        fi
        for tag_suffix in "${linux_suffixes[@]}"; do
            printf '%s %s\n' "$QUAY_COLLECTOR_REPO$fips$suffix:$stage$tag_suffix" "$digest"
        done > "$hub_dir/images_multiarch$fips.txt"
        for tag_suffix in '' -2019 -2022; do
            printf '%s %s\n' "$QUAY_COLLECTOR_REPO$fips-windows$suffix:$stage$tag_suffix" "$digest"
        done > "$hub_dir/images_windows_multiarch$fips.txt"
    done
}

git() {
    case "$1" in
        ls-remote)
            if [[ "${!#}" == refs/heads/main ]]; then
                [[ "${FAILED_MAIN_LOOKUP:-}" != true ]] || return 1
                printf '%s\trefs/heads/main\n' "$FIXTURE_MAIN_COMMIT"
                return 0
            fi
            if [[ "${!#}" == refs/tags/v1.2.3-rc.1 || "${!#}" == 'refs/tags/v1.2.3-rc.1^{}' ]]; then
                [[ "${FAILED_REF_LOOKUP:-}" != true ]] || return 1
                if [[ -f prerelease-tag ]]; then
                    printf '%040d\trefs/tags/v1.2.3-rc.1\n' 1
                    if [[ "${!#}" == 'refs/tags/v1.2.3-rc.1^{}' ]]; then
                        printf '%s\trefs/tags/v1.2.3-rc.1^{}\n' "$(cat prerelease-tag)"
                    fi
                fi
                return 0
            fi
            ;;
        cat-file)
            [[ "${!#}" == "$(printf '%040d' 100):CHANGELOG.md" && -f original-commit-fetched ]]
            return
            ;;
        fetch)
            [[ "${FAILED_SOURCE_FETCH:-}" != true && "${!#}" == refs/tags/v1.2.3 ]] || return 1
            touch original-commit-fetched
            return
            ;;
        show)
            [[ "${!#}" == "$(printf '%040d' 100):CHANGELOG.md" && -f original-commit-fetched ]] || return 1
            printf '## v1.2.3\nOriginal signed changelog\n'
            return
            ;;
    esac
    if [[ -f newer-tag && "${!#}" == 'refs/tags/v*' ]]; then
        printf 'newer-ref\trefs/tags/v1.2.4\n'
    fi
    if [[ -f github-tag && ( "${!#}" == 'refs/tags/v*' || "${!#}" == refs/tags/v1.2.3 ) ]]; then
        printf '%040d\trefs/tags/v1.2.3\n' 1
    fi
}

docker() {
    case "$1 $2 $3" in
        'buildx imagetools create')
            local tag="$5" source="$6"
            [[ "$tag" != "${FAILED_TARGET:-}" ]] || return 1
            printf '%s %s\n' "$tag" "${source##*@}" >> registry
            printf 'publish %s\n' "$tag" >> actions
            ;;
        'buildx imagetools inspect')
            local actual
            actual="$(awk -v ref="$6" '$1 == ref { digest = $2 } END { print digest }' registry)"
            [[ -n "$actual" ]] || return 1
            printf '{"digest":"%s"}\n' "$actual"
            ;;
        'login --username '*) cat > /dev/null ;;
        *) return 1 ;;
    esac
}

sleep() { :; }

fixture_github_release() {
    jq -cn --argjson draft "$1" --arg tag "${2:-v1.2.3}" '{id:1,tag_name:$tag,prerelease:false,draft:$draft}'
}

fixture_github_assets() {
    printf '%s\n' "$1" > github-assets.json
}

curl() {
    local url="${!#}" payload="" file="" request="" name checksum size
    local -a args=("$@")
    while (( $# > 1 )); do
        case "$1" in
            --data) payload="$2"; shift ;;
            --data-binary) file="${2#@}"; shift ;;
            --request) request="$2"; shift ;;
        esac
        shift
    done
    set -- "${args[@]}"
    case "$url" in
        */projects/200/deployments)
            [[ "$*" == *"JOB-TOKEN: $CI_JOB_TOKEN"* && "$*" == *"environment=$CI_ENVIRONMENT_NAME"* ]] || return 1
            case "${FIXTURE_DEPLOYMENT_HISTORY:-valid}" in
                unavailable) return 1 ;;
                malformed) printf 'null\n'; return ;;
                empty) printf '[]\n'; return ;;
                paginated)
                    if [[ "$*" == *" page=1 "* ]]; then
                        jq -cn --arg env "$CI_ENVIRONMENT_NAME" --arg sha "$CI_COMMIT_SHA" '
                            [range(1; 101) | {environment:{name:$env},sha:$sha,status:"created",deployable:{id:.,pipeline:{id:.}}}]
                        '
                        return
                    fi
                    [[ "$*" == *" page=2 "* ]] || return 1
                    ;;
            esac
            jq -cn --arg env "$CI_ENVIRONMENT_NAME" --arg sha "$CI_COMMIT_SHA" \
                --argjson job "$CI_JOB_ID" --argjson pipeline "$CI_PIPELINE_ID" \
                --argjson newer "${FIXTURE_NEWER_PIPELINE:-0}" --arg status "${FIXTURE_NEWER_STATUS:-success}" \
                --arg start "${FIXTURE_NEWER_START:-started}" '
                def deployment($job; $pipeline; $status; $start):
                    {environment:{name:$env},sha:$sha,status:$status,deployable:{id:$job,pipeline:{id:$pipeline}}}
                    | if $start == "missing" then .
                      elif $start == "never" then .deployable.started_at = null
                      else .deployable.started_at = "2026-11-23T16:00:00Z" end;
                [deployment($job; $pipeline; "running"; "started")]
                + (if $newer > 0 then [deployment(($job + 10); $newer; $status; $start)] else [] end)
            '
            ;;
        */git/tags/0000000000000000000000000000000000000001)
            jq -cn --arg commit "$(printf '%040d' 100)" --arg snapshot "$(cat github-tag)" \
                --argjson verified "${RELEASE_TAG_VERIFIED:-true}" \
                --arg signer "${RELEASE_TAG_SIGNER:-ssg-srv-gh-o11y-gdi@splunk.com}" \
                '{tag:"v1.2.3",object:{type:"commit",sha:$commit},tagger:{email:$signer},
                  verification:{verified:$verified},message:("Stable tag\n\nCollector-Release-Snapshot: " + $snapshot + "\n")}'
            ;;
        */jobs/*/artifacts\?job_token=fixture)
            local output job
            job="${url%/artifacts*}"
            job="${job##*/}"
            while [[ "$1" != --output ]]; do shift; done
            output="$2"
            cp "archives/$job.zip" "$output" || return
            echo "$job" >> artifact-downloads
            ;;
        */releases/latest)
            if [[ "${FAILED_API:-}" == true ]]; then
                printf '{}\n503'
            elif [[ -f latest-release ]]; then
                fixture_github_release false "$(cat latest-release)"
                printf '200'
            else
                printf '{}\n404'
            fi
            ;;
        */releases/tags/v1.2.3)
            if [[ -n "${RELEASE_STATUS:-}" ]]; then
                printf '{}\n%s' "$RELEASE_STATUS"
            elif [[ -f github-release ]]; then
                fixture_github_release false
                printf '200'
            else
                printf '{}\n404'
            fi
            ;;
        */releases\?per_page=100\&page=1)
            if [[ -f github-release ]]; then
                fixture_github_release false | jq -c '[.]'
            elif [[ -f github-draft ]]; then
                fixture_github_release true | jq -c '[.]'
            else
                printf '[]\n'
            fi
            ;;
        */releases)
            [[ "$request" == POST && -f github-tag && ! -e github-release && ! -e github-draft ]] || return 1
            jq -e '.tag_name == "v1.2.3" and .draft == true and .prerelease == false' <<< "$payload" > /dev/null || return
            jq -r '.body' <<< "$payload" > uploaded-release-notes
            fixture_github_assets '[]'
            touch github-draft
            echo 'create GitHub draft' >> actions
            fixture_github_release true
            ;;
        */releases/1/assets\?per_page=100\&page=1)
            [[ -f github-release || -f github-draft ]] || return 1
            if [[ "${INVALID_ASSET_INVENTORY:-}" == true ]]; then
                printf 'null\n'
            else
                cat github-assets.json
            fi
            if [[ "${NEWER_GITHUB_RELEASE_DURING_VERIFICATION:-}" == true ]]; then
                echo v1.2.4 > latest-release
            fi
            ;;
        */releases/assets/99)
            [[ "$request" == DELETE ]] || return 1
            jq -e 'any(.[]; .id == 99 and .state == "starter" and .size == 0)' github-assets.json >/dev/null || return
            fixture_github_assets "$(jq -c 'map(select(.id != 99))' github-assets.json)"
            echo 'delete starter' >> actions
            ;;
        https://uploads.github.com/repos/signalfx/splunk-otel-collector/releases/1/assets\?name=*)
            [[ -f github-release || -f github-draft ]] || return 1
            name="${url##*=}"
            [[ "$file" == "dist/assets/$name" ]] || return 1
            if [[ "${FAILED_UPLOAD:-}" == true && "$name" == collector.tgz ]]; then
                return 1
            fi
            jq -e --arg name "$name" 'all(.[]; .name != $name)' github-assets.json > /dev/null || return
            checksum="$(sha256sum -- "$file")"
            size="$(wc -c < "$file")"
            if [[ "${UNCONFIRMED_UPLOAD:-}" != true ]]; then
                fixture_github_assets "$(jq -c --arg name "$name" --arg digest "sha256:${checksum%% *}" \
                    --argjson size "${size//[[:space:]]/}" \
                    '. + [{id:(length+1),name:$name,digest:$digest,size:$size,state:"uploaded"}]' github-assets.json)"
            fi
            if [[ "$name" == collector.tgz ]]; then
                cp "$file" uploaded-asset
            fi
            echo "upload $name" >> actions
            ;;
        */releases/1)
            [[ "$request" == PATCH && -f github-draft && ! -e github-release ]] || return 1
            jq -e '.draft == false and (.make_latest == "true" or .make_latest == "false")' <<< "$payload" > /dev/null || return
            if [[ "$(jq -r '.make_latest' <<< "$payload")" == true ]]; then
                echo 'publish GitHub draft as latest' >> actions
                echo v1.2.3 > latest-release
            else
                echo 'publish GitHub draft without changing latest' >> actions
            fi
            touch github-release
            rm github-draft
            fixture_github_release false
            ;;
        *) return 1 ;;
    esac
}
export -f git docker sleep curl fixture_github_release fixture_github_assets

cat > .gitlab/create-github-tag.sh << 'EOF'
#!/bin/bash
if [[ $# == 2 ]]; then
    printf '%s\n' "$2" > prerelease-tag
else
    printf '%s\n' "$3" > github-tag
fi
echo tag >> actions
EOF
chmod +x .gitlab/create-github-tag.sh

mkdir -p dist/assets dist/github-release-tools dist/release-inputs archives
printf '#!/bin/bash\nexit 0\n' > dist/github-release-tools/multimod
chmod +x dist/github-release-tools/multimod

stage=1.2.3-staging-100
write_image_records "$stage"
echo 'original signed package' > dist/assets/collector.tgz
(cd dist/assets && sha256sum collector.tgz > checksums.txt)
echo 11 > dist/release-inputs/assets
echo 12 > dist/release-inputs/docker
echo 13 > dist/release-inputs/tools
zip -qr archives/11.zip dist/assets dist/release-inputs/assets
zip -qr archives/12.zip dist/docker dist/release-inputs/docker
zip -qr archives/13.zip dist/github-release-tools dist/release-inputs/tools
: > registry
: > actions
release=(bash .gitlab/publish-stable-collector-release.sh v1.2.3)

# The manual recovery job cannot start a release whose tag does not exist.
if "${release[@]}" "$stage" --recover-only; then
    echo 'Recovery unexpectedly accepted an untagged release' >&2
    exit 1
fi
if "${release[@]}" '' --recover-from-tag; then
    echo 'Fresh recovery unexpectedly accepted an untagged release' >&2
    exit 1
fi
[[ ! -s actions ]]

# An incomplete or conflicting platform record must stop publication before any writes.
records="$hub_dir/images_windows_multiarch-fips.txt"
for invalid_records in missing duplicate wrong_repository; do
    cp "$records" original-records
    case "$invalid_records" in
        missing) sed '/-2022 /d' original-records > "$records" ;;
        duplicate) head -n 1 original-records >> "$records" ;;
        wrong_repository) sed 's/-fips-windows:/-windows:/' original-records > "$records" ;;
    esac
    if "${release[@]}" "$stage"; then
        echo "Stable release unexpectedly accepted $invalid_records platform records" >&2
        exit 1
    fi
    [[ ! -s actions && ! -e github-tag ]]
    mv original-records "$records"
done

# Matching release-note files must also agree with the manifest being promoted.
printf '[%s@%s]\n' "$QUAY_COLLECTOR_REPO" "$other_digest" > "$quay_dir/multiarch_digest.txt"
printf '[%s@%s]\n' "$DOCKER_HUB_COLLECTOR_REPO" "$other_digest" > "$hub_dir/multiarch_digest.txt"
if "${release[@]}" "$stage"; then
    echo 'Stable release unexpectedly accepted stale release-note digests' >&2
    exit 1
fi
[[ ! -s actions && ! -e github-tag ]]
printf '[%s@%s]\n' "$QUAY_COLLECTOR_REPO" "$digest" > "$quay_dir/multiarch_digest.txt"
printf '[%s@%s]\n' "$DOCKER_HUB_COLLECTOR_REPO" "$digest" > "$hub_dir/multiarch_digest.txt"

# A partial registry failure leaves the version untagged on GitHub.
if FAILED_TARGET="$DOCKER_HUB_COLLECTOR_REPO:1.2.3" "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded with a failing registry' >&2
    exit 1
fi
[[ ! -e github-tag ]]
if grep -Eq '^(tag|create GitHub draft)$' actions; then
    echo 'GitHub was updated before both registries accepted the images' >&2
    exit 1
fi

# Retrying completes all image copies and creates the GitHub tag before upload.
: > actions
if FAILED_UPLOAD=true "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded after the upload failed' >&2
    exit 1
fi
[[ "$(grep -c '^publish ' actions)" == 34 ]]
[[ "$(tail -n 3 actions)" == $'tag\ncreate GitHub draft\nupload checksums.txt' ]]
[[ -f github-tag && -f github-draft && ! -f github-release ]]
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-fips-windows:1.2.3-2019" actions

# Recovery independently checks GitHub verification and the release bot identity.
: > actions
if RELEASE_TAG_VERIFIED=false "${release[@]}" "$stage" --recover-only; then
    echo 'Recovery unexpectedly accepted an unsigned tag' >&2
    exit 1
fi
if RELEASE_TAG_SIGNER=other@example.invalid "${release[@]}" "$stage" --recover-only; then
    echo 'Recovery unexpectedly accepted another signer' >&2
    exit 1
fi
if RELEASE_STATUS=503 "${release[@]}" "$stage" --recover-only; then
    echo 'Recovery unexpectedly succeeded when the release status check failed' >&2
    exit 1
fi
[[ ! -s actions && ! -e artifact-downloads ]]

# Retrying an incomplete draft uploads only the missing signed asset and publishes it.
: > actions
touch newer-tag
echo 'replacement signed package' > dist/assets/collector.tgz
(cd dist/assets && sha256sum collector.tgz > checksums.txt)
echo 21 > dist/release-inputs/assets
"${release[@]}" "$stage"
[[ "$(cat actions)" == $'upload collector.tgz\npublish GitHub draft as latest' ]]
[[ "$(cat uploaded-asset)" == 'original signed package' && "$(cat artifact-downloads)" == 11 ]]

# A standalone promotion or retry must recheck assets after the publication job succeeds.
complete_assets="$(cat github-assets.json)"
fixture_github_assets "$(jq -c 'map(select(.name != "collector.tgz"))' <<< "$complete_assets")"
: > actions
if bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"; then
    echo 'Standalone promotion unexpectedly accepted an incomplete GitHub release' >&2
    exit 1
fi
[[ ! -s actions ]]

# An existing public release can be incomplete even when the tag lookup returns HTTP 200.
if UNCONFIRMED_UPLOAD=true "${release[@]}" "$stage" --recover-only &&
        bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"; then
    echo 'Recovery unexpectedly promoted latest after an unconfirmed asset upload' >&2
    exit 1
fi
[[ "$(cat actions)" == 'upload collector.tgz' ]]
: > actions
"${release[@]}" "$stage" --recover-only
[[ "$(cat actions)" == 'upload collector.tgz' && "$(cat github-assets.json)" == "$complete_assets" ]]

# A failed upload can leave an empty placeholder. Recovery deletes only that placeholder.
fixture_github_assets "$(jq -c 'map(if .name == "collector.tgz" then {id:99,name:.name,state:"starter",size:0} else . end)' \
    <<< "$complete_assets")"
: > actions
"${release[@]}" "$stage"
[[ "$(cat actions)" == $'delete starter\nupload collector.tgz' && "$(cat github-assets.json)" == "$complete_assets" ]]

# A different checksum or unverified inventory must leave the published release untouched.
for inventory in checksum unexpected malformed; do
    invalid_inventory=false
    case "$inventory" in
        checksum) fixture_github_assets "$(jq -c --arg digest "$other_digest" '.[0].digest = $digest' <<< "$complete_assets")" ;;
        unexpected) fixture_github_assets "$(jq -c '. + [{id:99,name:"other.tgz"}]' <<< "$complete_assets")" ;;
        malformed) invalid_inventory=true ;;
    esac
    : > actions
    if INVALID_ASSET_INVENTORY="$invalid_inventory" "${release[@]}" "$stage" --recover-only; then
        echo "Recovery unexpectedly accepted $inventory GitHub assets" >&2
        exit 1
    fi
    if INVALID_ASSET_INVENTORY="$invalid_inventory" bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"; then
        echo "Standalone promotion unexpectedly accepted $inventory GitHub assets" >&2
        exit 1
    fi
    [[ ! -s actions ]]
    fixture_github_assets "$complete_assets"
done

# A newer tag alone does not prevent promotion, which restores the original Docker records.
echo 22 > dist/release-inputs/docker
printf '%s %s\n' "$QUAY_COLLECTOR_REPO:$stage" "$other_digest" > "$hub_dir/images_multiarch.txt"
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ "$(tail -n 1 artifact-downloads)" == 12 ]]
grep -Fx "publish $QUAY_COLLECTOR_REPO:latest-amd64" actions
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-fips-windows:latest-2019" actions

# If a newer release appears during verification, neither registry's latest tag moves.
: > actions
NEWER_GITHUB_RELEASE_DURING_VERIFICATION=true bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ ! -s actions && "$(cat latest-release)" == v1.2.4 ]]

# Manual recovery starts without any artifacts and fetches the pinned archives.
rm -r dist
rm github-release
echo v1.2.2 > latest-release
: > actions
: > artifact-downloads
"${release[@]}" "$stage" --recover-only
[[ "$(cat actions)" == $'create GitHub draft\nupload checksums.txt\nupload collector.tgz\npublish GitHub draft as latest' ]]
[[ "$(cat uploaded-asset)" == 'original signed package' && "$(cat artifact-downloads)" == $'11\n12' ]]
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO:latest-amd64" actions

# A complete published release is verified from the pinned assets and Docker records without writes.
rm -r dist
: > actions
: > artifact-downloads
"${release[@]}" "$stage" --recover-only
[[ ! -s actions && "$(cat artifact-downloads)" == $'11\n12' ]]
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ "$(cat artifact-downloads)" == $'11\n12' ]]
grep -Fx "publish $QUAY_COLLECTOR_REPO:latest-amd64" actions

# A new pipeline with a different checkout restores the archived release context.
rm -r dist
rm github-release
echo v1.2.2 > latest-release
: > actions
: > artifact-downloads
fresh=(env CI_PIPELINE_ID=99 "CI_COMMIT_SHA=$(printf '%040d' 999)"
    QUAY_COLLECTOR_REPO=quay.io/fixture/current DOCKER_HUB_COLLECTOR_REPO=docker.io/fixture/current)
if CI_PROJECT_ID=999 "${fresh[@]}" "${release[@]}" '' --recover-from-tag; then
    echo 'Fresh recovery unexpectedly accepted a tag from another GitLab project' >&2
    exit 1
fi
if RELEASE_TAG_VERIFIED=false "${fresh[@]}" "${release[@]}" '' --recover-from-tag; then
    echo 'Fresh recovery unexpectedly accepted an unsigned tag' >&2
    exit 1
fi
[[ ! -s actions && ! -s artifact-downloads ]]

if FAILED_SOURCE_FETCH=true "${fresh[@]}" "${release[@]}" '' --recover-from-tag; then
    echo 'Fresh recovery unexpectedly uploaded without the signed source changelog' >&2
    exit 1
fi
[[ ! -s actions ]]
"${fresh[@]}" "${release[@]}" '' --recover-from-tag
[[ "$(cat actions)" == $'create GitHub draft\nupload checksums.txt\nupload collector.tgz\npublish GitHub draft as latest' ]]
[[ "$(cat uploaded-asset)" == 'original signed package' && "$(cat artifact-downloads)" == $'11\n12' ]]
grep -Fx 'Original signed changelog' uploaded-release-notes

rm -r dist
fixture_github_assets "$(jq -c 'map(select(.name != "collector.tgz"))' <<< "$complete_assets")"
: > actions
: > artifact-downloads
if "${fresh[@]}" bash .gitlab/promote-stable-collector-latest.sh v1.2.3 '' --recover-from-tag; then
    echo 'Fresh promotion unexpectedly accepted an incomplete GitHub release' >&2
    exit 1
fi
[[ ! -s actions && "$(cat artifact-downloads)" == $'12\n11' ]]
rm -r dist
: > artifact-downloads
"${fresh[@]}" "${release[@]}" '' --recover-from-tag
[[ "$(cat actions)" == 'upload collector.tgz' && "$(cat artifact-downloads)" == $'11\n12' ]]
[[ "$(cat github-assets.json)" == "$complete_assets" ]]
"${fresh[@]}" bash .gitlab/promote-stable-collector-latest.sh v1.2.3 '' --recover-from-tag
[[ "$(cat artifact-downloads)" == $'11\n12' ]]
grep -Fx "publish $QUAY_COLLECTOR_REPO:latest-amd64" actions

# Once the newer release is public, the older upload can finish with latest unchanged.
: > actions
rm github-release
echo v1.2.4 > latest-release
"${release[@]}" "$stage"
[[ "$(cat actions)" == $'create GitHub draft\nupload checksums.txt\nupload collector.tgz\npublish GitHub draft without changing latest' ]]
[[ "$(cat latest-release)" == v1.2.4 ]]

: > actions
"${release[@]}" "$stage"
[[ ! -s actions ]]

# The older alias job succeeds without rolling back the newer version.
: > actions
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ ! -s actions ]]

# Another child cannot take over an already-tagged release.
: > actions
if CI_PIPELINE_ID=99 "${release[@]}" 1.2.3-staging-99; then
    echo 'Stale release unexpectedly succeeded' >&2
    exit 1
fi
[[ ! -s actions ]]

# An archive with the expected job ID but changed bytes fails closed.
echo 'unexpected signed package' > dist/assets/collector.tgz
if "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded after its signed assets changed' >&2
    exit 1
fi
[[ ! -s actions ]]
echo 'original signed package' > dist/assets/collector.tgz

# The pinned inputs must still agree with the public version tags.
printf '%s %s\n' "$DOCKER_HUB_COLLECTOR_REPO:1.2.3" "$other_digest" >> registry
if "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded with a changed public image tag' >&2
    exit 1
fi
[[ ! -s actions ]]
printf '%s %s\n' "$DOCKER_HUB_COLLECTOR_REPO:1.2.3" "$digest" >> registry

# The latest lookup must succeed before any public writes.
if FAILED_API=true "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded with a failed latest lookup' >&2
    exit 1
fi
[[ ! -s actions ]]

# An older version with no GitHub tag cannot be published for the first time.
rm github-tag
if "${release[@]}" "$stage"; then
    echo 'A new release older than an existing tag unexpectedly succeeded' >&2
    exit 1
fi
[[ ! -s actions ]]

mkdir -p tag-check/.gitlab
cp "$repo/.gitlab/create-github-tag.sh" tag-check/.gitlab/
cat > tag-check/.gitlab/common.sh << 'EOF'
setup_gpg() { exit 1; }
setup_git() { exit 1; }
git() {
    case "$1" in
        clone) mkdir -p repo-tmp ;;
        fetch | show-ref) : ;;
        rev-parse) printf '%s\n' "$CI_COMMIT_SHA" ;;
        *) return 1 ;;
    esac
}
EOF
(unset GITHUB_BOT_GPG_KEY; bash tag-check/.gitlab/create-github-tag.sh v1.2.3 "$CI_COMMIT_SHA")
if (unset GITHUB_BOT_GPG_KEY; bash tag-check/.gitlab/create-github-tag.sh v1.2.3 "$CI_COMMIT_SHA" '{}'); then
    echo 'Initial stable publication unexpectedly adopted a pre-existing tag' >&2
    exit 1
fi

# New prereleases promote the signed platform tags before creating the GitHub tag.
export GITHUB_BOT_GPG_KEY=fixture
unzip -oq archives/13.zip dist/github-release-tools/multimod
prerelease=(bash .gitlab/publish-prerelease-collector-release.sh v1.2.3-rc.1 1.2.3-rc.1-staging-100)
write_image_records 1.2.3-rc.1-staging-100
: > actions
if FAILED_REF_LOOKUP=true "${prerelease[@]}"; then
    echo 'Prerelease unexpectedly passed a failed GitHub tag lookup' >&2
    exit 1
fi
[[ ! -s actions && ! -e prerelease-tag ]]
if GITHUB_BOT_GPG_KEY='' "${prerelease[@]}"; then
    echo 'Prerelease unexpectedly published without the GitHub tag signing key' >&2
    exit 1
fi
[[ ! -s actions && ! -e prerelease-tag ]]
if FAILED_TARGET="$DOCKER_HUB_COLLECTOR_REPO:1.2.3-rc.1-amd64" "${prerelease[@]}"; then
    echo 'Prerelease unexpectedly passed a failed platform promotion' >&2
    exit 1
fi
[[ ! -e prerelease-tag ]]
: > actions
"${prerelease[@]}"
grep -Fx "publish $QUAY_COLLECTOR_REPO:1.2.3-rc.1-amd64" actions
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-fips-windows:1.2.3-rc.1-2019" actions
[[ "$(tail -n 1 actions)" == tag && "$(cat prerelease-tag)" == "$CI_COMMIT_SHA" ]]

# A fresh pipeline rejects an existing prerelease in preflight.
preflight="$(awk '
    /^manual_release_trigger:/ { found_job=1 }
    found_job && /^    - \|$/ { print_block=1; next }
    print_block && /^  rules:/ { exit }
    print_block
' "$repo/.gitlab/collector-pipeline.yml")"
printf 'module-sets:\n  splunk-otel-collector:\n    version: v1.2.3-rc.1\n' > versions.yaml
if RELEASE_VERSION=v1.2.3-rc.1 CI_PIPELINE_SOURCE=web bash -euc "$preflight" > preflight.log 2>&1; then
    echo 'Prerelease preflight unexpectedly accepted an existing tag' >&2
    exit 1
fi
grep -Fx 'v1.2.3-rc.1 is already tagged on GitHub.' preflight.log
DRY_RUN=true FAILED_REF_LOOKUP=true RELEASE_VERSION=v1.2.3-rc.1 CI_PIPELINE_SOURCE=web bash -euc "$preflight"

# Publication retries may verify the original tag, but cannot overwrite a tagged version.
: > actions
"${prerelease[@]}"
if CI_COMMIT_SHA="$(printf '%040d' 999)" "${prerelease[@]}"; then
    echo 'Prerelease unexpectedly took over a tag from a different commit' >&2
    exit 1
fi
printf '%s %s\n' "$DOCKER_HUB_COLLECTOR_REPO:1.2.3-rc.1-amd64" "$other_digest" >> registry
if "${prerelease[@]}"; then
    echo 'Prerelease unexpectedly accepted a changed public platform tag' >&2
    exit 1
fi
[[ ! -s actions ]]

# A delayed development pipeline still publishes its SHA alias under the promotion lock.
tag_setup="$(awk '/^\.set-docker-tags:/ { print_block=1; next } print_block && /^[^ ]/ { exit } print_block' "$repo/.gitlab/collector-pipeline.yml")"
dev_promotion="$(awk '
    /^promote-docker-latest:/ { found_job=1; next }
    found_job && /^promote-docker-stable-latest:/ { exit }
    found_job && /^    - \|$/ { print_block=1; next }
    print_block
' "$repo/.gitlab/collector-pipeline.yml")"
dev=(env RELEASE_VERSION='' CI_COMMIT_TAG='' CI_COMMIT_BRANCH=main CI_JOB_ID=400 "FIXTURE_MAIN_COMMIT=$CI_COMMIT_SHA"
    bash -euc "$tag_setup"$'\n'"$dev_promotion")
write_image_records "$CI_COMMIT_SHA-staging-$CI_PIPELINE_ID" -dev
CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" "${dev[@]}"
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-dev:latest" actions

: > actions
older_commit="$(printf '%040d' 99)"
write_image_records "$older_commit-staging-$CI_PIPELINE_ID" -dev
CI_COMMIT_SHA="$older_commit" CI_ENVIRONMENT_NAME="collector-development-images/$older_commit" "${dev[@]}"
grep -Fx "publish $QUAY_COLLECTOR_REPO-dev:$older_commit" actions
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-dev:$older_commit" actions
if grep -Eq ':latest(-|$)' actions; then
    echo 'An older development commit unexpectedly moved latest' >&2
    exit 1
fi

# Greater pipeline IDs for the same commit block writes even if publication was interrupted.
write_image_records "$CI_COMMIT_SHA-staging-$CI_PIPELINE_ID" -dev
for newer_status in success failed canceled; do
    : > actions
    CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" \
        FIXTURE_NEWER_PIPELINE=101 FIXTURE_NEWER_STATUS="$newer_status" "${dev[@]}"
    [[ ! -s actions ]]
done
# A newer deployment on a subsequent page still prevents any alias writes.
CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" \
    FIXTURE_DEPLOYMENT_HISTORY=paginated FIXTURE_NEWER_PIPELINE=101 "${dev[@]}"
[[ ! -s actions ]]
# A newer job that never started cannot have published aliases.
for status in created canceled failed; do
    : > actions
    CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" \
        FIXTURE_NEWER_PIPELINE=101 FIXTURE_NEWER_STATUS="$status" FIXTURE_NEWER_START=never "${dev[@]}"
    grep -Fx "publish $QUAY_COLLECTOR_REPO-dev:$CI_COMMIT_SHA" actions
    grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-dev:latest" actions
done
# An omitted start time cannot establish that a canceled job never ran.
: > actions
CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" FIXTURE_NEWER_PIPELINE=101 \
    FIXTURE_NEWER_STATUS=canceled FIXTURE_NEWER_START=missing "${dev[@]}"
[[ ! -s actions ]]

for history in unavailable malformed empty; do
    : > actions
    if CI_ENVIRONMENT_NAME="collector-development-images/$CI_COMMIT_SHA" FIXTURE_DEPLOYMENT_HISTORY="$history" "${dev[@]}"; then
        echo 'Development promotion unexpectedly accepted unverifiable deployment history' >&2
        exit 1
    fi
    [[ ! -s actions ]]
done

# A lookup failure or malformed branch tip must leave latest alone.
write_image_records "$CI_COMMIT_SHA-staging-$CI_PIPELINE_ID" -dev
for ref_lookup in FAILED_MAIN_LOOKUP=true FIXTURE_MAIN_COMMIT=malformed; do
    : > actions
    if env RELEASE_VERSION='' CI_COMMIT_TAG='' CI_COMMIT_BRANCH=main CI_JOB_ID=400 \
            "CI_ENVIRONMENT_NAME=collector-development-images/$CI_COMMIT_SHA" "FIXTURE_MAIN_COMMIT=$CI_COMMIT_SHA" "$ref_lookup" \
            bash -euc "$tag_setup"$'\n'"$dev_promotion"; then
        echo 'Development promotion unexpectedly accepted an unverifiable main tip' >&2
        exit 1
    fi
    if grep -Eq ':latest(-|$)' actions; then
        echo 'Development latest moved after a failed branch tip check' >&2
        exit 1
    fi
done

echo 'Collector Docker release retry tests passed'
