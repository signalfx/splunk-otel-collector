{% set install_fluentd = salt['pillar.get']('splunk-otel-collector:install_fluentd', 'True') %}
include:
{% if grains['os_family'] == 'Suse' or install_fluentd == 'False' %}
  - splunk-otel-collector.install
  - splunk-otel-collector.service_owner
  - splunk-otel-collector.config
  - splunk-otel-collector.collector_config
  - splunk-otel-collector.service
{% elif install_fluentd == 'True' %}
  - splunk-otel-collector.install
  - splunk-otel-collector.service_owner
  - splunk-otel-collector.config
  - splunk-otel-collector.collector_config
  - splunk-otel-collector.service
  - splunk-otel-collector.fluentd
  - splunk-otel-collector.fluentd_config
{% endif %}
