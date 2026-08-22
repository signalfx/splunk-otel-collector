# shellcheck shell=bash
export DEPLOYMENT_NAME="Puppet"
export DEPLOYMENT_CHANGELOG="deployments/puppet/CHANGELOG.md"
export DEPLOYMENT_PREFIX="puppet"

update_version_file() {
  local new_version="$1"
  # metadata.json uses standard JSON; target only the top-level "version" field.
  sed -i "s/\"version\": \"[^\"]*\"/\"version\": \"${new_version}\"/" "deployments/puppet/metadata.json"
}
