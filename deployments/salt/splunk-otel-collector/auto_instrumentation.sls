{% set auto_instrumentation_version = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_version', 'latest') %}
{% set auto_instrumentation_systemd = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_systemd', False) | to_bool %}
{% set auto_instrumentation_sdks = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_sdks', ['java', 'nodejs', 'dotnet']) %}
{% set auto_instrumentation_java_agent_path = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_java_agent_path', '/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar') %}
{% set auto_instrumentation_npm_path = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_npm_path', 'npm') %}
{% set auto_instrumentation_ld_so_preload = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_ld_so_preload') %}
{% set auto_instrumentation_resource_attributes = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_resource_attributes') %}
{% set auto_instrumentation_service_name = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_service_name') %}
{% set auto_instrumentation_generate_service_name = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_generate_service_name', True) | to_bool %}
{% set auto_instrumentation_disable_telemetry = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_disable_telemetry', False) | to_bool %}
{% set auto_instrumentation_enable_profiler = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_profiler', False) | to_bool %}
{% set auto_instrumentation_enable_profiler_memory = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_profiler_memory', False) | to_bool %}
{% set auto_instrumentation_enable_metrics = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_enable_metrics', False) | to_bool %}
{% set auto_instrumentation_otlp_endpoint = salt['pillar.get']('splunk-otel-collector:auto_instrumentation_otlp_endpoint', 'http://127.0.0.1:4317') %}
{% set with_new_instrumentation = auto_instrumentation_version == 'latest' or salt['pkg.version_cmp'](auto_instrumentation_version, '0.87.0') >= 0 %}
{% set dotnet_supported = (auto_instrumentation_version == 'latest' or salt['pkg.version_cmp'](auto_instrumentation_version, '0.99.0') >= 0) and grains['cpuarch'] in ['amd64', 'x86_64'] %}
{% set systemd_config_path = '/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf' %}
{% set old_instrumentation_config_path = '/usr/lib/splunk-instrumentation/instrumentation.conf' %}
{% set java_config_path = '/etc/splunk/zeroconfig/java.conf' %}
{% set nodejs_config_path = '/etc/splunk/zeroconfig/node.conf' %}
{% set dotnet_config_path = '/etc/splunk/zeroconfig/dotnet.conf' %}
{% set nodejs_prefix = '/usr/lib/splunk-instrumentation/splunk-otel-js' %}
{% set dotnet_home = '/usr/lib/splunk-instrumentation/splunk-otel-dotnet' %}

Install Splunk OpenTelemetry Auto Instrumentation:
  pkg.installed:
    - name: splunk-otel-auto-instrumentation
    - version: {{ auto_instrumentation_version }}
    - require:
      - pkg: splunk-otel-collector

{% if 'nodejs' in auto_instrumentation_sdks and with_new_instrumentation %}
{{ nodejs_prefix }}/node_modules:
  file.directory:
    - makedirs: True
    - require:
      - pkg: splunk-otel-auto-instrumentation

Install splunk-otel-js:
  cmd.run:
    - name: {{ auto_instrumentation_npm_path}} install --global=false /usr/lib/splunk-instrumentation/splunk-otel-js.tgz
    - cwd: {{ nodejs_prefix }}
    - require:
      - file: {{ nodejs_prefix }}/node_modules
{% endif %}

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
{{ systemd_config_path }}:
  file.managed:
    - contents:
        - "[Manager]"
        {% if 'java' in auto_instrumentation_sdks %}
        - DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:{{ auto_instrumentation_java_agent_path }}"
        {% endif %}
        {% if 'nodejs' in auto_instrumentation_sdks and with_new_instrumentation %}
        - DefaultEnvironment="NODE_OPTIONS=-r {{ nodejs_prefix }}/node_modules/@splunk/otel/instrument"
        {% endif %}
        {% if 'dotnet' in auto_instrumentation_sdks and dotnet_supported %}
        - DefaultEnvironment="CORECLR_ENABLE_PROFILING=1"
        - DefaultEnvironment="CORECLR_PROFILER={918728DD-259F-4A6A-AC2B-B85E1B658318}"
        - DefaultEnvironment="CORECLR_PROFILER_PATH={{ dotnet_home }}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
        - DefaultEnvironment="DOTNET_ADDITIONAL_DEPS={{ dotnet_home }}/AdditionalDeps"
        - DefaultEnvironment="DOTNET_SHARED_STORE={{ dotnet_home }}/store"
        - DefaultEnvironment="DOTNET_STARTUP_HOOKS={{ dotnet_home }}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll"
        - DefaultEnvironment="OTEL_DOTNET_AUTO_HOME={{ dotnet_home }}"
        - DefaultEnvironment="OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation"
        {% endif %}
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

{% for config in [java_config_path, nodejs_config_path, dotnet_config_path, old_instrumentation_config_path] %}
Delete {{ config }}:
  file.absent:
    - name: {{ config }}
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% endfor %}
{% else %}
Delete auto instrumentation systemd config:
  file.absent:
    - name: {{ systemd_config_path }}

{% if with_new_instrumentation %}
{% if 'java' in auto_instrumentation_sdks %}
{{ java_config_path }}:
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
Delete {{ java_config_path }}:
  file.absent:
    - name:
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% endif %}
{% if 'nodejs' in auto_instrumentation_sdks and with_new_instrumentation %}
{{ nodejs_config_path }}:
  file.managed:
    - contents:
        - NODE_OPTIONS=-r {{ nodejs_prefix }}/node_modules/@splunk/otel/instrument
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
Delete {{ nodejs_config_path }}:
  file.absent:
    - name: {{ nodejs_config_path }}
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% endif %}
{% if 'dotnet' in auto_instrumentation_sdks and dotnet_supported %}
{{ dotnet_config_path }}:
  file.managed:
    - contents:
        - CORECLR_ENABLE_PROFILING=1
        - CORECLR_PROFILER={918728DD-259F-4A6A-AC2B-B85E1B658318}
        - CORECLR_PROFILER_PATH={{ dotnet_home }}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so
        - DOTNET_ADDITIONAL_DEPS={{ dotnet_home }}/AdditionalDeps
        - DOTNET_SHARED_STORE={{ dotnet_home }}/store
        - DOTNET_STARTUP_HOOKS={{ dotnet_home }}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll
        - OTEL_DOTNET_AUTO_HOME={{ dotnet_home }}
        - OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation
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
Delete {{ dotnet_config_path }}:
  file.absent:
    - name: {{ dotnet_config_path }}
    - require:
      - pkg: splunk-otel-auto-instrumentation
{% endif %}
{% else %}
{{ old_instrumentation_config_path }}:
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
        - file: {{ systemd_config_path }}
