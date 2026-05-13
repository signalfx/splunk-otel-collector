# Splunk OpenTelemetry Instrumentation Automatic Configuration for Linux

The **Splunk OpenTelemetry Instrumentation Automatic Configuration for Linux** Debian/RPM package
(`splunk-otel-auto-instrumentation`) installs Splunk OpenTelemetry Auto Instrumentation agents, the `libsplunk.so`
shared object library, and default/sample configuration files to automatically instrument applications and services to
capture and report distributed traces and metrics to the [Splunk OpenTelemetry Collector](
https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then on to [Splunk APM](
https://docs.splunk.com/Observability/apm/intro-to-apm.html).

The `splunk-otel-auto-instrumentation` deb/rpm package installs and supports configuration of the following Auto
Instrumentation agents:

- [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/get-started.html)
- [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/get-started.html)
- [.NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/get-started.html)

For other languages or if the `splunk-otel-auto-instrumentation` deb/rpm package is not applicable for the target host
or applications/services, see [Instrument back-end applications to send spans to Splunk APM](
https://docs.splunk.com/Observability/en/gdi/get-data-in/application/application.html).

## Prerequisites/Requirements

- Check agent compatibility and requirements:
  - [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/java-otel-requirements.html)
  - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/nodejs-otel-requirements.html)
  - [.NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/dotnet-requirements.html)
- [Install and configure](https://docs.splunk.com/observability/en/gdi/opentelemetry/collector-linux/install-linux.html)
  the Splunk OpenTelemetry Collector.
- Debian or RPM based Linux distribution (amd64/x86_64 or arm64/aarch64).
  - **Note**: .NET only supported on amd64/x86_64

## Installation

### Installer Script

The [Linux Installer Script](../docs/getting-started/linux-installer.md) is available to automate the installation and
configuration of the Collector and Auto Instrumentation for supported platforms. See
[Auto Instrumentation](../docs/getting-started/linux-installer.md#auto-instrumentation) for details.

### Manual

1. [Install and configure](https://docs.splunk.com/observability/en/gdi/opentelemetry/collector-linux/install-linux.html)
   the Splunk OpenTelemetry Collector.
2. [Install](../docs/getting-started/linux-manual.md#auto-instrumentation-debianrpm-packages) the
   `splunk-otel-auto-instrumentation` deb/rpm package
3. If Auto Instrumentation for Node.js is required, [install](
   ../docs/getting-started/linux-manual.md#auto-instrumentation-post-install-configuration) the provided
   `/usr/lib/splunk-instrumentatgion/splunk-otel-js.tgz` Node.js package with `npm`.
4. [Activate and configure](#activation-and-configuration) Auto Instrumentation with the supported methods and options.

## Activation and Configuration

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
     NODE_OPTIONS=-r /usr/lib/splunk-instrumentation/splunk-otel-js/node_modules/@splunk/otel/instrument
     ```
   - `/etc/splunk/zeroconfig/dotnet.conf`:
     ```
     CORECLR_ENABLE_PROFILING=1
     CORECLR_PROFILER={918728DD-259F-4A6A-AC2B-B85E1B658318}
     CORECLR_PROFILER_PATH=/usr/lib/splunk-instrumentation/splunk-otel-dotnet/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so
     DOTNET_ADDITIONAL_DEPS=/usr/lib/splunk-instrumentation/splunk-otel-dotnet/AdditionalDeps
     DOTNET_SHARED_STORE=/usr/lib/splunk-instrumentation/splunk-otel-dotnet/store
     DOTNET_STARTUP_HOOKS=/usr/lib/splunk-instrumentation/splunk-otel-dotnet/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll
     OTEL_DOTNET_AUTO_HOME=/usr/lib/splunk-instrumentation/splunk-otel-dotnet
     OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation
     ```
   Configuration of the respective agents is supported by the adding/updating the following environment variables in
   each of these files (***any environment variable not in this list will be ignored***):
   - `OTEL_EXPORTER_OTLP_ENDPOINT`
   - `OTEL_EXPORTER_OTLP_PROTOCOL`
   - `OTEL_LOGS_EXPORTER`
   - `OTEL_METRICS_EXPORTER`
   - `OTEL_RESOURCE_ATTRIBUTES`
   - `OTEL_SERVICE_NAME`
   - `SPLUNK_METRICS_ENABLED`
   - `SPLUNK_PROFILER_ENABLED`
   - `SPLUNK_PROFILER_MEMORY_ENABLED`

   Check the following for details about these environment variables and default values:
   - [Java](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
   - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/configuration/advanced-nodejs-otel-configuration.html)
   - [.NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/configuration/advanced-dotnet-configuration.html)
3. Reboot the system or restart the applications/services for any changes to take effect. The `libsplunk.so` shared
   object library will then be preloaded for all subsequent processes and inject the environment variables from the
   `/etc/splunk/zeroconfig` configuration files for Java and Node.js processes.

