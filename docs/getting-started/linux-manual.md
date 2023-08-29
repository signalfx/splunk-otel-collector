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
- [Tar archive](#tar)

## Getting Started

All installation methods offer [default
configurations](https://github.com/signalfx/splunk-otel-collector/blob/main/cmd/otelcol/config/collector)
which can be configured via environment variables. How these variables are
configured depends on the installation method leveraged.

### Permissions

The installers below rely on using the setcap command (installed with libcap2)
to setup up granular permissions so the Collector doesn't have to be run with
root permissions to collect telemetry data. When the setcap command is
available on a Linux host, the Collector will be installed with the
[capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html) used
in the setcap command below. These capabilities are what are recommended to
allow the Collector to run with the least permissions needed regardless of
which user runs the Collector.
```sh
setcap CAP_SYS_PTRACE,CAP_DAC_READ_SEARCH=+eip /usr/bin/otelcol
```
**Note:** These permissions work well in most cases, however, a Collector
could need higher or more custom permission depending on the systems you are
monitoring.

If a user wanted to set custom permissions after the Collector was installed,
they could use the setcap command to do so.
```sh
setcap {CUSTOM_CAPABILITIES}=+eip /usr/bin/otelcol
```

### DEB and RPM Packages

#### Collector Package Repositories

If you prefer to install the Collector without the [installer script
](./linux-installer.md), we provide Debian and RPM package repositories that
you can make use of with the following commands (requires `root` privileges).

**Note:** The [SignalFx Smart Agent and collectd bundle](
https://github.com/signalfx/signalfx-agent/releases) is only supported and
installed on x86_64/amd64 platforms.

> IMPORTANT: `systemctl` is a requirement to run the Collector as a service.
> Otherwise, manually running the Collector is required.

1. Set up the package repository and install the Collector package:
    - Debian:
        ```sh
        curl -sSL https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg > /etc/apt/trusted.gpg.d/splunk.gpg
        echo 'deb https://splunk.jfrog.io/splunk/otel-collector-deb release main' > /etc/apt/sources.list.d/splunk-otel-collector.list
        apt-get update
        apt-get install -y splunk-otel-collector

        # Optional: install Splunk OpenTelemetry Auto Instrumentation for Java
        apt-get install -y splunk-otel-auto-instrumentation
        ```
    - RPM with `yum`:
        ```sh
        yum install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector

        cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo
        [splunk-otel-collector]
        name=Splunk OpenTelemetry Collector Repository
        baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
        gpgcheck=1
        gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
        enabled=1
        EOH

        yum install -y splunk-otel-collector

        # Optional: install Splunk OpenTelemetry Auto Instrumentation for Java
        yum install -y splunk-otel-auto-instrumentation
        ```
    - RPM with `dnf`:
        ```sh
        dnf install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector

        cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo
        [splunk-otel-collector]
        name=Splunk OpenTelemetry Collector Repository
        baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
        gpgcheck=1
        gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
        enabled=1
        EOH

        dnf install -y splunk-otel-collector

        # Optional: install Splunk OpenTelemetry Auto Instrumentation for Java
        dnf install -y splunk-otel-auto-instrumentation
        ```
    - RPM with `zypper`:
        ```sh
        zypper install -y libcap-progs  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector

        cat <<EOH > /etc/zypp/repos.d/splunk-otel-collector.repo
        [splunk-otel-collector]
        name=Splunk OpenTelemetry Collector Repository
        baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
        gpgcheck=1
        gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
        enabled=1
        EOH

        zypper install -y splunk-otel-collector

        # Optional: install Splunk OpenTelemetry Auto Instrumentation for Java
        zypper install -y splunk-otel-auto-instrumentation
        ```
1. See the [Collector Debian/RPM Post-Install
   Configuration](#collector-debianrpm-post-install-configuration) section.
1. If the optional Splunk OpenTelemetry Auto Instrumentation for Java package
   was installed, see the [Auto Instrumentation Post-Install
   Configuration](#auto-instrumentation-post-install-configuration) section.
1. If log collection is required, see the [Fluentd](#fluentd) section.
1. To upgrade the Collector, run the following commands:
    - Debian:
      ```sh
      sudo apt-get update
      sudo apt-get install --only-upgrade splunk-otel-collector
      ```
      **Note:** If the default configuration files in `/etc/otel/collector` have
      been modified after initial installation, you may be prompted to keep the
      existing files or overwrite the files from the new Collector package.
    - RPM:
        - `yum`
          ```sh
          sudo yum upgrade splunk-otel-collector
          ```
        - `dnf`:
          ```sh
          sudo dnf upgrade splunk-otel-collector
          ```
        - `zypper`
          ```sh
          sudo zypper refresh
          sudo zypper update splunk-otel-collector
          ```
      **Note:** If the default configuration files in `/etc/otel/collector` have
      been modified after initial installation, the existing files will be
      preserved and the files from the new Collector package may be installed
      with a `.rpmnew` extension.
1. To upgrade the Auto Instrumentation package, run the following commands:
   - Debian:
     ```sh
     sudo apt-get update
     sudo apt-get install --only-upgrade splunk-otel-auto-instrumentation
     ```
     **Note:** You may be prompted to keep or overwrite the configuration file
     at `/usr/lib/splunk-instrumentation/instrumentation.conf`.  Choosing to
     overwrite will revert this file to the default file provided by the new
     package.
   - RPM:
     - `yum`
       ```sh
       sudo yum upgrade splunk-otel-auto-instrumentation
       ```
     - `dnf`:
       ```sh
       sudo dnf upgrade splunk-otel-auto-instrumentation
       ```
     - `zypper`
       ```sh
       sudo zypper refresh
       sudo zypper update splunk-otel-instrumentation
       ```

#### Collector Debian/RPM Packages

If you prefer to install the Collector without the [installer script
](./linux-installer.md) or the [Debian/RPM Repositories
](#collector-debianrpm-packages) in the previous section, you can download the
individual Debian or RPM package from the [GitHub Releases](
https://github.com/signalfx/splunk-otel-collector/releases) page and install it
with the following commands (requires `root` privileges).

**Note:** The [SignalFx Smart Agent and collectd bundle](
https://github.com/signalfx/signalfx-agent/releases) is only supported and
installed on x86_64/amd64 platforms.

> IMPORTANT: `systemctl` is a requirement to run the Collector as a service.
> Otherwise, manually running the Collector is required.

1. Download the appropriate `splunk-otel-collector` Debian or RPM package for
   the target system from the [GitHub Releases](
   https://github.com/signalfx/splunk-otel-collector/releases) page.
1. Run the following commands to install the `setcap` dependency and the
   Collector package (replace `<path to splunk-otel-collector deb/rpm>` with
   the local path to the downloaded Collector package):
    - Debian:
        ```sh
        apt-get update && apt-get install -y libcap2-bin  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector
        dpkg -i <path to splunk-otel-collector deb>
        ```
    - RPM with `yum`:
        ```sh
        yum install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector
        rpm -ivh <path to splunk-otel-collector rpm>
        ```
    - RPM with `dnf`:
        ```sh
        dnf install -y libcap  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector
        rpm -ivh <path to splunk-otel-collector rpm>
        ```
    - RPM with `zypper`:
        ```sh
        zypper install -y libcap-progs  # Required for enabling cap_dac_read_search and cap_sys_ptrace capabilities on the Collector
        rpm -ivh <path to splunk-otel-collector rpm>
        ```
1. See the [Collector Debian/RPM Post-Install
   Configuration](#collector-debianrpm-post-install-configuration) section.
1. If log collection is required, see the [Fluentd](#fluentd) section.
1. To upgrade the Collector package, download the appropriate
   `splunk-otel-collector` Debian or RPM package for the target system from
   the [GitHub Releases](
   https://github.com/signalfx/splunk-otel-collector/releases) page and run the
   following commands (replace `<path to splunk-otel-collector deb/rpm>` with
   the local path to the downloaded Collector package):
   - Debian:
     ```sh
     sudo dpkg -i <path to splunk-otel-collector deb>
     ```
     **Note:** If the default configuration files in `/etc/otel/collector` have
     been modified after initial installation, you may be prompted to keep the
     existing files or overwrite the files from the new Collector package.
   - RPM
     ```sh
     sudo rpm -Uvh <path to splunk-otel-collector rpm>
     ```
     **Note:** If the default configuration files in `/etc/otel/collector` have
     been modified after initial installation, the existing files will be
     preserved and the files from the new Collector package may be installed
     with a `.rpmnew` extension.

#### Collector Debian/RPM Post-Install Configuration

1. A default configuration file will be installed to
   `/etc/otel/collector/agent_config.yaml` if it does not already exist.
1. The `/etc/otel/collector/splunk-otel-collector.conf` environment file is
   required to start the `splunk-otel-collector` systemd service (**Note**: The
   service will automatically start if this file exists during
   install/upgrade).  A sample environment file will be installed to
   `/etc/otel/collector/splunk-otel-collector.conf.example` that includes the
   required environment variables for the default config.  To utilize this
   sample file, set the variables as appropriate and save the file as
   `/etc/otel/collector/splunk-otel-collector.conf`.
1. Start/Restart the service with:
   ```sh
   sudo systemctl restart splunk-otel-collector
   ```
   **Note:** The service must be restarted for any changes to the config file
   or environment file to take effect.

1. Run the following command to check the `splunk-otel-collector` service
   status:
   ```sh
   sudo systemctl status splunk-otel-collector
   ```
1. The `splunk-otel-collector` service logs and errors can be viewed in the
   systemd journal:
   ```sh
   sudo journalctl -u splunk-otel-collector
   ```

#### Auto Instrumentation Debian/RPM Packages

If you prefer to install the Auto Instrumentation package without the
[installer script](./linux-installer.md) or the [Debian/RPM Repositories
](#collector-debianrpm-packages) in the previous section, you can download the
individual Debian or RPM package from the [GitHub Releases](
https://github.com/signalfx/splunk-otel-collector/releases) page
and install it with the following commands (requires `root` privileges).

1. Download the appropriate `splunk-otel-auto-instrumentation` Debian or RPM
   package for the target system from the [GitHub Releases](
   https://github.com/signalfx/splunk-otel-collector/releases) page.
1. Run the following commands to install the Auto Instrumentation package
   (replace `<path to splunk-otel-auto-instrumentation deb/rpm>` with
   the local path to the downloaded Auto Instrumentation package):
    - Debian:
        ```sh
        dpkg -i <path to splunk-otel-auto-instrumentation deb>
        ```
    - RPM:
        ```sh
        rpm -ivh <path to splunk-otel-auto-instrumentation rpm>
        ```
1. See the [Auto Instrumentation Debian/RPM Post-Install
   Configuration](#auto-instrumentation-post-install-configuration) section.
1. To upgrade the Auto Instrumentation package, download the appropriate
   `splunk-auto-auto-instrumentation` Debian or RPM
   package for the target system from the [GitHub Releases](
   https://github.com/signalfx/splunk-otel-collector/releases) page and run the
   following commands (replace `<path to splunk-otel-auto-instrumentation
   deb/rpm>` with the local path to the downloaded Auto Instrumentation
   package):
    - Debian:
      ```sh
      sudo dpkg -i <path to splunk-otel-auto-instrumentation deb>
      ```
      **Note:** You may be prompted to keep or overwrite the configuration file
      at `/usr/lib/splunk-instrumentation/instrumentation.conf`.  Choosing to
      overwrite will revert this file to the default file provided by the new
      package.
    - RPM
      ```sh
      sudo rpm -Uvh <path to splunk-otel-auto-instrumentation rpm>
      ```

#### Auto Instrumentation Post-Install Configuration

Choose one of the following methods to activate and configure Splunk
OpenTelemetry Auto Instrumentation ***globally*** with either the provided
`libsplunk.so` shared object library or sample `systemd` drop-in files. To
activate and configure auto instrumentation for individual services or
applications, see
[Instrument back-end applications to send spans to Splunk APM](
https://docs.splunk.com/Observability/gdi/get-data-in/application/application.html).

1. Preload method
   - The `/usr/lib/splunk-instrumentation/libsplunk.so` shared object library
     can be added to the [`/etc/ld.so.preload`](
     https://man7.org/linux/man-pages/man8/ld.so.8.html#FILES) file to activate
     auto instrumentation for ***all*** supported processes.
   - The `/usr/lib/splunk-instrumentation/instrumentation.conf` configuration
     file can be configured for resource attributes and other supported
     parameters. By default, this file will contain the `java_agent_jar`
     parameter set to the path of the installed [Java Instrumentation Agent](
     https://github.com/signalfx/splunk-otel-java)
     (`/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`).
   - See [Linux Java Auto Instrumentation](../../instrumentation/libsplunk.md)
     for more details.

2. `Systemd` method
   - The sample `systemd` drop-in files in the
     `/usr/lib/splunk-instrumentation/examples/systemd/` directory can be
     configured and copied to the host's [`systemd` configuration
     directory](
     https://www.freedesktop.org/software/systemd/man/systemd-system.conf.html),
     for example `/usr/lib/systemd/system.conf.d/`, to activate and
     configure auto instrumentation for ***all*** supported applications
     running as `systemd` services.
   - See [Splunk OpenTelemetry Zero Configuration Auto Instrumentation for
     Systemd](../../instrumentation/systemd.md) for more details.

**Note:** After installation/upgrade or any configuration changes, reboot the
system or restart the application(s) on the host for automatic instrumentation
to take effect and/or to source the updated values.

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
1. The Fluentd service logs and errors can be viewed in
   `/var/log/td-agent/td-agent.log`.
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

### Tar

We offer for convenience a tar.gz archive of the distribution.

To use the archive:

1. Unarchive it to a directory of your choice on the target system.
```bash
tar xzf splunk-otel-collector_<version>_<arch>.tar.gz
```

2. On amd64 systems, go into the unarchived `agent-bundle` directory and run `bin/patch-interpreter $(pwd)`. 
This ensures that the binaries in the bundle have the right loader set on them since your host's loader may not be compatible.

The tar archive contains the default agent and gateway configuration files.
Both refer to environment variables described in the [Other](#Other) section above.
If you are running the Collector from a non-default location, the Smart Agent receiver and agent configuration file require that you set two environment variables currently used in the Smart Agent extension:
- `SPLUNK_BUNDLE_DIR` (`/usr/lib/splunk-otel-collector/agent-bundle` default): The path to the Smart Agent bundle, e.g. `/opt/my/environment/splunk-otel-collector/agent-bundle`
- `SPLUNK_COLLECTD_DIR` (`/usr/lib/splunk-otel-collector/agent-bundle/run/collectd` default): The path to the collectd config directory for the Smart Agent, e.g. `/opt/my/environment/splunk-otel-collector/agent-bundle/run/collectd`

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
    --set=service.telemetry.logs.level=debug
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
      verbosity: detailed
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
