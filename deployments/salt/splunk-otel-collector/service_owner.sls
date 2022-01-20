{% set splunk_service_user = salt['pillar.get']('splunk-otel-collector:splunk_service_user', 'splunk-otel-collector') %}

{% set splunk_service_group = salt['pillar.get']('splunk-otel-collector:splunk_service_group', 'splunk-otel-collector') %}

splunk_service_group:
  group.present:
    - name: {{ splunk_service_group }}
    - system: True
    - unless: getent group {{ splunk_service_group }}

splunk_service_user:
  user.present:
    - name: {{ splunk_service_user }}
    - system: True
{%- if grains['os_family'] == 'Debian' %}
    - shell: /usr/sbin/nologin
{% else %}
    - shell: /sbin/nologin
{% endif %}
    - createhome: /etc/otel/collector
    - groups:
      - {{ splunk_service_group }}
    - unless: getent passwd {{ splunk_service_user }}
    - watch:
      - group: splunk_service_group

/etc/tmpfiles.d/splunk-otel-collector.conf:
  file.managed:
    - contents: |
        D /run/splunk-otel-collector 0755 {{ splunk_service_user }} {{ splunk_service_group }} - -
    - makedirs: True
    - mode: '0644'
    - watch:
      - user: splunk_service_user
      - group: splunk_service_group

Set systemd service owner for Splunk Otel Collector:
  file.managed:
    - name: /etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf
    - contents: |
        [Service]
        User={{ splunk_service_user }}
        Group={{ splunk_service_group }}
    - makedirs: True
    - mode: '0644'
    - watch:
      - user: splunk_service_user
      - group: splunk_service_group

Stop Service:
  service.dead:
    - name: splunk-otel-collector
    - onchanges:
      - file: /etc/tmpfiles.d/splunk-otel-collector.conf
      - file: /etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf

init tmpfile:
  cmd.run:
    - name: systemd-tmpfiles --create --remove /etc/tmpfiles.d/splunk-otel-collector.conf
    - onchanges:
      - file: /etc/tmpfiles.d/splunk-otel-collector.conf

Reload service:
  cmd.run:
    - name: systemctl daemon-reload
    - onchanges:
      - file: /etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf
