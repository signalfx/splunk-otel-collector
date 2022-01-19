{% set splunk_fluentd_config = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config', '/etc/otel/collector/fluentd/fluent.conf') %}

Start FluentD service:
  service.running:
    - name: td-agent
    - enable: True
    - require:
      - pkg: Install FluentD
    - watch:
      - pkg: td-agent
      - file: {{ splunk_fluentd_config }}
