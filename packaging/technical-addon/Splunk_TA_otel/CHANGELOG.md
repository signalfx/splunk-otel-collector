# Changelog

## Unreleased

- Removed `splunk_otel_log_file` configuration option. In unmanaged scenarios, this file
  would grow in case of errors and would not be rotated or truncated by operators, leading
  to problems with disk space use. Now, by default, the OTel Collector will log errors
  to `splunkd.log` on all platforms. If you want more detailed logging or if you want to log to a different file,
  you can modify the `service.telemetry.logs` section in the YAML files under the `configs/` directory of the add-on.

## v1.5.0

Updates addon to use collector version v0.131.0
Fixes Smart Agent bundle extraction
Fixes windows orphaned OTel Collector processes when SplunkD exits ungracefully

## v1.4.4

Updates addon to use collector version v0.130.0

## v1.4.3

Updates addon to use collector version v0.128.0
