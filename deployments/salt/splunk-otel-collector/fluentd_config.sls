{% set splunk_fluentd_config = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config', '/etc/otel/collector/fluentd/fluent.conf') %}

{% set splunk_fluentd_config_source = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config_source') %}

{% if splunk_fluentd_config_source != '' %}

Push custom FluentD config, if provided:
  file.managed:
    - name: {{ splunk_fluentd_config }}
    - source: {{ splunk_fluentd_config_source }}
    - template: jinja
    - mode: '0644'
    - makedirs: true
    - user: td-agent
    - group: td-agent

{% endif %}
