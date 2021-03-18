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

<details>
<summary>
Optional environment variables
</summary>

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/splunk_config_linux.yaml`): Which configuration to load.
- `SPLUNK_BALLAST_SIZE_MIB` (no default): How much memory to allocate to the ballast.
- `SPLUNK_MEMORY_TOTAL_MIB` (default = `512`): Total memory allocated to the Collector.

> `SPLUNK_MEMORY_TOTAL_MIB` automatically configures the ballast and memory limit.
> If `SPLUNK_BALLAST_SIZE_MIB` is also defined, it will override the value calculated
> by `SPLUNK_MEMORY_TOTAL_MIB`.
</details>

### Docker

Deploy the latest Docker image.

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```

> Use of a SemVer tag over `latest` is highly recommended.

A docker-compose example is also available [here](../../examples/docker-compose). To use it:

- Go to the example directory
- Edit the `.env` appropriately for your environment
- Run `docker-compose up`

### Standalone

Run as a binary on the local system:

```bash
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector
$ make install-tools
$ make otelcol
$ SPLUNK_ACCESS_TOKEN=12345 SPLUNK_REALM=us0 ./bin/otelcol
```

## Advanced Configuration

### Command Line Arguments

After the binary command or Docker container command line arguments can be
specified.

> IMPORTANT: Command line arguments take precedence over environment variables.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest \
        --log-level=DEBUG
```

> Use `--help` to see all available CLI arguments.
> Use of a SemVer tag over `latest` is highly recommended.

### Custom Configuration

When changes to the default configuration file is needed, a custom
configuration file can specified and used instead. Use the `SPLUNK_CONFIG`
environment variable or the `--config` command line argument to provide a
custom configuration.

> Command line arguments take precedence over environment variables. This
> applies to `--config` and `--mem-ballast-size-mib`.

For example in Docker:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -e SPLUNK_CONFIG=/etc/collector.yaml -p 13133:13133 -p 14250:14250 \
    -p 14268:14268 -p 4317:4317 -p 6060:6060 -p 7276:7276 -p 8888:8888 \
    -p 9411:9411 -p 9943:9943 -v "${PWD}/collector.yaml":/etc/collector.yaml:ro \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```

> Use of a SemVer tag over `latest` is highly recommended.

In the case of Docker, a volume mount may be required to load custom
configuration as shown above.

If the custom configuration includes a `memory_limiter` processor then the
`ballast_size_mib` parameter should be the same as the
`SPLUNK_BALLAST_SIZE_MIB` environment variable. See
[splunk_config_linux.yaml](cmd/otelcol/config/collector/splunk_config_linux.yaml)
as an example.
