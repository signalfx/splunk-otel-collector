// Code generated by monitor-code-gen. DO NOT EDIT.

package atlas

import (
	"github.com/signalfx/golib/v3/datapoint"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

const monitorType = "mongodb-atlas"

const (
	groupHardware = "hardware"
	groupMongodb  = "mongodb"
)

var groupSet = map[string]bool{
	groupHardware: true,
	groupMongodb:  true,
}

const (
	assertsMsg                              = "asserts.msg"
	assertsRegular                          = "asserts.regular"
	assertsUser                             = "asserts.user"
	assertsWarning                          = "asserts.warning"
	backgroundFlushAvg                      = "background_flush_avg"
	cacheBytesReadInto                      = "cache.bytes.read_into"
	cacheBytesWrittenFrom                   = "cache.bytes.written_from"
	cacheDirtyBytes                         = "cache.dirty_bytes"
	cacheUsedBytes                          = "cache.used_bytes"
	connectionsCurrent                      = "connections.current"
	cursorsTimedOut                         = "cursors.timed_out"
	cursorsTotalOpen                        = "cursors.total_open"
	dataSize                                = "data_size"
	diskPartitionIopsRead                   = "disk.partition.iops.read"
	diskPartitionIopsTotal                  = "disk.partition.iops.total"
	diskPartitionIopsWrite                  = "disk.partition.iops.write"
	diskPartitionLatencyRead                = "disk.partition.latency.read"
	diskPartitionLatencyWrite               = "disk.partition.latency.write"
	diskPartitionSpaceFree                  = "disk.partition.space.free"
	diskPartitionSpacePercentFree           = "disk.partition.space.percent_free"
	diskPartitionSpacePercentUsed           = "disk.partition.space.percent_used"
	diskPartitionSpaceUsed                  = "disk.partition.space.used"
	diskPartitionUtilization                = "disk.partition.utilization"
	documentMetricsDeleted                  = "document.metrics.deleted"
	documentMetricsInserted                 = "document.metrics.inserted"
	documentMetricsReturned                 = "document.metrics.returned"
	documentMetricsUpdated                  = "document.metrics.updated"
	extraInfoPageFaults                     = "extra_info.page_faults"
	globalLockCurrentQueueReaders           = "global_lock.current_queue.readers"
	globalLockCurrentQueueTotal             = "global_lock.current_queue.total"
	globalLockCurrentQueueWriters           = "global_lock.current_queue.writers"
	indexSize                               = "index_size"
	memMapped                               = "mem.mapped"
	memResident                             = "mem.resident"
	memVirtual                              = "mem.virtual"
	networkBytesIn                          = "network.bytes_in"
	networkBytesOut                         = "network.bytes_out"
	networkNumRequests                      = "network.num_requests"
	opExecutionTimeCommands                 = "op.execution.time.commands"
	opExecutionTimeReads                    = "op.execution.time.reads"
	opExecutionTimeWrites                   = "op.execution.time.writes"
	opcounterCommand                        = "opcounter.command"
	opcounterDelete                         = "opcounter.delete"
	opcounterGetmore                        = "opcounter.getmore"
	opcounterInsert                         = "opcounter.insert"
	opcounterQuery                          = "opcounter.query"
	opcounterReplCommand                    = "opcounter.repl.command"
	opcounterReplDelete                     = "opcounter.repl.delete"
	opcounterReplInsert                     = "opcounter.repl.insert"
	opcounterReplUpdate                     = "opcounter.repl.update"
	opcounterUpdate                         = "opcounter.update"
	operationsScanAndOrder                  = "operations_scan_and_order"
	oplogMasterLagTimeDiff                  = "oplog.master.lag_time_diff"
	oplogMasterTime                         = "oplog.master.time"
	oplogRate                               = "oplog.rate"
	oplogSlaveLagMasterTime                 = "oplog.slave.lag_master_time"
	processCPUKernel                        = "process.cpu.kernel"
	processCPUUser                          = "process.cpu.user"
	processNormalizedCPUChildrenKernel      = "process.normalized.cpu.children_kernel"
	processNormalizedCPUChildrenUser        = "process.normalized.cpu.children_user"
	processNormalizedCPUKernel              = "process.normalized.cpu.kernel"
	processNormalizedCPUUser                = "process.normalized.cpu.user"
	queryExecutorScanned                    = "query.executor.scanned"
	queryExecutorScannedObjects             = "query.executor.scanned_objects"
	queryTargetingScannedObjectsPerReturned = "query.targeting.scanned_objects_per_returned"
	queryTargetingScannedPerReturned        = "query.targeting.scanned_per_returned"
	storageSize                             = "storage_size"
	systemCPUGuest                          = "system.cpu.guest"
	systemCPUIowait                         = "system.cpu.iowait"
	systemCPUIrq                            = "system.cpu.irq"
	systemCPUKernel                         = "system.cpu.kernel"
	systemCPUNice                           = "system.cpu.nice"
	systemCPUSoftirq                        = "system.cpu.softirq"
	systemCPUSteal                          = "system.cpu.steal"
	systemCPUUser                           = "system.cpu.user"
	systemNormalizedCPUGuest                = "system.normalized.cpu.guest"
	systemNormalizedCPUIowait               = "system.normalized.cpu.iowait"
	systemNormalizedCPUIrq                  = "system.normalized.cpu.irq"
	systemNormalizedCPUKernel               = "system.normalized.cpu.kernel"
	systemNormalizedCPUNice                 = "system.normalized.cpu.nice"
	systemNormalizedCPUSoftirq              = "system.normalized.cpu.softirq"
	systemNormalizedCPUSteal                = "system.normalized.cpu.steal"
	systemNormalizedCPUUser                 = "system.normalized.cpu.user"
	ticketsAvailableReads                   = "tickets.available.reads"
	ticketsAvailableWrite                   = "tickets.available.write"
)

var metricSet = map[string]monitors.MetricInfo{
	assertsMsg:                              {Type: datapoint.Count, Group: groupMongodb},
	assertsRegular:                          {Type: datapoint.Count, Group: groupMongodb},
	assertsUser:                             {Type: datapoint.Count, Group: groupMongodb},
	assertsWarning:                          {Type: datapoint.Count, Group: groupMongodb},
	backgroundFlushAvg:                      {Type: datapoint.Count, Group: groupMongodb},
	cacheBytesReadInto:                      {Type: datapoint.Count, Group: groupMongodb},
	cacheBytesWrittenFrom:                   {Type: datapoint.Count, Group: groupMongodb},
	cacheDirtyBytes:                         {Type: datapoint.Gauge, Group: groupMongodb},
	cacheUsedBytes:                          {Type: datapoint.Gauge, Group: groupMongodb},
	connectionsCurrent:                      {Type: datapoint.Gauge, Group: groupMongodb},
	cursorsTimedOut:                         {Type: datapoint.Count, Group: groupMongodb},
	cursorsTotalOpen:                        {Type: datapoint.Gauge, Group: groupMongodb},
	dataSize:                                {Type: datapoint.Gauge, Group: groupMongodb},
	diskPartitionIopsRead:                   {Type: datapoint.Count, Group: groupHardware},
	diskPartitionIopsTotal:                  {Type: datapoint.Count, Group: groupHardware},
	diskPartitionIopsWrite:                  {Type: datapoint.Count, Group: groupHardware},
	diskPartitionLatencyRead:                {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionLatencyWrite:               {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionSpaceFree:                  {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionSpacePercentFree:           {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionSpacePercentUsed:           {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionSpaceUsed:                  {Type: datapoint.Gauge, Group: groupHardware},
	diskPartitionUtilization:                {Type: datapoint.Gauge, Group: groupHardware},
	documentMetricsDeleted:                  {Type: datapoint.Count, Group: groupMongodb},
	documentMetricsInserted:                 {Type: datapoint.Count, Group: groupMongodb},
	documentMetricsReturned:                 {Type: datapoint.Count, Group: groupMongodb},
	documentMetricsUpdated:                  {Type: datapoint.Count, Group: groupMongodb},
	extraInfoPageFaults:                     {Type: datapoint.Count, Group: groupMongodb},
	globalLockCurrentQueueReaders:           {Type: datapoint.Gauge, Group: groupMongodb},
	globalLockCurrentQueueTotal:             {Type: datapoint.Gauge, Group: groupMongodb},
	globalLockCurrentQueueWriters:           {Type: datapoint.Gauge, Group: groupMongodb},
	indexSize:                               {Type: datapoint.Gauge, Group: groupMongodb},
	memMapped:                               {Type: datapoint.Gauge, Group: groupMongodb},
	memResident:                             {Type: datapoint.Gauge, Group: groupMongodb},
	memVirtual:                              {Type: datapoint.Gauge, Group: groupMongodb},
	networkBytesIn:                          {Type: datapoint.Gauge, Group: groupMongodb},
	networkBytesOut:                         {Type: datapoint.Gauge, Group: groupMongodb},
	networkNumRequests:                      {Type: datapoint.Count, Group: groupMongodb},
	opExecutionTimeCommands:                 {Type: datapoint.Gauge, Group: groupMongodb},
	opExecutionTimeReads:                    {Type: datapoint.Gauge, Group: groupMongodb},
	opExecutionTimeWrites:                   {Type: datapoint.Gauge, Group: groupMongodb},
	opcounterCommand:                        {Type: datapoint.Count, Group: groupMongodb},
	opcounterDelete:                         {Type: datapoint.Count, Group: groupMongodb},
	opcounterGetmore:                        {Type: datapoint.Count, Group: groupMongodb},
	opcounterInsert:                         {Type: datapoint.Count, Group: groupMongodb},
	opcounterQuery:                          {Type: datapoint.Count, Group: groupMongodb},
	opcounterReplCommand:                    {Type: datapoint.Count, Group: groupMongodb},
	opcounterReplDelete:                     {Type: datapoint.Count, Group: groupMongodb},
	opcounterReplInsert:                     {Type: datapoint.Count, Group: groupMongodb},
	opcounterReplUpdate:                     {Type: datapoint.Count, Group: groupMongodb},
	opcounterUpdate:                         {Type: datapoint.Count, Group: groupMongodb},
	operationsScanAndOrder:                  {Type: datapoint.Count, Group: groupMongodb},
	oplogMasterLagTimeDiff:                  {Type: datapoint.Gauge, Group: groupMongodb},
	oplogMasterTime:                         {Type: datapoint.Gauge, Group: groupMongodb},
	oplogRate:                               {Type: datapoint.Gauge, Group: groupMongodb},
	oplogSlaveLagMasterTime:                 {Type: datapoint.Gauge, Group: groupMongodb},
	processCPUKernel:                        {Type: datapoint.Gauge, Group: groupHardware},
	processCPUUser:                          {Type: datapoint.Gauge, Group: groupHardware},
	processNormalizedCPUChildrenKernel:      {Type: datapoint.Gauge, Group: groupHardware},
	processNormalizedCPUChildrenUser:        {Type: datapoint.Gauge, Group: groupHardware},
	processNormalizedCPUKernel:              {Type: datapoint.Gauge, Group: groupHardware},
	processNormalizedCPUUser:                {Type: datapoint.Gauge, Group: groupHardware},
	queryExecutorScanned:                    {Type: datapoint.Count, Group: groupMongodb},
	queryExecutorScannedObjects:             {Type: datapoint.Count, Group: groupMongodb},
	queryTargetingScannedObjectsPerReturned: {Type: datapoint.Gauge, Group: groupMongodb},
	queryTargetingScannedPerReturned:        {Type: datapoint.Gauge, Group: groupMongodb},
	storageSize:                             {Type: datapoint.Gauge, Group: groupMongodb},
	systemCPUGuest:                          {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUIowait:                         {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUIrq:                            {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUKernel:                         {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUNice:                           {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUSoftirq:                        {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUSteal:                          {Type: datapoint.Gauge, Group: groupHardware},
	systemCPUUser:                           {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUGuest:                {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUIowait:               {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUIrq:                  {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUKernel:               {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUNice:                 {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUSoftirq:              {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUSteal:                {Type: datapoint.Gauge, Group: groupHardware},
	systemNormalizedCPUUser:                 {Type: datapoint.Gauge, Group: groupHardware},
	ticketsAvailableReads:                   {Type: datapoint.Gauge, Group: groupMongodb},
	ticketsAvailableWrite:                   {Type: datapoint.Gauge, Group: groupMongodb},
}

var defaultMetrics = map[string]bool{
	connectionsCurrent:            true,
	dataSize:                      true,
	diskPartitionIopsRead:         true,
	diskPartitionIopsWrite:        true,
	extraInfoPageFaults:           true,
	globalLockCurrentQueueReaders: true,
	globalLockCurrentQueueWriters: true,
	indexSize:                     true,
	memResident:                   true,
	memVirtual:                    true,
	networkBytesIn:                true,
	networkBytesOut:               true,
	networkNumRequests:            true,
	opcounterCommand:              true,
	opcounterDelete:               true,
	opcounterGetmore:              true,
	opcounterInsert:               true,
	opcounterQuery:                true,
	opcounterUpdate:               true,
	oplogMasterLagTimeDiff:        true,
	processCPUUser:                true,
	storageSize:                   true,
}

var groupMetricsMap = map[string][]string{
	groupHardware: {
		diskPartitionIopsRead,
		diskPartitionIopsTotal,
		diskPartitionIopsWrite,
		diskPartitionLatencyRead,
		diskPartitionLatencyWrite,
		diskPartitionSpaceFree,
		diskPartitionSpacePercentFree,
		diskPartitionSpacePercentUsed,
		diskPartitionSpaceUsed,
		diskPartitionUtilization,
		processCPUKernel,
		processCPUUser,
		processNormalizedCPUChildrenKernel,
		processNormalizedCPUChildrenUser,
		processNormalizedCPUKernel,
		processNormalizedCPUUser,
		systemCPUGuest,
		systemCPUIowait,
		systemCPUIrq,
		systemCPUKernel,
		systemCPUNice,
		systemCPUSoftirq,
		systemCPUSteal,
		systemCPUUser,
		systemNormalizedCPUGuest,
		systemNormalizedCPUIowait,
		systemNormalizedCPUIrq,
		systemNormalizedCPUKernel,
		systemNormalizedCPUNice,
		systemNormalizedCPUSoftirq,
		systemNormalizedCPUSteal,
		systemNormalizedCPUUser,
	},
	groupMongodb: {
		assertsMsg,
		assertsRegular,
		assertsUser,
		assertsWarning,
		backgroundFlushAvg,
		cacheBytesReadInto,
		cacheBytesWrittenFrom,
		cacheDirtyBytes,
		cacheUsedBytes,
		connectionsCurrent,
		cursorsTimedOut,
		cursorsTotalOpen,
		dataSize,
		documentMetricsDeleted,
		documentMetricsInserted,
		documentMetricsReturned,
		documentMetricsUpdated,
		extraInfoPageFaults,
		globalLockCurrentQueueReaders,
		globalLockCurrentQueueTotal,
		globalLockCurrentQueueWriters,
		indexSize,
		memMapped,
		memResident,
		memVirtual,
		networkBytesIn,
		networkBytesOut,
		networkNumRequests,
		opExecutionTimeCommands,
		opExecutionTimeReads,
		opExecutionTimeWrites,
		opcounterCommand,
		opcounterDelete,
		opcounterGetmore,
		opcounterInsert,
		opcounterQuery,
		opcounterReplCommand,
		opcounterReplDelete,
		opcounterReplInsert,
		opcounterReplUpdate,
		opcounterUpdate,
		operationsScanAndOrder,
		oplogMasterLagTimeDiff,
		oplogMasterTime,
		oplogRate,
		oplogSlaveLagMasterTime,
		queryExecutorScanned,
		queryExecutorScannedObjects,
		queryTargetingScannedObjectsPerReturned,
		queryTargetingScannedPerReturned,
		storageSize,
		ticketsAvailableReads,
		ticketsAvailableWrite,
	},
}

var monitorMetadata = monitors.Metadata{
	MonitorType:     "mongodb-atlas",
	DefaultMetrics:  defaultMetrics,
	Metrics:         metricSet,
	SendUnknown:     false,
	Groups:          groupSet,
	GroupMetricsMap: groupMetricsMap,
	SendAll:         false,
}
