# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Java

**Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Java** installs and enables the [Splunk
OpenTelemetry Auto Instrumentation Java Agent](https://github.com/signalfx/splunk-otel-java) to automatically instrument
Java applications running as `systemd` services on Linux, send the captured distributed traces to the locally running
[Splunk OpenTelemetry Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then
on to [Splunk APM](https://docs.splunk.com/Observability/apm/intro-to-apm.html).

> ***Note***: The `libsplunk.so` shared object file and related enablement/configuration processes are
> ***deprecated***. The `splunk-otel-auto-instrumentation` deb/rpm package is also ***deprecated*** and replaced by the
> `splunk-otel-systemd-auto-instrumentation` deb/rpm package.
> 
> [Installation](#installation) of `splunk-otel-systemd-auto-instrumentation` will automatically:
> - Uninstall the `splunk-otel-auto-instrumentation` deb/rpm package (if installed).
> - Remove `/usr/lib/splunk-instrumentation/libsplunk.so` from `/etc/ld.so.preload` (if it exists).
> - Enable the Java agent ***only*** for `systemd` services via the
>   `JAVA_TOOL_OPTIONS-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar` `systemd` environment
>   variable.
>
> For reference only, the documentation for `libsplunk.so` has been moved to [libsplunk.md](./libsplunk.md).

## Prerequisites/Requirements

- Check [Java agent compatibility and requirements](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/java/java-otel-requirements.html)
- Debian or RPM based Linux distribution with the `systemd` service manager
- Java application(s) running as `systemd` services
- [Install](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html) the Splunk OpenTelemetry
  Collector

For Linux distributions that do ***not*** support Debian/RPM packages or Java applications ***not*** running as
`systemd` services, see the [Instructions for app servers](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/instrumentation/java-servers-instructions.html)
for how to manually enable and configure the Java agent.

## Installation

The `splunk-otel-systemd-auto-instrumentation` deb/rpm package provides the following files to enable and configure the
Java agent for `systemd` services:
- [`/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf`](#systemd-environment-variables): Drop-in file with the
  following default environment variables for `systemd` services:
  - `JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`
  - `OTEL_JAVAAGENT_CONFIGURATION_FILE=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties`
- `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`: The
  [Splunk OpenTelemetry Auto Instrumentation Java Agent](https://github.com/signalfx/splunk-otel-java).
- [`/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties`](#configuration-file): The default system
  properties file to [configure the Splunk OpenTelemetry Auto Instrumentation Java Agent](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html).

Install these packages from [package repositories](#debian-and-rpm-repositories) or download them from
[GitHub Releases](#debian-and-rpm-packages).

After installation, restart the applicable `systemd` services or reboot the system to enable the Java agent with the
default configuration. Optionally, see [Configuration](#configuration) for details about configuring the agent.

### Debian and RPM Repositories

Set up the package repository and install the `splunk-otel-systemd-auto-instrumentation` package (requires `root`
privileges):
- Debian:
    ```shell
    curl -sSL https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg > /etc/apt/trusted.gpg.d/splunk.gpg && \
    echo 'deb https://splunk.jfrog.io/splunk/otel-collector-deb release main' > /etc/apt/sources.list.d/splunk-otel-collector.list && \
    apt-get update && \
    apt-get install -y splunk-otel-systemd-auto-instrumentation
    ```
- RPM with `yum`:
    ```shell
    cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo && yum install -y splunk-otel-systemd-auto-instrumentation
    [splunk-otel-collector]
    name=Splunk OpenTelemetry Collector Repository
    baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
    gpgcheck=1
    gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
    enabled=1
    EOH
    ```
- RPM with `dnf`:
    ```shell
    cat <<EOH > /etc/yum.repos.d/splunk-otel-collector.repo && dnf install -y splunk-otel-systemd-auto-instrumentation
    [splunk-otel-collector]
    name=Splunk OpenTelemetry Collector Repository
    baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
    gpgcheck=1
    gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
    enabled=1
    EOH
    ```
- RPM with `zypper`:
    ```shell
    cat <<EOH > /etc/zypp/repos.d/splunk-otel-collector.repo && zypper install -y splunk-otel-systemd-auto-instrumentation
    [splunk-otel-collector]
    name=Splunk OpenTelemetry Collector Repository
    baseurl=https://splunk.jfrog.io/splunk/otel-collector-rpm/release/\$basearch
    gpgcheck=1
    gpgkey=https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
    enabled=1
    EOH
    ```

### Debian and RPM Packages

Download and install the `splunk-otel-systemd-auto-instrumentation` package ***without*** setting up a
[package repository](#debian-and-rpm-repositories) (requires `root` privileges):
1. Download the appropriate `splunk-otel-systemd-auto-instrumentation` deb/rpm package for the target system
   (amd64/x86_64 or arm64/aarch64) from [GitHub Releases](https://github.com/signalfx/splunk-otel-collector/releases).
2. Download and install the public key for package signature verification:
   - Debian:
       ```shell
       curl -sSL https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg > /etc/apt/trusted.gpg.d/splunk.gpg
       ```
   - RPM:
       ```shell
       rpm --import https://splunk.jfrog.io/splunk/otel-collector-rpm/splunk-B3CD4420.pub
       ```
3. Install the package with the following command (replace `<path>` with the local path to the downloaded package):
   - Debian:
       ```shell
       $ dpkg -i <path>
       ```
   - RPM:
       ```shell
       $ rpm -ivh <path>
       ```

## Configuration

See the [Advanced Configuration Guide](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
for details about supported options and defaults for the Java agent. These options can be configured via
[environment variables](#systemd-environment-variables) or their corresponding [system properties](#configuration-file)
after installation.

> ### Configuration Priority
> The Java agent can consume configuration from one or more of the following sources (ordered from highest to lowest
> priority):
> 1. Java system properties (`-D` flags) passed directly to the agent. For example,
>      ```shell
>      JAVA_TOOL_OPTIONS="-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
>      ```
> 2. Environment variables
> 3. Configuration files
> 
> Before making any changes, check the configuration of the system or individual services for potential conflicts.

### Systemd environment variables

The default [`/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf`](./packaging/fpm/00-splunk-otel-javaagent.conf)
`systemd` drop-in file defines the following environment variables to enable the Java agent and sets the path to the
default system properties file for agent configuration, respectively:
- `JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`
- `OTEL_JAVAAGENT_CONFIGURATION_FILE=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties`

Any changes to this file will affect ***all*** `systemd` services, unless overriden by [higher-priority](
#configuration-priority) system or service configurations.

> ***Note***: `Systemd` supports many options/methods for configuring environment variables at the system level or for
> individual services, and are not limited to the examples below. Consult the documentation specific to your Linux
> distribution or service before making any changes. For general details about `systemd`, see the [`systemd` man page](
> https://www.freedesktop.org/software/systemd/man/index.html).

To add/modify [supported environment variables](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
defined in `/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf` (requires `root` privileges):
1. **Option A**: Update `DefaultEnvironment` within `/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf` for the
   desired environment variables (space-separated `"key=value"` pairs). For example:
     ```shell
     $ cat <<EOH > /etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service" "OTEL_JAVAAGENT_CONFIGURATION_FILE=/my/custom/javaagent.properties" "SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option B**: Create/Modify a higher-priority drop-in file for ***all*** services to add or override the environment
   variables defined in `/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For example:
     ```shell
     $ cat <<EOH > /etc/systemd/system.conf.d/01-my-custom-env-vars.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service" "OTEL_JAVAAGENT_CONFIGURATION_FILE=/my/custom/javaagent.properties" "SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option C**: Create/Modify a higher-priority drop-in file for a ***specific*** service to add or override the
   environment variables defined in `/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For example:
     ```shell
     $ cat <<EOH > /etc/systemd/system/my-service.d/01-my-custom-env-vars.conf
     [Service]
     Environment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service" "OTEL_JAVAAGENT_CONFIGURATION_FILE=/my/custom/javaagent.properties" "SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
2. After any configuration changes, reboot the system or run the following commands to restart the applicable services
   for the changes to take effect:
     ```shell
     $ systemctl daemon-reload
     $ systemctl restart <service-name>   # replace "<service-name>" and run for each applicable service
     ```

### Configuration file

The Java agent is configured by default (via the `OTEL_JAVAAGENT_CONFIGURATION_FILE` [`systemd` environment variable](
#systemd-environment-variables)) to consume system properties from the
[`/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties`](./packaging/fpm/splunk-otel-javaagent.properties)
configuration file.

Any changes to this file will affect ***all*** `systemd` services, unless overriden by [higher-priority](
#configuration-priority) system or service configurations.

To add/modify [supported system properties](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
in `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties` (requires `root` privileges):
1. Update `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties` for the desired system properties. For
   example:
     ```shell
     $ cat <<EOH > /usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties
     # This is a comment
     otel.service.name=my-service
     otel.resource.attributes=deployment.environment=my-environment
     splunk.metrics.enabled=true
     splunk.profiler.enabled=true
     splunk.profiler.memory.enabled=true
     EOH
     ```
2. After any configuration changes, reboot the system or run the following command to restart the applicable services
   for the changes to take effect:
     ```shell
     $ systemctl restart <service-name>   # replace "<service-name>" and run for each applicable service
     ```

## Uninstall

1. If necessary, back up `/etc/systemd/system.conf.d/00-splunk-otel-javaagent.conf` and
   `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.properties`.
2. Run the following command to uninstall the `splunk-otel-systemd-auto-instrumentation` package (requires `root`
   privileges):
   - Debian:
       ```shell
       dpkg -P splunk-otel-systemd-auto-instrumentation
       ```
   - RPM:
       ```shell
       rpm -e splunk-otel-systemd-auto-instrumentation
       ```
3. Reboot the system or run the following command to restart the applicable services:
     ```shell
     $ systemctl restart <service-name>  # replace "<service-name>" and run for each applicable service
     ```
