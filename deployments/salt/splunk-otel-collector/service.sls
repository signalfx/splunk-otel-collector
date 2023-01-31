{% set splunk_otel_collector_config = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config', '/etc/otel/collector/agent_config.yaml') %}

{% set splunk_service_user = salt['pillar.get']('splunk-otel-collector:splunk_service_user', 'splunk-otel-collector') %}

{% set splunk_service_group = salt['pillar.get']('splunk-otel-collector:splunk_service_group', 'splunk-otel-collector') %}

Ensure /etc/otel/collector is owned by the service user/group:
  file.directory:
    - name: /etc/otel/collector
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - recurse:
      - user
      - group
    - require:
      - pkg: splunk-otel-collector
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group

Restart Splunk Otel Collector service:
  service.running:
    - name: splunk-otel-collector
    - enable: True
    - require:
      - pkg: splunk-otel-collector
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group
      - file: {{ splunk_otel_collector_config }}
      - file: /etc/otel/collector/splunk-otel-collector.conf
