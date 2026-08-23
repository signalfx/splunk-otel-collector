# Splunk App Observer

The Splunk App Observer reports each direct child directory under `apps_path` as
an observer endpoint. It is intended for use with `receiver_creator` so receivers
can be started when Splunk apps appear and stopped when those app directories are
removed.

By default, `apps_path` is `$SPLUNK_HOME/etc/apps`. The observer polls the
directory every `refresh_interval`, which defaults to `10s`.

```yaml
extensions:
  splunk_app_observer:
    apps_path: /opt/splunk/etc/apps

receivers:
  receiver_creator:
    watch_observers: [splunk_app_observer]
    receivers:
      filelog/splunk_app:
        rule: type == "splunk_app" && name == "Splunk_TA_example"
        config:
          include:
            - '`endpoint`/local/*.log'
```

Endpoint environment fields include:

| Field | Value |
| --- | --- |
| `type` | `container` |
| `name` | Splunk app directory name |
| `image` | `splunk_app` |
| `command` | Full Splunk app directory path |
| `endpoint` | Full Splunk app directory path |
| `container_id` | Full Splunk app directory path |
| `splunk_app_name` | Splunk app directory name |
| `splunk_app_path` | Full Splunk app directory path |
| `labels["splunk_app_name"]` | Splunk app directory name |
| `labels["splunk_app_path"]` | Full Splunk app directory path |
