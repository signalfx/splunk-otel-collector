#!/bin/bash

set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

mkdir -p .gitlab packaging/release
cp "$repo"/.gitlab/{collector-docker-common,collector-release-snapshot,promote-collector-docker-images,promote-stable-collector-latest,publish-stable-collector-release}.sh .gitlab/
cp "$repo/packaging/release/gh-release-notes.sh" packaging/release/
printf '## v1.2.3\nTest release\n' > CHANGELOG.md

export QUAY_COLLECTOR_REPO=quay.io/splunk/collector DOCKER_HUB_COLLECTOR_REPO=docker.io/splunk/collector
export CIRCLECI_QUAY_PASSWORD=test DOCKER_HUB_TOKEN=test GITHUB_TOKEN=test
export CI_PIPELINE_ID=100 CI_PROJECT_ID=200 CI_JOB_TOKEN=fixture CI_API_V4_URL=https://gitlab.invalid/api/v4
CI_COMMIT_SHA="$(printf '%040d' 100)"
export CI_COMMIT_SHA
unset DRY_RUN
quay_dir="dist/docker/$QUAY_COLLECTOR_REPO"
hub_dir="dist/docker/$DOCKER_HUB_COLLECTOR_REPO"
mkdir -p "$quay_dir" "$hub_dir"
digest="sha256:$(printf '%064d' 1)"
other_digest="sha256:$(printf '%064d' 2)"
printf '[%s@%s]\n' "$QUAY_COLLECTOR_REPO" "$digest" > "$quay_dir/multiarch_digest.txt"
printf '[%s@%s]\n' "$DOCKER_HUB_COLLECTOR_REPO" "$digest" > "$hub_dir/multiarch_digest.txt"

write_image_records() {
    local stage="$1" fips
    for fips in '' -fips; do
        printf '%s %s\n' "$QUAY_COLLECTOR_REPO$fips:$stage" "$digest" \
            "$QUAY_COLLECTOR_REPO$fips:$stage-amd64" "$digest" > "$hub_dir/images_multiarch$fips.txt"
        printf '%s %s\n' "$QUAY_COLLECTOR_REPO$fips-windows:$stage" "$digest" \
            "$QUAY_COLLECTOR_REPO$fips-windows:$stage-2019" "$digest" > "$hub_dir/images_windows_multiarch$fips.txt"
    done
}

git() {
    case "$1" in
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

curl() {
    case "${!#}" in
        */git/tags/0000000000000000000000000000000000000001)
            jq -cn --arg commit "$(printf '%040d' 100)" --arg snapshot "$(cat github-tag)" \
                --argjson verified "${RELEASE_TAG_VERIFIED:-true}" \
                --arg signer "${RELEASE_TAG_SIGNER:-ssg-srv-gh-o11y-gdi@splunk.com}" \
                '{tag:"v1.2.3",object:{type:"commit",sha:$commit},tagger:{email:$signer},
                  verification:{verified:$verified},message:("Stable tag\n\nCollector-Release-Snapshot: " + $snapshot + "\n")}'
            ;;
        */jobs/*/artifacts\?job_token=fixture)
            local output url="${!#}" job
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
                printf '{"tag_name":"%s"}\n200' "$(cat latest-release)"
            else
                printf '{}\n404'
            fi
            ;;
        */releases/tags/v1.2.3)
            if [[ -n "${RELEASE_STATUS:-}" ]]; then
                printf '%s' "$RELEASE_STATUS"
            elif [[ -f github-release ]]; then
                printf '200'
            else
                printf '404'
            fi
            ;;
        */releases\?per_page=100\&page=1)
            if [[ -f github-release ]]; then
                printf '[{"id":1,"tag_name":"v1.2.3","draft":false}]\n'
            elif [[ -f github-draft ]]; then
                printf '[{"id":1,"tag_name":"v1.2.3","draft":true}]\n'
            else
                printf '[]\n'
            fi
            ;;
        */releases/1)
            [[ "$*" == *'--request PATCH'* && "$*" == *'{"draft":false,"make_latest":"false"}'* ]] || return 1
            echo 'publish GitHub draft without changing latest' >> actions
            touch github-release
            ;;
        *) return 1 ;;
    esac
}
export -f git docker sleep curl

cat > .gitlab/create-github-tag.sh << 'EOF'
#!/bin/bash
printf '%s\n' "$3" > github-tag
echo tag >> actions
EOF
chmod +x .gitlab/create-github-tag.sh

mkdir -p dist/assets dist/github-release-tools dist/release-inputs archives
cat > dist/github-release-tools/ghr << 'EOF'
#!/bin/bash
set -euo pipefail
cp dist/assets/collector.tgz uploaded-asset
args=("$@")
while [[ "$1" != -b ]]; do shift; done
printf '%s\n' "$2" > uploaded-release-notes
if [[ "${FAILED_UPLOAD:-}" == true ]]; then
    exit 1
