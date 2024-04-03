# Changelog

## Unreleased

## puppet-v0.14.0

- For Splunk Otel Collector version `0.97.0` or greater, `GOMEMLIMIT` env var is introduced. The default is set to 90% of the `SPLUNK_TOTAL_MEM_MIB`. For more information regarding the usage, please follow the instructions ([here](https://github.com/signalfx/splunk-otel-collector?tab=readme-ov-file#from-0961-to-0970)).
- The `splunk_ballast_size_mib` option is deprecated and no longer effective. It is only applicable for Splunk OpenTelemetry Collector version < `0.97.0`.

## puppet-v0.13.0

- On Windows the `SPLUNK_*` environment variables were moved from the machine scope to the collector service scope.
  It is possible that some instrumentations are relying on the machine-wide environment variables set by the installation. ([#3930](https://github.com/signalfx/splunk-otel-collector/pull/3930))
- Initial support for [Splunk OpenTelemetry for Node.js](https://github.com/signalfx/splunk-otel-js) Auto
  Instrumentation on Linux:
  - The Node.js SDK is installed and activated by default if the `with_auto_instrumentation` option is set to `true`
    and `npm` is found on the node with the `bash -c 'command -v npm'` shell command.
  - Set the `with_auto_instrumentation_sdks` option to only `['java']` to skip Node.js auto instrumentation.
  - Use the `auto_instrumentation_npm_path` option to specify a custom path for `npm`.
  - **Note:** This cookbook does not manage the installation/configuration of Node.js, `npm`, or Node.js applications.

## puppet-v0.12.0

- **Deprecations**: The `auto_instrumentation_generate_service_name` and `auto_instrumentation_disable_telemetry`
  options are deprecated and only applicable if the `auto_instrumentation_version` option is < `0.87.0`.
- Support Splunk OpenTelemetry Auto Instrumentation for Linux [v0.87.0](
  https://github.com/signalfx/splunk-otel-collector/releases/tag/v0.87.0) and newer (Java only).
- Support activation and configuration of auto instrumentation for only `systemd` services.
- Support setting the OTLP exporter endpoint for auto instrumentation (default: `http://127.0.0.1:4317`). Only
  applicable if the `auto_instrumentation_version` option is `latest` or >= `0.87.0`.

## puppet-v0.11.0

- Add support for `splunk_listen_interface` used by default configurations as `SPLUNK_LISTEN_INTERFACE` environment variable (only populated if set).
- Update fluentd url for Windows.

## puppet-v0.10.0

- **Breaking Changes**: Fluentd installation ***disabled*** by default.
  - Specify the `with_fluentd => true` option to enable installation

## puppet-v0.9.0

- Add support for additional options for Splunk OpenTelemetry Auto Instrumentation for Java (Linux only)

## puppet-v0.8.0

- Add `collector_additional_env_vars` option to allow passing additional environment variables to the collector service
- Add support for Windows 2022

## puppet-v0.7.0

- Initial support for [Linux Java Auto Instrumentation](https://github.com/signalfx/splunk-otel-collector/tree/main/instrumentation)
- Bump default td-agent version to 4.3.2

## puppet-v0.6.0

- Add support for Ubuntu 22.04

## puppet-v0.5.0

- Add support for Debian 11, remove support for Debian 8

## puppet-v0.4.0

- Bump default td-agent version to 4.3.0

## puppet-v0.3.0

- Initial support for Suse 12 and 15

## puppet-v0.2.1

- Install `libcap` dependency for RPM distros

## puppet-v0.2.0

- Initial support for Windows

## puppet-v0.1.0

- Initial release
