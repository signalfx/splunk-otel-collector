# Persistent queue connector

The persistent queue connector stores data into a queue. It reads from the queue to send data down to exporters.

As such, this processor is asynchronous and can enforce a bandwidth limit.
It should be placed as a connector at the end of the processing pipeline. It can be placed after the batch processor. Batches should
be less than the bandwidth.

The queue retries indefinitely, and is strictly applying FIFO order. If it is reused across multiple pipelines, it will
apply the limit and FIFO across all signals.

# Configuration

| Name | Description | Default value |
| -- | -- | -- |
| path | The path at which the queue will be stored on disk | "" (required) |
| throughput_limit | The limit in bytes of data to read from the queue. If 0, no limit applies. | 0 |

# Example

```yaml
connectors:
  persistent_queue:
    path: /var/queue
    throughput_limit: 256768 # 256 Kb

receivers:
  filelog:
    include: ["/var/log/my.log"]

exporters:
  debug:

service:
  pipelines:
    logs:
      receivers: [filelog]
      exporters: [persistent_queue]
    logs/queue_out:
      receivers: [persistent_queue]
      exporters: [debug]
```