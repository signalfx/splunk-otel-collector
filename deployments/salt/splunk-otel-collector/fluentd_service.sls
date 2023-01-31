{% set splunk_fluentd_config = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config', '/etc/otel/collector/fluentd/fluent.conf') %}

Ensure /etc/otel/collector/fluentd is owned by the td-agent user/group:
  file.directory:
    - name: /etc/otel/collector/fluentd
    - user: td-agent
    - group: td-agent
    - recurse:
      - user
      - group
    - require:
      - pkg: splunk-otel-collector
      - pkg: td-agent
    - watch:
      - pkg: splunk-otel-collector
      - pkg: td-agent

Start FluentD service:
  service.running:
    - name: td-agent
    - enable: True
    - require:
      - pkg: Install FluentD
    - watch:
      - pkg: td-agent
      - file: {{ splunk_fluentd_config }}
