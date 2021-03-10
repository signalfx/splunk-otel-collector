#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_ZOOKEEPER_CONNECT" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_ZOOKEEPER_CONNECT"
    exit 1
fi
"$KAFKA_BIN"/kafka-topics.sh \
    --create \
    --zookeeper "$KAFKA_ZOOKEEPER_CONNECT" \
    --replication-factor 1 \
    --partitions 1 \
    --topic sfx-employee \
    --config flush.messages=1
