> The official Splunk documentation for this page is [Install on Linux](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

# Linux Manual

The easiest and recommended way to get started is with the [Linux installer
script](./linux-installer.md). Alternatively, binary installation is available.
All Intel, AMD and ARM systemd-based operating systems are supported including
CentOS, Debian, Oracle, Red Hat and Ubuntu. This installation method is useful
for containerized environments or users wanting other common deployment
options.

The following deployment options are supported:

- [DEB and RPM Packages](#deb-and-rpm-packages)
- [Docker](#docker)
- [Binary](#binary)

## Getting Started

All installation methods offer [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which can be configured via environment variables. How these variables are
configured depends on the installation method leveraged.

### DEB and RPM Packages

#### Collector

If you prefer to install the collector without the [installer script
](./linux-installer.md), we provide Debian and RPM package repositories that
you can make use of with the following commands (requires `root` privileges).

**Note:** The [SignalFx Smart Agent and collectd bundle](
https://github.com/signalfx/signalfx-agent/releases) is only supported and
installed on x86_64/amd64 platforms.

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
- RPM with `yum`:
  ```sh
  yum install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the collector

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
- RPM with `dnf`:
  ```sh
  dnf install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the collector

  cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo
  [splunk-otel-collector]
  name=Splunk OpenTelemetry Collector Repository
  baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
  gpgcheck=1
  gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
  enabled=1
  EOH

  dnf install -y splunk-otel-collector
  ```
- RPM with `zypper`:
  ```sh
  zypper install -y libcap-progs  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the collector

  cat <<EOH > /etc/zypp/repos.d/splunk-otel-collector.repo
  [splunk-otel-collector]
  name=Splunk OpenTelemetry Collector Repository
  baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
  gpgcheck=1
  gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
  enabled=1
  EOH

  zypper install -y splunk-otel-collector
  ```
2. A default configuration file will be installed to
   `/etc/otel/collector/agent_config.yaml` if it does not already exist.
3. The `/etc/otel/collector/splunk-otel-collector.conf` environment file is
   required to start the `splunk-otel-collector` systemd service (**Note**: The
   service will automatically start if this file exists during
   install/upgrade).  A sample environment file will be installed to
   `/etc/otel/collector/splunk-otel-collector.conf.example` that includes the
   required environment variables for the default config.  To utilize this
   sample file, set the variables as appropriate and save the file as
   `/etc/otel/collector/splunk-otel-collector.conf`.
4. Start/Restart the service with:
   ```sh
   sudo systemctl restart splunk-otel-collector
   ```
   **Note:** The service must be restarted for any changes to the config file
   or environment file to take effect.

Run the following command to check the `splunk-otel-collector` service status:
```sh
sudo systemctl status splunk-otel-collector
```

The `splunk-otel-collector` service logs and errors can be viewed in the
systemd journal:
```sh
sudo journalctl -u splunk-otel-collector
```

#### Fluentd

If log collection is required, perform the following steps to install Fluentd
and forward collected log events to the Collector (requires `root` privileges):

1. Install, configure, and start the Collector as described in the previous
   section.  The Collector's default configuration file
   (`/etc/otel/collector/agent_config.yaml`) listens for log events on
   `127.0.0.1:8006` and sends them to the Splunk Observability Cloud.
1. Check [https://docs.fluentd.org/installation](
   https://docs.fluentd.org/installation) to install the `td-agent` package
   appropriate for the Linux distribution/version of the target system.
1. If necessary, check [https://docs.fluentd.org/deployment/linux-capability](
   https://docs.fluentd.org/deployment/linux-capability) to install the
   `capng_c` plugin and dependencies for enabling Linux capabilities, e.g.
   `cap_dac_read_search` and/or `cap_dac_override`.  Requires `td-agent`
   version 4.1 or newer.
1. If necessary, check
   [https://github.com/fluent-plugin-systemd/fluent-plugin-systemd](
   https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) to install
   the `fluent-plugin-systemd` plugin to collect log events from the systemd
   journal.
1. Configure Fluentd to collect log events and forward them to the Collector:
   - Option 1: Update the default config file at `/etc/td-agent/td-agent.conf`
     provided by the Fluentd package to collect the desired log events and
     [forward](https://docs.fluentd.org/output/forward) them to
     `127.0.0.1:8006`.
   - Option 2: The installed Collector package provides a custom Fluentd config
     file (`/etc/otel/collector/fluentd/fluent.conf`) to collect log events
     from many popular services (`/etc/otel/collector/fluentd/conf.d/*.conf`)
     and forwards them to `127.0.0.1:8006`. To utilize these files, copy the
     `/etc/otel/collector/fluentd/splunk-otel-collector.conf` systemd
     environment file to
     `/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf` in
     order to override the default config file path for the Fluentd service.
1. Ensure that the `td-agent` service user/group has permissions to access to
   the config file(s) from the previous step.
1. Apply the changes by running the following command to restart the Fluentd
   service:
   ```sh
   systemctl restart td-agent
   ```
   **Note**: The `td-agent` service must be restarted in order for any changes
   made to the Fluentd config files to take effect.
1. The Fluentd logs and errors can be viewed in the systemd journal:
   ```sh
   journalctl -u td-agent
   ```
1. See [https://docs.fluentd.org/configuration](
   https://docs.fluentd.org/configuration) for general Fluentd configuration
   details.

### Other

The remaining installation methods support environmental variables to configure
the Collector. The following environmental variables are required:

- `SPLUNK_REALM` (no default): Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN` (no default): Access token to authenticate requests

<details>
<summary>
Optional environment variables
</summary>

- `SPLUNK_CONFIG` (default = `/etc/otel/collector/gateway_config.yaml`): Which configuration to load.
- `SPLUNK_BALLAST_SIZE_MIB` (no default): How much memory to allocate to the ballast.
- `SPLUNK_MEMORY_TOTAL_MIB` (default = `512`): Total memory allocated to the Collector.

> `SPLUNK_MEMORY_TOTAL_MIB` automatically configures the ballast and memory limit.
> If `SPLUNK_BALLAST_SIZE_MIB` is also defined, it will override the value calculated
> by `SPLUNK_MEMORY_TOTAL_MIB`.
</details>

### Docker

Deploy the latest Docker image.

```bash
docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060 \
    -p 8888:8888 -p 9080:9080 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```

> Use of a SemVer tag over `latest` is highly recommended.

A docker-compose example is also available [here](../../examples/docker-compose). To use it:

- Go to the example directory
- Edit the `.env` appropriately for your environment
- Run `docker-compose up`

### Binary

Run as a binary on the local system:

```bash
git clone https://github.com/signalfx/splunk-otel-collector.git
cd splunk-otel-collector
make install-tools
make otelcol
SPLUNK_ACCESS_TOKEN=12345 SPLUNK_REALM=us0 ./bin/otelcol
```

## Advanced Configuration

### Command Line Arguments

After the binary command or Docker container command line arguments can be
specified.

> IMPORTANT: Command line arguments take precedence over environment variables.

For example in Docker:

```bash
docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 4317:4317 -p 6060:6060 \
    -p 8888:8888 -p 9080:9080 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest \
    --log-level=DEBUG
```

> Use `--help` to see all available CLI arguments.
> Use of a SemVer tag over `latest` is highly recommended.

### Custom Configuration

When changes to the default configuration YAML file are needed, create a
custom configuration file. Use environment variable `SPLUNK_CONFIG` or
command line argument `--config` to provide the path to this file.

Also, you can use environment variable `SPLUNK_CONFIG_YAML` to specify 
your custom configuration YAML at the command line. This is useful in
environments where access to the underlying file system is not readily
available. For example, in AWS Fargate you can store your custom configuration
YAML in a parameter in AWS Systems Manager Parameter Store, then in
your container definition specify `SPLUNK_CONFIG_YAML` to get the
configuration from the parameter.

> Command line arguments take precedence over environment variables. This
> applies to `--config` and `--mem-ballast-size-mib`. `SPLUNK_CONFIG` 
> takes precedence over `SPLUNK_CONFIG_YAML`

For example in Docker:

```bash
docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_REALM=us0 \
    -e SPLUNK_CONFIG=/etc/collector.yaml -p 13133:13133 -p 14250:14250 \
    -p 14268:14268 -p 4317:4317 -p 6060:6060 -p 8888:8888 \
    -p 9080:9080 -p 9411:9411 -p 9943:9943 \
    -v "${PWD}/collector.yaml":/etc/collector.yaml:ro \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```

> Use of a SemVer tag over `latest` is highly recommended.

In the case of Docker, a volume mount may be required to load the custom
configuration file, as shown above.

If the custom configuration includes a `memory_limiter` processor, then the
`ballast_size_mib` parameter should be the same as the
`SPLUNK_BALLAST_SIZE_MIB` environment variable. See
[gateway_config.yaml](../../cmd/otelcol/config/collector/gateway_config.yaml)
as an example.

The following example shows the `SPLUNK_CONFIG_YAML` environment variable:
```bash
CONFIG_YAML=$(cat <<-END
receivers:
   hostmetrics:
      collection_interval: 1s
      scrapers:
         cpu:
exporters:
   logging:
      logLevel: debug
service:
   pipelines:
      metrics:
         receivers: [hostmetrics]
         exporters: [logging]
END
)

docker run --rm \
    -e SPLUNK_CONFIG_YAML=${CONFIG_YAML} \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```
The configuration YAML above is for collecting and logging cpu
metrics. The YAML is assigned to parameter `CONFIG_YAML` for
convenience in the first command. In the subsequent `docker run`
command, parameter `CONFIG_YAML` is expanded and assigned to
environment variable `SPLUNK_CONFIG_YAML`. Note that YAML 
requires whitespace indentation to be maintained.