fi
if [[ -f github-release ]]; then
    echo 'ghr already published' >> actions
elif [[ "${args[0]}" == --draft ]]; then
    echo 'ghr draft' >> actions
    touch github-draft
else
    echo ghr >> actions
    touch github-release
    echo v1.2.3 > latest-release
fi
EOF
printf '#!/bin/bash\nexit 0\n' > dist/github-release-tools/multimod
chmod +x dist/github-release-tools/{ghr,multimod}

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

# A partial registry failure leaves the version untagged on GitHub.
if FAILED_TARGET="$DOCKER_HUB_COLLECTOR_REPO:1.2.3" "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded with a failing registry' >&2
    exit 1
fi
[[ ! -e github-tag ]]
if grep -Eq '^(tag|ghr)$' actions; then
    echo 'GitHub was updated before both registries accepted the images' >&2
    exit 1
fi

# Retrying completes all image copies and creates the GitHub tag before upload.
: > actions
if FAILED_UPLOAD=true "${release[@]}" "$stage"; then
    echo 'Release unexpectedly succeeded after the upload failed' >&2
    exit 1
fi
[[ "$(grep -c '^publish ' actions)" == 16 ]]
[[ "$(tail -n 1 actions)" == tag && -f github-tag && ! -f github-release ]]
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

# Retries restore the original assets and tools after replacement attempts.
: > actions
touch newer-tag
echo 'replacement signed package' > dist/assets/collector.tgz
(cd dist/assets && sha256sum collector.tgz > checksums.txt)
echo 21 > dist/release-inputs/assets
echo 23 > dist/release-inputs/tools
echo 'exit 1' > dist/github-release-tools/ghr
"${release[@]}" "$stage"
[[ "$(cat actions)" == ghr && "$(cat uploaded-asset)" == 'original signed package' ]]
[[ "$(cat artifact-downloads)" == $'11\n13' ]]

# A newer tag alone does not prevent promotion, which restores the original Docker records.
echo 22 > dist/release-inputs/docker
printf '%s %s\n' "$QUAY_COLLECTOR_REPO:$stage" "$other_digest" > "$hub_dir/images_multiarch.txt"
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ "$(tail -n 1 artifact-downloads)" == 12 ]]
grep -Fx "publish $QUAY_COLLECTOR_REPO:latest-amd64" actions
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO-fips-windows:latest-2019" actions

# Manual recovery starts without any artifacts and fetches the pinned archives.
rm -r dist
rm github-release
echo v1.2.2 > latest-release
: > actions
: > artifact-downloads
"${release[@]}" "$stage" --recover-only
[[ "$(cat actions)" == ghr && "$(cat uploaded-asset)" == 'original signed package' ]]
[[ "$(cat artifact-downloads)" == $'11\n12\n13' ]]
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
grep -Fx "publish $DOCKER_HUB_COLLECTOR_REPO:latest-amd64" actions

# For a published release, manual recovery only needs the pinned Docker records.
rm -r dist
: > actions
: > artifact-downloads
"${release[@]}" "$stage" --recover-only
[[ ! -s actions && ! -s artifact-downloads ]]
bash .gitlab/promote-stable-collector-latest.sh v1.2.3 "$stage"
[[ "$(cat artifact-downloads)" == 12 ]]
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
[[ "$(cat actions)" == ghr && "$(cat uploaded-asset)" == 'original signed package' ]]
[[ "$(cat artifact-downloads)" == $'11\n12\n13' ]]
grep -Fx 'Original signed changelog' uploaded-release-notes

rm -r dist
: > actions
: > artifact-downloads
"${fresh[@]}" "${release[@]}" '' --recover-from-tag
[[ ! -s actions && ! -s artifact-downloads ]]
"${fresh[@]}" bash .gitlab/promote-stable-collector-latest.sh v1.2.3 '' --recover-from-tag
[[ "$(cat artifact-downloads)" == 12 ]]
grep -Fx "publish $QUAY_COLLECTOR_REPO:latest-amd64" actions

# Once the newer release is public, the older upload can finish with latest unchanged.
: > actions
rm github-release
echo v1.2.4 > latest-release
"${release[@]}" "$stage"
[[ "$(cat actions)" == $'ghr draft\npublish GitHub draft without changing latest' ]]
[[ "$(cat latest-release)" == v1.2.4 ]]

: > actions
"${release[@]}" "$stage"
[[ "$(cat actions)" == 'ghr already published' ]]

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

echo 'Collector Docker release retry tests passed'
