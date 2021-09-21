> The official Splunk documentation for this page is [](). For instructions on how to contribute to the docs, see [CONTRIBUTE.md](../CONTRIBUTING#documentation.md).

# FAQ

- **What’s new in latest splunk-otel-collector release?** See the
  [CHANGELOG.md](../CHANGELOG.md).
- **What does "beta" mean?** See the [beta definition](beta-definition.md).
- **Can upstream opentelemetry-collector or opentelemetry-collector-contrib be
  used instead?** Definitely, however Splunk only provides best-effort support.
  See the [beta definition](beta-definition.md).
- **What’s different between Splunk OpenTelemetry Connector and OpenTelemetry
  Collector?** Supported by Splunk, better defaults for Splunk products,
  bundled FluentD for log capture, tools to support migration from SignalFx
  products. Note, we take an upstream-first approach, Splunk OpenTelemetry
  Connector allow us to move fast.
- **Can AWS Distro for OpenTelemetry be used?** For Splunk APM and
  Infrastructure Monitoring if the data goes through the AWS OTel Collector
  then yes. Splunk Log Observer is not supported today. The AWS Lambda
  instrumentation included in the AWS Distro of OpenTelemetry does not support
  sending to the OTel Collector or directly to Splunk today.
- **Can I deploy Splunk OpenTelemetry Connector without fluentd?** Yes, manual
  installation does not include fluentd. The installer script offers a way to
  exclude fluentd via a runtime parameter.
- **Can I deploy fluentd without Splunk OpenTelemetry Connector?** Yes using
  the already available open-source project with manual configuration.
