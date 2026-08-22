{% set os_family = grains['os_family'] %}
{% set splunk_repo_base_url = salt['pillar.get']('splunk-otel-collector:repo_base_url', 'https://splunk.jfrog.io/splunk') %}
{% set package_stage = salt['pillar.get']('splunk-otel-collector:package_stage', 'release') %}
{% set collector_version = salt['pillar.get']('splunk-otel-collector:collector_version', 'latest') %}
{% set debian_gpg_key_path = '/etc/apt/keyrings/splunk-otel-collector.gpg' %}
{% set local_artifact_testing_enabled = salt['pillar.get']('splunk-otel-collector:local_artifact_testing_enabled', False) | to_bool %}
{% set collector_package_source = salt['pillar.get']('splunk-otel-collector:collector_package_source', '') %}
{% set zypper_local_artifact_testing_enabled = local_artifact_testing_enabled and salt['cmd.has_exec']('zypper') %}

# Repository configuration.

{% if os_family == 'RedHat' %}

{% if not local_artifact_testing_enabled %}
Add Splunk OpenTelemetry Collector repo to yum source list:
  pkgrepo.managed:
    - name: 'splunk-otel-collector-yum-repo'
    - humanname: Splunk OpenTelemetry Collector Repository
    - baseurl: {{ splunk_repo_base_url }}/otel-collector-rpm/{{ package_stage }}/$basearch/
    - gpgkey: {{ splunk_repo_base_url }}/otel-collector-rpm/splunk-B3CD4420.pub
    - gpgcheck: 1
    - enabled: 1
{% endif %}

Install setcap via yum package manager:
  pkg.latest:
    - pkgs:
      - libcap

{% elif os_family == 'Debian' %}

/etc/apt/keyrings:
  file.directory:
    - user: root
    - group: root
    - mode: '0755'

{% if not local_artifact_testing_enabled %}
Add Splunk OpenTelemetry Collector repo to apt source list:
  pkgrepo.managed:
    - name: deb [signed-by={{ debian_gpg_key_path }}] {{ splunk_repo_base_url }}/otel-collector-deb {{ package_stage }} main
    - file: /etc/apt/sources.list.d/splunk-otel-collector.list
    - key_url: {{ splunk_repo_base_url }}/otel-collector-deb/splunk-B3CD4420.gpg
    - aptkey: False
    - refresh: True
    - gpgcheck: 1
    - enabled: 1
    - require:
      - file: /etc/apt/keyrings
{% endif %}

Install apt dependencies for secure transport:
  pkg.latest:
    - pkgs:
      - apt-transport-https
      - gnupg
      - libcap2-bin

{% elif os_family == 'Suse' %}

Install setcap via zypper package manager:
  pkg.latest:
    - pkgs:
      - libcap-progs
    - refresh: True

{% if not local_artifact_testing_enabled %}
Import the Splunk GPG key:
  cmd.run:
    - name: rpm --import {{ splunk_repo_base_url }}/otel-collector-rpm/splunk-B3CD4420.pub

Add Splunk OpenTelemetry Collector repo to zypper source list:
  file.managed:
    - name: /etc/zypp/repos.d/splunk-otel-collector.repo
    - contents: |
        [splunk-otel-collector]
        autorefresh = 0
        baseurl = {{ splunk_repo_base_url }}/otel-collector-rpm/{{ package_stage }}/$basearch/
        enabled = 1
        gpgcheck = 1
        gpgkey = {{ splunk_repo_base_url }}/otel-collector-rpm/splunk-B3CD4420.pub
        name = Splunk OpenTelemetry Collector Repository
        type = rpm-md
    - makedirs: True
{% endif %}

{% endif %}

# Installation of splunk-otel-collector package and starting of service.

{% if zypper_local_artifact_testing_enabled %}
Install local splunk-otel-collector package:
  cmd.run:
    - name: zypper --non-interactive --no-gpg-checks install -y --allow-unsigned-rpm {{ collector_package_source }}
    - unless: rpm -q splunk-otel-collector

{% endif %}

splunk-otel-collector:
  pkg.installed:
{% if local_artifact_testing_enabled %}
{% if zypper_local_artifact_testing_enabled %}
    - name: splunk-otel-collector
    - require:
      - cmd: Install local splunk-otel-collector package
{% else %}
    - sources:
      - splunk-otel-collector: {{ collector_package_source }}
    - skip_verify: True
{% endif %}
{% else %}
    - name: splunk-otel-collector
    - version: {{ collector_version }}
{% endif %}
