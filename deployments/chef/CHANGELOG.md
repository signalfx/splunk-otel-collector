# Changelog

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
