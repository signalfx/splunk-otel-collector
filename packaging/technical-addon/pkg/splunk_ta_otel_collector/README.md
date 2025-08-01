# Splunk Add-on for OpenTelemetry Collector

This technical add-on for Splunk Universal Forwarder helps in deploying the OpenTelemetry Collector alongside Universal Forwarder. 

Splunk Add-on for OpenTelemetry Collector is supported on Linux and Windows on `amd64` (`x86_64`)

## Contents

The package contains the following folders and files -

------------|
                 |------windows_x86_64

------------|
                 |------linux_x86_64
                 
------------|
                 |------README
                 
------------|
                 |------default
                 
------------|
                 |------configs
            
------------|
                 |------README.md


The Windows and Linux folders contain platform specific binaries. The default folder contains the `app.conf` and `inputs.conf` files. The configs folder contains sample configuration files for collecting host metrics and traces using the OpenTelemetry Collector.

## Install the add-on

The installation process for the Splunk Add-on for OpenTelemetry Collector differs depending on the deployment method.

### Splunk Deployment Server

Follow these steps to install the add-on using Splunk Deployment Server:

1. Extract the `Splunk_TA_otel.tgz` file to the `$SPLUNK_HOME/etc/deployment-apps` folder. 
2. Edit the configuration files.
3. Follow the instructions for deploying apps. See https://docs.splunk.com/Documentation/Splunk/8.1.2/Updating/Updateconfigurations. 
4. Make sure to activate the restart of Universal Forwarder after deployment.

### Splunk Universal Forwarder

To install the add-on directly on Universal Forwarder, follow these steps:

1. Log in to Universal Forwarder.
2. Copy the tar file on the server.
3. Extract the package to the `$SPLUNK_HOME/etc/apps` folder:

```
tar -zxf Splunk_TA_otel.tgz
```

## Configure the add-on

Before using the add-on, create a `local` folder and write the contents of your Splunk  Observability Cloud access token to this file.
**WE STRONGLY RECOMMEND** that you restrict read and write to this token to the minimal possible permissions.
Alternatively, you can set the `SPLUNK_ACCESS_TOKEN` environment variable to avoid needing to write this.

```sh
cd /opt/splunkforwarder/etc/apps/Splunk_TA_otel/
mkdir local
cp -R config local
touch local/access_token
# make the contents of the access_token file your o11y access token.
```

After all the steps are completed, restart Splunk Universal Forwarder by running `$SPLUNK_HOME/bin/splunk.exe restart`

## Configure the OpenTelemetry Collector

To configure the OpenTelemetry Collector within the add-on, follow these steps:

1. Create a new configuration file in YAML format for the Collector. For more information, see [Configure the Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/configure-the-collector.html) in the Splunk Observability Cloud documentation.
2. Edit the inputs.conf file inside default to point to the new configuration file.
3. Restart Splunk Universal Forwarder.

Various settings in the `inputs.conf.spec` are "pass through" [environment variables](https://github.com/signalfx/splunk-otel-collector/blob/main/internal/settings/settings.go#L37-L64) to the default splunk distribution of the opentelemetry collector's agent configuration.  We do not currently pass through things related to `gateway` nor for the splunk hec related settings.  That said, the binary vended in the TA is identical to upstream

Note that we additionally configure the following environment variables in the hopes of making configuration easier.
- `$SPLUNK_OTEL_TA_HOME` References the location of the TA, ex `/opt/splunk/etc/apps/Splunk_TA_Otel`
- `$SPLUNK_OTEL_TA_PLATFORM_HOME` references the location of the platform-specific configuration for the TA, ex `/opt/splunk/etc/apps/Splunk_TA_Otel/linux_x86_64`.  By default the `splunk_bundle_dir` and `splunk_collectd_dir` options in `inputs.conf` references this environment variable.

Finally, setting `$SPLUNK_OTEL_TA_DEBUG` to anything other than an empty string will provide detailed logging messages during TA start up.


## Check operational status

Both the add-on and the OpenTelemetry Collector generate log files to indicate operational status. 

You can find the log files in the `$SPLUNK_HOME/var/log/splunk/` folder.

- The Splunk add-on log file is `Splunk_TA_otel.log`.
- By default, the OpenTelemetry Collector log file is `otel.log` (although you can override this path to whatever you please).

## Explore metrics and traces

You can browse the metrics and traces collected by the add-on in  in Splunk Observability Cloud. See the [Splunk Observability Cloud documentation for more information](https://docs.splunk.com/Observability).

If you're using the default configuration, you can search [metrics finder](https://docs.splunk.com/observability/en/metrics-and-metadata/metrics-finder-metadata-catalog.html#metrics-finder-and-metadata-catalog) for [`splunk.distribution:otel-ta`](https://app.signalfx.com/#/metrics?sources%5B%5D=splunk.distribution:otel-ta).

*note*: if you're using your own configuration, *please* continue to include the telemetry in the configuration.


Under `processors`
```
  resource/telemetry:
    attributes:
      - action: insert
        key: splunk.distribution
        value: otel-ta
```

Under `receivers`
```
  # This section is used to collect the OpenTelemetry Collector metrics
  # Even if just a Splunk APM customer, these metrics are included
  prometheus/internal:
    config:
      scrape_configs:
      - job_name: 'otel-collector'
        scrape_interval: 10s
        static_configs:
        - targets: ["${env:SPLUNK_LISTEN_INTERFACE}:8888"]
        metric_relabel_configs:
          - source_labels: [ __name__ ]
            regex: '.*grpc_io.*'
            action: drop
```

Under `pipelines`, assuming you're using a `signalfx` exporter (only `otlp` is supported for our internal telemetry metrics)
```
    metrics/telemetry:
      receivers: [prometheus/internal]
      processors: [memory_limiter, batch, resourcedetection, resource/telemetry]
      exporters: [signalfx]
```

# Binary File Declaration
`linux_x86_64/bin/otelcol_linux_amd64`
`windows_x86_64/bin/otelcol_windows_amd64.exe`
