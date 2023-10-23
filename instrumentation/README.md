# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux

The **Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux** Debian/RPM package
(`splunk-otel-auto-instrumentation`) installs Splunk OpenTelemetry Auto Instrumentation agents, the `libsplunk.so`
shared object library, and default/sample configuration files to automatically instrument applications and services to
capture and report distributed traces and metrics to the [Splunk OpenTelemetry Collector](
https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then on to [Splunk APM](
https://docs.splunk.com/Observability/apm/intro-to-apm.html).

The `splunk-otel-auto-instrumentation` deb/rpm package installs and supports configuration of the following Auto
Instrumentation agents:

- [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/get-started.html)
- [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/get-started.html)

For other languages or if the `splunk-otel-auto-instrumentation` deb/rpm package is not applicable for the target host
or applications/services, see [Instrument back-end applications to send spans to Splunk APM](
https://docs.splunk.com/Observability/en/gdi/get-data-in/application/application.html).

## Prerequisites/Requirements

- Check agent compatibility and requirements:
  - [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/java-otel-requirements.html)
  - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/nodejs-otel-requirements.html)
- [Install and configure](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html) the Splunk
  OpenTelemetry Collector.
- Debian or RPM based Linux distribution (amd64/x86_64 or arm64/aarch64).

## Installation

### Installer Script

The [Linux Installer Script](../docs/getting-started/linux-installer.md) is available to automate the installation and
configuration of the Collector and Auto Instrumentation for supported platforms. See
[Auto Instrumentation](../docs/getting-started/linux-installer.md#auto-instrumentation) for details.

### Manual

1. [Install and configure](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html) the Splunk
   OpenTelemetry Collector
2. [Install](../docs/getting-started/linux-manual.md#auto-instrumentation-debianrpm-packages) the
   `splunk-otel-auto-instrumentation` deb/rpm package
3. If Auto Instrumentation for Node.js is required, [install](
   ../docs/getting-started/linux-manual.md#auto-instrumentation-post-install-configuration) the provided
   `/usr/lib/splunk-instrumentatgion/splunk-otel-js.tgz` Node.js package with `npm`.
4. [Activate and configure](#activation-and-configuration) Auto Instrumentation with the supported methods and options.

## Activation and Configuration

The following methods are supported to manually activate and configure Auto Instrumentation after installation of the
`splunk-otel-auto-instrumentation` deb/rpm package (requires `root` privileges):

- [System-wide](#system-wide)
- [`Systemd` services only](#systemd-services-only)

> **Note**: To prevent conflicts and duplicate traces/metrics, only one method should be activated on the target system.

### System-wide

1. Add the path of the provided `/usr/lib/splunk-instrumentation/libsplunk.so` shared object library to the
   [`/etc/ld.so.preload`](https://man7.org/linux/man-pages/man8/ld.so.8.html#FILES) file to activate Auto
   Instrumentation for ***all*** supported processes on the system. For example:
   ```
   echo /usr/lib/splunk-instrumentation/libsplunk.so >> /etc/ld.so.preload
   ```
2. The default configuration files in the `/etc/splunk/zeroconfig` directory includes the required environment variables
   to activate the respective agents with the default options:
   - `/etc/splunk/zeroconfig/java.conf`:
     ```
     JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar
     ```
   - `/etc/splunk/zeroconfig/node.conf`:
     ```
     NODE_OPTIONS=-r @splunk/otel/instrument
     ```
   Configuration of the respective agents is supported by the adding/updating the following environment variables in
   each of these files (***any environment variable not in this list will be ignored***):
   - `OTEL_EXPORTER_OTLP_ENDPOINT`
   - `OTEL_RESOURCE_ATTRIBUTES`
   - `OTEL_SERVICE_NAME`
   - `SPLUNK_METRICS_ENABLED`
   - `SPLUNK_PROFILER_ENABLED`
   - `SPLUNK_PROFILER_MEMORY_ENABLED`

   Check the following for details about these environment variables and default values:
   - [Java](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
   - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/configuration/advanced-nodejs-otel-configuration.html)
3. Reboot the system or restart the applications/services for any changes to take effect. The `libsplunk.so` shared
   object library will then be preloaded for all subsequent processes and inject the environment variables from the
   `/etc/splunk/zeroconfig` configuration files for Java and Node.js processes.

### `Systemd` services only

> **Note**: The following steps utilize a sample `systemd` drop-in file to activate/configure the provided agents for
> all `systemd` services via default environment variables. `Systemd` supports many options, methods, and paths for
> configuring environment variables at the system level or for individual services, and are not limited to the steps
> below. Before making any changes, it is recommended to consult the documentation specific to your Linux distribution
> or service, and check the existing configurations of the system and individual services for potential conflicts or to
> override an environment variable for a particular service. For general details about `systemd`, see the
> [`systemd` man page](https://www.freedesktop.org/software/systemd/man/index.html).

1. Copy the provided sample `systemd` drop-in file
   `/usr/lib/splunk-instrumentation/examples/systemd/00-splunk-otel-auto-instrumentation.conf` to the host's `systemd`
   [drop-in configuration directory](https://www.freedesktop.org/software/systemd/man/systemd-system.conf.html) to
   activate Auto Instrumentation for ***all*** supported applications running as `systemd` services. For example:
   ```
   mkdir -p /usr/lib/systemd/system.conf.d/ && cp /usr/lib/splunk-instrumentation/examples/systemd/00-splunk-otel-auto-instrumentation.conf /usr/lib/systemd/system.conf.d/
   ```
   This file includes the required environment variables to activate the respective agents with the default options:
   - Java:
     ```
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
     ```
   - Node.js:
     ```
     DefaultEnvironment="NODE_OPTIONS=-r @splunk/otel/instrument"
     ```
2. To configure the activated agents, add/update [`DefaultEnvironment`](
   https://www.freedesktop.org/software/systemd/man/systemd-system.conf.html#DefaultEnvironment=) within the target file
   from the previous step for the desired environment variables. For example:
   ```
   cat <<EOH >> /usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf
   DefaultEnvironment="OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317"
   DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=deployment.environment=my_deployment_environment"
   DefaultEnvironment="OTEL_SERVICE_NAME=my_service_name"
   DefaultEnvironment="SPLUNK_METRICS_ENABLED=true"
   DefaultEnvironment="SPLUNK_PROFILER_ENABLED=true"
   DefaultEnvironment="SPLUNK_PROFILER_MEMORY_ENABLED=true"
   EOH
   ```
   Check the following for all supported environment variables and default values:
   - [Java](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
   - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/configuration/advanced-nodejs-otel-configuration.html)
3. Reboot the system, or run `systemctl daemon-reload` and then restart the applicable `systemd` services for any
   changes to take effect.
