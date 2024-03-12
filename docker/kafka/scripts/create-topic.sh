#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_BROKER" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_BROKER"
    exit 1
fi
"$KAFKA_BIN"/kafka-topics.sh \
    --create \
    --bootstrap-server "$KAFKA_BROKER" \
    --replication-factor 1 \
    --partitions 1 \
    --topic sfx-employee \
    --config flush.messages=1
