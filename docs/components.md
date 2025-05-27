> The official Splunk documentation for this page is [Components](https://docs.splunk.com/Observability/gdi/opentelemetry/components.html).
> For instructions on how to contribute to the docs, see [CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

# Components

The distribution offers support for the following components.

> Each component has a link to its configuration documentation.

<div style="display: grid;grid-template-columns: auto auto auto auto;">

<div>

| Receivers                                                                                                                                                          | Stability        |
|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-----------------|
| [active_directory_ds](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/activedirectorydsreceiver)                              | [beta]           |
| [apache](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/apachereceiver)                                                      | [alpha]          |
| [apachespark](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/apachesparkreceiver)                                            | [alpha]          |
| [awscontainerinsights](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awscontainerinsightreceiver)                           | [beta]           |
| [awsecscontainermetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/awsecscontainermetricsreceiver)                      | [beta]           |
| [azureblob](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/azureblobreceiver)                                                | [alpha]          |
| [azureeventhub](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/azureeventhubreceiver)                                        | [alpha]          |
| [azuremonitor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/azuremonitorreceiver)                                          | [in development] |
| [carbon](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/carbonreceiver)                                                      | [alpha]          |
| [chrony](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/chronyreceiver)                                                      | [beta]           |
| [cloudfoundry](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/cloudfoundryreceiver)                                          | [beta]           |
| [collectd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/collectdreceiver)                                                  | [beta]           |
| [discovery](../internal/receiver/discoveryreceiver)                                                                                                                | [in development] |
| [elasticsearch](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/elasticsearchreceiver)                                        | [beta]           |
| [filelog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)                                                    | [beta]           |
| [filestats](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filestatsreceiver)                                                | [beta]           |
| [fluentforward](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/fluentforwardreceiver)                                        | [beta]           |
| [googlecloudpubsub](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/googlecloudpubsubreceiver)                                | [beta]           |
| [haproxy](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/haproxyreceiver)                                                    | [beta]           |
| [hostmetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver)                                            | [beta]           |
| [httpcheck](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/httpcheckreceiver)                                                | [in development] |
| [iis](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/iisreceiver)                                                            | [beta]           |
| [influxdb](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/influxdbreceiver)                                                       | [beta]           |
| [jaeger](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jaegerreceiver)                                                      | [beta]           |
| [jmx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver)                                                            | [alpha]          |
| [journald](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/journaldreceiver)                                                  | [alpha]          |
| [k8s_cluster](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sclusterreceiver)                                             | [beta]           |
| [k8s_events](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver)                                               | [alpha]          |
| [k8sobjects](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8sobjectsreceiver)                                              | [alpha]          |
| [kafka](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkareceiver)                                                        | [beta]           |
| [kafkametrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver)                                          | [beta]           |
| [kubeletstats](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver)                                          | [beta]           |
| [lightprometheus](../internal/receiver/lightprometheusreceiver)                                                                                                    | [in development] |
| [mongodb](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbreceiver)                                                    | [beta]           |
| [mongodbatlas](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbatlasreceiver)                                          | [beta]           |
| [mysql](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mysqlreceiver)                                                        | [beta]           |
| [nginx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/nginxreceiver)                                                        | [beta]           |
| [nop](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/nopreceiver)                                                                    | [beta]           |
| [oracledb](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/oracledbreceiver)                                                  | [alpha]          |
| [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver)                                                                  | [stable]         |
| [postgresql](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/postgresqlreceiver)                                              | [beta]           |
| [prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver)                                              | [beta]           |
| [prometheus_simple](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver)                                 | [beta]           |
| [purefa](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/purefareceiver)                                                      | [alpha]          |
| [rabbitmq](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/rabbitmqreceiver)                                                  | [beta]           |
| [receiver_creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator)                                           | [beta]           |
| [redis](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/redisreceiver)                                                        | [beta]           |
| [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sapmreceiver)                                                          | [beta]           |
| [scripted_inputs](../internal/receiver//scriptedinputsreceiver)                                                                                                    | [in development] |
| [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/signalfxreceiver)                                                  | [stable]         |
| [signalfxgatewayprometheusremotewrite](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/signalfxgatewayprometheusremotewritereceiver) | [in development] |
| [smartagent](../pkg/receiver/smartagentreceiver)                                                                                                                   | [beta]           |
| [snowflake](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/snowflakereceiver)                                                | [beta]           |
| [solace](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/solacereceiver)                                                      | [beta]           |
| [splunkenterprise](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/splunkenterprisereceiver)                                  | [beta]           |
| [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/splunkhecreceiver)                                               | [beta]           |
| [sqlquery](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sqlqueryreceiver)                                                  | [alpha]          |
| [sqlserver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sqlserverreceiver)                                                | [beta]           |
| [sshcheck](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/sshcheckreceiver)                                                  | [alpha]          |
| [statsd](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver)                                                      | [beta]           |
| [syslog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver)                                                      | [alpha]          |
| [tcplog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/tcplogreceiver)                                                      | [alpha]          |
| [udplog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/udplogreceiver)                                                      | [alpha]          |
| [vcenter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/vcenterreceiver)                                                    | [alpha]          |
| [wavefront](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/wavefrontreceiver)                                                | [beta]           |
| [windowseventlog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowseventlogreceiver)                                    | [alpha]          |
| [windowsperfcounters](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowsperfcountersreceiver)                            | [beta]           |
| [zipkin](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/zipkinreceiver)                                                      | [beta]           |

</div>

<div>

| Processors                                                                                                                                   | Stability        |
|:---------------------------------------------------------------------------------------------------------------------------------------------|:-----------------|
| [attributes](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/attributesprocessor)                      | [alpha]          |
| [batch](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)                                        | [beta]           |
| [cumulativetodelta](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/cumulativetodeltaprocessor)        | [beta]           |
| [filter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor)                              | [alpha]          |
| [groupbyattrs](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)                  | [beta]           |
| [k8sattributes](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor)                | [beta]           |
| [logstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logstransformprocessor)                | [in development] |
| [memory_limiter](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/memorylimiterprocessor)                       | [beta]           |
| [metricsgeneration](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricsgenerationprocessor)        | [alpha]          |
| [metricstransform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)          | [beta]           |
| [probabilistic_sampler](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) | [beta]           |
| [redaction](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/redactionprocessor)                        | [beta]           |
| [resource](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourceprocessor)                          | [beta]           |
| [resourcedetection](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor)        | [beta]           |
| [routing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/routingprocessor)                            | [beta]           |
| [span](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/spanprocessor)                                  | [alpha]          |
| [tail_sampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor)                 | [beta]           |
| [timestamp](../pkg/processor/timestampprocessor)                                                                                             | [in development] |
| [transform](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor)                        | [alpha]          |

</div>

<div>

| Exporters                                                                                                               		| Stability        |
|:--------------------------------------------------------------------------------------------------------------------------------------|:-----------------|
| [awss3](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/awss3exporter)                		| [alpha]          |
| [googlecloudpubsub](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/googlecloudpubsubexporter)   | [beta] 
| [debug](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/debugexporter)                         		| [in development] |
| [file](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/fileexporter)                   		| [alpha]          |
| [kafka](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/kafkaexporter)                 		| [beta]           |
| [loadbalancing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/loadbalancingexporter) 		| [beta]           |
| [logging](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)                     		| [deprecated]     |
| [nop](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/nopexporter)                             		| [beta]           |
| [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter)                           		| [stable]         |
| [otlphttp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)                   		| [stable]         |
| [pulsar](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/pulsarexporter)               		| [alpha]          |
| [sapm](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/sapmexporter)                   		| [beta]           |
| [signalfx](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter)           		| [beta]           |
| [splunk_hec](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/splunkhecexporter)        		| [beta]           |

</div>

<div>

| Extensions                                                                                                                          | Stability |
|:------------------------------------------------------------------------------------------------------------------------------------|:----------|
| [ack](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/ackextension)                           | [alpha]   |
| [basicauth](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/basicauthextension)               | [beta]    |
| [bearertokenauth](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/bearertokenauthextension)   | [beta]    |
| [docker_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/dockerobserver)    | [beta]    |
| [ecs_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecsobserver)          | [beta]    |
| [ecs_task_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecstaskobserver) | [beta]    |
| [file_storage](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage)           | [beta]    |
| [headers_setter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/headerssetterextension)      | [alpha]   |
| [health_check](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/healthcheckextension)          | [beta]    |
| [host_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/hostobserver)        | [beta]    |
| [http_forwarder](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/httpforwarderextension)      | [beta]    |
| [k8s_observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/k8sobserver)          | [beta]    |
| [oauth2client](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/oauth2clientauthextension)     | [beta]    |
| [opamp](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/opampextension)                       | [alpha]   |
| [pprof](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/pprofextension)                       | [beta]    |
| [smartagent](../pkg/extension/smartagentextension)                                                                                  | [beta]    |
| [zpages](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension)                             | [beta]    |

</div>

<div>

| Connectors                                                                                                                | Stability        |
| :------------------------------------------------------------------------------------------------------------------------ | :--------------- |
| [count](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/countconnector)             | [in development] |
| [forward](https://github.com/open-telemetry/opentelemetry-collector/tree/main/connector/forwardconnector)                 | [beta]           |
| [routing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/routingconnector)         | [alpha]          |
| [spanmetrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/spanmetricsconnector) | [alpha]          |
| [sum](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/sumconnector)                 | [alpha]          |

</div>
</div>

[stable]: https://github.com/open-telemetry/opentelemetry-collector#stable
[beta]: https://github.com/open-telemetry/opentelemetry-collector#beta
[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[in development]: https://github.com/open-telemetry/opentelemetry-collector#development
[deprecated]: https://github.com/open-telemetry/opentelemetry-collector#deprecated

