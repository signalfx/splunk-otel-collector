//go:build linux
// +build linux

package kafkaconsumer

var defaultMBeanYAML = `
records-lag-max:
  objectName: "kafka.consumer:client-id=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.consumer.records-lag-max"
    type: "gauge"
    table: false
    attribute: "records-lag-max"

bytes-consumed-rate:
  objectName: "kafka.consumer:client-id=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.consumer.bytes-consumed-rate"
    type: "gauge"
    table: false
    attribute: "bytes-consumed-rate"

records-consumed-rate:
  objectName: "kafka.consumer:client-id=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.consumer.records-consumed-rate"
    type: "gauge"
    table: false
    attribute: "records-consumed-rate"

fetch-rate:
  objectName: "kafka.consumer:client-id=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.consumer.fetch-rate"
    type: "gauge"
    table: false
    attribute: "fetch-rate"

fetch-size-avg:
  objectName: "kafka.consumer:client-id=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.consumer.fetch-size-avg"
    type: "gauge"
    table: false
    attribute: "fetch-size-avg"

bytes-consumed-rate-per-topic:
  objectName: "kafka.consumer:client-id=*,topic=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.consumer.bytes-consumed-rate"
    type: "gauge"
    table: false
    attribute: "bytes-consumed-rate"

records-consumed-rate-per-topic:
  objectName: "kafka.consumer:client-id=*,topic=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.consumer.records-consumed-rate"
    type: "gauge"
    table: false
    attribute: "records-consumed-rate"

fetch-size-avg-per-topic:
  objectName: "kafka.consumer:client-id=*,topic=*,type=consumer-fetch-manager-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.consumer.fetch-size-avg"
    type: "gauge"
    table: false
    attribute: "fetch-size-avg"
`
