#!/usr/bin/env bash
set -e

if [[ -z "$START_AS" ]]; then
    echo "ERROR: missing mandatory config: START_AS"
    exit 1
fi

if [[ -z "$KAFKA_BIN" ]]; then
    echo "ERROR: missing mandatory config: KAFKA_BIN"
    exit 1
fi

case "$START_AS" in
 "broker" )
   echo "running broker"
   scripts/run-broker.sh
   ;;
 "consumer" )
   echo "running consumer"
   scripts/run-consumer.sh
   ;;
 "producer" )
   echo "running producer"
   scripts/run-producer.sh
   ;;
 "create-topic" )
   echo "creating topic"
   scripts/create-topic.sh
   ;;
  * )
   echo "Valid options include broker, consumer, producer, create-topic"
   exit 1
   ;;
esac
