{% if grains['os_family'] not in ['Debian', 'RedHat', 'Suse'] %}
Check OS family:
  test.fail_without_changes:
    - name: "OS family ({{ grains['os_family'] }}) is not supported!"
    - failhard: True
{% elif not salt['pillar.get']('splunk-otel-collector:splunk_access_token') %}
Check splunk_access_token:
  test.fail_without_changes:
    - name: "splunk_access_token is not specified!"
    - failhard: True
{% elif not salt['pillar.get']('splunk-otel-collector:splunk_realm') %}
Check splunk_realm:
  test.fail_without_changes:
    - name: "splunk_realm is not specified!"
    - failhard: True
{% endif %}

{% set install_fluentd = salt['pillar.get']('splunk-otel-collector:install_fluentd', True) | to_bool %}
{% set install_auto_instrumentation = salt['pillar.get']('splunk-otel-collector:install_auto_instrumentation', False) | to_bool %}

include:
{% if grains['os_family'] == 'Suse' or install_fluentd == False %}
  - splunk-otel-collector.install
  - splunk-otel-collector.service_owner
  - splunk-otel-collector.config
  - splunk-otel-collector.collector_config
  - splunk-otel-collector.service
{% elif install_fluentd == True %}
  - splunk-otel-collector.install
  - splunk-otel-collector.service_owner
  - splunk-otel-collector.config
  - splunk-otel-collector.collector_config
  - splunk-otel-collector.service
  - splunk-otel-collector.fluentd
  - splunk-otel-collector.fluentd_config
  - splunk-otel-collector.fluentd_service
{% endif %}
{% if install_auto_instrumentation %}
  - splunk-otel-collector.auto_instrumentation
{% endif %}
