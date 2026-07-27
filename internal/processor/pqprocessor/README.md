# Persistent queue processor

The persistent queue processor stores data into a queue. It reads from the queue to send data down to exporters.

As such, this processor is asynchronous and can enforce a bandwidth limit.

The queue retries indefinitely, and is strictly applying FIFO order. As such, it cannot be shared across multiple pipelines.

Example:
```yaml
persistent_queue:
  folder: /var/queue
  bandwidth: 256768 # 256 Kb
```