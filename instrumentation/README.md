# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux

**Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux** (`splunk-otel-auto-instrumentation`)
Debian/RPM package installs Splunk OpenTelemetry Auto Instrumentation agent(s), the `libsplunk.so` shared object
library, and sample configuration files to automatically instrument applications and services to capture and report
distributed traces and metrics to the [Splunk OpenTelemetry Collector](
https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html), and then on to [Splunk APM](
https://docs.splunk.com/Observability/apm/intro-to-apm.html).

Currently, `splunk-otel-auto-instrumentation` installs and supports configuration of the following Auto Instrumentation
agent(s):

- [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/get-started.html)

## Prerequisites/Requirements

- Check agent compatibility and requirements:
  - [Java](https://docs.splunk.com/Observability/gdi/get-data-in/application/java/java-otel-requirements.html)
- [Install and configure](https://docs.splunk.com/Observability/gdi/opentelemetry/install-linux.html) the Splunk
  OpenTelemetry Collector.
- Debian or RPM based Linux distribution (amd64/x86_64 or arm64/aarch64).

## Zero Configuration Options

The following options are supported to enable the installed Auto Instrumentation agent(s):

- **[`libsplunk.so`](./libsplunk.md)**: The provided `/usr/lib/splunk-instrumentation/libsplunk.so` shared object
  library can be added to the `/etc/ld.so.preload` file to enable and configure Auto Instrumentation for ***all***
  supported processes on the system, or it can be assigned to the `LD_PRELOAD` environment variable for specific
  applications/services. Configuration of the installed agent(s) is supported by the
  [`/usr/lib/splunk-instrumentation/instrumentation.conf`](./libsplunk.md#configuration-file) file.

- **[`systemd`](./systemd.md)**: The provided `systemd` drop-in file(s) in the
  `/usr/lib/splunk-instrumentation/examples/systemd/` directory can be copied to the host's `systemd` configuration
  directory, e.g. `/usr/lib/systemd/system.conf.d/`, to enable and configure Auto Instrumentation for ***all***
  supported applications running as `systemd` services. Configuration of the installed agent(s) is supported by
  modifying these files or adding custom drop-in files with the desired environment variables.

Alternatively, [manually install and configure](
https://docs.splunk.com/Observability/gdi/get-data-in/application/application.html)
Auto Instrumentation if neither option is appropriate for the target host or application.

> ### Notes
> 
> 1. To prevent conflicts and duplicate traces/metrics, only one option should be enabled on the target system.
> 2. The configuration files and the options defined within are only applicable for the respective option that is
>    configured. For example, `/usr/lib/splunk-instrumentation/instrumentation.conf` is only applicable with
>    `libsplunk.so`, and the systemd drop-in file is not applicable for `libsplunk.so`.
> 3. The [`splunk.linux-autoinstr.executions`](./libsplunk.md#disable_telemetry-optional) telemetry
>    metric is currently only provided with `libsplunk.so`.
