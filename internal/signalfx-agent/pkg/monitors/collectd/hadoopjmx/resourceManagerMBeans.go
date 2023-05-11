//go:build linux
// +build linux

package hadoopjmx

var defaultResourceManagerMBeanYAML = `
hadoop-resourceManager-jvm:
  objectName: "Hadoop:service=ResourceManager,name=JvmMetrics"
  values:
  - instancePrefix: "hadoop-resourceManager-heap-used"
    type: "gauge"
    table: false
    attribute: "MemHeapUsedM"
  - instancePrefix: "hadoop-resourceManager-heap-max"
    type: "gauge"
    table: false
    attribute: "MemHeapMaxM"

hadoop-resourceManager-cluster-metrics:
  objectName: "Hadoop:service=ResourceManager,name=ClusterMetrics"
  values:
  - instancePrefix: "hadoop-resourceManager-active-nms"
    type: "gauge"
    table: false
    attribute: "NumActiveNMs"
    
hadoop-resourceManager-queue-metrics:
  objectName: "Hadoop:service=ResourceManager,name=QueueMetrics,q0=root"
  values:
  - instancePrefix: "hadoop-resourceManager-active-apps"
    type: "gauge"
    table: false
    attribute: "ActiveApplications"
  - instancePrefix: "hadoop-resourceManager-available-memory"
    type: "gauge"
    table: false
    attribute: "AvailableMB"
  - instancePrefix: "hadoop-resourceManager-allocated-memory"
    type: "gauge"
    table: false
    attribute: "AllocatedMB"
  - instancePrefix: "hadoop-resourceManager-active-users"
    type: "gauge"
    table: false
    attribute: "ActiveUsers"
  - instancePrefix: "hadoop-resourceManager-available-vcores"
    type: "gauge"
    table: false
    attribute: "AvailableVCores"
  - instancePrefix: "hadoop-resourceManager-allocated-vcores"
    type: "gauge"
    table: false
    attribute: "AllocatedVCores"
  - instancePrefix: "hadoop-resourceManager-allocated-containers"
    type: "gauge"
    table: false
    attribute: "AllocatedContainers"
`
