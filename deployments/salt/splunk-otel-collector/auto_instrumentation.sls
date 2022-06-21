{% set auto_instrumentation_version = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_version', 'latest') %}
{% set auto_instrumentation_java_agent_path = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_java_agent_path', '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar') %}
{% set auto_instrumentation_ld_so_preload = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_ld_so_preload') %}
{% set auto_instrumentation_resource_attributes = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_resource_attributes') %}
{% set auto_instrumentation_service_name = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_service_name') %}

Install Splunk OpenTelemetry Auto Instrumentation:
  pkg.installed:
    - name: splunk-otel-auto-instrumentation
    - version: {{ auto_instrumentation_version }}
    - require:
      - pkg: splunk-otel-collector

/etc/ld.so.preload:
  file.managed:
    - contents:
        - /usr/lib/splunk-instrumentation/libsplunk.so
{% if auto_instrumentation_ld_so_preload %}
        - {{ auto_instrumentation_ld_so_preload }}
{% endif %}
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation

/usr/lib/splunk-instrumentation/instrumentation.conf:
  file.managed:
    - contents:
        - java_agent_jar={{ auto_instrumentation_java_agent_path }}
{% if auto_instrumentation_resource_attributes %}
        - resource_attributes={{ auto_instrumentation_resource_attributes }}
{% endif %}
{% if auto_instrumentation_service_name %}
        - service_name={{ auto_instrumentation_service_name }}
{% endif %}
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation
