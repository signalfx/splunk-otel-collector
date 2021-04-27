# Components

The distribution offers support for the following components.

## Beta

These components are considered stable. While in beta, breaking changes may be
introduced in a new release. In addition, any of these components may be
removed prior to the 1.0 release.

### Receivers

- [fluentforward](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver)
- [hostmetrics](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/hostmetricsreceiver)
- [jaeger](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/jaegerreceiver)
- [k8s_cluster](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sclusterreceiver)
- [kubeletstats](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver)
- [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver)
- [receiver_creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator)
- [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sapmreceiver)
- [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/signalfxreceiver)
- [simpleprometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver)
- [smartagent](../internal/receiver/smartagentreceiver)
- [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/splunkhecreceiver)
- [zipkin](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/zipkinreceiver)
 
### Processors

- [attributes](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/attributesprocessor)
- [batch](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)
- [filter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/filterprocessor)
- [k8s_tagger](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sprocessor)
- [memorylimiter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiter)
- [metrictransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)
- [probabilisticsampler](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/probabilisticsamplerprocessor)
- [resource](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/resourceprocessor)
- [resourcedetection](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor)
- [span](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/spanprocessor)

### Exporters

- [file](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/fileexporter)
- [logging](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)
- [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter)
- [otlphttp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)    
- [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/sapmexporter)  
- [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter)
- [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter)

### Extensions  

- [healthcheck](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/healthcheckextension)
- [httpforwarder](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/httpforwarder)
- [host_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/hostobserver)
- [k8s_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/k8sobserver)
- [pprof](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/pprofextension)
- [smartagent](../internal/extension/smartagentextension)
- [zpages](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension)

## Alpha

These components may or may not be stable. In addition, the may be limited in
their capabilities (for example the Kafka receiver/exporter only offers tracing
support at this time). While in alpha, breaking changes may be introduced in a
new release. In addition, any of these components may be removed prior to the
1.0 release.

### Receivers

- [carbon](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/carbonreceiver)
- [collectd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/collectdreceiver)
- [filelog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [kafka](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/kafkareceiver)
- [kafkametrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver)
- [statsd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver)

### Exporters

- [kafka](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/kafkaexporter)
