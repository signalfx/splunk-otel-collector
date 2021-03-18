# FAQ

- **What’s new in latest splunk-otel-collector release?** Check the
  [CHANGELOG.md](../CHANGELOG.md)
- **What does "beta" mean?** Beta means that breaking changes could be
  introduced in a future release of this distribution. While in beta, Splunk
  provides full support for this distribution; customers use this distribution
  in production today.
- **Can upstream opentelemetry-collector or opentelemetry-collector-contrib be
  used instead?** Definitely, however we will not supply hands-on technical support for customers using the upstream distribution.
- **What’s different between Splunk distro and upstream?** Supported by Splunk, better
  defaults for Splunk products, usability improvements, tools to support
  migration from SignalFx products. Note, we take an upstream-first approach,
  distros allow us to move fast.
- **Can AWS Distro for OpenTelemetry be used?** For Splunk APM and
  Infrastructure Monitoring if the data goes through the AWS OTel Collector
  then yes. Splunk Log Observer is not supported today. The AWS Lambda
  instrumentation included in the AWS Distro of OpenTelemetry does not support
  sending to the OTel Collector or directly to Splunk today.
- **Can I deploy splunk-otel-collector without fluentd?** Yes, manual
  installation does not include fluentd. The installer script offers a way to
  exclude fluentd via a runtime parameter.
- **Can I deploy fluentd without splunk-otel-collector?** This is not Splunk
  supported today, but fluentd is open-source and can be manually configured.
  The fluentd configuration that the installer script configures can be found
  [here](../internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd).
