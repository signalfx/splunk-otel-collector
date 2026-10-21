{% set splunk_access_token = salt['pillar.get']('splunk-otel-collector:splunk_access_token', '') | trim %}
{% set splunk_platform_url = salt['pillar.get']('splunk-otel-collector:splunk_platform_url', '') | trim %}
{% set splunk_platform_token = salt['pillar.get']('splunk-otel-collector:splunk_platform_token', '') | trim %}
{% set splunk_platform_logs_index = salt['pillar.get']('splunk-otel-collector:splunk_platform_logs_index', '') | trim %}
{% set splunk_realm = salt['pillar.get']('splunk-otel-collector:splunk_realm', '') | trim %}

{% if grains['os_family'] not in ['Debian', 'RedHat', 'Suse'] %}
Check OS family:
  test.fail_without_changes:
    - name: "OS family ({{ grains['os_family'] }}) is not supported!"
    - failhard: True
{% elif (splunk_platform_url and not splunk_platform_token) or (splunk_platform_token and not splunk_platform_url) %}
Check Splunk Platform HEC settings:
  test.fail_without_changes:
    - name: "splunk_platform_url and splunk_platform_token must be provided together."
    - failhard: True
{% elif splunk_platform_logs_index and not splunk_platform_url %}
Check Splunk Platform logs index:
  test.fail_without_changes:
    - name: "splunk_platform_logs_index requires splunk_platform_url."
    - failhard: True
{% elif not splunk_access_token and not (splunk_platform_url and splunk_platform_token) %}
Check collector credentials:
  test.fail_without_changes:
    - name: "Provide splunk_access_token or both splunk_platform_url and splunk_platform_token."
    - failhard: True
{% elif splunk_access_token and not splunk_realm %}
Check splunk_realm:
  test.fail_without_changes:
    - name: "splunk_realm is not specified!"
    - failhard: True
{% endif %}

{% set install_auto_instrumentation = salt['pillar.get']('splunk-otel-collector:install_auto_instrumentation', False) | to_bool %}

include:
  - splunk-otel-collector.install
  - splunk-otel-collector.service_owner
  - splunk-otel-collector.config
  - splunk-otel-collector.collector_config
  - splunk-otel-collector.service
{% if install_auto_instrumentation %}
  - splunk-otel-collector.auto_instrumentation
{% endif %}
