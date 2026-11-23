#!/bin/bash

set -euo pipefail

source .gitlab/collector-docker-common.sh

version="$1"
staging_tag="$2"
if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-(beta|rc)\.[0-9]+$ ]]; then
    echo "Invalid collector prerelease version $version" >&2
    exit 1
fi
: "${CI_COMMIT_SHA:?}"

tagged_commit="$(collector_release_tag_commit "$version")"
if [[ -n "$tagged_commit" ]]; then
    if [[ "$tagged_commit" != "$CI_COMMIT_SHA" ]]; then
        echo "$version is already tagged at $tagged_commit instead of $CI_COMMIT_SHA" >&2
        exit 1
    fi
    # A retry can verify an existing publication without rewriting its tags.
    bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}" verify
    exit 0
fi

tools_dir="$PWD/dist/github-release-tools"
: "${GITHUB_TOKEN:?}" "${GITHUB_BOT_GPG_KEY:?}"
test -x "$tools_dir/multimod"
export PATH="$tools_dir:$PATH"
bash .gitlab/promote-collector-docker-images.sh "$staging_tag" "${version#v}"
./.gitlab/create-github-tag.sh "$version" "$CI_COMMIT_SHA"
