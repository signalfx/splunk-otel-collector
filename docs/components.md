# Components

The distribution offers support for the following components.

## Beta

These components are considered stable. While in beta, breaking changes may be
introduced in a new release. In addition, any of these components may be
removed prior to the 1.0 release.

| Receivers        | Processors        | Exporters | Extensions    |
| :--------------: | :--------:        | :-------: | :--------:    |
| fluentforward    | attributes        | file      | healthcheck   |
| hostmetrics      | batch             | logging   | httpforwarder |
| jaeger           | filter            | otlp      | host_observer |
| k8s_cluster      | k8s_tagger        | sapm      | k8s_observer  |
| kubeletstats     | memorylimiter     | signalfx  | pprof         |
| opencensus       | metrictransform   | splunk_hec | zpages        |
| otlp             | resource          |           |               |
| receiver_creator | resourcedetection |           |               |
| sapm             | span              |           |               |
| signalfx         |                   |           |               |
| simpleprometheus |                   |           |               |
| smartagent       |                   |           |               |
| splunk_hec       |                   |           |               |
| zipkin           |                   |           |               |

## Alpha

These components may or may not be stable. In addition, the may be limited in
their capabilities (for example the Kafka receiver/exporter only offers tracing
support at this time). While in alpha, breaking changes may be introduced in a
new release. In addition, any of these components may be removed prior to the
1.0 release.

| Receivers      | Processors | Exporters | Extensions |
| :-------:      | :--------: | :-------: | :--------: |
| carbon         |            | kafka     |            |
| collectd       |            |           |            |
| kafka          |            |           |            |
| statsd         |            |           |            |
