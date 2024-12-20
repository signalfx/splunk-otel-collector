# CollectD example

This example shows how to connect a CollectD daemon running on a host to an OpenTelemetry Collector.

For the purpose of this example, the host is represented by a Ubuntu 24.04 docker image.

On this image, we install collectd as a Debian package, using stock instructions.

We proceed to add configuration to CollectD to instruct it to have an active behavior:
* We give it directions to ingest free disk related metrics through `collectd/metrics.conf`
* We instruct CollectD to send data over HTTP using `collectd/http.conf`

We also set up a collector to listen over HTTP for traffic from the collectD daemon.

To do so, we set up the [collectd receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/collectdreceiver):

```yaml
    collectd:
      endpoint: "0.0.0.0:8081"
```

Note we use `0.0.0.0` to make sure we expose the 8081 port over the Docker network interface so the 2 Docker containers may interact.

We run the example with the instruction to start the docker-compose setup, building the collectd container:

```bash
$> docker compose up --build
```

We check that the collector is indeed receiving metrics and logging them to stdout via the debug exporter:

```bash
$> docker logs otelcollector
```

A typical example of output is:
```
StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-12-20 19:55:44.006000128 +0000 UTC
Value: 38.976566
Metric #17
Descriptor:
     -> Name: percent_bytes.reserved
     -> Description: 
     -> Unit: 
     -> DataType: Gauge
NumberDataPoints #0
Data point attributes:
     -> plugin: Str(df)
     -> plugin_instance: Str(etc-hosts)
     -> host: Str(ea1d62c7a229)
     -> dsname: Str(value)
StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-12-20 19:55:44.006000128 +0000 UTC
Value: 5.102245
        {"kind": "exporter", "data_type": "metrics", "name": "debug"}
```