# Changelog

## Unreleased

### 💡 Enhancements 💡

- Add support for the `collector_command_line_args` option to
  configure the command line arguments for the Splunk OpenTelemetry Collector
  service. On Windows, this option is only supported by versions `>= 0.127.0`.

### 🚩 Deprecations 🚩

- Fluentd support has been deprecated and will be removed in a future release.
  Please refer to [deprecation documentation](../../docs/deprecations/fluentd-support.md) for more information ([#6339](https://github.com/signalfx/splunk-otel-collector/pull/6339))

## chef-v0.17.0

- Bug fix: Fix bug that caused custom variables to not be set when running on Windows
  with a Splunk OTel Collector version >= `0.98.0`.
- Remove the option `with_signalfx_dotnet_auto_instrumentation` used to
install the deprecated `SignalFx Auto Instrumentation for .NET` on Windows.

## chef-v0.16.0

- Add support for the `auto_instrumentation_logs_exporter` option to configure the `OTEL_LOGS_EXPORTER` environment variable.

## chef-v0.15.0

- Breaking Change: The default for the `auto_instrumentation_otlp_endpoint` option has been changed from
  `http://127.0.0.1:4317` to `''` (empty), i.e. defer to the default `OTEL_EXPORTER_OTLP_ENDPOINT` value for each
  activated SDK.
- Add support for the `auto_instrumentation_otlp_endpoint_protocol` and `auto_instrumentation_metrics_exporter` options
  to configure the `OTEL_EXPORTER_OTLP_PROTOCOL` and `OTEL_METRICS_EXPORTER` environment variables, respectively.

## chef-v0.14.0

- Initial support for [Splunk OpenTelemetry for .NET](https://github.com/signalfx/splunk-otel-dotnet) Auto
  Instrumentation on Linux (x86_64/amd64 only):
  - The .NET SDK is activated by default if the `with_auto_instrumentation` option is set to `true` and
    `auto_instrumentation_version` is `latest` or >= `0.99.0`.
  - To skip .NET auto instrumentation, configure the `with_auto_instrumentation_sdks` option without `dotnet`.

## chef-v0.13.0

- Only copy the collector configuration file to `ProgramData` if the source exists.

## chef-v0.12.0

- `splunk_ballast_size_mib` is deprecated and removed. For Splunk Otel Collector version `0.97.0` or greater, `GOMEMLIMIT` env var is introduced. The default is set to 90% of the `SPLUNK_TOTAL_MEM_MIB`. For more information regarding the usage, please follow the instructions ([here](https://github.com/signalfx/splunk-otel-collector?tab=readme-ov-file#from-0961-to-0970)).

## chef-v0.11.0

- On Windows the `SPLUNK_*` environment variables were moved from the machine scope to the collector service scope.
  It is possible that some instrumentations are relying on the machine-wide environment variables set by the installation. ([#3930](https://github.com/signalfx/splunk-otel-collector/pull/3930))

- Initial support for [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js) Auto
  Instrumentation on Linux:
  - The Node.js SDK is installed and activated by default if the `with_auto_instrumentation` option is set to `true`
    and `npm` is found on the node with the `bash -c 'command -v npm'` shell command.
  - Set the `with_auto_instrumentation_sdks` option to only `%w(java)` to skip Node.js auto instrumentation.
  - Use the `auto_instrumentation_npm_path` option to specify a custom path for `npm`.
  - **Note:** This recipe does not manage the installation/configuration of Node.js, `npm`, or Node.js applications.

## chef-v0.9.0

- **Deprecations**: The `auto_instrumentation_generate_service_name` and `auto_instrumentation_disable_telemetry`
  options are deprecated and only applicable if the `auto_instrumentation_version` option is < `0.87.0`.
- Support Splunk OpenTelemetry Auto Instrumentation for Linux [v0.87.0](
  https://github.com/signalfx/splunk-otel-collector/releases/tag/v0.87.0) and newer (Java only).
- Support activation and configuration of auto instrumentation for only `systemd` services.
- Support setting the OTLP exporter endpoint for auto instrumentation (default: `http://127.0.0.1:4317`). Only
  applicable if the `auto_instrumentation_version` option is `latest` or >= `0.87.0`.

## chef-v0.8.0

- Update `splunk_listen_interface` default to only set target SPLUNK_LISTEN_INTERFACE environment variable if
  configured.

## chef-v0.7.0

- Add support for the `splunk_listen_interface` option to configure the network interface the collector receivers
  will listen on (default: `0.0.0.0`)
- Add support for SignalFx .NET Auto Instrumentation on Windows (disabled by default)

## chef-v0.6.0

- **Breaking Changes**: Fluentd installation ***disabled*** by default.
  - Specify the `with_fluentd: true` option to enable installation

## chef-v0.5.0

- Add support for additional options for Splunk OpenTelemetry Auto Instrumentation for Java (Linux only)

## chef-v0.4.0

- Add `collector_additional_env_vars` option to allow passing additional environment variables to the collector service

## chef-v0.3.0

- Update default `td-agent` version to 4.3.2 to support log collection with fluentd on Ubuntu 22.04
- Initial support for Splunk OpenTelemetry Auto Instrumentation for Java (Linux only)

## chef-v0.2.0

- Initial release
