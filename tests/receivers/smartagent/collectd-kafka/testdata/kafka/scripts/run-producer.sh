#!/usr/bin/env bash
set -e

if [[ -z "$KAFKA_BROKER" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_BROKER"
    exit 1
fi

function produce() {
    local i=0
    while true; do
        echo "Hello World $i"
        (( i += 1))
        sleep $((1 + RANDOM % 10))
    done
}

produce | "$KAFKA_BIN"/kafka-console-producer.sh --broker-list "$KAFKA_BROKER" --topic sfx-employee