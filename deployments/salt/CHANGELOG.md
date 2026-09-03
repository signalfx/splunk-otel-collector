# Changelog

## Unreleased

### 🛑 Breaking changes 🛑

- Linux auto-instrumentation now uses the official OpenTelemetry Injector for package versions newer than `0.159.0` (or `latest`). The state manages `/etc/opentelemetry/injector/` and `libotelinject.so`, removes the legacy `/etc/splunk/zeroconfig/` configuration during upgrades, and continues to support the legacy layout for package versions through `0.159.0`. .NET auto-instrumentation is also supported on arm64 for newer package versions.

  Existing values are not migrated automatically. Before upgrading, copy custom `OTEL_*` and `SPLUNK_*` values from `/etc/splunk/zeroconfig/{java,node,dotnet}.conf` to `/etc/opentelemetry/injector/default_env.conf`; variables with other prefixes, including `JAVA_TOOL_OPTIONS`, `NODE_OPTIONS`, `CORECLR_*`, and `DOTNET_*`, are ignored. The shared `default_env.conf` applies to all runtimes, while runtime- or service-specific values must be configured in the relevant application or service environment. The `/etc/ld.so.preload` entry changes from `/usr/lib/splunk-instrumentation/libsplunk.so` to `/usr/lib/splunk-instrumentation/libotelinject.so`, and legacy configuration files are removed during the upgrade. Restart instrumented applications or services, or reboot, after updating the preload entry or configuration.
