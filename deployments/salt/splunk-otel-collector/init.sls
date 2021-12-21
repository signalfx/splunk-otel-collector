{% if not salt['pillar.get']('splunk-otel-collector:splunk_access_token') %}
Check splunk_access_token:
  cmd.run:
    - name: |
        echo "splunk_access_token is is not specified";
        exit 1
{% elif not salt['pillar.get']('splunk-otel-collector:splunk_realm') %}
Check splunk_realm:
  cmd.run:
    - name: |
        echo "splunk_realm is is not specified";
        exit 1
{% else %}

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

{% endif %}
