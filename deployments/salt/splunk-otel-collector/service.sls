{% set splunk_otel_collector_config = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config', '/etc/otel/collector/agent_config.yaml') %}

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
