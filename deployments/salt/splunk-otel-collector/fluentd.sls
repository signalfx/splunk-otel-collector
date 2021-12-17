{% set splunk_fluentd_config = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config', '/etc/otel/collector/fluentd/fluent.conf') %}

{% set fluentd_repo_base = salt['pillar.get']('splunk-otel-collector:fluentd_repo_base', 'http://packages.treasuredata.com') %}

{%- if grains['oscodename'] == 'jessie' %}
{% set td_agent_version = salt['pillar.get']('splunk-otel-collector:td_agent_version', '3.3.0-1') %}
{%- elif grains['oscodename'] == 'stretch' %}
{% set td_agent_version = salt['pillar.get']('splunk-otel-collector:td_agent_version', '3.7.1-0') %}
{%- elif grains['os_family'] == 'Debian' %}
{% set td_agent_version = salt['pillar.get']('splunk-otel-collector:td_agent_version', '4.1.1-1') %}
{%- else %}
{% set td_agent_version = salt['pillar.get']('splunk-otel-collector:td_agent_version', '4.1.1') %}
{%- endif %}

{% set td_agent_major_version = td_agent_version.split('.')[0] %}

{%- if grains['os_family'] == 'Debian' %}
Setup FluentD repository:
  pkgrepo.managed:
    - file: /etc/apt/sources.list.d/treasure-data.list
    - name: deb {{ fluentd_repo_base }}/{{ td_agent_major_version }}/{{ grains['os']|lower }}/{{ grains['oscodename'] }}/ {{ grains['oscodename'] }} contrib
    - refresh: True
    - key_url: {{ fluentd_repo_base }}/GPG-KEY-td-agent

Install FluentD Linux capability module dependencies:
  pkg.latest:
    - pkgs:
      - build-essential
      - libcap-ng0
      - libcap-ng-dev
      - pkg-config

{%- endif %}

{%- if grains['os_family'] == 'RedHat' %}
Setup FluentD repository:
  pkgrepo.managed:
    - name: treasuredata
    - humanname: TreasureData
{%- if grains['os'] in ['RedHat', 'CentOS'] %}
    - baseurl: {{ fluentd_repo_base }}/{{ td_agent_major_version }}/redhat/$basearch
{%- elif grains['os'] == 'Amazon' %}
    - baseurl: {{ fluentd_repo_base }}/{{ td_agent_major_version }}/amazon/2/$basearch
{%- endif %}
    - gpgkey: {{ fluentd_repo_base }}/GPG-KEY-td-agent
    - gpgcheck: 1
  cmd.run:
    - name: rpm --import {{ fluentd_repo_base }}/GPG-KEY-td-agent

Install Developement Tools:
  pkg.group_installed:
    - name: Development Tools

Install FluentD Linux capability module dependencies:
  pkg.latest:
    - pkgs:
      - libcap-ng
      - libcap-ng-devel
      - pkgconfig
    - require:
      - pkg: Install Developement Tools
{%- endif %}

Install FluentD:
  pkg.installed:
    - name: td-agent
    - version: {{ td_agent_version }}
    - require:
      - pkgrepo: Setup FluentD repository

/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf:
  file.managed:
    - contents: |
        [service]
        Environment=FLUENT_CONF={{ splunk_fluentd_config }}
    - makedirs: True
    - mode: '0644'

Install capng_c fluentd plugin:
  cmd.run:
    - name: td-agent-gem install capng_c -v 0.2.2
    - require:
      - pkg: Install FluentD Linux capability module dependencies

Install fluent-plugin-systemd:
  cmd.run:
    - name: td-agent-gem install fluent-plugin-systemd -v 1.0.1
    - require:
      - pkg: Install FluentD Linux capability module dependencies

Start FluentD service:
  service.running:
    - name: td-agent
    - require:
      - pkg: Install FluentD
