#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_BROKER" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_BROKER"
    exit 1
fi

"$KAFKA_BIN"/kafka-console-consumer.sh \
   --bootstrap-server "$KAFKA_BROKER" \
   --topic sfx-employee \
   --max-messages $((10 + RANDOM % 100))
