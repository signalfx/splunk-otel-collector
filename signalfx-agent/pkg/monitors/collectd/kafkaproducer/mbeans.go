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

package kafkaproducer

var defaultMBeanYAML = `
response-rate:
  objectName: "kafka.producer:client-id=*,type=producer-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.producer.response-rate"
    type: "gauge"
    table: false
    attribute: "response-rate"

request-rate:
  objectName: "kafka.producer:client-id=*,type=producer-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.producer.request-rate"
    type: "gauge"
    table: false
    attribute: "request-rate"

request-latency-avg:
  objectName: "kafka.producer:client-id=*,type=producer-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.producer.request-latency-avg"
    type: "gauge"
    table: false
    attribute: "request-latency-avg"

outgoing-byte-rate:
  objectName: "kafka.producer:client-id=*,type=producer-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.producer.outgoing-byte-rate"
    type: "gauge"
    table: false
    attribute: "outgoing-byte-rate"

io-wait-time-ns-avg:
  objectName: "kafka.producer:client-id=*,type=producer-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  values:
  - instancePrefix: "kafka.producer.io-wait-time-ns-avg"
    type: "gauge"
    table: false
    attribute: "io-wait-time-ns-avg"

byte-rate-per-topic:
  objectName: "kafka.producer:client-id=*,topic=*,type=producer-topic-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.producer.byte-rate"
    type: "gauge"
    table: false
    attribute: "byte-rate"

compression-rate:
  objectName: "kafka.producer:client-id=*,topic=*,type=producer-topic-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.producer.compression-rate"
    type: "gauge"
    table: false
    attribute: "compression-rate"

record-error-rate:
  objectName: "kafka.producer:client-id=*,topic=*,type=producer-topic-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.producer.record-error-rate"
    type: "gauge"
    table: false
    attribute: "record-error-rate"

record-retry-rate:
  objectName: "kafka.producer:client-id=*,topic=*,type=producer-topic-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.producer.record-retry-rate"
    type: "gauge"
    table: false
    attribute: "record-retry-rate"

record-send-rate:
  objectName: "kafka.producer:client-id=*,topic=*,type=producer-topic-metrics"
  instancePrefix: "all"
  dimensions:
  - client-id
  - topic
  values:
  - instancePrefix: "kafka.producer.record-send-rate"
    type: "gauge"
    table: false
    attribute: "record-send-rate"
`
