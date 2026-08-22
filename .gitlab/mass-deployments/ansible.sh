# shellcheck shell=bash
export DEPLOYMENT_NAME="Ansible"
export DEPLOYMENT_CHANGELOG="deployments/ansible_collections/signalfx/splunk_otel_collector/CHANGELOG.md"
export DEPLOYMENT_PREFIX="ansible"

update_version_file() {
  local new_version="$1"
  sed -i "s/^version: .*/version: ${new_version}/" \
    "deployments/ansible_collections/signalfx/splunk_otel_collector/galaxy.yml"
}
