# Migrating Fluentd's Position File to Otel Checkpoints

This package provides a simple program to read Fluentd's position information 
from position file and store it in the format of Otel's checkpoints database 
in `file_storage` extension.

## Intended Use Case

Its primary intended use is as a initContainer in kubernetes for users who have 
been using fluentd as a logging agent but wants to use OpenTelemetry Collector. 

## Configuration

Environment variables are used to configure it. These are currently available 
variables and its default values.
```yaml
- name: CONTAINER_LOG_PATH_FLUENTD
  value: "/var/log/splunk-fluentd-containers.log.pos"
- name: CONTAINER_LOG_PATH_OTEL
  value: "/var/lib/otel_pos/receiver_filelog_"
- name: CUSTOM_LOG_PATH_FLUENTD
  value: "/var/log/splunk-fluentd-*.pos"
- name: CUSTOM_LOG_PATH_OTEL
  value: "/var/lib/otel_pos/receiver_filelog_"
- name: CUSTOM_LOG_CAPTURE_REGEX
  value: "\\/splunk\\-fluentd\\-(?P<name>[\\w0-9-_]+)\\.pos"
- name: JOURNALD_LOG_PATH_FLUENTD
  value: "/var/log/splunkd-fluentd-journald-*.pos.json"
- name: JOURNALD_LOG_PATH_OTEL
  value: "/var/lib/otel_pos/receiver_journald_"
- name: JOURNALD_LOG_CAPTURE_REGEX
  value: "\\/splunkd\\-fluentd\\-journald\\-(?P<name>[\\w0-9-_]+)\\.pos\\.json"
```