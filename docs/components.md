> The official Splunk documentation for this page is [Use Splunk Distribution of OpenTelemetry Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/resources.html). For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

# Components

The distribution offers support for the following components.

> Each component has a link to its configuration documentation.

<div style="display: grid;grid-template-columns: auto auto auto auto;">

<div>

| Receivers                                                                                                                               | Stability        |
|:----------------------------------------------------------------------------------------------------------------------------------------|:-----------------|
| [azureeventhub](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/azureeventhubreceiver)             | [alpha]          |
| [carbon](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/carbonreceiver)                           | [alpha]          |
| [cloudfoundry](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/cloudfoundryreceiver)               | [beta]           |
| [collectd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/collectdreceiver)                       | [alpha]          |
| [databricks](../internal/receiver/databricksreceiver)                                                                                   | [alpha]          |
| [discovery](../internal/receiver/discoveryreceiver)                                                                                     | [in development] |
| [filelog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)                         | [beta]           |
| [fluentforward](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver)             | [beta]           |
| [hostmetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver)                 | [beta]           |
| [jaeger](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jaegerreceiver)                           | [beta]           |
| [jmx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver)                                 | [alpha]          |
| [journald](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/journaldreceiver)                       | [alpha]          |
| [k8s_cluster](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sclusterreceiver)                  | [beta]           |
| [k8s_events](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver)                    | [alpha]          |
| [k8sobjects](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sobjectsreceiver)                   | [alpha]          |
| [kafkametrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver)               | [beta]           |
| [kafka](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkareceiver)                             | [beta]           |
| [kubeletstats](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver)               | [beta]           |
| [lightprometheus](../internal/receiver/lightprometheusreceiver)                                                                         | [in development] |
| [mongodbatlas](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbatlasreceiver)               | [beta]           |
| [oracledb](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/oracledbreceiver)                       | [alpha]          |
| [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver)                                       | [stable]         |
| [postgres](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/postgresqlreceiver)                     | [beta]           |
| [prometheusexecreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusexecreceiver)   | [deprecated]     |
| [prometheusreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver)           | [beta]           |
| [receiver_creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator)                | [beta]           |
| [redis](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/redisreceiver)                             | [beta]           |
| [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sapmreceiver)                               | [beta]           |
| [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/signalfxreceiver)                       | [stable]         |
| [simpleprometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver)       | [beta]           |
| [smartagent](../pkg/receiver/smartagentreceiver)                                                                                        | [beta]           |
| [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/splunkhecreceiver)                    | [beta]           |
| [sqlquery](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sqlqueryreceiver)                       | [alpha]          |
| [statsd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver)                           | [beta]           |
| [syslog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver)                           | [alpha]          |
| [tcplog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/tcplogreceiver)                           | [alpha]          |
| [windowseventlog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowseventlogreceiver)         | [alpha]          |
| [windowsperfcounters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowsperfcountersreceiver) | [beta]           |
| [zipkin](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/zipkinreceiver)                           | [beta]           |

</div>

<div>

| Processors                                                                                                                                  | Stability                   |
|:--------------------------------------------------------------------------------------------------------------------------------------------|:----------------------------|
| [attributes](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor)                     | [alpha]                     |
| [batch](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)                                       | [beta]                      |
| [filter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor)                             | [alpha]                     |
| [groupbyattrs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)                 | [beta]                      |
| [k8sattributes](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor)               | [beta]                      |
| [logstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logstransformprocessor)               | [in development]            |
| [memorylimiter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/memorylimiterprocessor)                       | [beta]                      |
| [metricstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)         | [beta]                      |
| [probabilisticsampler](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) | [beta] |
| [resourcedetection](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor)       | [beta]                      |
| [resource](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor)                         | [beta]                      |
| [routing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/routingprocessor)                           | [beta]                      |
| [span_metrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/spanmetricsprocessor)                  | [in development]            |
| [span](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/spanprocessor)                                 | [alpha]                     |
| [tail_sampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor)                | [beta]                      |
| [timestamp](../pkg/processor/timestampprocessor)                                                                                                     | [in development]            |
| [transform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)                       | [alpha]                     |

</div>

<div>

| Exporters                                                                                                            | Stability        |
|:---------------------------------------------------------------------------------------------------------------------|:-----------------|
| [file](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/fileexporter)            | [alpha]          |
| [kafka](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/kafkaexporter)          | [beta]           |
| [httpsink](../internal/exporter/httpsinkexporter)                                                                    | [in development] |
| [logging](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)              | [in development] |
| [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter)                    | [stable]         |
| [otlphttp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)            | [stable]         |
| [pulsar](../internal/exporter/pulsarexporter)                                                                        | [deprecated]     |
| [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter)    | [beta]           |
| [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/sapmexporter)            | [beta]           |
| [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter) | [beta]           |

</div>

<div>

| Extensions                                                                                                                          | Stability |
|:------------------------------------------------------------------------------------------------------------------------------------|:----------|
| [docker_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/dockerobserver)    | [beta]    |
| [ecs_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecsobserver)          | [beta]    |
| [ecs_task_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecstaskobserver) | [beta]    |
| [healthcheck](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/healthcheckextension)           | [beta]    |
| [httpforwarder](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/httpforwarder)                | [beta]    |
| [host_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/hostobserver)        | [beta]    |
| [k8s_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/k8sobserver)          | [beta]    |
| [pprof](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/pprofextension)                       | [beta]    |
| [smartagent](../pkg/extension/smartagentextension)                                                                                  | [beta]    |
| [zpages](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension)                             | [beta]    |
| [file_storage](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage)           | [beta]    |
| [ballast](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/ballastextension)                           | [beta]    |

</div>

<div>

| Connectors                                                                                                                 | Stability        |
|:---------------------------------------------------------------------------------------------------------------------------|:-----------------|
| [count](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/countconnector)              | [in development] |
| [forward](https://github.com/open-telemetry/opentelemetry-collector/tree/main/connector/forwardconnector)                  | [beta]           |
| [spanmetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/spanmetricsconnector)  | [alpha]          |

</div>
</div>

[stable]: https://github.com/open-telemetry/opentelemetry-collector#stable
[beta]: https://github.com/open-telemetry/opentelemetry-collector#beta
[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[in development]: https://github.com/open-telemetry/opentelemetry-collector#development
[deprecated]: https://github.com/open-telemetry/opentelemetry-collector#deprecated

