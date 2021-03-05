#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_BROKER" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_BROKER"
    exit 1
fi

if [ "$KAFKA_VERSION" = "2.0.0" ]; then
 "$KAFKA_BIN"/kafka-console-consumer.sh \
    --bootstrap-server "$KAFKA_BROKER" \
    --topic sfx-employee \
    --max-messages $((10 + RANDOM % 100))
else
 "$KAFKA_BIN"/kafka-console-consumer.sh \
    --bootstrap-server "$KAFKA_BROKER" \
    --new-consumer \
    --topic sfx-employee \
    --max-messages $((10 + RANDOM % 100))
fi
