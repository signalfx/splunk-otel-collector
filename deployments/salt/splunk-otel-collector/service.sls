{% set splunk_otel_collector_config = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config', '/etc/otel/collector/agent_config.yaml') %}

{% set splunk_service_user = salt['pillar.get']('splunk-otel-collector:splunk_service_user', 'splunk-otel-collector') %}

{% set splunk_service_group = salt['pillar.get']('splunk-otel-collector:splunk_service_group', 'splunk-otel-collector') %}
{% set splunk_platform_url = salt['pillar.get']('splunk-otel-collector:splunk_platform_url', '') | trim %}

Ensure /etc/otel/collector is owned by the service user/group:
  file.directory:
    - name: /etc/otel/collector
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - mode: '0755'
    - recurse:
      - user
      - group
    - require:
      - pkg: splunk-otel-collector
      - service: Stop Service
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group

Ensure /var/lib/otelcol is owned by the service user/group:
  file.directory:
    - name: /var/lib/otelcol
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - mode: '0755'
    - recurse:
      - user
      - group
    - require:
      - pkg: splunk-otel-collector
      - service: Stop Service
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group

{% if splunk_platform_url %}
Ensure Splunk Platform log file storage directory:
  file.directory:
    - name: /var/lib/otelcol/filelogs
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - mode: '0700'
    - require:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group
    - watch:
      - pkg: splunk-otel-collector
      - user: splunk_service_user
      - group: splunk_service_group
{% endif %}

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
      - file: Ensure /etc/otel/collector is owned by the service user/group
      - file: Ensure /var/lib/otelcol is owned by the service user/group
{% if splunk_platform_url %}
      - file: Ensure Splunk Platform log file storage directory
{% endif %}
      - file: {{ splunk_otel_collector_config }}
      - file: /etc/otel/collector/splunk-otel-collector.conf
