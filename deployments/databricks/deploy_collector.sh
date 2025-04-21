# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash

set -euo pipefail

# This script is used to deploy the Splunk distribution of the OpenTelemetry Collector
# on the current node of a Databricks cluster. Through UI configuration the script will
# be distributed and run on every node of the cluster.

# Required Variables:
# - SPLUNK_ACCESS_TOKEN: Splunk o11y access token for sending data to Splunk o11y backend.
# - DATABRICKS_CLUSTER_HOSTNAME: Hostname of the Databricks compute resource. Use the Server
#   Hostname described here:
#   https://docs.databricks.com/en/integrations/compute-details.html
# - DATABRICKS_ACCESS_TOKEN: Databricks personal access token (PAT) used to connect to the Apache Spark API.
#   Directions for creating a PAT: https://docs.databricks.com/en/dev-tools/auth/pat.html

# Optional Variables:
# - SPLUNK_OTEL_VERSION: Version of the Splunk OpenTelemetry Collector to deploy as a part of this release.
#   Default: "latest". Valid version must be >=0.119.0.
# - SCRIPT_DIR: Installation path for the Collector and its config
# - SPLUNK_REALM: Splunk o11y realm to send data to. Default: us0

DATABRICKS_CLUSTER_HOSTNAME=${DATABRICKS_CLUSTER_HOSTNAME:-}
SPLUNK_ACCESS_TOKEN=${SPLUNK_ACCESS_TOKEN:-}
DATABRICKS_ACCESS_TOKEN=${DATABRICKS_ACCESS_TOKEN:-}

if [ -z "${DATABRICKS_CLUSTER_HOSTNAME}" ]; then
  echo "environment variable 'DATABRICKS_CLUSTER_HOSTNAME' must be set, exiting."
  exit 1
fi

if [ -z "${SPLUNK_ACCESS_TOKEN}" ]; then
  echo "environment variable 'SPLUNK_ACCESS_TOKEN' must be set, exiting."
  exit 1
fi

if [ -z "${DATABRICKS_ACCESS_TOKEN}" ]; then
  echo "environment variable 'DATABRICKS_ACCESS_TOKEN' must be set, exiting."
  exit 1
fi

SPLUNK_OTEL_VERSION=${SPLUNK_OTEL_VERSION:-latest}
OS="linux_amd64"
SPLUNK_OTEL_BINARY_NAME="splunk_otel_collector"
SPLUNK_OTEL_DOWNLOAD_BASE_URL="https://github.com/signalfx/splunk-otel-collector/releases"
SPLUNK_OTEL_API_URL="https://api.github.com/repos/signalfx/splunk-otel-collector/releases/latest"
SCRIPT_DIR=${SCRIPT_DIR:-/tmp/collector_download}
CONFIG_FILENAME="config.yaml"
SPLUNK_OTEL_BINARY_FILE="$SCRIPT_DIR/$SPLUNK_OTEL_BINARY_NAME"
CONFIG_FILE="$SCRIPT_DIR/$CONFIG_FILENAME"
SERVICE_PATH="/etc/systemd/system/"
SERVICE_FILE="$SERVICE_PATH/$SPLUNK_OTEL_BINARY_NAME.service"

if [ $SPLUNK_OTEL_VERSION = "latest" ]; then
        SPLUNK_OTEL_VERSION=$(curl --silent "$SPLUNK_OTEL_API_URL" |    # Get latest Collector release from GitHub api
            grep '"tag_name":' |                          # Get tag name line
            sed -E 's/.*"([^"]+)".*/\1/')                 # Pluck latest release version
        if [ -z "$SPLUNK_OTEL_VERSION" ]; then
            echo "Failed to get tag_name for latest release from $SPLUNK_OTEL_VERSION/latest" >&2
            exit 1
        fi
fi

SPLUNK_OTEL_BINARY_DOWNLOAD_URL="${SPLUNK_OTEL_DOWNLOAD_BASE_URL}/download/${SPLUNK_OTEL_VERSION}/otelcol_${OS}"
mkdir -p "$SCRIPT_DIR"
INSTALLED_SPLUNK_OTEL_VERSION=""

