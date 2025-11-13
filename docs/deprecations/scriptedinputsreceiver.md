# Scripted Inputs Receiver Deprecation

[The Scripted Inputs Receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/scriptedinputsreceiver) has been deprecated and will be removed in a future release.

## Replacement guidance

The Scripted Inputs Receiver was designed to replicate log collection behavior of the Splunk Universal Forwarder when the [Unix and Linux Technical Add-on](https://docs.splunk.com/Documentation/AddOns/released/UnixLinux/About) is installed. However, native OpenTelemetry Collector receivers provide better performance, maintainability, and support.


### Additional Resources

- [File Log Receiver Documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [Splunk OpenTelemetry Collector Configuration Examples](https://github.com/signalfx/splunk-otel-collector/tree/main/examples)

## Timeline

- **Deprecation Notice**: Current release
- **Planned Removal**: Future release (to be announced)

Please plan to migrate to the recommended alternatives at your earliest convenience.