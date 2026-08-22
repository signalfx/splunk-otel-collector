# shellcheck shell=bash
export DEPLOYMENT_NAME="Chef"
export DEPLOYMENT_CHANGELOG="deployments/chef/CHANGELOG.md"
export DEPLOYMENT_PREFIX="chef"

update_version_file() {
  local new_version="$1"
  sed -i "s/^version '.*'/version '${new_version}'/" "deployments/chef/metadata.rb"
}
