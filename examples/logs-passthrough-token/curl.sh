#!/bin/sh

exitFlag=0

trap ctrl_c INT

function ctrl_c() {
  exitFlag=1
  echo "exit"
}

while [ $exitFlag -eq 0 ]
do
  curl -k http://otelcollector:8088/services/collector -d '{"event": "event"}' -H "Authorization: Splunk 00000000-0000-0000-0000-0000000000123"
  curl -k http://otelcollector:8088/services/collector/raw -d 'my raw event' -H "Authorization: Splunk 00000000-0000-0000-0000-0000000000123"
  sleep 5
done