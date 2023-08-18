# Removal of deprecated prometheusexec receiever
## Why is this happening?
There exist [security concerns](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6722) using the promethusexec receiver.  All prometheus specific functionality should be supported in the "normal" prometheus (scraping) receiver, along with others.

## If I'm using the prometheusexec receiver, what should I do?

We recommend you migrate your configuration to use the [prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) opentelemetry receiver.

### Natively your metrics for scraping
The exact method of your migration will depend on why you needed to use the `prometheus_exec` scraper in the first place.


#### Prometheus exporters
The intended use case for the `prometheus_exec` receiver was to leverage the rich existing community of prometheus exporters (not to be confused with the opentelemetry-collector-contrib `prometheusexporter`). If this was your use case, we first reccomend checking to see if there is a native open-telemetry collector receiver in [core](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver), [contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver), or our [splunk distribution](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver).

If no equivalent reciever exists for a given prometheus exporter, or if you simply wish to continue using prometheus exporters, you will need to manually run the relevant prometheus exporter ([list](https://prometheus.io/docs/instrumenting/exporters/)) on your own infrastructure.  While the details of how you decide to deploy and manage your infrastructure is beyond the scope of this migration, we have provided an [example configuration](../../examples/prometheusexec-migration/README.md) using the [node exporter](https://github.com/prometheus/node_exporter#readme) in `examples/prometheusexec-migration` in this git repo.

#### Non-standard usage
If you needed to federate between a prometheus server and otel, you can [enable federation](https://prometheus.io/docs/prometheus/latest/federation/) on your prometheus server, you can see our [prometheus-federation](./examples/prometheus-federation/README.md) example under [`examples/prometheus-federation`](./examples/prometheus-federation) in this git repository.  Ensure your endpoint is accessible from your otel collector, and feel free to ask in the [#otel-prometheus-wg](https://cloud-native.slack.com/archives/C01LSCJBXDZ), [#otel-collector](https://cloud-native.slack.com/archives/C01N6P7KR6W), or [#prometheus](https://cloud-native.slack.com/archives/C167KFM6C) slack channels for any help.  You can also feel free to cut us an issue, or reach out to your support representative.

For more novel use cases such as running arbitrary code, this is no longer supported.

### Moving configuration to the prometheus receiver

You may reuse the `scrape_interval` and `port` when migrating to the `prometheus` opentelemetry receiver.
Assuming your endpoint is on `localhost:8080`, your config may look something like this

```yaml
receivers:
    prometheus:
        config:
            scrape_configs:
              - job_name: 'scrape-prometheus'
                scrape_interval: 5s
                static_configs:
                  - targets: ['localhost:8080']
```

For further configuration options, please see the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver#readme).
