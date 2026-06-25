DEPLOYMENT_NAME="Chef"
DEPLOYMENT_CHANGELOG="deployments/chef/CHANGELOG.md"
DEPLOYMENT_PREFIX="chef"

update_version_file() {
  local new_version="$1"
  sed -i "s/^version '.*'/version '${new_version}'/" "deployments/chef/metadata.rb"
}
