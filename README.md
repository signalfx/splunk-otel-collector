# Splunk distribution of OpenTelemetry Collector

The Splunk distribution of [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector) provides
a binary that can be deployed as a standalone service (also known as a gateway)
that can receive, process and export trace and metric data.

The collector is an optional component deployed either:

- Between the [Smart Agent](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-smart-agent.html) and SaaS ingestion
- Between a serverless function and SaaS ingestion

> :construction: This project is currently in **BETA**.

## Getting Started

This distribution comes with a [default
configuration](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector/splunk_config.yaml)
which requires the following environment variables:

- `SPLUNK_REALM` (no default): Which realm to send the data to (for example: `us0`)
- `SPLUNK_TOKEN` (no default): Access token to authenticate requests
- `SPLUNK_BALLAST_SIZE_MIB` (no default): How much memory to allocate to the ballast. This should be set to 1/3 to 1/2 of configured memory.

In addition, the following environment variables are optional:

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk-config.yaml`): Which configuration to load.
- `SPLUNK_MEMORY_LIMIT_PERCENTAGE` (default = `90`): Maximum amount of total memory targeted to be allocated by the process heap.
- `SPLUNK_MEMORY_SPIKE_PERCENTAGE` (default = `20`): Maximum spike expected between the measurements of memory usage.

When running on a non-linux system, the following environment variables are required:

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk-config.yaml`): Which configuration to load.
- `SPLUNK_MEMORY_LIMIT_MIB` (no default): Maximum amount of total memory targeted to be allocated by the process heap.
- `SPLUNK_MEMORY_SPIKE_MIB` (no default): Maximum spike expected between the measurements of memory usage.

Deploy the collector as outlined in the below. More information
about deploying and configuring the collector can be found
[here](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-opentelemetry-collector.html)

### Docker

Deploy from a Docker container (replace `0.1.0-otel-0.11.0` with the latest
stable version number if necessary):

```bash
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_BALLAST_SIZE_MIB=683 \
  docker run -p 13133 -p 14250 -p 14268 -p 55678-55680 -p 6060 -p 7276 -p 8888 -p 9411 -p 9943 \
    --name otelcol signalfx/splunk-otel-collector:0.1.0-otel-0.11.0
```

### Kubernetes

To deploy the OpenTelemetry Collector in Kubernetes, create a configuration
file that defines a ConfigMap, Service, and Deployment for the cluster. For
more information about creating a configuration file, see the example
[signalfx-k8s.yaml](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/exporter/sapmexporter/examples/signalfx-k8s.yaml)
file on GitHub.

### Standalone

```bash
$ make otelcol
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_BALLAST_SIZE_MIB=683 \
  ./bin/otelcol
```

## Advanced Configuration

### Command Line Arguments

Following the binary command or Docker container command line arguments can be
specified. Command line arguments take priority over environment variables.

For example in Docker:

```bash
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_BALLAST_SIZE_MIB=683 \
  docker run -p 13133 -p 14250 -p 14268 -p 55678-55680 -p 6060 -p 7276 -p 8888 -p 9411 -p 9943 \
    -v collector.yaml:/etc/collector.yaml:ro \
    --name otelcol signalfx/splunk-otel-collector:0.1.0-otel-0.11.0 \
        --log-level=DEBUG
```

### Custom Configuration

In addition to using the default configuration, a custom configuration can also
be provided.

For example in Docker:

```bash
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_BALLAST_SIZE_MIB=683 SPLUNK_CONFIG=/etc/collector.yaml \
  docker run -p 13133 -p 14250 -p 14268 -p 55678-55680 -p 6060 -p 7276 -p 8888 -p 9411 -p 9943 \
    -v collector.yaml:/etc/collector.yaml:ro \
    --name otelcol signalfx/splunk-otel-collector:0.1.0-otel-0.11.0
```

Note that if the configuration includes a memorylimiter processor then it must set the
value of `ballast_size_mib` setting of the processor to the env variable SPLUNK_BALLAST_SIZE_MIB.
See for example splunk_config.yaml on how to do it.
