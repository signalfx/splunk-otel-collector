# Splunk distribution of OpenTelemetry Collector

The Splunk distribution of [OpenTelemetry
Collector](https://github.com/open-telemetry/opentelemetry-collector) provides
a binary that can be deployed as a standalone service (also known as a gateway)
that can receive, process and export trace, metric and log data. The Collector
is an optional component that currently supports:

- [Splunk APM](https://www.splunk.com/en_us/software/splunk-apm.html) via the
  [`sapm`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/sapmexporter).
  More information available
  [here](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-opentelemetry-collector.html).
- [Splunk Infrastructure
  Monitoring](https://www.splunk.com/en_us/software/infrastructure-monitoring.html)
  via the [`signalfx`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/signalfxexporter).
  More information available
  [here](https://docs.signalfx.com/en/latest/otel/imm-otel-collector.html).
- [Splunk Cloud](https://www.splunk.com/en_us/software/splunk-cloud.html) or
  [Splunk
  Enterprise](https://www.splunk.com/en_us/software/splunk-enterprise.html) via
  the [`splunk_hec`
  exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/splunkhecexporter).

The Collector is supported on and packaged for a variety of platforms including:

- Kubernetes: Helm and YAML
- Linux: All Intel, AMD and ARM systemd-based operating systems are supported
  including CentOS, Debian, Oracle, Red Hat and Ubuntu. DEB and RPM packages
  are also provided.
- Windows: EXE and MSI

> :construction: This project is currently in **BETA**.

## Getting Started

This distribution comes with [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which require the following environment variables:

- `SPLUNK_REALM` (no default): Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN` (no default): Access token to authenticate requests
- `SPLUNK_BALLAST_SIZE_MIB` (no default): How much memory to allocate to the ballast. This should be set to 1/3 to 1/2 of configured memory.

In addition, the following environment variables are optional:

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk-config_linux.yaml`): Which configuration to load.
- `SPLUNK_MEMORY_LIMIT_PERCENTAGE` (default = `90`): Maximum total memory to be allocated by the process heap.
- `SPLUNK_MEMORY_SPIKE_PERCENTAGE` (default = `20`): Maximum spike between the measurements of memory usage.

When running on a non-linux system, the following environment variables are required:

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk-config_non_linux.yaml`): Configuration to load.
- `SPLUNK_MEMORY_LIMIT_MIB` (no default): Maximum total memory to be allocated by the process heap.
- `SPLUNK_MEMORY_SPIKE_MIB` (no default): Maximum spike between the measurements of memory usage.

The following sections describe how to deploy the Collector in supported environments.

### Docker

Deploy from a Docker container. Replace `0.1.0` with the latest stable version number:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_BALLAST_SIZE_MIB=683 \
    -e SPLUNK_REALM=us0 -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 55678-55680:55678-55680 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:0.1.0
```

### Kubernetes

To deploy in Kubernetes, create a configuration file that defines a ConfigMap,
Service, and Deployment for the cluster. For more information about creating a
configuration file, see the example
[signalfx-k8s.yaml](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/exporter/sapmexporter/examples/signalfx-k8s.yaml)
file on GitHub.

### Standalone

```bash
$ make otelcol
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_BALLAST_SIZE_MIB=683 \
    ./bin/otelcol
```

### Linux Installer Script

For non-containerized Linux environments, a convenience script is available for
installing the collector package and [TD Agent
(Fluentd)](https://www.fluentd.org/).

Run the following command on your host (replace `SPLUNK_REALM`,
`SPLUNK_BALLAST_SIZE`, and `SPLUNK_ACCESS_TOKEN` appropriately for your
environment):

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh;
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --ballast SPLUNK_BALLAST_SIZE \
    -- SPLUNK_ACCESS_TOKEN
```

You can view the [source](internal/buildscripts/packaging/installer/install.sh)
for more details and other options.

Currently, only the following Linux distributions and versions are supported:

- Amazon Linux: 2
- CentOS / Red Hat / Oracle: 7, 8
- Debian: 8, 9, 10
- Ubuntu: 16.04, 18.04, 20.04

## Sizing

The OpenTelemetry Collector can be scaled up or out as needed. Sizing is based
on the amount of data per data source and requires 1 CPU core per:

- Traces: 10,000 spans per second
- Metrics: 20,000 data points per second

> If a Collector handles both trace and metric data then both must be accounted
> for when sizing. For example, 5K spans per second plus 10K data points per
> second would require 1 CPU core.

The recommendation is to use a ratio of 1:2 for CPU:memory and to allocate at
least a CPU core per Collector. Multiple Collectors can deployed behind a
simple round-robin load balancer. Each Collector runs independently, so scale
increases linearly with the number of Collectors you deploy.

> The Collector does not persist data to disk so no disk space is required.

## Advanced Configuration

### Command Line Arguments

Following the binary command or Docker container command line arguments can be
specified. Command line arguments take priority over environment variables.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_BALLAST_SIZE_MIB=683 \
    -e SPLUNK_REALM=us0 -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 55678-55680:55678-55680 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:0.1.0 \
        --log-level=DEBUG
```

> Use `--help` to see all available CLI arguments.

### Custom Configuration

In addition to using the default configuration, a custom configuration can also
be provided. This can be done via the `SPLUNK_CONFIG` environment variable as
well as the `--config` command line argument.

> Command line arguments take precedence over environment variables. This
> applies to `--config` and `--mem-ballast-size-mib`.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_BALLAST_SIZE_MIB=683 \
    -e SPLUNK_REALM=us0 -e SPLUNK_CONFIG=/etc/collector.yaml -p 13133:13133 -p 14250:14250 \
    -p 14268:14268 -p 55678-55680:55678-55680 -p 6060:6060 -p 7276:7276 -p 8888:8888 \
    -p 9411:9411 -p 9943:9943 -v collector.yaml:/etc/collector.yaml:ro \
    --name otelcol quay.io/signalfx/splunk-otel-collector:0.1.0
```

In the case of Docker, a volume mount may be required to load custom
configuration as shown above.

If the custom configuration includes a `memory_limiter` processor then the
`ballast_size_mib` parameter should be the same as the
`SPLUNK_BALLAST_SIZE_MIB` environment variable. See
[splunk_config.yaml](cmd/otelcol/config/collector/splunk_config.yaml) as an
example.

## Monitoring

The default configuration automatically scrapes the Collector's own metrics and
sends the data using the `signalfx` exporter. A built-in dashboard provides
information about the health and status of Collector instances.

## Troubleshooting

See the [Collector troubleshooting
documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/troubleshooting.md).
