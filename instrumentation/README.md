# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux

**Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux** installs and configures Splunk OpenTelemetry
Auto Instrumentation agent(s) that automatically instrument applications and services to capture and report distributed
traces and metrics to the [Splunk OpenTelemetry Collector](
https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then on to [Splunk APM](
https://docs.splunk.com/Observability/apm/intro-to-apm.html).

Currently, the following Auto Instrumentation agent(s) are supported:

- [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/get-started.html)

## Debian/RPM Package Options

- **[`splunk-otel-auto-instrumentation`](./splunk-otel-auto-instrumentation)**: Installs the Auto Instrumentation
  agent(s) and the `libsplunk.so` shared object library, which is automatically added to `/etc/ld.so.preload` to enable
  and configure Auto Instrumentation for ***all*** supported processes on the system. Configuration of the installed
  agent(s) is supported by the [`/usr/lib/splunk-instrumentation/instrumentation.conf`](
  ./splunk-otel-auto-instrumentation/README.md#configuration-file) file.

- **[`splunk-otel-systemd-auto-instrumentation`](./splunk-otel-systemd-auto-instrumentation)**: Installs the Auto
  Instrumentation agent(s) and a `systemd` drop-in file, which define environment variables to automatically enable and
  configure Auto Instrumentation for supported applications running as `systemd` services. Configuration of the
  installed agent(s) is supported by modifying the installed drop-in file or adding custom drop-in files with the
  desired environment variables.

Alternatively, [manually install and configure](
https://docs.splunk.com/Observability/gdi/get-data-in/application/application.html)
Auto Instrumentation if neither package is appropriate.

> ### Notes
> 
> 1. To prevent conflicts and duplicate traces/metrics, only one package should be installed on the target system.
> 2. The configuration files and the options defined within are only applicable for the respective package that is
>    installed. For example, `/usr/lib/splunk-instrumentation/instrumentation.conf` is only applicable with
>    `splunk-otel-auto-instrumentation`, and `` is only
>    applicable with `splunk-otel-systemd-auto-instrumentation`. Furthermore, migration from one package to the other
>    ***does not*** automatically migrate the options from the old configuration file to the new one.
> 3. The [`splunk.linux-autoinstr.executions`](./splunk-otel-auto-instrumentation#disable_telemetry-optional) telemetry
>    metric is currently only provided by the `libsplunk.so` shared object library from the
>    `splunk-otel-auto-instrumentation` package.
