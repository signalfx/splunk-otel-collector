> The official Splunk documentation for this page is [Install on Linux](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING#documentation.md).

# Linux Installer Script

For non-containerized Linux environments, an installer script is available. The
script deploys and configures:

- Splunk OpenTelemetry Collector for Linux (**x86_64/amd64 and aarch64/arm64 platforms only**)
- [SignalFx Smart Agent and collectd bundle](https://github.com/signalfx/signalfx-agent/releases) (**x86_64/amd64 platforms only**)
- [Fluentd (via the TD Agent)](https://www.fluentd.org/)
  - Optional, **enabled** by default for supported Linux distributions
  - See the [Fluentd Configuration](#fluentd-configuration) section for additional information, including how to skip installation.
- [Splunk OpenTelemetry Auto Instrumentation for Java](https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#linux-java-auto-instrumentation)
  - Optional, **disabled** by default
  - See the [Auto Instrumentation](#auto-instrumentation) section for additional information, including how to enable installation.

> IMPORTANT: systemd is required to use this script.

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2, 2023 (**Note:** Log collection with Fluentd not currently supported for Amazon Linux 2023.)
- CentOS / Red Hat / Oracle: 7, 8, 9
- Debian: 9, 10, 11 (**Note:** Log collection with Fluentd is not supported for Debian 9 aarch64.)
- SUSE: 12, 15 (**Note:** Only for Collector versions v0.34.0 or higher.  Log collection with Fluentd not currently supported.)
- Ubuntu: 16.04, 18.04, 20.04, 22.04 (**Note:** Log collection with Fluentd is not supported for Ubuntu 16.04 aarch64.)

## Getting Started

Download the latest release of the installer script and view all available
options by running the script with the `-h` flag.

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sh /tmp/splunk-otel-collector.sh -h
```

Run the following command on your host to begin installation with the default
options.  Replace these variables:

- `SPLUNK_REALM`: Which realm to send the data to (for example: `us0`)
- `SPLUNK_ACCESS_TOKEN`: Access token to authenticate requests

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM -- SPLUNK_ACCESS_TOKEN
```

After successful installation, run the following command to check the
`splunk-otel-collector` service status:
```sh
sudo systemctl status splunk-otel-collector
```

The `splunk-otel-collector` service logs and errors can be viewed in the
systemd journal:
```sh
sudo journalctl -u splunk-otel-collector
```

You can view the [source](../../internal/buildscripts/packaging/installer/install.sh)
for more details and available options.

## Advanced Configuration

### Additional Script Options

Additional configuration options supported by the script can be found by
running the script with the `-h` flag.

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sh /tmp/splunk-otel-collector.sh -h
```

One additional parameter that may need to be changed is `--memory` in order to
configure the memory allocation.

> By default, this variable is set to `512`. If you have allocated more memory
> to the Collector then you must increase this setting.

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --memory SPLUNK_MEMORY_TOTAL_MIB \
    -- SPLUNK_ACCESS_TOKEN
```

By default, apt/yum/zypper repo definition files will be created to download
the Collector and Fluentd deb/rpm packages from
[https://splunk.jfrog.io/splunk](https://splunk.jfrog.io/splunk) and
[https://packages.treasuredata.com](https://packages.treasuredata.com),
respectively.  To skip these steps and use pre-configured repos on the target
system that provide the `splunk-otel-collector` and `td-agent` deb/rpm
packages, specify the `--skip-collector-repo` and/or
`--skip-fluentd-repo` options.  For example:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --skip-collector-repo --skip-fluentd-repo \
    -- SPLUNK_ACCESS_TOKEN
```

### Collector Configuration

The Collector comes with a default configuration which can be found at
`/etc/otel/collector/agent_config.yaml`. This configuration can be
modified as needed. Possible configuration options can be found in the
`receivers`, `processors`, `exporters`, and `extensions` folders of either:

- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector)
- [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)

To use an existing Collector configuration file instead of the default, run
the installer script with the `--collector-config PATH_TO_CONFIG` option,
replacing `PATH_TO_CONFIG` with the absolute path to the desired configuration
file on the host.  For example:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --realm SPLUNK_REALM --collector-config /etc/my-config.yaml \
    -- SPLUNK_ACCESS_TOKEN
```

If the Collector configuration file includes references to custom environment
variables, these variables and values will need to be manually added to the
`/etc/otel/collector/splunk-otel-collector.conf` systemd environment file
after installation in order for the `splunk-otel-collector` systemd service to
expand these variables. For example, if the Collector configuration file
references `${MY_CUSTOM_VAR1}` and `${MY_CUSTOM_VAR2}`, add the following to
`/etc/otel/collector/splunk-otel-collector.conf`:
```
MY_CUSTOM_VAR1=my_custom_value1
MY_CUSTOM_VAR2=my_custom_value2
```
See [EnvironmentFile](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#EnvironmentFile=)
for more details about the systemd environment file.

If the Collector configuration file or
`/etc/otel/collector/splunk-otel-collector.conf` is modified after
installation, the Collector service needs to be restarted for the changes to
take effect:

```sh
sudo systemctl restart splunk-otel-collector
```

The `splunk-otel-collector` service logs and errors can be viewed in the
systemd journal:

```sh
sudo journalctl -u splunk-otel-collector
```

### Collector Upgrade

To upgrade the Collector, run the following commands on your system (requires
`root` privileges):
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
  - `dnf`
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
  preserved and the files from the new Collector package may be installed with
  a `.rpmnew` extension.

### Fluentd Configuration

> If log collection is not required, run the installer script with the
> `--without-fluentd` option to skip installation of Fluentd and the
> plugins/dependencies listed below.

By default, the Fluentd service will be installed and configured to forward log
events with the `@SPLUNK` label to the Collector (see below for how to add
custom Fluentd log sources), and the Collector will send these events to the
HEC ingest endpoint determined by the `--realm SPLUNK_REALM` option, e.g.
`https://ingest.SPLUNK_REALM.signalfx.com/v1/log`.

The following Fluentd plugins will also be installed:

- [capng_c](https://github.com/fluent-plugins-nursery/capng_c) for enabling [Linux capabilities](https://docs.fluentd.org/deployment/linux-capability)
- [fluent-plugin-systemd](https://github.com/fluent-plugin-systemd/fluent-plugin-systemd) for systemd journal log collection

Additionally, the following dependencies will be installed as prerequisites for
the Fluentd plugins:

- Debian-based systems:
  - `build-essential`
  - `libcap-ng0`
  - `libcap-ng-dev`
  - `pkg-config`

- RPM-based systems:
  - `Development Tools`
  - `libcap-ng`
  - `libcap-ng-devel`
  - `pkgconfig`

To configure the Collector to send log events to a custom HEC endpoint URL, you
can specify the following parameters for the installer script:

- `--hec-url URL`
- `--hec-token TOKEN`

The main Fluentd configuration file will be installed to
`/etc/otel/collector/fluentd/fluent.conf`. Custom Fluentd source config files
can be added to the `/etc/otel/collector/fluentd/conf.d` directory after 
installation. Please note:

- By default, Fluentd will be configured to collect log events from many
  popular services, like the systemd journal.  Check the `.conf` files in this
  directory for the default configuration of the included sources.  **Note:**
  The paths defined within these sources may need to be updated for the system
  or service.
- Any new source added to this directory should have a `.conf` extension and
  have the `@SPLUNK` label to automatically forward log events to the
  Collector.
- All files with a `.conf` extension in this directory will automatically be
  included when the Fluentd service starts/restarts.
- The Fluentd service runs as the `td-agent` user/group.  If adding/modifying
  any configuration file, ensure that the `td-agent` user/group has permissions
  to access the configuration file and the path(s) defined within.
- After any configuration modification, the Fluentd service needs to be
  restarted:
  ```sh
  sudo systemctl restart td-agent
  ```
- The Fluentd service logs and errors can be viewed in
  `/var/log/td-agent/td-agent.log`.
- See [https://docs.fluentd.org/configuration](
  https://docs.fluentd.org/configuration) for general Fluentd configuration
  details.

**Note:** If the `td-agent` package is upgraded after initial installation, [Linux
capabilities](https://docs.fluentd.org/deployment/linux-capability) may need
to be set for the new version by performing the following steps (only
applicable for `td-agent` versions 4.1 or newer):

1. Check for the enabled capabilities:
  ```sh
  sudo /opt/td-agent/bin/fluent-cap-ctl --get -f /opt/td-agent/bin/ruby
  ```
  The output should be:
  ```sh
  Capabilities in '/opt/td-agent/bin/ruby',
  Effective:   dac_override, dac_read_search
  Inheritable: dac_override, dac_read_search
  Permitted:   dac_override, dac_read_search
  ```

2. If the output from the previous command does not include `dac_override` and
   `dac_read_search` as shown above, run the following commands:
  ```sh
  sudo td-agent-gem install capng_c
  sudo /opt/td-agent/bin/fluent-cap-ctl --add "dac_override,dac_read_search" -f /opt/td-agent/bin/ruby
  sudo systemctl daemon-reload
  sudo systemctl restart td-agent
  ```

### Auto Instrumentation

>To see all supported options and defaults **before** installation, run:
>```sh
>curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
>sh /tmp/splunk-otel-collector.sh -h
>```

#### Installation

To install the Collector and the Splunk OpenTelemetry Auto Instrumentation for
Java packages, run the installer script with the `--with-instrumentation`
option:
```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --with-instrumentation --realm SPLUNK_REALM -- SPLUNK_ACCESS_TOKEN
```

To automatically define the optional `deployment.environment` resource
attribute at installation time, run the installer script with the
`--deployment-environment VALUE` option (replace `VALUE` with the desired
attribute value, for example, `prod`):
```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --with-instrumentation --deployment-environment VALUE --realm SPLUNK_REALM -- SPLUNK_ACCESS_TOKEN
```

**Note:** After successful installation, the Java application(s) on the host
need to be manually started/restarted for automatic instrumentation to take
effect.

#### Post-Install Configuration

- The `/etc/ld.so.preload` file will be automatically created/updated with the
  default path to the installed instrumentation library
  (`/usr/lib/splunk-instrumentation/libsplunk.so`).  If necessary, custom
  library paths can be manually added to this file.
- The `/usr/lib/splunk-instrumentation/instrumentation.conf` configuration file
  can be manually configured for resource attributes and other parameters.  By
  default, this file will contain the `java_agent_jar` parameter set to the
  path of the installed [Java Instrumentation Agent](
  https://github.com/signalfx/splunk-otel-java)
  (`/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`).  If the
  `--deployment-environment VALUE` installer script option was specified,
  the `deployment.environment=VALUE` resource attribute will be automatically
  added to this file.

See [Linux Java Auto Instrumentation](https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation#linux-java-auto-instrumentation)
for more details.

**Note:** After any configuration changes, the Java application(s) on the host
need to be manually started/restarted to source the updated values from the
configuration file.

#### Upgrade

To upgrade the Auto Instrumentation package, run the following commands on your
system (requires `root` privileges):
- Debian:
  ```sh
  sudo apt-get update
  sudo apt-get install --only-upgrade splunk-otel-auto-instrumentation
  ```
  **Note:** You may be prompted to keep or overwrite the configuration file at
  `/usr/lib/splunk-instrumentation/instrumentation.conf`.  Choosing to
  overwrite will revert this file to the default file provided by the new
  package.
- RPM:
  - `yum`
    ```sh
    sudo yum upgrade splunk-otel-auto-instrumentation
    ```
  - `dnf`
    ```sh
    sudo dnf upgrade splunk-otel-auto-instrumentation
    ```
  - `zypper`
    ```sh
    sudo zypper refresh
    sudo zypper update splunk-otel-auto-instrumentation
    ```

**Note:** After successful upgrade, the Java application(s) on the host need to
be manually started/restarted in order for the changes to take effect.

### Uninstall

If you wish to uninstall the Collector, Fluentd, and Auto Instrumentation
packages, you can run:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh > /tmp/splunk-otel-collector.sh && \
sudo sh /tmp/splunk-otel-collector.sh --uninstall
```

> Note that configuration files may be left on the filesystem.  On RPM-based
> systems, modified configuration files will be renamed with the `.rpmsave`
> extension and can be manually deleted if they are no longer needed.  On
> Debian-based systems, modified configuration files will persist and should
> be manually deleted before re-running the installer script if you do not
> intend on re-using these configuration files.
