{% set splunk_fluentd_config = salt['pillar.get']('splunk-otel-collector:splunk_fluentd_config', '/etc/otel/collector/fluentd/fluent.conf') %}

{% set fluentd_repo_base = salt['pillar.get']('splunk-otel-collector:fluentd_repo_base', 'https://packages.treasuredata.com') %}

{% set td_agent_version = salt['pillar.get']('splunk-otel-collector:fluentd_repo_base') %}

{% if td_agent_version  %}
{{ td_agent_version }}
{%- elif grains['oscodename'] == 'stretch' %}
{% set td_agent_version = '3.7.1-0' %}
{%- elif grains['os_family'] == 'Debian' %}
{% set td_agent_version = '4.1.1-1' %}
{%- else %}
{% set td_agent_version = '4.1.1' %}
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

{%- if grains['os'] == 'Amazon' %}
{% set distro =  'amazon' %}
{%- else %}
{% set distro =  'redhat' %}
{%- endif %}

Import td-agent GPG Key:
  cmd.run:
    - name: rpm --import {{ fluentd_repo_base }}/GPG-KEY-td-agent

Setup FluentD repository:
  file.managed:
    - name: /etc/yum.repos.d/td-agent.repo
    - contents: |
        [td-agent]
        name = TreasureData Repository
        baseurl = {{ fluentd_repo_base }}/{{ td_agent_major_version }}/{{ distro }}/$releasever/$basearch
        gpgcheck = 1
        gpgkey = {{ fluentd_repo_base }}/GPG-KEY-td-agent
        enabled = 1
    - makedirs: True
    - mode: '0644'

Install Developement Tools:
  pkg.group_installed:
    - name: Development Tools

Install FluentD Linux capability module dependencies:
  pkg.latest:
    - pkgs:
      - libcap-ng
      - libcap-ng-devel
{%- if grains['osmajorrelease'] == 8 %}
      - pkgconf-pkg-config
{%- else %}
      - pkgconfig
{%- endif %}
    - require:
      - pkg: Install Developement Tools
{%- endif %}

Install FluentD:
  pkg.installed:
    - name: td-agent
    - version: {{ td_agent_version }}

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

Install FluentD systemd plugin:
  cmd.run:
    - name: td-agent-gem install fluent-plugin-systemd -v 1.0.1
    - require:
      - pkg: Install FluentD Linux capability module dependencies

Reload td-agent service:
  cmd.run:
    - name: systemctl daemon-reload
    - onchanges:
      - file: /etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf

Start FluentD service:
  service.running:
    - name: td-agent
    - require:
      - pkg: Install FluentD
