#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_ZOOKEEPER_CONNECT" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_ZOOKEEPER_CONNECT"
    exit 1
fi
"$KAFKA_BIN"/kafka-server-start.sh /opt/kafka_2.11-"$KAFKA_VERSION"/config/server.properties \
    --override zookeeper.connect="$KAFKA_ZOOKEEPER_CONNECT"
