> The official Splunk documentation for this page is [Install and configure Splunk Distribution of OpenTelemetry Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/opentelemetry.html). For instructions on how to contribute to the docs, see [CONTRIBUTE.md](../CONTRIBUTE.md).

# Sending logs directly from Fluentd to Splunk Log Observer  

> WARNING: The following is not supported or tested today. Performance testing,
> sizing guidelines, and latency impact are not provided. Multiline messages
> may not work properly. Other limitations or issues may exist.

Use-case: If you have an existing signalfx-agent collecting metrics and/or 
traces and want to try Splunk Log Observer without migrating to Splunk 
OpenTelemetry Connector, you can use fluentd to send logs directly to Splunk Log
Observer. 

## New fluentd deployment on Linux

1. Follow the official documentation to install fluentd:
    https://docs.fluentd.org/installation.
   
    Recommended way is to follow a Linux distribution specific installation.
   
1. Install Splunk HEC plugin for Fluentd. 
   
    If td-agent was installed with a package manager, use the following command:
    ```sh
    sudo td-agent-gem install fluent-plugin-splunk-hec
    ```

    If fluentd was installed with ruby gem:
    ```sh
    gem install fluent-plugin-splunk-hec
    ```

1. Give the `td-agent` user access to systemd logs

    ```sh
    sudo usermod -a -G systemd-journal td-agent
    ```
    
    Assuming that the `/var/log/journal` path is only readable by the 
    `systemd-journal` group, otherwise change the command accordingly

1. Set the following fluentd configuration in the td-agent config file
    **/etc/td-agent/td-agent.conf** (or **./fluent/fluent.conf** in case of ruby 
    gem installation):

    ```
    <source>
      @id journald
      @type systemd
      @label @SPLUNK
      tag "journald"
      path "/var/log/journal"
      read_from_head false
      <storage>
        @type local
        persistent true
        path /var/log/td-agent/journald.pos.json
      </storage>
      <entry>
        field_map {"_HOSTNAME": "host.name"}
        fields_strip_underscores true
        fields_lowercase true
      </entry>
    </source>

    <label @SPLUNK>
      <match **>
        @type splunk_hec
        hec_host "ingest.<SPLUNK_REALM>.signalfx.com"
        hec_port 443
        hec_token "<SPLUNK_TOKEN>"
        data_type event
        source ${tag}
        sourcetype _json
        <buffer>
          @type memory
          total_limit_size 600m
          chunk_limit_size 1m
          chunk_limit_records 100000
          flush_interval 5s
          flush_thread_count 1
          overflow_action block
          retry_max_times 3
        </buffer>
      </match>
    </label>

    <system>
      log_level info
    </system>
    ```

    Make sure to replace `<SPLUNK_REALM>` with a Splunk realm you are going to 
    send data to and `<SPLUNK_TOKEN>` with your Splunk Access Token.

    With this configuration fluentd will be collecting systemd logs and sending 
    them to Splunk Log Observer.
    
    If you want to collect logs from other sources, add specific sections from 
    https://github.com/signalfx/splunk-otel-collector/tree/main/internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd/conf.d
    to the config file.

1. Restart fluentd service to apply the changes

   For Ubuntu deployment: `sudo service td-agent restart`

   For CentOS deployment: `sudo /etc/init.d/td-agent start`

   For systemd-based deployment: `sudo systemctl start td-agent.service`

## Kubernetes deployment

Using [Splunk OpenTelemetry Helm chart](https://github.com/signalfx/splunk-otel-collector-chart#disable-particular-types-of-telemetry)
you can disable logs and traces collection by using the following configuration options:

```
metricsEnabled: false
tracesEnabled: false
```