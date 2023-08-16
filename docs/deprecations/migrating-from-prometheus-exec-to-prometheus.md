# Removal of deprecated prometheusexec receiever
## Why is this happening?
There exist [security concerns](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6722) using the promethusexec receiver.  All prometheus specific functionality should be supported in the "normal" prometheus (scraping) receiver, along with others

## If I'm using the prometheusexec receiver, what should I do?

Migrate your configuration to use one of the following receivers

- [prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) (reccomended)
- [lightprometheus](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/lightprometheusreceiver) -- splunk specific, lighter resource usage)
- [simpleprometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver) -- recommended to use lightprometheusreceiver instead
- [sfx gateway prometheus remote write](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/signalfxgatewayprometheusremotewritereceiver) -- unlikely to be needed, but supports a simplified remote-write endpoint, should you need to "invert" your scraper from pull to push.  Does not fully support histograms.  Writes in a signalfx compatible format.  Prefer to use native otlp for a push/server based receiver if at all possible


### Moving configuration to the prometheus receiver

`scrape_interval` and `port` are the only consistent parameters across `prometheus`, `lightprometheus`, and `simpleprometheus` receivers. They retain the same semantic meaning.
The functionality and responsibility of instantiating and exposing and endpoint to be scraped from now lies with the end user, and is not supported in any of the current prometheus receivers.