#!/bin/bash -eux
set -o pipefail

# customize TA
BUILD_DIR="$(realpath "$BUILD_DIR")"

# Set ta-agent-config.yml with mysql config
TEMP_DIR="$BUILD_DIR/ci-cd/collectd"
mkdir -p "$TEMP_DIR"
tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR" 
cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
echo "$OLLY_ACCESS_TOKEN" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"
sed -i '/splunk_config/d' "$TEMP_DIR/Splunk_TA_otel/local/inputs.conf"
mkdir "$TEMP_DIR/Splunk_TA_otel/configs/mysql"
cp "$SOURCE_DIR/packaging-scripts/cicd-tests/mysql/ta-agent-config.yaml" "$TEMP_DIR/Splunk_TA_otel/configs/mysql"
echo "splunk_config=\$SPLUNK_OTEL_TA_HOME/configs/mysql/ta-agent-config.yaml" >> "$TEMP_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$TEMP_DIR" -hcz -f "$TEMP_DIR/Splunk_TA_otel.tgz" "Splunk_TA_otel"

# Create ORCA container & grab id
splunk_orca -vvv --cloud kubernetes --printer sdd-json --deployment-file "$BUILD_DIR/orca_deployment_collectd.json" --ansible-log "$BUILD_DIR/ansible-local.log" create --env SPLUNK_CONNECTION_TIMEOUT=600 --so 1 --splunk-version "${UF_VERSION}" --platform x64_centos_7 --local-apps "$BUILD_DIR/ci-cd/collectd/Splunk_TA_otel.tgz" --cloud-instance-type-mapping '{"universal_forwarder": "c5.large", "heavy_forwarder": "c5.large", "indexer": "c5.large", "eventgen_standalone": "c5.large", "testrunner": "c5.large", "license_master": "c5.large", "deployment_server": "c5.large", "deployer": "c5.large", "eventgen_server": "c5.large", "cluster_master": "c5.large", "search_head": "c5.large", "eventgen_controller": "c5.large", "standalone": "c5.large"}' --playbook "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml" --custom-services "$SOURCE_DIR/packaging-scripts/orca-playbook-mysql.yml"
deployment_id="$(grep "orca_deployment_id" "$BUILD_DIR/orca_deployment_collectd.json" | awk -F ':' '{print $2}' | awk -F '"' '{print $2}')"
echo "deployment_id: $deployment_id"
splunk_orca --cloud kubernetes show containers --deployment-id "${deployment_id}" >> "$BUILD_DIR/container_details_collectd.txt"
sed -i '/custom/d' "$BUILD_DIR/container_details_collectd.txt"
orca_container=$(grep "Container Name" "$BUILD_DIR/container_details_collectd.txt" | awk -F " " '{print $8}')

# Change to the host running mysql in the config yaml
custom_ip="$(jq -r '.server_roles.custom[0].host' "$BUILD_DIR/orca_deployment_collectd.json")"
echo "custom IP: $custom_ip"
sed -i "s/127.0.0.1/$custom_ip/g" "$TEMP_DIR/Splunk_TA_otel/configs/mysql/ta-agent-config.yaml"
splunk_orca --cloud kubernetes copy to "${orca_container}" ${ORCA_OPTION} --source "$TEMP_DIR/Splunk_TA_otel/configs/mysql/ta-agent-config.yaml" --destination /opt/splunk/etc/apps/Splunk_TA_otel/configs/mysql
splunk_orca --cloud kubernetes provision "${deployment_id}" --playbooks "$SOURCE_DIR/packaging-scripts/orca-playbook-linux-restart.yml"
sleep 90s
splunk_orca --cloud kubernetes copy from "${orca_container}" ${ORCA_OPTION} --source /opt/splunk/var/log/splunk/otel.log --destination "$BUILD_DIR/splunklog"  
grep -q "Failed to connect to database mysql" "$BUILD_DIR/splunklog/otel.log" && exit 1
grep -q "Collectd died when it was supposed to be running" "$BUILD_DIR/splunklog/otel.log" && exit 1

# clean up orca container
splunk_orca --cloud kubernetes destroy "${deployment_id}"

mv "$BUILD_DIR/orca_deployment_collectd.json" "$BUILD_DIR/container_details_collectd.txt" "$BUILD_DIR/ansible-local.log" "$TEMP_DIR"
rm "$TEMP_DIR/Splunk_TA_otel.tgz"
exit 0
