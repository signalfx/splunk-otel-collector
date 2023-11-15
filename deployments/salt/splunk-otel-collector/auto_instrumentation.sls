{% set auto_instrumentation_version = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_version', 'latest') %}
{% set auto_instrumentation_systemd = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_systemd', False) | to_bool %}
{% set auto_instrumentation_java_agent_path = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_java_agent_path', '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar') %}
{% set auto_instrumentation_ld_so_preload = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_ld_so_preload') %}
{% set auto_instrumentation_resource_attributes = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_resource_attributes') %}
{% set auto_instrumentation_service_name = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_service_name') %}
{% set auto_instrumentation_generate_service_name = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_generate_service_name', True) | to_bool %}
{% set auto_instrumentation_disable_telemetry = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_disable_telemetry', False) | to_bool %}
{% set auto_instrumentation_enable_profiler = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_profiler', False) | to_bool %}
{% set auto_instrumentation_enable_profiler_memory = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_profiler_memory', False) | to_bool %}
{% set auto_instrumentation_enable_metrics = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_metrics', False) | to_bool %}
{% set auto_instrumentation_otlp_endpoint = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_otlp_endpoint', 'http://127.0.0.1:4317') %}

Install Splunk OpenTelemetry Auto Instrumentation:
  pkg.installed:
    - name: splunk-otel-auto-instrumentation
    - version: {{ auto_instrumentation_version }}
    - require:
      - pkg: splunk-otel-collector

/etc/ld.so.preload:
  file.managed:
    - contents: |
        {% if not auto_instrumentation_systemd %}
        /usr/lib/splunk-instrumentation/libsplunk.so
        {% endif %}
        {% if auto_instrumentation_ld_so_preload != "" %}
        {{ auto_instrumentation_ld_so_preload }}
        {% endif %}
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation

{% if auto_instrumentation_systemd %}
/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf:
  file.managed:
    - contents:
        - "[Manager]"
        - DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:{{ auto_instrumentation_java_agent_path }}"
        {% if auto_instrumentation_resource_attributes != "" %}
        - DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }}-systemd,{{ auto_instrumentation_resource_attributes }}"
        {% else %}
        - DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }}-systemd"
        {% endif %}
        {% if auto_instrumentation_service_name != "" %}
        - DefaultEnvironment="OTEL_SERVICE_NAME={{ auto_instrumentation_service_name }}"
        {% endif %}
        - DefaultEnvironment="SPLUNK_PROFILER_ENABLED={{ auto_instrumentation_enable_profiler | string | lower }}"
        - DefaultEnvironment="SPLUNK_PROFILER_MEMORY_ENABLED={{ auto_instrumentation_enable_profiler_memory | string | lower }}"
        - DefaultEnvironment="SPLUNK_METRICS_ENABLED={{ auto_instrumentation_enable_metrics | string | lower }}"
        - DefaultEnvironment="OTEL_EXPORTER_OTLP_ENDPOINT={{ auto_instrumentation_otlp_endpoint }}"
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% else %}
Delete auto instrumentation systemd config:
  file.absent:
    - name: /usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf

{% if auto_instrumentation_version == 'latest' or salt['pkg.version_cmp'](auto_instrumentation_version, '0.87.0') >= 0 %}
/etc/splunk/zeroconfig/java.conf:
  file.managed:
    - contents:
        - JAVA_TOOL_OPTIONS=-javaagent:{{ auto_instrumentation_java_agent_path }}
        {% if auto_instrumentation_resource_attributes != "" %}
        - OTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }},{{ auto_instrumentation_resource_attributes }}
        {% else %}
        - OTEL_RESOURCE_ATTRIBUTES=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }}
        {% endif %}
        {% if auto_instrumentation_service_name != "" %}
        - OTEL_SERVICE_NAME={{ auto_instrumentation_service_name }}
        {% endif %}
        - SPLUNK_PROFILER_ENABLED={{ auto_instrumentation_enable_profiler | string | lower }}
        - SPLUNK_PROFILER_MEMORY_ENABLED={{ auto_instrumentation_enable_profiler_memory | string | lower }}
        - SPLUNK_METRICS_ENABLED={{ auto_instrumentation_enable_metrics | string | lower }}
        - OTEL_EXPORTER_OTLP_ENDPOINT={{ auto_instrumentation_otlp_endpoint }}
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% else %}
/usr/lib/splunk-instrumentation/instrumentation.conf:
  file.managed:
    - contents:
        - java_agent_jar={{ auto_instrumentation_java_agent_path }}
        {% if auto_instrumentation_resource_attributes != "" %}
        - resource_attributes=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }},{{ auto_instrumentation_resource_attributes }}
        {% else %}
        - resource_attributes=splunk.zc.method=splunk-otel-auto-instrumentation-{{ auto_instrumentation_version }}
        {% endif %}
        {% if auto_instrumentation_service_name %}
        - service_name={{ auto_instrumentation_service_name }}
        {% endif %}
        - generate_service_name={{ auto_instrumentation_generate_service_name | string | lower }}
        - disable_telemetry={{ auto_instrumentation_disable_telemetry | string | lower }}
        - enable_profiler={{ auto_instrumentation_enable_profiler | string | lower }}
        - enable_profiler_memory={{ auto_instrumentation_enable_profiler_memory | string | lower }}
        - enable_metrics={{ auto_instrumentation_enable_metrics | string | lower }}
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% endif %}
{% endif %}

Reload systemd:
  cmd.run:
    - name: systemctl daemon-reload
    - onchanges:
        - file: /usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf
