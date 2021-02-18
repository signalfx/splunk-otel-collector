# Linux Standalone

The easiest and recommended way to get started is with the [Linux installer
script](./linux-installer.md). Alternatively, standalone installation is available.
All Intel, AMD and ARM systemd-based operating systems are supported including
CentOS, Debian, Oracle, Red Hat and Ubuntu. This installation method is useful
for containerized environments or users wanting other common deployment
options.

The following deployment options are supported:

- [DEB and RPM Packages](#deb-and-rpm-packages)
- [Docker](#docker)
- [Standalone](#standalone)

## Getting Started

All installation methods offer [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which can be configured via environment variables. How these variables are
configured depends on the installation method leveraged.

<details>
<summary>
Accessible endpoints
</summary>
With the Collector configured, the following endpoints are accessible:

- `http(s)://<collectorFQDN>:13133/` Health endpoint useful for load balancer monitoring
- `http(s)://<collectorFQDN>:[14250|14268]` Jaeger [gRPC|Thrift HTTP] receiver
- `http(s)://<collectorFQDN>:55678` OpenCensus gRPC and HTTP receiver
- `http(s)://localhost:55679/debug/[tracez|pipelinez]` zPages monitoring
- `http(s)://<collectorFQDN>:55680` OpenTelemetry gRPC receiver
- `http(s)://<collectorFQDN>:6060` HTTP Forwarder used to receive Smart Agent `apiUrl` data
- `http(s)://<collectorFQDN>:7276` SignalFx Infrastructure Monitoring gRPC receiver
- `http(s)://localhost:8888/metrics` Prometheus metrics for the Collector
- `http(s)://<collectorFQDN>:9411/api/[v1|v2]/spans` Zipkin JSON (can be set to proto)receiver
- `http(s)://<collectorFQDN>:9943/v2/trace` SignalFx APM receiver
</details>

### DEB and RPM Packages

If you prefer to install the collector without the [installer script
](./linux-installer.md), we provide Debian and RPM package repositories that
you can make use of with the following commands (requires `root` privileges).

> IMPORTANT: `systemctl` is a requirement to run the collector as a service.
> Otherwise, manually running the collector is required.

1. Set up the package repository and install the collector package:
- Debian:
```sh
curl -sSL https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg > /etc/apt/trusted.gpg.d/splunk.gpg
echo 'deb https://splunk.jfrog.io/splunk/otel-collector-deb release main' > /etc/apt/sources.list.d/splunk-otel-collector.list
apt-get update
apt-get install -y splunk-otel-collector
```
- RPM:
```sh
cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo
[splunk-otel-collector]
name=Splunk OpenTelemetry Collector Repository
baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
gpgcheck=1
gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
enabled=1
EOH

yum install -y splunk-otel-collector
```
2. A default configuration file will be installed to
   `/etc/otel/collector/splunk_config_linux.yaml` if it does not already exist.
3. The `/etc/otel/collector/splunk_env` environment file is required to start
   the `splunk-otel-collector` systemd service.  A sample environment file will
   be installed to `/etc/otel/collector/splunk_env.example` that includes the
   required environment variables for the default config.  To utilize this
   sample file, set the variables as appropriate and save the file as
   `/etc/otel/collector/splunk_env`.
4. Start/Restart the service with
   `sudo systemctl restart splunk-otel-collector.service`.

### Other

The remaining installation methods support environmental variables to configure
the Collector. The following environmental variables are required:

- `SPLUNK_REALM` (no default): Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN` (no default): Access token to authenticate requests
- `SPLUNK_MEMORY_TOTAL_MIB` (no default): Total memory allocated to the Collector.

<details>
<summary>
Optional environment variables
</summary>

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk_config_linux.yaml`): Which configuration to load.
- `SPLUNK_BALLAST_SIZE_MIB` (no default): How much memory to allocate to the ballast.
- For Linux systems:
  - `SPLUNK_MEMORY_LIMIT_PERCENTAGE` (default = `90`): Maximum total memory to be allocated by the process heap.
  - `SPLUNK_MEMORY_SPIKE_PERCENTAGE` (default = `20`): Maximum spike between the measurements of memory usage.
- For non-Linux systems:
  - `SPLUNK_MEMORY_LIMIT_MIB` (no default): Maximum total memory to be allocated by the process heap.
  - `SPLUNK_MEMORY_SPIKE_MIB` (no default): Maximum spike between the measurements of memory usage.

> `SPLUNK_MEMORY_TOTAL_MIB` automatically configures the ballast, memory limit,
> and memory spike. If the optional environment variables are defined, they
> will override the value calculated from `SPLUNK_MEMORY_TOTAL_MIB`.
</details>

### Docker

Deploy from a Docker container. Replace `0.1.0` with the latest stable version number:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_MEMORY_TOTAL_MIB=1024 \
    -e SPLUNK_REALM=us0 -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 55678-55680:55678-55680 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:0.1.0
```

### Standalone

Run as a binary on the local system:

```bash
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector
$ make install-tools
$ make otelcol
$ SPLUNK_REALM=us0 SPLUNK_ACCESS_TOKEN=12345 SPLUNK_MEMORY_TOTAL_MIB=1024 \
    ./bin/otelcol
```

## Advanced Configuration

### Command Line Arguments

After the binary command or Docker container command line arguments can be
specified.

> IMPORTANT: Command line arguments take precedence over environment variables.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_MEMORY_TOTAL_MIB=1024 \
    -e SPLUNK_REALM=us0 -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 55678-55680:55678-55680 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:0.1.0 \
        --log-level=DEBUG
```

> Use `--help` to see all available CLI arguments.

### Custom Configuration

When changes to the default configuration file is needed, a custom
configuration file can specified and used instead. Use the `SPLUNK_CONFIG`
environment variable or the `--config` command line argument to provide a
custom configuration.

> Command line arguments take precedence over environment variables. This
> applies to `--config` and `--mem-ballast-size-mib`.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_MEMORY_TOTAL_MIB=1024 \
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
[splunk_config_linux.yaml](cmd/otelcol/config/collector/splunk_config_linux.yaml)
as an example.
