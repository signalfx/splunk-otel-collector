# Changelog

## Unreleased

- Removed `splunk_otel_log_file` configuration option. By default, the OTel Collector will log errors
to `splunkd.log` on all platforms. If you want more detailed logging or if you want to log to a different file,
you can modify the `service.telemetry.logs` section in the YAML files under the `configs/` directory of the addon.

## v1.5.0

Updates addon to use collector version v0.131.0
Fixes Smart Agent bundle extraction
Fixes windows orphaned OTel Collector processes when SplunkD exits ungracefully

## v1.4.4

Updates addon to use collector version v0.130.0

## v1.4.3

Updates addon to use collector version v0.128.0