if [ -f $SPLUNK_OTEL_BINARY_FILE ]; then
  # Output of `otelcol --version` is of the form:
  # otelcol version vX.X.X
  # Capture output of `otelcol --version` as an array, then extract version
  INSTALLED_SPLUNK_OTEL_VERSION=($($SPLUNK_OTEL_BINARY_FILE --version))
  INSTALLED_SPLUNK_OTEL_VERSION=${INSTALLED_SPLUNK_OTEL_VERSION[2]}
fi

if [ ! -f $SPLUNK_OTEL_BINARY_FILE ] || [ $INSTALLED_SPLUNK_OTEL_VERSION != $SPLUNK_OTEL_VERSION ]; then
  # If the binary is already installed it's the wrong version and it needs to be removed before downloading again
  # to avoid "Text file busy" errors
  if [ -f $SPLUNK_OTEL_BINARY_FILE ]; then
    rm "$SPLUNK_OTEL_BINARY_FILE"
  fi

  # Download Splunk's distribution of the OpenTelemetry Collector
  curl -L --output "$SPLUNK_OTEL_BINARY_FILE" $SPLUNK_OTEL_BINARY_DOWNLOAD_URL || { echo "Failed to download $OTEL_BINARY_DOWNLOAD_URL"; exit 1; }
  chmod +x "$SPLUNK_OTEL_BINARY_FILE"
else
  echo "Splunk OpenTelemetry Collector '${SPLUNK_OTEL_VERSION}' is already installed"
fi

# The Spark receiver should only be run in one instance per-Cluster. Run
# it on the driver node, as there's one per-cluster.
# More info on Databricks init script environment variables:
# https://docs.databricks.com/en/init-scripts/environment-variables.html#use-secrets-in-init-scripts
if [ $DB_IS_DRIVER = "TRUE" ]; then
        OPTIONAL_SPARK_RECEIVER=", apachespark"
else
        OPTIONAL_SPARK_RECEIVER=""
fi

collector_config="
extensions:
  bearertokenauth:
    token: $DATABRICKS_ACCESS_TOKEN

receivers:
  apachespark:
    # https://community.databricks.com/t5/data-engineering/how-to-obtain-the-server-url-for-using-spark-s-rest-api/td-p/83410
    endpoint: https://$DATABRICKS_CLUSTER_HOSTNAME/driver-proxy-api/o/0/$DB_CLUSTER_ID/40001
    auth:
      authenticator: bearertokenauth
  # TODO: Identify any additional scrapers that are necessary and useful
  hostmetrics:
    scrapers:
      cpu:
      memory:
      network:

processors:
  batch:
    send_batch_size: 10000
    timeout: 10s
  resourcedetection:
    detectors: [system]
  resource:
    attributes:
      - key: databricks.cluster.name
        value: \"$DB_CLUSTER_NAME\"
        action: upsert
      - key: databricks.cluster.id
        value: \"$DB_CLUSTER_ID\"
        action: upsert
      - key: databricks.node.driver
        value: \"$DB_IS_DRIVER\"
        action: upsert

exporters:
  signalfx:
    access_token: $SPLUNK_ACCESS_TOKEN
    realm: ${SPLUNK_REALM:-us0}

service:
  extensions: [bearertokenauth]
  pipelines:
    metrics:
      receivers: [hostmetrics$OPTIONAL_SPARK_RECEIVER]
      processors: [batch, resourcedetection, resource]
      exporters: [signalfx]
"

echo "$collector_config" > "$CONFIG_FILE"

collector_service="
[Unit]
Description=Splunk distribution of the OpenTelemetry Collector
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
ExecStart=$SPLUNK_OTEL_BINARY_FILE --config $CONFIG_FILE

[Install]
WantedBy=multi-user.target
"

echo "$collector_service" > $SERVICE_FILE
chmod 755 $SERVICE_FILE
systemctl daemon-reload

# The collector is run as a service on the current node
systemctl start $SPLUNK_OTEL_BINARY_NAME

exit 0