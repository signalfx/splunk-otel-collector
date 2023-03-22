//go:build linux
// +build linux

package activemq

var defaultMBeanYAML = `
activemq-broker:
  objectName: "org.apache.activemq:type=Broker,brokerName=*"
  instanceFrom:
  - brokerName
  values:
  - type: counter
    instancePrefix: amq.TotalConnectionsCount
    table: false
    attribute: TotalConnectionsCount
  - type: gauge
    instancePrefix: amq.TotalConsumerCount
    table: false
    attribute: TotalConsumerCount
  - type: gauge
    instancePrefix: amq.TotalDequeueCount
    table: false
    attribute: TotalDequeueCount
  - type: gauge
    instancePrefix: amq.TotalEnqueueCount
    table: false
    attribute: TotalEnqueueCount
  - type: gauge
    instancePrefix: amq.TotalMessageCount
    table: false
    attribute: TotalMessageCount
  - type: gauge
    instancePrefix: amq.TotalProducerCount
    table: false
    attribute: TotalProducerCount

activemq-queue:
  objectName: "org.apache.activemq:type=Broker,brokerName=*,destinationType=Queue,destinationName=*"
  instanceFrom:
  - brokerName
  - destinationName
  values:
  - type: gauge
    instancePrefix: amq.queue.QueueSize
    table: false
    attribute: QueueSize
  - type: gauge
    instancePrefix: amq.queue.AverageMessageSize
    table: false
    attribute: AverageMessageSize
  - type: gauge
    instancePrefix: amq.queue.ConsumerCount
    table: false
    attribute: ConsumerCount
  - type: gauge
    instancePrefix: amq.queue.ProducerCount
    table: false
    attribute: ProducerCount
  - type: gauge
    instancePrefix: amq.queue.DequeueCount
    table: false
    attribute: DequeueCount
  - type: gauge
    instancePrefix: amq.queue.EnqueueCount
    table: false
    attribute: EnqueueCount
  - type: gauge
    instancePrefix: amq.queue.ExpiredCount
    table: false
    attribute: ExpiredCount
  - type: gauge
    instancePrefix: amq.queue.ForwardCount
    table: false
    attribute: ForwardCount
  - type: gauge
    instancePrefix: amq.queue.InFlightCount
    table: false
    attribute: InFlightCount
  - type: gauge
    instancePrefix: amq.queue.AverageBlockedTime
    table: false
    attribute: AverageBlockedTime
  - type: gauge
    instancePrefix: amq.queue.AverageEnqueueTime
    table: false
    attribute: AverageEnqueueTime
  - type: gauge
    instancePrefix: amq.queue.BlockedSends
    table: false
    attribute: BlockedSends
  - type: gauge
    instancePrefix: amq.queue.TotalBlockedTime
    table: false
    attribute: TotalBlockedTime
  
activemq-topic:
  objectName: "org.apache.activemq:type=Broker,brokerName=*,destinationType=Topic,destinationName=*"
  instanceFrom:
  - brokerName
  - destinationName
  values:
  - type: gauge
    instancePrefix: amq.topic.QueueSize
    table: false
    attribute: QueueSize
  - type: gauge
    instancePrefix: amq.topic.AverageMessageSize
    table: false
    attribute: AverageMessageSize
  - type: gauge
    instancePrefix: amq.topic.ConsumerCount
    table: false
    attribute: ConsumerCount
  - type: gauge
    instancePrefix: amq.topic.ProducerCount
    table: false
    attribute: ProducerCount
  - type: gauge
    instancePrefix: amq.topic.DequeueCount
    table: false
    attribute: DequeueCount
  - type: gauge
    instancePrefix: amq.topic.EnqueueCount
    table: false
    attribute: EnqueueCount
  - type: gauge
    instancePrefix: amq.topic.ExpiredCount
    table: false
    attribute: ExpiredCount
  - type: gauge
    instancePrefix: amq.topic.ForwardCount
    table: false
    attribute: ForwardCount
  - type: gauge
    instancePrefix: amq.topic.InFlightCount
    table: false
    attribute: InFlightCount
  - type: gauge
    instancePrefix: amq.topic.AverageBlockedTime
    table: false
    attribute: AverageBlockedTime
  - type: gauge
    instancePrefix: amq.topic.AverageEnqueueTime
    table: false
    attribute: AverageEnqueueTime
  - type: gauge
    instancePrefix: amq.topic.BlockedSends
    table: false
    attribute: BlockedSends
  - type: gauge
    instancePrefix: amq.topic.TotalBlockedTime
    table: false
    attribute: TotalBlockedTime
`
