// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
