#!/bin/bash -eux
set -o pipefail

# customize TA
BUILD_DIR="$(realpath "$BUILD_DIR")"

# Set gateway
TEMP_DIR="$BUILD_DIR/ci-cd/gateway"
mkdir -p "$TEMP_DIR"
tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR" 
cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
echo "$OLLY_ACCESS_TOKEN" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"
echo "splunk_config=\$SPLUNK_OTEL_TA_HOME/configs/ta-gateway-config.yaml" >> "$TEMP_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$TEMP_DIR" -hcz -f "$TEMP_DIR/Splunk_TA_otel.tgz" "Splunk_TA_otel"

# Set gateway
TEMP_DIR="$BUILD_DIR/ci-cd/gateway-agent"
mkdir -p "$TEMP_DIR"
tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR" 
cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
echo "$OLLY_ACCESS_TOKEN" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"
echo "splunk_config=\$SPLUNK_OTEL_TA_HOME/configs/ta-agent-to-gateway-config.yaml" >> "$TEMP_DIR/Splunk_TA_otel/local/inputs.conf"
echo "gateway_url=localhost" >> "$TEMP_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$TEMP_DIR" -hcz -f "$TEMP_DIR/Splunk_TA_otel_agent_to_gateway.tgz" "Splunk_TA_otel"

splunk_orca -v --cloud kubernetes --ansible-log "$BUILD_DIR/ansible-local-gateway.log" create --env SPLUNK_CONNECTION_TIMEOUT=600 --platform x64_centos_7 --local-apps "$BUILD_DIR/ci-cd/gateway/Splunk_TA_otel.tgz" --local-apps "$BUILD_DIR/ci-cd/gateway-agent/Splunk_TA_otel_agent_to_gateway.tgz" --playbook "$SOURCE_DIR/packaging-scripts/orca-playbook-linux.yml,site.yml"
rm "$BUILD_DIR/ci-cd/gateway/Splunk_TA_otel.tgz"
rm "$TEMP_DIR/Splunk_TA_otel_agent_to_gateway.tgz"
exit 0
