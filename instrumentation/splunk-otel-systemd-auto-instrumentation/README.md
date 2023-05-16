# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Systemd

**Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Systemd** installs and configures Splunk
OpenTelemetry Auto Instrumentation agent(s) to automatically instrument supported processes running as `systemd`
services, send the captured traces and metrics to the [Splunk OpenTelemetry Collector](
https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then on to [Splunk APM](
https://docs.splunk.com/Observability/apm/intro-to-apm.html).

Currently, the following Auto Instrumentation agent(s) are supported:

- [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/get-started.html)

## Prerequisites/Requirements

- Check agent compatibility and requirements:
  - [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/java-otel-requirements.html)
- [Install and configure](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html) the Splunk
  OpenTelemetry Collector.
- Debian or RPM based Linux distribution (amd64/x86_64 or arm64/aarch64).
- Supported applications running as `systemd` services.

## Installation

The `splunk-otel-systemd-auto-instrumentation` deb/rpm package provides the following files to enable and configure Auto
Instrumentation agent(s) for `systemd` services:
- [`/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`](#systemd-environment-variables): `systemd` drop-in
  file to enable and configure environment variables for the Java Auto Instrumentation agent.
- `/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`: The [Splunk OpenTelemetry Auto Instrumentation Java
  Agent](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/splunk-java-otel-distribution.html).

Install this package from [Debian/RPM Package Repositories](
../../docs/getting-started/linux-manual.md#auto-instrumentation-debianrpm-package-repositories), or manually download
and install the individual [Debian/RPM Packages](
../../docs/getting-started/linux-manual.md#auto-instrumentation-debianrpm-packages).

After installation, restart the applicable services or reboot the system to enable the Auto Instrumentation agent(s)
with the default configuration. Optionally, see [Configuration](#configuration) for details about configuring the
installed agent(s).

## Configuration

> Before making any changes, it is recommended to check the configuration of the system or individual services for
> potential conflicts.

- **Java**: See the [Advanced Configuration Guide](
  https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
  for details about supported options and defaults for the Java agent. These options can be configured via
  environment variables or their corresponding system properties after installation.

  > **Configuration Priority**:
  > 
  > The Java agent can consume configuration options from one or more of the following sources (ordered from highest to
  > lowest priority):
  > 1. Java system properties (`-D` flags) passed directly to the agent. For example,
  >      ```shell
  >      JAVA_TOOL_OPTIONS="-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
  >      ```
  > 2. Environment variables
  > 3. Property files

### Systemd Environment Variables

The default [`/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`](
./packaging/00-splunk-otel-javaagent.conf) drop-in file defines the following environment variables to
enable the installed agent(s) to auto instrument supported Java applications running as `systemd` services:

- `JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`

Any changes to this file will affect ***all*** `systemd` services, unless overriden by higher-priority system or service
configurations.

> ***Note***: `systemd` supports many options/methods for configuring environment variables at the system level or for
> individual services, and are not limited to the examples below. Consult the documentation specific to your Linux
> distribution or service before making any changes. For general details about `systemd`, see the [`systemd` man page](
> https://www.freedesktop.org/software/systemd/man/index.html).

To add/modify/override supported environment variables defined in
`/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf` (requires `root` privileges):
1. **Option A**: Add/Update `DefaultEnvironment` within
   `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf` for the desired environment variables. For
   example:
     ```shell
     $ cat <<EOH > /usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     DefaultEnvironment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option B**: Create/Modify a higher-priority drop-in file for ***all*** services to add or override the environment
   variables defined in `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For example:
     ```shell
     $ cat <<EOH >> /usr/lib/systemd/system.conf.d/99-my-custom-env-vars.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     DefaultEnvironment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option C**: Create/Modify a higher-priority drop-in file for a ***specific*** service to add or override the
   environment variables defined in `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For
   example:
     ```shell
     $ cat <<EOH >> /usr/lib/systemd/system/my-service.d/99-my-custom-env-vars.conf
     [Service]
     Environment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     Environment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     Environment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
2. After any configuration changes, reboot the system or run the following commands to restart the applicable services
   for the changes to take effect:
     ```shell
     $ systemctl daemon-reload
     $ systemctl restart <service-name>   # replace "<service-name>" and run for each applicable service
     ```

## Uninstall

1. Run the following command to uninstall the `splunk-otel-systemd-auto-instrumentation` package (requires `root`
   privileges):
   - Debian:
       ```shell
       dpkg -P splunk-otel-systemd-auto-instrumentation
       ```
   - RPM:
       ```shell
       rpm -e splunk-otel-systemd-auto-instrumentation
       ```
2. The `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf` drop-in file may persist if it was
   modified after initial installation. Manually delete this file and any custom drop-in files as necessary.
3. Reboot the system or restart the applicable services for the changes to take effect.
