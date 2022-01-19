{% set splunk_service_user = salt['pillar.get']('splunk-otel-collector:service_user', 'splunk-otel-collector') %}

{% set splunk_service_group = salt['pillar.get']('splunk-otel-collector:service_group', 'splunk-otel-collector') %}

{% set splunk_access_token = salt['pillar.get']('splunk-otel-collector:splunk_access_token') %}

{% set splunk_realm = salt['pillar.get']('splunk-otel-collector:splunk_realm', 'us0') %}

{% set splunk_api_url = salt['pillar.get']('splunk-otel-collector:splunk_api_url', 'https://api.' + splunk_realm + '.signalfx.com' ) %}

{% set splunk_ingest_url = salt['pillar.get']('splunk-otel-collector:splunk_ingest_url', 'https://ingest.' + splunk_realm + '.signalfx.com') %}

{% set splunk_trace_url = salt['pillar.get']('splunk-otel-collector:splunk_trace_url', splunk_ingest_url + '/v2/trace') %}

{% set splunk_hec_url = salt['pillar.get']('splunk-otel-collector:splunk_hec_url', splunk_ingest_url + '/v1/log') %}

{% set splunk_hec_token = salt['pillar.get']('splunk-otel-collector:splunk_hec_token', splunk_access_token) %}

{% set splunk_otel_collector_config = salt['pillar.get']('splunk-otel-collector:splunk_otel_collector_config', '/etc/otel/collector/agent_config.yaml') %}

{% set splunk_collectd_dir = salt['pillar.get']('splunk-otel-collector:splunk_collectd_dir', '/usr/lib/splunk-otel-collector/agent-bundle/run/collectd') %}

{% set splunk_bundle_dir = salt['pillar.get']('splunk-otel-collector:splunk_bundle_dir', '/usr/lib/splunk-otel-collector/agent-bundle') %}

{% set splunk_memory_total_mib = salt['pillar.get']('splunk-otel-collector:splunk_memory_total_mib', '512') %}

{% set splunk_ballast_size_mib = salt['pillar.get']('splunk-otel-collector:splunk_ballast_size_mib', '') %}

/etc/otel/collector/splunk-otel-collector.conf:
  file.managed:
    - contents: |
        SPLUNK_CONFIG={{ splunk_otel_collector_config }}
        SPLUNK_ACCESS_TOKEN={{ splunk_access_token }}
        SPLUNK_REALM={{ splunk_realm }}
        SPLUNK_API_URL={{ splunk_api_url }}
        SPLUNK_INGEST_URL={{ splunk_ingest_url }}
        SPLUNK_TRACE_URL={{ splunk_trace_url }}
        SPLUNK_HEC_URL={{ splunk_hec_url }}
        SPLUNK_HEC_TOKEN={{ splunk_hec_token }}
        SPLUNK_MEMORY_TOTAL_MIB={{ splunk_memory_total_mib }}
        SPLUNK_BALLAST_SIZE_MIB={{ splunk_ballast_size_mib }}
        SPLUNK_BUNDLE_DIR={{ splunk_bundle_dir }}
        SPLUNK_COLLECTD_DIR={{ splunk_collectd_dir }}
    - mode: '0600'
    - makedirs: True
    - user: {{ splunk_service_user }}
    - group: {{ splunk_service_group }}
    - watch:
      - user: splunk_service_user
      - group: splunk_service_group
