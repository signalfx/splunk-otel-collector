//go:build linux
// +build linux

package hadoopjmx

var defaultNameNodeMBeanYAML = `
hadoop-namenode-fsNameSystem:
  objectName: "Hadoop:service=NameNode,name=FSNamesystem"
  values:
  - instancePrefix: "hadoop-namenode-capacity-remaining"
    type: "gauge"
    table: false
    attribute: "CapacityRemaining"
  - instancePrefix: "hadoop-namenode-capacity-used"
    type: "gauge"
    table: false
    attribute: "CapacityUsed"
  - instancePrefix: "hadoop-namenode-files-total"
    type: "counter"
    table: false
    attribute: "FilesTotal"
  - instancePrefix: "hadoop-namenode-stale-datanodes"
    type: "gauge"
    table: false
    attribute: "StaleDataNodes"
  - instancePrefix: "hadoop-namenode-total-load"
    type: "counter"
    table: false
    attribute: "TotalLoad"
  - instancePrefix: "hadoop-namenode-corrupt-blocks"
    type: "gauge"
    table: false
    attribute: "CorruptBlocks"
  - instancePrefix: "hadoop-namenode-missing-blocks"
    type: "gauge"
    table: false
    attribute: "MissingBlocks"
  - instancePrefix: "hadoop-namenode-under-replicated-blocks"
    type: "gauge"
    table: false
    attribute: "UnderReplicatedBlocks"
  - instancePrefix: "hadoop-namenode-blocks-with-corrupt-replicas"
    type: "gauge"
    table: false
    attribute: "CorruptBlocks"

hadoop-namenode-fsNameSystemState:
  objectName: "Hadoop:service=NameNode,name=FSNamesystemState"
  values:
  - instancePrefix: "hadoop-namenode-capacity-total"
    type: "gauge"
    table: false
    attribute: "CapacityTotal"
  - instancePrefix: "hadoop-namenode-live-datanodes"
    type: "gauge"
    table: false
    attribute: "NumLiveDataNodes"
  - instancePrefix: "hadoop-namenode-dead-datanodes"
    type: "gauge"
    table: false
    attribute: "NumDeadDataNodes"
  - instancePrefix: "hadoop-namenode-volume-failures"
    type: "counter"
    table: false
    attribute: "VolumeFailuresTotal"

hadoop-namenode-jvmMetrics:
  objectName: "Hadoop:service=NameNode,name=JvmMetrics"
  values:
  - instancePrefix: "hadoop-namenode-max-heap"
    type: "gauge"
    table: false
    attribute: "MemHeapMaxM"
  - instancePrefix: "hadoop-namenode-current-heap-used"
    type: "gauge"
    table: false
    attribute: "MemHeapUsedM"
  - instancePrefix: "hadoop-namenode-gc-count"
    type: "counter"
    table: false
    attribute: "GcCount"
  - instancePrefix: "hadoop-namenode-gc-time"
    type: "counter"
    table: false
    attribute: "GcTimeMillis"

hadoop-namenode-rpc:
  objectName: "Hadoop:service=NameNode,name=RpcActivityForPort*"
  values:
  - instancePrefix: "hadoop-namenode-rpc-avg-process-time"
    type: "gauge"
    table: false
    attribute: "RpcProcessingTimeAvgTime"
  - instancePrefix: "hadoop-namenode-rpc-total-calls"
    type: "counter"
    table: false
    attribute: "RpcProcessingTimeNumOps"
  - instancePrefix: "hadoop-namenode-rpc-avg-queue"
    type: "gauge"
    table: false
    attribute: "RpcQueueTimeAvgTime"

hadoop-namenode-info:
  objectName: "Hadoop:name=NameNodeInfo,service=NameNode"
  values:
  - instancePrefix: "hadoop-namenode-percent-remaining"
    type: "gauge"
    table: false
    attribute: "PercentRemaining"
  - instancePrefix: "hadoop-namenode-percent-dfs-used"
    type: "gauge"
    table: false
    attribute: "PercentUsed"
  - instancePrefix: "hadoop-namenode-dfs-free"
    type: "gauge"
    table: false
    attribute: "Free"
`
