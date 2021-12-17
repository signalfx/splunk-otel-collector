Start Splunk Otel Collector service:
  service.running:
    - name: splunk-otel-collector
    - enable: True
    - require:
      - pkg: splunk-otel-collector
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group
