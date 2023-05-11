//go:build linux
// +build linux

package hadoopjmx

var defaultDataNodeMBeanYAML = `
hadoop-datanode-FsVolume:
  objectName: "Hadoop:name=DataNodeVolume-/hadoop/hdfs/data,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-dataFile-io-avg"
    type: "gauge"
    table: false
    attribute: "DataFileIoRateAvgTime"
  - instancePrefix: "hadoop-datanode-total-metadata-ops"
    type: "counter"
    table: false
    attribute: "TotalMetadataOperations"
  - instancePrefix: "hadoop-datanode-metadata-ops"
    type: "gauge"
    table: false
    attribute: "MetadataOperationRateNumOps"
  - instancePrefix: "hadoop-datanode-metadata-ops-avg"
    type: "gauge"
    table: false
    attribute: "MetadataOperationRateAvgTime"
  - instancePrefix: "hadoop-datanode-dataFile-io"
    type: "counter"
    table: false
    attribute: "TotalDataFileIos"
  - instancePrefix: "hadoop-datanode-dataFile-io-ops"
    type: "gauge"
    table: false
    attribute: "DataFileIoRateNumOps"
  - instancePrefix: "hadoop-datanode-flush-io-num-ops"
    type: "gauge"
    table: false
    attribute: "FlushIoRateNumOps"
  - instancePrefix: "hadoop-datanode-flush-io-num-ops"
    type: "gauge"
    table: false
    attribute: "FlushIoRateAvgTime"
  - instancePrefix: "hadoop-datanode-sync-io-num-ops"
    type: "gauge"
    table: false
    attribute: "SyncIoRateNumOps"
  - instancePrefix: "hadoop-datanode-sync-io-avg"
    type: "gauge"
    table: false
    attribute: "SyncIoRateAvgTime"
  - instancePrefix: "hadoop-datanode-read-io-num-ops"
    type: "gauge"
    table: false
    attribute: "ReadIoRateNumOps"
  - instancePrefix: "hadoop-datanode-read-io-avg"
    type: "gauge"
    table: false
    attribute: "ReadIoRateAvgTime"
  - instancePrefix: "hadoop-datanode-write-io-num-ops"
    type: "gauge"
    table: false
    attribute: "WriteIoRateNumOps"
  - instancePrefix: "hadoop-datanode-write-io-avg"
    type: "gauge"
    table: false
    attribute: "WriteIoRateAvgTime"
  - instancePrefix: "hadoop-datanode-total-file-io-errors"
    type: "counter"
    table: false
    attribute: "TotalFileIoErrors"
  - instancePrefix: "hadoop-datanode-file-io-num-ops"
    type: "gauge"
    table: false
    attribute: "FileIoErrorRateNumOps"
  - instancePrefix: "hadoop-datanode-file-io-avg"
    type: "gauge"
    table: false
    attribute: "FileIoErrorRateAvgTime"

hadoop-datanode-activity:
  objectName: "Hadoop:name=DataNodeActivity-*,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-bytes-written"
    type: "counter"
    table: false
    attribute: "BytesWritten"
  - instancePrefix: "hadoop-datanode-bytes-read"
    type: "counter"
    table: false
    attribute: "BytesRead"
  - instancePrefix: "hadoop-datanode-blocks-written"
    type: "counter"
    table: false
    attribute: "BlocksWritten"
  - instancePrefix: "hadoop-datanode-blocks-read"
    type: "counter"
    table: false
    attribute: "BlocksRead"

hadoop-datanode-fs-data-set-state:
  objectName: "Hadoop:name=FSDatasetState,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-fs-capacity"
    type: "gauge"
    table: false
    attribute: "Capacity"
  - instancePrefix: "hadoop-datanode-fs-dfs-used"
    type: "gauge"
    table: false
    attribute: "DfsUsed"
  - instancePrefix: "hadoop-datanode-fs-dfs-remaining"
    type: "gauge"
    table: false
    attribute: "Remaining"

hadoop-datenode-jvm:
  objectName: "Hadoop:name=JvmMetrics,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-jvm-non-heap-used"
    type: "gauge"
    table: false
    attribute: "MemNonHeapUsedM"
  - instancePrefix: "hadoop-datanode-jvm-heap-used"
    type: "gauge"
    table: false
    attribute: "MemHeapUsedM"

hadoop-datenode-info:
  objectName: "Hadoop:name=DataNodeInfo,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-info-xceiver"
    type: "gauge"
    table: false
    attribute: "XceiverCount"

hadoop-datanode-rpc:
  objectName: "Hadoop:name=RpcActivityForPort*,service=DataNode"
  values:
  - instancePrefix: "hadoop-datanode-rpc-queue-time-avg"
    type: "gauge"
    table: false
    attribute: "RpcQueueTimeAvgTime"
  - instancePrefix: "hadoop-datanode-rpc-processing-avg"
    type: "gauge"
    table: false
    attribute: "RpcProcessingTimeAvgTime"
  - instancePrefix: "hadoop-datanode-rpc-open-connections"
    type: "gauge"
    table: false
    attribute: "NumOpenConnections"
  - instancePrefix: "hadoop-datanode-rpc-call-queue-length"
    type: "gauge"
    table: false
    attribute: "CallQueueLength"
`
