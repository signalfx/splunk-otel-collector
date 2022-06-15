{% set os_family = grains['os_family'] %}
{% set splunk_repo_base_url = salt['pillar.get']('splunk-otel-collector:repo_base_url', 'https://splunk.jfrog.io/splunk') %}
{% set package_stage = salt['pillar.get']('splunk-otel-collector:package_stage', 'release') %}
{% set collector_version = salt['pillar.get']('splunk-otel-collector:collector_version', 'latest') %}

# Repository configuration.

{% if os_family == 'RedHat' %}

Add Splunk OpenTelemetry Collector repo to yum source list:
  pkgrepo.managed:
    - name: 'splunk-otel-collector-yum-repo'
    - humanname: Splunk OpenTelemetry Collector Repository
    - baseurl: {{ splunk_repo_base_url }}/otel-collector-rpm/{{ package_stage }}/$basearch/
    - gpgkey: {{ splunk_repo_base_url }}/otel-collector-rpm/splunk-B3CD4420.pub
    - gpgcheck: 1
    - enabled: 1

Install setcap via yum package manager:
  pkg.latest:
    - pkgs:
      - libcap

{% elif os_family == 'Debian' %}

Add Splunk OpenTelemetry Collector repo to apt source list:
  pkgrepo.managed:
    - name: deb {{ splunk_repo_base_url }}/otel-collector-deb {{ package_stage }} main
    - file: /etc/apt/sources.list.d/splunk-otel-collector.list
    - key_url: {{ splunk_repo_base_url }}/otel-collector-deb/splunk-B3CD4420.gpg
    - refresh: True
    - gpgcheck: 1
    - enabled: 1

Install apt dependencies for secure transport:
  pkg.latest:
    - pkgs:
      - apt-transport-https
      - gnupg

{% elif os_family == 'Suse' %}

Install setcap via zypper package manager:
  pkg.latest:
    - pkgs:
      - libcap-progs
    - refresh: True

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

# Installation of splunk-otel-collector package and starting of service.

Install Splunk OpenTelemetry Collector:
  pkg.installed:
    - name: splunk-otel-collector
    - version: {{ collector_version }}
