//go:build linux
// +build linux

package kafka

var defaultMBeanYAML = `
kafka-all-messages:
  objectName: "kafka.server:type=BrokerTopicMetrics,name=MessagesInPerSec"
  instancePrefix: "all"
  values:
  - instancePrefix: "kafka-messages-in"
    type: "counter"
    table: false
    attribute: "Count"

kafka-all-bytes-in:
  objectName: "kafka.server:type=BrokerTopicMetrics,name=BytesInPerSec"
  instancePrefix: "all"
  values:
  - instancePrefix: "kafka-bytes-in"
    type: "counter"
    table: false
    attribute: "Count"

kafka-all-bytes-out:
  objectName: "kafka.server:type=BrokerTopicMetrics,name=BytesOutPerSec"
  instancePrefix: "all"
  values:
  - instancePrefix: "kafka-bytes-out"
    type: "counter"
    table: false
    attribute: "Count"

kafka-active-controllers:
  objectName: "kafka.controller:type=KafkaController,name=ActiveControllerCount"
  values:
  - type: "gauge"
    table: false
    attribute: "Value"
    instancePrefix: "kafka-active-controllers"

kafka-offline-partitions-count:
  objectName: "kafka.controller:type=KafkaController,name=OfflinePartitionsCount"
  values:
  - type: "gauge"
    table: false
    attribute: "Value"
    instancePrefix: "kafka-offline-partitions-count"

kafka-underreplicated-partitions:
  objectName: "kafka.server:type=ReplicaManager,name=UnderReplicatedPartitions"
  values:
  - type: "gauge"
    table: false
    attribute: "Value"
    instancePrefix: "kafka-underreplicated-partitions"

kafka-isr-shrinks:
  objectName: "kafka.server:type=ReplicaManager,name=IsrShrinksPerSec"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka-isr-shrinks"

kafka-isr-expands:
  objectName: "kafka.server:type=ReplicaManager,name=IsrExpandsPerSec"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka-isr-expands"

kafka-max-lag:
  objectName: "kafka.server:type=ReplicaFetcherManager,name=MaxLag,clientId=Replica"
  values:
  - type: "gauge"
    table: false
    attribute: "Value"
    instancePrefix: "kafka-max-lag"

kafka-leader-election-rate:
  objectName: "kafka.controller:type=ControllerStats,name=LeaderElectionRateAndTimeMs"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka-leader-election-rate"

kafka-unclean-elections:
  objectName: "kafka.controller:type=ControllerStats,name=UncleanLeaderElectionsPerSec"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka-unclean-elections-rate"

kafka-request-queue:
  objectName: "kafka.network:type=RequestChannel,name=RequestQueueSize"
  values:
  - type: "gauge"
    table: false
    attribute: "Value"
    instancePrefix: "kafka-request-queue"

kafka.fetch-consumer.total-time:
  objectName: "kafka.network:type=RequestMetrics,name=TotalTimeMs,request=FetchConsumer"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka.fetch-consumer.total-time.count"
  - type: "gauge"
    table: false
    attribute: "50thPercentile"
    instancePrefix: "kafka.fetch-consumer.total-time.median"
  - type: "gauge"
    table: false
    attribute: "99thPercentile"
    instancePrefix: "kafka.fetch-consumer.total-time.99th"

kafka.fetch-follower.total-time:
  objectName: "kafka.network:type=RequestMetrics,name=TotalTimeMs,request=FetchFollower"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka.fetch-follower.total-time.count"
  - type: "gauge"
    table: false
    attribute: "50thPercentile"
    instancePrefix: "kafka.fetch-follower.total-time.median"
  - type: "gauge"
    table: false
    attribute: "99thPercentile"
    instancePrefix: "kafka.fetch-follower.total-time.99th"

kafka.produce.total-time:
  objectName: "kafka.network:type=RequestMetrics,name=TotalTimeMs,request=Produce"
  values:
  - type: "counter"
    table: false
    attribute: "Count"
    instancePrefix: "kafka.produce.total-time.count"
  - type: "gauge"
    table: false
    attribute: "50thPercentile"
    instancePrefix: "kafka.produce.total-time.median"
  - type: "gauge"
    table: false
    attribute: "99thPercentile"
    instancePrefix: "kafka.produce.total-time.99th"

kafka.log.flush:
  objectName: "kafka.log:type=LogFlushStats,name=LogFlushRateAndTimeMs"
  values:
  - type: counter
    table: false
    attribute: Count
    instancePrefix: kafka.logs.flush-time.count
  - type: gauge
    table: false
    attribute: 50thPercentile
    instancePrefix: kafka.logs.flush-time.median
  - type: gauge
    table: false
    attribute: 99thPercentile
    instancePrefix: kafka.logs.flush-time.99th
`
