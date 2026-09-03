# Splunk OpenTelemetry Instrumentation Automatic Configuration for Linux

The **Splunk OpenTelemetry Instrumentation Automatic Configuration for Linux** Debian/RPM package
(`splunk-otel-auto-instrumentation`) installs Splunk OpenTelemetry Auto Instrumentation agents, the
[OpenTelemetry injector](https://github.com/open-telemetry/opentelemetry-injector) (`libotelinject.so`) shared object
library, and default/sample configuration files to automatically instrument applications and services to capture and
report distributed traces and metrics to the [Splunk OpenTelemetry Collector](
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

1. Add the path of the provided `/usr/lib/splunk-instrumentation/libotelinject.so` shared object library to the
   [`/etc/ld.so.preload`](https://man7.org/linux/man-pages/man8/ld.so.8.html#FILES) file to activate Auto
   Instrumentation for ***all*** supported processes on the system. For example:

   ```bash
   echo /usr/lib/splunk-instrumentation/libotelinject.so >> /etc/ld.so.preload
   ```

2. The default configuration file `/etc/opentelemetry/injector/injector.conf` includes the required settings, i.e. the
   paths to the respective auto-instrumentation agents per runtime:

   ```
   jvm_auto_instrumentation_agent_path=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar
   nodejs_auto_instrumentation_agent_path=/usr/lib/splunk-instrumentation/splunk-otel-js/node_modules/@splunk/otel/instrument.js
   dotnet_auto_instrumentation_agent_path_prefix=/usr/lib/splunk-instrumentation/splunk-otel-dotnet
   ```

   Configuration of the respective agents is supported by adding/updating environment variables in the
   `/etc/opentelemetry/injector/default_env.conf` file, which by default only sets:

   ```
   OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation
   ```

   You can add/update the following environment variables in this file (***any environment variable that does not
   start with `OTEL_` or `SPLUNK_` will be ignored***):
   - `OTEL_EXPORTER_OTLP_ENDPOINT`
   - `OTEL_EXPORTER_OTLP_PROTOCOL`
   - `OTEL_LOGS_EXPORTER`
   - `OTEL_METRICS_EXPORTER`
   - `OTEL_RESOURCE_ATTRIBUTES`
   - `OTEL_SERVICE_NAME`
   - `SPLUNK_METRICS_ENABLED`
   - `SPLUNK_PROFILER_ENABLED`
   - `SPLUNK_PROFILER_MEMORY_ENABLED`

   If an environment variable is set both in an application or service environment and in `default_env.conf`, the
   existing application or service value takes precedence.

   Check the following for details about these environment variables and default values:
   - [Java](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
   - [Node.js](https://docs.splunk.com/Observability/en/gdi/get-data-in/application/nodejs/configuration/advanced-nodejs-otel-configuration.html)
   - [.NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/configuration/advanced-dotnet-configuration.html)

   See the [OpenTelemetry injector README](https://github.com/open-telemetry/opentelemetry-injector#readme) for
   additional configuration options, such as selectively enabling/disabling auto-instrumentation for specific runtimes
   or programs, and Kubernetes-related resource attribute mapping.
3. Reboot the system or restart the applications/services for any changes to take effect. The `libotelinject.so`
   shared object library will then be preloaded for all subsequent processes and inject the environment variables from
   the `/etc/opentelemetry/injector/` configuration files for Java, Node.js, and .NET processes.

## Running the `auto-instrumentation` CI Workflow Locally

The [`auto-instrumentation.yml`](../.github/workflows/auto-instrumentation.yml) workflow builds the collector binary
and the `splunk-otel-auto-instrumentation` package, then runs `packaging/tests/instrumentation/instrumentation_test.py`
against them in distro containers. To reproduce a single `test-package (<distro>, <arch>, <testcase>)` job locally
(e.g. `test-package (debian-bullseye, arm64, dotnet)`):

1. Build the collector binary for the target arch (from the repo root):

   ```bash
   make binaries-linux_arm64   # or binaries-linux_amd64
   ```

   Produces `bin/otelcol_linux_<arch>`.

2. Check out the pinned injector source and build the deb/rpm package (from `instrumentation/`):

   ```bash
   cd instrumentation
   make checkout-injector          # clones open-telemetry/opentelemetry-injector at the version in packaging/injector-release.txt
   make deb-package ARCH=arm64     # or rpm-package, matching the target distro's package type
   cd ..
   ```

   Produces `instrumentation/dist/*.deb` (or `.rpm`).

3. Install the test dependencies:

   ```bash
   python3 -m venv .venv && source .venv/bin/activate
   pip install -r packaging/tests/requirements.txt
   ```

4. Run pytest with the same `-k` filter CI uses (distro, arch, testcase):

   ```bash
   python3 -u -m pytest -s --verbose \
     -k "debian-bullseye and arm64 and (dotnet or uninstall)" \
     packaging/tests/instrumentation/instrumentation_test.py
   ```

Notes:

- Docker must be running; tests spin up `--privileged` systemd containers defined under
  `packaging/tests/instrumentation/images/{deb,rpm}/Dockerfile.<distro>`.
- Run on a host matching the target `arch` to avoid needing QEMU emulation.
- The test looks up `bin/otelcol_linux_<arch>` and a matching package in `instrumentation/dist/`, so steps 1 and 2
  must complete first.
