# Sizing

The OpenTelemetry Collector can be scaled up or out as needed. Sizing is based
on the amount of data per data source and requires 1 CPU core per:

- Traces: 10,000 spans per second
- Metrics: 20,000 data points per second

If a Collector handles both trace and metric data then both must be accounted
for when sizing. For example, 5K spans per second plus 10K data points per
second would require 1 CPU core.

The recommendation is to use a ratio of 1 CPU to 2 GB of memory. By default, the
Collector is configured to use 512 MB of memory.

> The Collector does not persist data to disk so no disk space is required.

## Agent

For Agent instances, scale up resources as needed. Typically only a single
agent runs per application or host so properly sizing the agent is important.
Multiple independent agents could be deployed on a given application or host
depending on the use-case. For example, a privileged agent could be deployed
alongside an unprivileged agent.

## Gateway

For Gateway instances, allocate at least a CPU core per Collector. Note that
multiple Collectors can deployed behind a simple round-robin load balancer for
availability and performance reasons. Each Collector runs independently, so
scale increases linearly with the number of Collectors you deploy.

The recommendation is to configure at least N+1 redundancy, which means a load
balancer and a minimum of two Collector instances should be configured
initially.
