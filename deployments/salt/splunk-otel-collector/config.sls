{% set splunk_otel_collector_config = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config', '/etc/otel/collector/agent_config.yaml') %}

{% set splunk_otel_collector_config_source = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config_source') %}

{% set splunk_service_user = salt['pillar.get']('splunk-otel-collector:service_user', 'splunk-otel-collector') %}

{% set splunk_service_group = salt['pillar.get']('splunk-otel-collector:service_group', 'splunk-otel-collector') %}

{% if splunk_otel_collector_config_source  != '' %}

Push custom config for Splunk Otel Collector, if provided:
  file.managed:
    - name: {{ splunk_otel_collector_config }}
    - source: {{ splunk_otel_collector_config_source }}
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - mode: '0644'
    - makedirs: True
    - template: jinja
    - watch:
      - user: splunk_service_user
      - group: splunk_service_group

{% else %}

Copy default config for Splunk Otel Collector:
  file.copy:
    - name: {{ splunk_otel_collector_config }}
    - source: /etc/otel/collector/agent_config.yaml
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - mode: '0644'
    - makedirs: True
    - template: jinja
    - watch:
      - user: splunk_service_user
      - group: splunk_service_group

{% endif %}
