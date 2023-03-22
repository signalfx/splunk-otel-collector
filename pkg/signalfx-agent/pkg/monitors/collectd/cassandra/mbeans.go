//go:build linux
// +build linux

package cassandra

var defaultMBeanYAML = `
cassandra-client-read-latency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Latency
  values:
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Read.Latency.50thPercentile
    attribute: 50thPercentile
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Read.Latency.Max
    attribute: Max
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Read.Latency.99thPercentile
    attribute: 99thPercentile
  - type: counter
    instancePrefix: cassandra.ClientRequest.Read.Latency.Count
    attribute: Count


cassandra-client-read-totallatency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=TotalLatency
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Read.TotalLatency.Count
    attribute: Count


cassandra-client-casread-latency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=CASRead,name=Latency
  values:
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASRead.Latency.50thPercentile
    attribute: 50thPercentile
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASRead.Latency.Max
    attribute: Max
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASRead.Latency.99thPercentile
    attribute: 99thPercentile
  - type: counter
    instancePrefix: cassandra.ClientRequest.CASRead.Latency.Count
    attribute: Count


cassandra-client-casread-totallatency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=CASRead,name=TotalLatency
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.CASRead.TotalLatency.Count
    attribute: Count


cassandra-client-read-timeouts:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Timeouts
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Read.Timeouts.Count
    attribute: Count


cassandra-client-read-unavailables:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Unavailables
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Read.Unavailables.Count
    attribute: Count


cassandra-client-rangeslice-latency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Latency
  values:
  - type: gauge
    instancePrefix: cassandra.ClientRequest.RangeSlice.Latency.50thPercentile
    attribute: 50thPercentile
  - type: gauge
    instancePrefix: cassandra.ClientRequest.RangeSlice.Latency.Max
    attribute: Max
  - type: gauge
    instancePrefix: cassandra.ClientRequest.RangeSlice.Latency.99thPercentile
    attribute: 99thPercentile
  - type: counter
    instancePrefix: cassandra.ClientRequest.RangeSlice.Latency.Count
    attribute: Count


cassandra-client-rangeslice-totallatency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=TotalLatency
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.RangeSlice.TotalLatency.Count
    attribute: Count


cassandra-client-rangeslice-timeouts:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Timeouts
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.RangeSlice.Timeouts.Count
    attribute: Count


cassandra-client-rangeslice-unavailables:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Unavailables
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.RangeSlice.Unavailables.Count
    attribute: Count


cassandra-client-write-latency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Latency
  values:
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Write.Latency.50thPercentile
    attribute: 50thPercentile
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Write.Latency.Max
    attribute: Max
  - type: gauge
    instancePrefix: cassandra.ClientRequest.Write.Latency.99thPercentile
    attribute: 99thPercentile
  - type: counter
    instancePrefix: cassandra.ClientRequest.Write.Latency.Count
    attribute: Count


cassandra-client-write-totallatency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=TotalLatency
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Write.TotalLatency.Count
    attribute: Count


cassandra-client-caswrite-latency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=CASWrite,name=Latency
  values:
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASWrite.Latency.50thPercentile
    attribute: 50thPercentile
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASWrite.Latency.Max
    attribute: Max
  - type: gauge
    instancePrefix: cassandra.ClientRequest.CASWrite.Latency.99thPercentile
    attribute: 99thPercentile
  - type: counter
    instancePrefix: cassandra.ClientRequest.CASWrite.Latency.Count
    attribute: Count


cassandra-client-caswrite-totallatency:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=CASWrite,name=TotalLatency
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.CASWrite.TotalLatency.Count
    attribute: Count


cassandra-client-write-timeouts:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Timeouts
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Write.Timeouts.Count
    attribute: Count


cassandra-client-write-unavailables:
  objectName: org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Unavailables
  values:
  - type: counter
    instancePrefix: cassandra.ClientRequest.Write.Unavailables.Count
    attribute: Count


cassandra-storage-exceptions:
  objectName: org.apache.cassandra.metrics:type=Storage,name=Exceptions
  values:
  - type: counter
    instancePrefix: cassandra.Storage.Exceptions.Count
    attribute: Count


cassandra-storage-load:
  objectName: org.apache.cassandra.metrics:type=Storage,name=Load
  values:
  - type: counter
    instancePrefix: cassandra.Storage.Load.Count
    attribute: Count


cassandra-storage-hints:
  objectName: org.apache.cassandra.metrics:type=Storage,name=TotalHints
  values:
  - type: counter
    instancePrefix: cassandra.Storage.TotalHints.Count
    attribute: Count


cassandra-storage-hints-in-progress:
  objectName: org.apache.cassandra.metrics:type=Storage,name=TotalHintsInProgress
  values:
  - type: counter
    instancePrefix: cassandra.Storage.TotalHintsInProgress.Count
    attribute: Count


cassandra-compaction-pending-tasks:
  objectName: org.apache.cassandra.metrics:type=Compaction,name=PendingTasks
  values:
  - type: gauge
    instancePrefix: cassandra.Compaction.PendingTasks.Value
    attribute: Value


cassandra-compaction-total-completed:
  objectName: org.apache.cassandra.metrics:type=Compaction,name=TotalCompactionsCompleted
  values:
  - type: counter
    instancePrefix: cassandra.Compaction.TotalCompactionsCompleted.Count
    attribute: Count
`
