## Enhanced Collector Sizing

As an application deployed in customer environments, the collector should advertise its resource requirements for its
capabilities. This would aid in avoiding:

1. Failure scenarios and inefficiencies resulting from under and over provisioning.
2. Guesswork from the user and developer in assessing requirements and their variation resulting from arbitrary use
cases and changesets (scaling events, upgrades, functional changes, etc.).

Since the collector is an assortment of components with specialized purposes and features, many offering infinite
configuration permutations that influence utilization, it is generally impossible to accurately predict the requirements
for a deployment and engagement. Despite this, if individual components make their requirements known somehow, then
estimated tolerances are possible (though susceptible to unforeseen environmental and nonlinear scaling effects).

### Current Guidance

The current sizing guidelines for the collector are advertised
[here](https://docs.splunk.com/observability/en/gdi/opentelemetry/sizing.html). The quantified guidance provided is:

```
With a single CPU core, the Collector can receive, process, or export the following:

- If handling traces, 15,000 spans per second.
- If handling metrics, 20,000 data points per second.
- If handling logs, 10,000 log records per second, including Fluentd td-agent, which forwards logs to the fluentforward
  receiver in the Collector.
```

and 

```
Use a ratio of one CPU to 2 GB of memory.

If the Collector handles both trace and metrics data, consider both types of data when planning your deployment. For
example, 7.5K spans per second plus 10K data points per second requires 1 CPU core.

The Collector does not persist data to disk so no disk space is required.
```

Unfortunately these statements aren't based on advisable heuristics and don't account for the characteristics of default
deployments. The remaining guidance details some use case concerns to consider without providing grounded examples or
usable suggestions. This leaves the effort as a daunting exercise for the user.

The telemetry per second guidance was generated from modified contrib testbed load tests as described
[here](https://github.com/signalfx/splunk-otel-collector/pull/226). The
[contrib testbed](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/testbed/README.md)
provides a helpful platform for resource utilization sanity tests. However, it's currently not a reliable origin for
the per-core usage given its lack of setting processor affinity with taskset/cpuset. It also doesn't use real world
telemetry in its load generator, though it has this data provider capability. It also determines resource utilization
from procfs so is unable to provide package-level utilization figures. It also doesn't include disk or network resource
monitoring.

The 1:2 CPU to GB memory guidance appears to be [inherited from the OpenCensus
agent](https://github.com/open-telemetry/opentelemetry-collector/commit/740700747450ecca9e579c21816613dfa723e41b#diff-118f0ef20c6b4cf62284949363699de7c7412fddb16c693a22f07610ee14fec8R38)
and doesn't accurately reflect real world, per component, or use case determined requirements.

### Proposed Guidance

I propose that each component's real world resource requirements need to be ascertainable in order to provide valid
sizing guidance. This includes:

1. Processor requirements resembling nanoseconds per core per unit of work (scrapes, batches, unit telemetry, etc.)
   over 1..N cores. This should also include runtime reported os threads and goroutine info.
1. Memory requirements resembling bytes per unit of work for different runtime phases (start up, idle, etc.). This
   includes kernel reported vm size, rss, pss, as well as self-reported heap info from the pprof extension.
1. Network throughput resembling bytes per unit of work per request and general socket and buffer stats by throughput.
1. Disk requirements when using file storage extension resembling bytes per unit of work.

These measurements would ideally be available for an arbitrary receiver, processor, exporter, and extension using
arbitrary configuration without requiring intensive, upfront instrumentation. I have been unable to find a robust tool
in the go ecosystem offering these capabilities. If you know of one please share.

### Proposed Roadmap

There are a few immediate possible adoptions and improvements to the testbed that would be helpful in devising more
accurate guidance:

1. Create Splunk distro benchmarking suite that uses the contrib testbed helpers directly with Splunk-specific testdata 
   not suitable for upstream hosting. All improvements to the testbed will target upstream however, if accepted.
1. Use taskset or similar utility to limit collector process to a configurable number and arrangement of cores.
1. Model from or use recorded telemetry from actual host and cloud services instead of generic load generator content.
1. Exercise and incorporate exporter helper sending queue configuration to suite (`num_consumers`, queue size tuning,
   etc.)
   evaluating procfs and stop evaluating before teardown).

After these are delivered our guidance can be more readily updated and evaluated per release. However, meeting more
exacting requirements of the general aims is more complicated. I propose we develop a "utilization extension" that uses
the available stdlib runtime diagnostic features to present a more complete performance snapshot
to a modified contrib testbed.

This extension would use the same features used by the pprof and related helpers, but
also include a profile parser and stats aggregator similar to
[godeltaprof](https://github.com/grafana/pyroscope-go/tree/main/godeltaprof) allowing per package/component info. It
would also be able to analyze in-process heapdumps using the parser from
[heapspurs](https://github.com/adamroach/heapspurs) to assess complete heap utilization for arbitrary packages to
provide arbitrarily granular memory analysis. This could prove greatly useful to developers in evaluating the impact
from their changes and also provide tooling to immediately flag memory leaks during development and testing.

```yaml
extensions:
  utilization:
    # provides json-speaking endpoints at `endpoint`/util
    endpoint: localhost:4321
    # `endpoint`/util/cpu
    cpu:
      enabled: false # default is true
      collection_interval: 10s
    # `endpoint`/util/heap provides /debug/pprof/heap level info for the lifetime of the extension
    # `endpoint`/util/heap(?gc=N) option to run N runtime.GC() invocations before providing snapshot breakdown
    heap:
      enabled: false
      collection_interval: 10s
    # `endpoint`/util/heapdump provides stop the world debug.WriteHeapDump() snapshot analysis.
    # `endpoint`/util/heapdump(?gc=N) option to run N runtime.GC() before providing heapdump breakdown
    heapdump:
    # `endpoint`/util/net provides aggregated /proc/self/net/netstat and similar stats for the lifetime of the process.
    # I suspect aggregating by caller/library-level info is available w/ eBPF but I am not versed in these capabilities.
    net:
      enabled: true
      collection_interval: 10s
    # `endpoint`/util/disk provides aggregated lsfd similar metrics for the lifetime of the process
    # I suspect aggregating by caller/library-level info is available w/ eBPF but I am not versed in these capabilities.
    disk:
      enabled: true
      collection_interval: 10s
```

Where applicable filters are supported for requests to filter analysis based on matching and captured callstack roots:

```
The metric views provided in the response will be keyed by parents with the shortest match.
{"filter": "go.opentelemetry.io/collector/(?P<ComponentType>[^/]*)/(?P<Component>[^/]*)/.*"}
```

TODO: Example responses per endpoint based on available data and collection methods.
