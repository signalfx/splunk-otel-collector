# Removal of deprecated prometheusexec receiever
## Why is this happening?
There exist [security concerns](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6722) using the promethusexec receiver.  All prometheus specific functionality should be supported in the "normal" prometheus (scraping) receiver, along with others

## If I'm using the prometheusexec receiver, what should I do?

We recommend you migrate your configuration to use the [prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) opentelemetry receiver.

### Natively your metrics for scraping
The functionality and responsibility of instantiating and exposing an endpoint to be scraped from now lies with the end user, and is not supported in any of the current prometheus receivers.  The exact method of your migration
will depend on why you needed to use the `prometheusexec` scraper in the first place.

The easiest path would likely be to [enable federation](https://prometheus.io/docs/prometheus/latest/federation/) on your prometheus server, you can see our [prometheus-federation](./examples/prometheus-federation/README.md) example under [`examples/prometheus-federation`](./examples/prometheus-federation) in this git repository.  Ensure your endpoint is accessible from your otel collector, and feel free to ask in the [#otel-prometheus-wg](https://cloud-native.slack.com/archives/C01LSCJBXDZ), [#otel-collector](https://cloud-native.slack.com/archives/C01N6P7KR6W), or [#prometheus](https://cloud-native.slack.com/archives/C167KFM6C) for any help.  You can also feel free to cut us an issue, or reach our to your support rep.

If it's a matter of your data not natively being in the prometheus format, and you were using a prometheus exporter (not to be confused with the opentelemetry-collector-contrib `prometheusexporter`) to extract data, you can either run the relevant prometheus exporter ([list](https://prometheus.io/docs/instrumenting/exporters/)) on your own infrastructure, or better yet, you could check if there's already a native [opentelemetry receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver) for your use case!

### Moving configuration to the prometheus receiver

You may reuse the `scrape_interval` and `port` when migrating to the `prometheus` opentelemetry receiver.
Assuming your endpoint is on `localhost:8080`, your config may look something like this

```
receivers:
    prometheus:
        collection_interval: 10s
        endpoint: prometheus:8080
        metrics_path: /metrics # optional
```