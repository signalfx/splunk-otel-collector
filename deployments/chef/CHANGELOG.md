# Changelog

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
