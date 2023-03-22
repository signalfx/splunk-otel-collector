//go:build linux
// +build linux

package hadoopjmx

var defaultNodeManagerMBeanYAML = `
hadoop-nodeManager-metrics:
  objectName: "Hadoop:name=NodeManagerMetrics,service=NodeManager"
  values:
  - instancePrefix: "hadoop-nodeManager-allocated-memory"
    type: "gauge"
    table: false
    attribute: "AllocatedGB"
  - instancePrefix: "hadoop-nodeManager-available-memory"
    type: "gauge"
    table: false
    attribute: "AvailableGB"
  - instancePrefix: "hadoop-nodeManager-containers-launched"
    type: "counter"
    table: false
    attribute: "ContainersLaunched"
  - instancePrefix: "hadoop-nodeManager-containers-failed"
    type: "counter"
    table: false
    attribute: "ContainersFailed"
  - instancePrefix: "hadoop-nodeManager-allocated-vcores"
    type: "gauge"
    table: false
    attribute: "AllocatedVCores"
  - instancePrefix: "hadoop-nodeManager-available-vcores"
    type: "gauge"
    table: false
    attribute: "AvailableVCores"
`
