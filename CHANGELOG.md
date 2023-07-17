# Changelog

## Unreleased

### 🧰 Bug fixes 🧰

- (Splunk) Discovery mode: Ensure all successful observers are used in resulting receiver creator instance ([#3391](https://github.com/signalfx/splunk-otel-collector/pull/3391))

## v0.81.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.81.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.81.0) and the [opentelemetry-collector-contrib v0.81.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.81.0) releases where appropriate.

### 🛑 Breaking changes 🛑
- (Core) `service`: Remove 'service.connectors' featuregate ([#7952](https://github.com/open-telemetry/opentelemetry-collector/pull/7952))
- (Contrib) `receiver/mongodbatlas`: Change the types of `Config.PrivateKey` and `Config.Alerts.Secret` to be `configopaque.String` ([#17273](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/17273))

### 🚩 Deprecations 🚩

- `mysqlreceiver`: set `mysql.locked_connects` as optional in order to remove it in next release ([#14138](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14138), [#23274](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23274))

### 💡 Enhancements 💡

- (Splunk) Package default discovery configuration in reference form in `/etc/otel/collector/config.d` ([#3311](https://github.com/signalfx/splunk-otel-collector/pull/3311))
- (Splunk) Add bundled collectd/nginx Smart Agent receiver discovery rules ([#3321](https://github.com/signalfx/splunk-otel-collector/pull/3321))
- (Splunk) Support custom `--discovery-properties` file ([#3334](https://github.com/signalfx/splunk-otel-collector/pull/3334))
- (Splunk) Add `--discovery` to the Linux installer script ([#3365](https://github.com/signalfx/splunk-otel-collector/pull/3365))
- (Splunk) Starting from this version the logs pipeline is split in the default configuration in a way that profiling 
  data is always sent to Splunk Observability endpoint while other logs can be sent to another hec endpoint configured
  with `SPLUNK_HEC_URL` and `SPLUNK_HEC_TOKEN` environment variables ([#3330](https://github.com/signalfx/splunk-otel-collector/pull/3330))
- (Core) `HTTPServerSettings`: Add zstd support to HTTPServerSettings ([#7927](https://github.com/open-telemetry/opentelemetry-collector/pull/7927))
  This adds ability to decompress zstd-compressed HTTP requests to| all receivers that use HTTPServerSettings.
- (Core) `confighttp`: Add `response_headers` configuration option on HTTPServerSettings. It allows for additional headers to be attached to each HTTP response sent to the client ([#7328](https://github.com/open-telemetry/opentelemetry-collector/issues/7328))
- (Core) `otlpreceiver, otlphttpexporter, otlpexporter, configgrpc`: Upgrade github.com/mostynb/go-grpc-compression and switch to nonclobbering imports ([#7920](https://github.com/open-telemetry/opentelemetry-collector/issues/7920))
  consumers of this library should not have their grpc codecs overridden
- (Core) `otlphttpexporter`: Treat partial success responses as errors ([#6686](https://github.com/open-telemetry/opentelemetry-collector/issues/6686))
- (Contrib) `sqlqueryreceiver`: Add support of Start and End Timestamp Column in Metric Configuration. ([#18925](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18925), [#14146](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14146))
- (Contrib) `filelogreceiver`: Add support for tracking the current file in filelogreceiver ([#22998](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22998))
- (Contrib) `pkg/ottl`: Adds new `Time` converter to convert a string to a Golang time struct based on a supplied format ([#22007](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22007))
- (Contrib) `hostmetricsreceiver`: Add new Windows-exclusive process.handles metric. ([#21379](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21379))
- (Contrib) `resourcedetectionprocessor`: Adds a way to configure the list of added resource attributes by the processor ([#21482](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21482))
  Users can now configure what resource attributes are gathered by specific detectors.
  Example configuration:

  ```
  resourcedetection:
    detectors: [system, ec2]
    system:
      resource_attributes:
        host.name:
          enabled: true
        host.id:
          enabled: false
    ec2:
      resource_attributes:
        host.name:
          enabled: false
        host.id:
          enabled: true
  ```

  For example, this config makes `host.name` being set by `system` detector, and `host.id` by `ec2` detector.
  Moreover:
  - Existing behavior remains unaffected as all attributes are currently enabled by default.
  - The default attributes 'enabled' values are defined in `metadata.yaml`.
  - Future releases will introduce changes to resource_attributes `enabled` values.
  - Users can tailor resource detection process to their needs and environment.
- (Contrib) `k8sclusterreceiver`: Switch k8s.pod and k8s.container metrics to use pdata. ([#23441](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23441))

### 🧰 Bug fixes 🧰

- (Contrib) `k8sclusterreceiver`: Add back all other vendor-specific node conditions, and report them even if missing, as well as all allocatable node metrics if present,  to the list of Kubernetes node metrics available, which went missing during the pdata translation ([#23839](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23839))
- (Contrib) `k8sclusterreceiver`: Add explicitly `k8s.node.allocatable_pods` to the list of Kubernetes node metrics available, which went missing during the pdata translation ([#23839](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23839))
- (Contrib) `receiver/kafkametricsreceiver`: Updates certain metrics in kafkametricsreceiver to function as non-monotonic sums. ([#4327](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4327))
  Update the metrics type in KafkaMetricsReceiver from "gauge" to "nonmonotonic sum". Changes metrics are, kafka.brokers, kafka.topic.partitions, kafka.partition.replicas, kafka.partition.replicas_in_sync, kafka.consumer_group.members.
- (Contrib) `windowseventlogreceiver`: Fix buffer overflow when ingesting large raw Events ([#23677](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23677))
- (Contrib) `pkg/stanza`: adding octet counting event breaking for syslog parser ([#23577](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23577))

## v0.80.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.80.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.80.0) and the [opentelemetry-collector-contrib v0.80.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.80.0) releases where appropriate.

### 🛑 Breaking changes 🛑
- (Contrib) `redisreceiver`: Updates metric unit from no unit to Bytes. ([#23454](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23454))
  Affected metrics can be found below.
  - redis.clients.max_input_buffer
  - redis.clients.max_output_buffer
  - redis.replication.backlog_first_byte_offset
  - redis.replication.offset
- (Splunk) Embed observer configuration in `observer.discovery.yaml` `config` mapping. This is only a breaking change if you have written your own custom discovery mode observer configuration ([#3277](https://github.com/signalfx/splunk-otel-collector/pull/3277)).

### 💡 Enhancements 💡

- (Contrib) `resourcedetectionprocessor`: use opentelemetry-go library for `host.id` detection in the `system` detector ([#18533](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/18533))
- (Contrib) `k8sattributesprocessor`: Store only necessary ReplicaSet and Pod data ([#23226](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23226))
- (Contrib) `k8sclusterreceiver`: Do not store unused data in the k8s API cache to reduce RAM usage ([#23433](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23433))
- (Contrib) `pkg/ottl`: Add new `IsString` and `IsMap` functions to facilitate type checking. ([#22750](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22750))
  Especially useful for checking log body type before parsing.
- (Contrib) `pkg/ottl`: Adds `StandardFuncs` and `StandardConverters` to facilitate function map generation. ([#23190](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23190))
  This change means that new functions added to ottlfuncs get automatically added to Cotnrib components that use OTTL
- (Contrib) `pkg/ottl`: Change replacement functions to accept a path expression as a replacement ([#22787](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22787))
  The following replacement functions now accept a path expression as a replacement:
  - replace_match
  - replace_pattern
  - replace_all_matches
  - replace_all_patterns
- (Contrib) `sapmexporter`: sapm exporter now supports `compression` config option to specify either gzip or zstd compression to use. ([#23257](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23257))
- (Contrib) `sapmreceiver`: sapm receiver now accepts requests in compressed with zstd. ([#23257](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23257))
- (Contrib) `exporter/signalfx`: Do not drop container.cpu.time metric in the default translations so it can be enabled in the include_metrics config. ([#23403](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23403))
- (Contrib) `sqlqueryreceiver`: Add support for logs ([#20284](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/20284))
- (Contrib) `k8sclusterreceiver`: Switch k8s.deployment metrics to use pdata. ([#23416](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23416))
- (Contrib) `k8sclusterreceiver`: Switch k8s.hpa metrics to use pdata. ([#18250](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/18250))
- (Contrib) `k8sclusterreceiver`: Switch k8s.namespace metrics to use pdata. ([#23437](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23437))
- (Contrib) `k8sclusterreceiver`: Switch k8s.node metrics to use pdata. ([#23438](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23438))
- (Contrib) `k8sclusterreceiver`: Switch k8s.rq metrics to use pdata. ([#23419](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23419))
- (Contrib) `k8sclusterreceiver`: Switch k8s.ss metrics to use pdata. ([#23420](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23420))
- (Contrib) `carbonreceiver`: Remove use of opencensus model in carbonreceiver ([#20759](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/20759))
- (Core) `service`: Added dry run flag to validate config file without running collector. ([#4671](https://github.com/open-telemetry/opentelemetry-collector/issues/4671))
- (Core) `configtls`: Allow TLS Settings to be provided in memory in addition to filepath. ([#7313](https://github.com/open-telemetry/opentelemetry-collector/issues/7313))
- (Core) `connector`: Updates the way connector nodes are built to always pass a fanoutconsumer to their factory functions. ([#7672](https://github.com/open-telemetry/opentelemetry-collector/issues/7672), [#7673](https://github.com/open-telemetry/opentelemetry-collector/issues/7673))
- (Core) `otlp`: update otlp protos to v0.20.0 ([#7839](https://github.com/open-telemetry/opentelemetry-collector/issues/7839))
- (Core) `config`: Split config into more granular modules ([#7895](https://github.com/open-telemetry/opentelemetry-collector/issues/7895))
- (Core) `connector`: Split connector into its own module ([#7895](https://github.com/open-telemetry/opentelemetry-collector/issues/7895))
- (Core) `extension`: split extension and `extension/auth` into its own module ([#7306](https://github.com/open-telemetry/opentelemetry-collector/issues/7306), [#7054](https://github.com/open-telemetry/opentelemetry-collector/issues/7054))
- (Core) `processor`: Split the processor into its own go module ([#7307](https://github.com/open-telemetry/opentelemetry-collector/issues/7307))
- (Core) `confighttp`: Avoid re-creating the compressors for every request. ([#7859](https://github.com/open-telemetry/opentelemetry-collector/issues/7859))
- (Core) `otlpexporter`: Treat partial success responses as errors ([#6686](https://github.com/open-telemetry/opentelemetry-collector/issues/6686))
- (Core) `service/pipelines`: Add pipelines.Config to remove duplicate of the pipelines configuration ([#7854](https://github.com/open-telemetry/opentelemetry-collector/issues/7854))

### 🧰 Bug fixes 🧰

- (Contrib) `otel-collector`: Fix cri-o log format time layout ([#23027](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23027))
- (Contrib) `receiver/hostmetricsreceiver`: Fix not sending `process.cpu.utilization` when `process.cpu.time` is disabled. ([#23450](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23450))
- (Contrib) `receiver/kafkametricsreceiver`: Updates certain metrics in kafkametricsreceiver to function as non-monotonic sums. ([#4327](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/4327))
  Update the metric type in KafkaMetricsReceiver from "gauge" to "nonmonotonic sum".
- (Contrib) `receiver/hostmetrics`: Fix issue where receiver fails to read parent-process information for some processes on Windows ([#14679](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/14679))
- (Contrib) `k8sclusterreceiver`: Fix empty k8s.namespace.name attribute in k8s.namespace.phase metric ([#23452](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23452))
- (Contrib) `splunkhecexporter`: Apply multi-metric merge at the level of the whole batch rather than within events emitted for one metric. ([#23365](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23365))

## v0.79.1

### 🛑 Breaking changes 🛑

- (Contrib) Set `pkg.translator.prometheus.NormalizeName` feature gate back to Alpha state since it was enabled
  prematurely. Metrics coming from Prometheus receiver will not be normalized by default, specifically `_total` suffix
  will not be removed from metric names. To maintain the current behavior (drop the `_total` suffix), you can enable
  the feature gate using the `--feature-gates=pkg.translator.prometheus.NormalizeName` command argument. However, note
  that the translation in the prometheus receiver is a subject to possible future changes.
  ([#23229](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23229))

### 💡 Enhancements 💡

- (Splunk) Add spanmetric and count connectors ([#3300](https://github.com/signalfx/splunk-otel-collector/pull/3300))
- (Splunk) Upgrade builds to use golang 1.20.5 ([#3299](https://github.com/signalfx/splunk-otel-collector/pull/3299))
- (Splunk) `receiver/smartagent`: Add `scrapeFailureLogLevel` config field to `prometheus-exporter` and its sourcing monitors to determine the log level for reported scrape failures ([#3260](https://github.com/signalfx/splunk-otel-collector/pull/3260))

### 🧰 Bug fixes 🧰

- (Splunk) Correct imported Contrib `pkg/translator/prometheus` dependency for `pkg.translator.prometheus.NormalizeName` Alpha state ([#3303](https://github.com/signalfx/splunk-otel-collector/pull/3303))

## v0.79.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.79.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.79.0) and the [opentelemetry-collector-contrib v0.79.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.79.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) ~~Set `pkg.translator.prometheus.NormalizeName` feature gate back to Alpha state since it was enabled prematurely.~~ edit: This was an incomplete adoption, addressed in release v0.79.1.
- (Contrib) `attributesprocessor`: Enable SHA-256 as hashing algorithm by default for attributesprocessor hashing action ([#4759](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4759))
- (Contrib) `windowseventlogreceiver`: Emit raw Windows events as strings instead of byte arrays ([#22704](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22704))
- (Contrib) `pkg/ottl`: Removes `StandardTypeGetter` in favor of `StandardStringGetter`, `StandardIntGetter`, `StandardFloatGetter`, and `StandardPMapGetter`, which handle converting pcommon.Values of the proper type. ([#22763](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22763))
  This is only a breaking change for users using OTTL in custom components. For all Contrib components this is an enhancement.
- (Contrib) `postgresqlreceiver`: Remove resource attribute feature gates ([#22479](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22479))

### 💡 Enhancements 💡

- (Splunk) `smartagentreceiver`: Add `kubernetes-cluster` config option to sync node labels as properties on the `kubernetes_node` dimension ([#3267](https://github.com/signalfx/splunk-otel-collector/pull/3267))
- (Splunk) Discovery mode: Support `splunk.discovery` mapping in properties.discovery.yaml ([#3238](https://github.com/signalfx/splunk-otel-collector/pull/3238))
- (Splunk) Upgrade to the latest Java agent version [v1.25.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.25.0) ([#3272](https://github.com/signalfx/splunk-otel-collector/pull/3272))
- (Contrib) `jmxreceiver`: Add the JMX metrics gatherer version 1.26.0-alpha to the supported jars hash list ([#22042](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22042))
- (Contrib) `receivers`: Adding `initial_delay` to receivers to control when scraping interval starts ([#23030](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23030))
  The updated receivers are:
  - `oracledb`
  - `postgresql`
  - `sqlquery`
  - `windowsperfcounters`
- (Contrib) `oracledbreceiver`: Add a simpler alternative configuration option ([#22087](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22087))
- (Contrib) `pkg/ottl`: Add `body.string` accessor to ottllog ([#22786](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22786))
- (Contrib) `pkg/ottl`: Allow indexing map and slice log bodies ([#17396](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/17396), [#22068](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22068))
- (Contrib) `pkg/ottl`: Add hash converters/functions for OTTL ([#22725](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22725))
- (Contrib) `splunkhecreceiver`: Support different strategies for splitting payloads when receiving a request with the Splunk HEC receiver ([#22788](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22788))
- (Contrib) `exporter/splunk_hec`: Apply compression to Splunk HEC payload unconditionally if it's enabled in the config ([#22969](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22969), [#22018](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22018))
  The compression used to be enabled only if the payload size was greater than 1.5KB which significantly
  complicated the logic and made it hard to test. This change makes the compression unconditionally applied to
  the payload if it's enabled in the config. The benchmarking shows improvements in the throughput and CPU usage for
  large payloads and expected degradation for small payloads which is acceptable given that it's not a common case.
- (Core) `otelcol`: Add connectors to output of the `components` command ([#7809](https://github.com/open-telemetry/opentelemetry-collector/pull/7809))
- (Core) `scraperhelper`: Will start calling scrapers on component start. ([#7635](https://github.com/open-telemetry/opentelemetry-collector/pull/7635))
  The change allows scrapes to perform their initial scrape on component start
  and provide an initial delay. This means that scrapes will be delayed by `initial_delay`
  before first scrape and then run on `collection_interval` for each consecutive interval.
- (Core) `batchprocessor`: Change multiBatcher to use sync.Map, avoid global lock on fast path ([#7714](https://github.com/open-telemetry/opentelemetry-collector/pull/7714))
- (Core, Contrib, Splunk) Third-party dependency updates.

### 🧰 Bug fixes 🧰

- (Splunk) `smartagentreceiver` add missing `monitorID` logger field to `http` monitor ([#3261](https://github.com/signalfx/splunk-otel-collector/pull/3261))
- (Contrib) `jmxreceiver`: Fixed the issue where the JMX receiver's subprocess wasn't canceled upon shutdown, resulting in a rogue java process. ([#23051](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23051))
- (Contrib) `internal/filter/filterlog`: fix filtering non-string body by bodies property ([#22736](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22736))
  Affects `filterprocessor` and `attributesprocessor`.
- (Contrib) `prometheusreceiver`: Remove sd_file validations from config.go in Prometheus Receiver to avoid failing Collector with error as this behaviour is incompatible with the Prometheus. ([#21509](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21509))
- (Contrib) `fileexporter`: Fixes broken lines when rotation is set. ([#22747](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22747))
- (Contrib) `exporter/splunk_hec`: Make sure the `max_event_size` option is used to drop events larger than `max_event_size` instead of using it for batch size. ([#18066](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18066))
- (Contrib) `postgresqlreceiver`: Fix race condition when capturing errors from multiple requests simultaneously ([#23026](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23026))
- (Contrib) `prometheusreceiver`: The prometheus receiver now sets a full, versioned user agent. ([#21910](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21910))
- (Contrib) `splunkhecreceiver`: Fix reusing the same splunkhecreiver between logs and metrics ([#22848](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22848))
- (Core) `connectors`: When replicating data to connectors, consider whether the next pipeline will mutate data ([#7776](https://github.com/open-telemetry/opentelemetry-collector/issues/7776))

## v0.78.1

### 🧰 Bug fixes 🧰

- (Contrib) `receiver/filelog` Account for empty files ([#22815](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22815))

### 💡 Enhancements 💡
- (Core, Contrib, Splunk) Third-party dependency updates.

## v0.78.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.78.2](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.78.2) and the [opentelemetry-collector-contrib v0.78.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.78.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `receiver/mongodbatlas`: Update emitted Scope name to "otelcol/mongodbatlasreceiver" ([#21382](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21382))
- (Contrib) `receivers`: Updating receivers that run intervals to use standard interval by default ([#22138](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22138))
- (Contrib) `pkg/ottl`: Updates the `Int` converter to use a new `IntLikeGetter` which will error if the value cannot be converted to an int. ([#22059](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22059))
  Affected components: transformprocessor, filterprocessor, routingprocessor, tailsamplingprocessor, countconnector. It is HIGHLY recommended to use each component's error_mode configuration option to handle errors returned by `Int`.

### 💡 Enhancements 💡

- (Splunk) Add `enabled` field support to `*.discovery.yaml` config ([#3207](https://github.com/signalfx/splunk-otel-collector/pull/3207))
- (Contrib) `jmxreceiver`: Add the JMX metrics gatherer version 1.26.0-alpha to the supported jars hash list ([#22042](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22042))
- (Contrib) `receivercreator`: add logs and traces support to receivercreator ([#19205](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19205), [#19206](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19206))
- (Contrib) `pkg/ottl`: Add Log function ([#18076](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18076))
- (Contrib) `oracledbreceiver`: Adds support for `consistent gets` and `db block gets` metrics. Disabled by default. ([#21215](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21215))
- (Contrib) `pkg/batchperresourceattr`: Mark as not mutating as it does defensive copying. ([#21885](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21885))
- (Contrib) `receiver/kafkareceiver`: Support configuration of initial offset strategy to allow consuming form latest or earliest offset ([#14976](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14976))
- (Contrib) `internal/filter`: Add `Log`, `UUID`, and `ParseJSON` converters to filterottl standard functions ([#21970](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21970))
- (Contrib) `pkg/stanza`: aggregate the latter part of the split-log due to triggering the size limit ([#21241](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21241))
- (Contrib) `cmd/mdatagen`: Allow setting resource_attributes without introducing the metrics builder. ([#21516](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21516))
- (Contrib) `receiver/mongodbatlasreceiver`: Allow collection of MongoDB Atlas Access Logs as a new feature of the MongoDBAtlas receiver. ([#21182](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21182))
- (Contrib) `pkg/ottl`: Add `FloatLikeGetter` and `FloatGetter` to facilitate float retrival for functions. ([#21896](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21896))
- (Contrib) `pkg/ottl`: Add access to get and set span kind using a string ([#21773](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21773))
- (Contrib) `processor/routingprocessor`: Instrument the routing processor with non-routed spans/metricpoints/logrecords counters (OTel SDK). ([#21476](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21476))
- (Contrib) `exporter/splunkhec`: Improve performance and reduce memory consumption. ([#22018](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22018))
- (Contrib) `processor/transform`: Add access to the Log function ([#22014](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22014))
- (Core) `batchprocessor`: Add support for batching by metadata keys. ([#4544](https://github.com/open-telemetry/opentelemetry-collector/issues/4544))
- (Core) `service`: Add feature gate `telemetry.useOtelWithSDKConfigurationForInternalTelemetry` that will add support for configuring the export of internal telemetry to additional destinations in future releases ([#7678](https://github.com/open-telemetry/opentelemetry-collector/pull/7678), [#7641](https://github.com/open-telemetry/opentelemetry-collector/pull/7641))
- (Core) `forwardconnector`: Promote to beta ([#7579](https://github.com/open-telemetry/opentelemetry-collector/pull/7579))
- (Core) `featuregate`: Promote `featuregate` to the stable module-set ([#7693](https://github.com/open-telemetry/opentelemetry-collector/pull/7693))
- (Core, Contrib, Splunk) Third-party dependency updates.

### 🧰 Bug fixes 🧰

- (Splunk) Evaluate `--set` properties as yaml values ([#3175](https://github.com/signalfx/splunk-otel-collector/pull/3175))
- (Splunk) Evaluate config converter updates to `--dry-run` content ([#3176](https://github.com/signalfx/splunk-otel-collector/pull/3176))
- (Splunk) Support config provider uris in `--config` option values ([#3182](https://github.com/signalfx/splunk-otel-collector/pull/3182))
- (Splunk) `receiver/smartagent`: Don't attempt to expand observer `endpoint` fields if host and port are unsupported ([#3187](https://github.com/signalfx/splunk-otel-collector/pull/3187))
- (Splunk) Replace deprecated `loglevel: debug` logging exporter field with `verbosity: detailed` in default configs ([#3189](https://github.com/signalfx/splunk-otel-collector/pull/3189))
- (Contrib) `statsdreceiver`: Handles StatsD server not running when shutting down to avoid NPE ([#22004](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22004))
- (Contrib) `pkg/ottl`: Fix the factory name for the limit function ([#21920](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21920))
- (Contrib) `processor/filter`: Fix issue where the OTTL function `HasAttributeKeyOnDatapoint` was not usable. ([#22057](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22057))
- (Contrib) `pkg/ottl`: Allow using capture groups in `replace_all_patterns` when replacing map keys ([#22094](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22094))
- (Contrib) `exporter/splunkhec`: Fix a bug causing incorrect data in the partial error returned by the exporter ([#21720](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/21720))
- (Core) `batchprocessor`: Fix return error for batch processor when consuming Metrics and Logs ([#7711](https://github.com/open-telemetry/opentelemetry-collector/pull/7711))
- (Core) `batchprocessor`: Fix start/stop logic for batch processor ([#7708](https://github.com/open-telemetry/opentelemetry-collector/pull/7708))
- (Core) `featuregate`: Fix issue where `StageDeprecated` was not usable ([#7586](https://github.com/open-telemetry/opentelemetry-collector/pull/7586))
- (Core) `exporterhelper`: Fix persistent storage behaviour with no available space on device ([#7198](https://github.com/open-telemetry/opentelemetry-collector/issues/7198))

## v0.77.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.77.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.77.0) and the [opentelemetry-collector-contrib v0.77.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.77.0) releases where appropriate.

### 💡 Enhancements 💡

- `connector/forward` - Add support for the forward connector ([#3100](https://github.com/signalfx/splunk-otel-collector/pull/3100))
- `receiver/signalfxgatewayprometheusremotewritereceiver` - Add new receiver that aims to be an otel-native version of
  the SignalFx [Prometheus remote write](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go)
  [gateway](https://github.com/signalfx/gateway/blob/main/README.md) ([#3064](https://github.com/signalfx/splunk-otel-collector/pull/3064))
- `signalfx-agent`: Relocate to be internal to the collector ([#3052](https://github.com/signalfx/splunk-otel-collector/pull/3052))

## v0.76.1

### 💡 Enhancements 💡

- `receiver/jmxreceiver`: Add OpenTelemetry JMX receiver to the distribution ([#3068](https://github.com/signalfx/splunk-otel-collector/pull/3068))
- Update Java auto-instrumentation library to 1.23.1 ([#3055](https://github.com/signalfx/splunk-otel-collector/pull/3055))
- Update installer script to check system architecture ([#2888](https://github.com/signalfx/splunk-otel-collector/pull/2888))

## v0.76.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.76.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.76.1) and the [opentelemetry-collector-contrib v0.76.3](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.76.3) releases where appropriate.

### 💡 Enhancements 💡

- `receiver/lightprometheus`: Limit default resource attributes ([#3042](https://github.com/signalfx/splunk-otel-collector/pull/3042))
- `receiver/discovery`: exposed JSON-encoded evaluated statement zap fields ([#3004](https://github.com/signalfx/splunk-otel-collector/pull/3004), [#3032](https://github.com/signalfx/splunk-otel-collector/pull/3032))
- `receiver/smartagent`: Update bundled python to 3.11.3 ([#3002](https://github.com/signalfx/splunk-otel-collector/pull/3002))
- Update token verification failure message for installer scripts ([#2991](https://github.com/signalfx/splunk-otel-collector/pull/2991))
- `exporter/httpsink`: Add support for metrics and filtering ([#2959](https://github.com/signalfx/splunk-otel-collector/pull/2959))
- `--discovery`: Add `k8sobserver` support for `smartagent/postgresql` ([#3023](https://github.com/signalfx/splunk-otel-collector/pull/3023))
- `--discovery`: Append discovered components to existing metrics pipeline ([#2986](https://github.com/signalfx/splunk-otel-collector/pull/2986))
- `receiver/smartagent`: add `isolatedCollectd` option for native collectd monitors ([#2957](https://github.com/signalfx/splunk-otel-collector/pull/2957))
- Third party dependency updates

### 🧰 Bug fixes 🧰

- `receiver/smartagent`: Don't set `monitorID` attribute if set by monitor ([#3031](https://github.com/signalfx/splunk-otel-collector/pull/3031))
- `receiver/smartagent`: set `sql` monitor logger type from config ([#3001](https://github.com/signalfx/splunk-otel-collector/pull/3001))

## v0.75.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.75.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.75.0) and the [opentelemetry-collector-contrib v0.75.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.75.0) releases where appropriate.

### 💡 Enhancements 💡

- New [light prometheus receiver](https://github.com/signalfx/splunk-otel-collector/pull/2921) we're prototyping

### 🧰 Bug fixes 🧰

- Cherry-pick [fluentforward receiver fix](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/20721)
  from upstream which fixes a performance regression introduced in v0.73.0.
- Fixed sendLoadState, sendSubState and sendActiveState options for [systemd metadata](https://github.com/signalfx/splunk-otel-collector/pull/2929)


## v0.74.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.74.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.74.0) and the [opentelemetry-collector-contrib v0.74.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.74.0) releases where appropriate.

### 💡 Enhancements 💡
- [Relocate agent codebase into pkg/signalfx-agent](https://github.com/signalfx/splunk-otel-collector/pull/2717)
- [Tanzu Tile implementation and documentation](https://github.com/signalfx/splunk-otel-collector/pull/2726)
- [Mark our internal pulsar exporter as deprecated](https://github.com/signalfx/splunk-otel-collector/pull/2873)

### 🧰 Bug fixes 🧰
- [Add shutdown method to hostmetadata monitor](https://github.com/signalfx/splunk-otel-collector/pull/2917)
- [Support core file and env config provider directive resolution](https://github.com/signalfx/splunk-otel-collector/pull/2893)

## v0.73.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.73.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.73.0) and the [opentelemetry-collector-contrib v0.73.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.73.0) releases where appropriate.

### 💡 Enhancements 💡
- [Build experimental linux arm64 agent-bundle](https://github.com/signalfx/splunk-otel-collector/pull/2671)
- Added profiling, JVM metrics, and service name generation options for zero configuration auto instrumentation of Java apps (Linux only):
  - [Installer script](https://github.com/signalfx/splunk-otel-collector/pull/2718)
  - [Ansible v0.16.0](https://github.com/signalfx/splunk-otel-collector/pull/2729)
  - [Chef v0.5.0](https://github.com/signalfx/splunk-otel-collector/pull/2733)
  - [Puppet v0.9.0](https://github.com/signalfx/splunk-otel-collector/pull/2734)
  - [Salt](https://github.com/signalfx/splunk-otel-collector/pull/2735)
- [update translation rule to use a copy of system.cpu.time and leave the original one intact](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/19743)

## v0.72.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.72.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.72.0) and the [opentelemetry-collector-contrib v0.72.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.72.0) releases where appropriate.

### 💡 Enhancements 💡
- [Added discoverybundler, initial embedded bundle.d and enabled properties for discovery mode](https://github.com/signalfx/splunk-otel-collector/pull/2601)
- [Updated pulsarexporter configuration to prepare for using exporter from contrib](https://github.com/signalfx/splunk-otel-collector/pull/2650)
- [Corrected module names for directory locations in examples](https://github.com/signalfx/splunk-otel-collector/pull/2665)
- [Built linux and windows amd64 agent bundles](https://github.com/signalfx/splunk-otel-collector/pull/2649)
- Third party dependency updates

## v0.71.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.71.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.71.0) and the [opentelemetry-collector-contrib v0.71.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.71.0) releases where appropriate.

### 💡 Enhancements 💡
- [Added the fluentforwarder receiver to the default ECS/EC2 configuration.](https://github.com/signalfx/splunk-otel-collector/pull/2537)
- [Added the PostgreSQL receiver](https://github.com/signalfx/splunk-otel-collector/pull/2564)
- [Zero config support added for always on profiling.](https://github.com/signalfx/splunk-otel-collector/pull/2538)
- [Upgraded to include changes from SignalFx Smart Agent v5.27.3](https://github.com/signalfx/signalfx-agent/releases/tag/v5.27.3)
- [Upgraded to the latest Java agent version v1.21.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.21.0)
- Third party dependency updates.

### 🧰 Bug fixes 🧰
- [Added the smartagent extension to the default agent config to properly source environment variables.](https://github.com/signalfx/splunk-otel-collector/pull/2599)

## v0.70.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.70.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.70.0) and the [opentelemetry-collector-contrib v0.70.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.70.0) releases where appropriate.

### 💡 Enhancements 💡

- Initial [Discovery properties provider](https://github.com/signalfx/splunk-otel-collector/pull/2494) and config incorporation for the discovery mode.
- Third-party dependency updates.

### 🧰 Bug fixes 🧰

- [Addressed SignalFx exporter deferred metadata client initialization](https://github.com/open-telemetry/opentelemetry-collector-contrib/commit/f607cb47c8d972febb9d9d215e0029b3e8cb9884) causing [issues in the Smart Agent receiver](https://github.com/signalfx/splunk-otel-collector/issues/2508).

## v0.69.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.69.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.69.1) and the [opentelemetry-collector-contrib v0.69.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.69.0) releases where appropriate.

### 💡 Enhancements 💡
- Upgraded to the latest [Java agent version (v1.20.0)](https://github.com/signalfx/splunk-otel-collector/pull/2487)
- Upgrade to include changes from [SignalFx Smart Agent v5.27.2](https://github.com/signalfx/signalfx-agent/releases/tag/v5.27.2)
- [Added a variable for Ansible deployments to set NO_PROXY](https://github.com/signalfx/splunk-otel-collector/pull/2482)
- [Updated configuration file for the upstream Collector to enable sync of host metadata](https://github.com/signalfx/splunk-otel-collector/pull/2491)

### 🛑 Breaking changes 🛑
Resource detection for `gke`/`gce` have been combined into the `gcp` resource detector.  While the Splunk Distribution of the Opentelemetry Collector will currently automatically detect and translate any "deprecated" configuration using `gke`/`gce`, [we recommend users with affected configurations specify the new `gcp` detector](https://github.com/signalfx/splunk-otel-collector/pull/2488)

### 🧰 Bug fixes 🧰

- [Added check for nil for k8s attribute, fixing issue causing a core dump on startup](https://github.com/signalfx/splunk-otel-collector/pull/2489)
- [Removed containerd override to address CVE](https://github.com/signalfx/splunk-otel-collector/pull/2466)
- [Updated golang to 1.19.4 to address CVE](https://github.com/signalfx/splunk-otel-collector/pull/2493)

## v0.68.1

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.68.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.68.0) and the [opentelemetry-collector-contrib v0.68.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.68.0) releases where appropriate.

### 💡 Enhancements 💡

- [Added the Windows Log Event Receiver](https://github.com/signalfx/splunk-otel-collector/pull/2449)
- [Ensure config values aren't expanded in discovery mode](https://github.com/signalfx/splunk-otel-collector/pull/2445)
- [Added an example of how to use the recombine operator](https://github.com/signalfx/splunk-otel-collector/pull/2451)

### 🧰 Bug fixes 🧰

- [Fixed link to Java instrumentation agent](https://github.com/signalfx/splunk-otel-collector/pull/2458)

## v0.68.0 (Broken)

### Instrumentation packages are incomplete. Please use release v0.68.1 instead.

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.68.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.68.0) and the [opentelemetry-collector-contrib v0.68.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.68.0) releases where appropriate.

### 💡 Enhancements 💡

- [Moved to upstream Oracle DB receiver(alpha) that captures telemetry such as instance and session specific metrics from an Oracle Database](https://github.com/signalfx/splunk-otel-collector/pull/2381)
- [Upgraded to the latest Java agent version (v1.19.0) for zero configuration auto instrumentation via the Collector](https://github.com/signalfx/splunk-otel-collector/pull/2375)
- [Ensuring the Collector dry run option does not provide expanded final config values](https://github.com/signalfx/splunk-otel-collector/pull/2439)
- [Added capability to disable service name generation for zero configuration auto instrumentation via the Collector](https://github.com/signalfx/splunk-otel-collector/pull/2410)
- [Added upstream Redis receiver (alpha) along with an example; supports TLS](https://github.com/signalfx/splunk-otel-collector/pull/2096)

### 🧰 Bug fixes 🧰

- [Downgrading gopsutil to v3.22.10](https://github.com/signalfx/splunk-otel-collector/pull/2400)
- [Fixed a warning for Salt deployments to set the ballast memory size under an extension instead of memory_limiter processor](https://github.com/signalfx/splunk-otel-collector/pull/2379)

## v0.67.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.67.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.67.0) and the [opentelemetry-collector-contrib v0.67.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.67.0) releases where appropriate.

### 💡 Enhancements 💡

- [add README to packaging/choco directory](https://github.com/signalfx/splunk-otel-collector/pull/2328)
- [Add Azure Eventhub receiver](https://github.com/signalfx/splunk-otel-collector/pull/2342)
- [add support for proxy as part of bosh deployment](https://github.com/signalfx/splunk-otel-collector/pull/2273)
- [PPC support](https://github.com/signalfx/splunk-otel-collector/pull/2308)
- [Add logstransformprocessor from contrib](https://github.com/signalfx/splunk-otel-collector/pull/2246)

### 🧰 Bug fixes 🧰

- [fix image filter to regex match the tag](https://github.com/signalfx/splunk-otel-collector/pull/2357)
- [Rework command line arguments parsing](https://github.com/signalfx/splunk-otel-collector/pull/2343)
- [Temporarily add a no-op flag --metrics-addr](https://github.com/signalfx/splunk-otel-collector/pull/2363)
- [Remove handling of unsupported --mem-ballast-size-mib command line argument](https://github.com/signalfx/splunk-otel-collector/pull/2339)
- [fix digest artifact path](https://github.com/signalfx/splunk-otel-collector/pull/2301)

## v0.66.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.65.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.65.0), which has the same content as [opentelemetry-collector v0.66.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.66.0), the [opentelemetry-collector-contrib v0.65.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.65.0), and the [opentelemetry-collector-contrib v0.66.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.66.0) releases where appropriate.

### 💡 Enhancements 💡

- Add alpha `k8sobjects` receiver [#2270](https://github.com/signalfx/splunk-otel-collector/pull/2270)
- Add Windows 2022 Docker image support [#2269](https://github.com/signalfx/splunk-otel-collector/pull/2269)
- Update internal config source logic better adopt upstream components [#2267](https://github.com/signalfx/splunk-otel-collector/pull/2267) and [#2271](https://github.com/signalfx/splunk-otel-collector/pull/2271)
- Third-party dependency updates

## v0.65.0 (Skipped)

There is no Splunk OpenTelemetry Collector release v0.65.0. The Contrib project [retracted this release](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/16457) for mismatched component dependency versions.

## v0.64.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.64.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.64.0), the [opentelemetry-collector v0.64.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.64.1), and the [opentelemetry-collector-contrib v0.64.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.64.0) releases where appropriate.

### 💡 Enhancements 💡

- Add Zero Config support for installing signalfx-dotnet-tracing instrumentation (#2068)
- Upgrade to Smart Agent release 5.26.0 (#2251)
- Remove usage of opentelemetry-collector experimental config source package (#2267)
- Third-party dependency updates

## v0.63.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.63.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.63.0) and the [opentelemetry-collector-contrib v0.63.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.63.0) releases, and the [opentelemetry-collector v0.63.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.63.1) and the [opentelemetry-collector-contrib v0.63.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.63.1) releases where appropriate.

### 💡 Enhancements 💡

- Experimental --discovery and --dry-run functionality [#2195](https://github.com/signalfx/splunk-otel-collector/pull/2195)
- Upgrade to smart agent release 5.25.0 (#2226)
- unify <ANY> and <VERSION_FROM_BUILD> values and checks[#2179](https://github.com/signalfx/splunk-otel-collector/pull/2179)
- Fix example config for Pulsar exporter, units are nanoseconds [#2185](https://github.com/signalfx/splunk-otel-collector/pull/2185)
- Fix-sa-receiver-link [#2193](https://github.com/signalfx/splunk-otel-collector/pull/2193)
- make dependabot updates weekly [#2215](https://github.com/signalfx/splunk-otel-collector/pull/2215)

## v0.62.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.62.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.62.0) and the [opentelemetry-collector-contrib v0.62.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.62.0) releases.

### 💡 Enhancements 💡

- Increase number of queue consumers in gateway default configuration (#2084)
- Add a new Oracle database receiver (#2011)
- Upgrade to java agent 1.17 (#2161)
- Upgrade to smart agent release 5.24.0 (#2161)
- Update include config source to beta (#2093)

## v0.61.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.61.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.61.0) and the [opentelemetry-collector-contrib v0.61.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.61.0) releases.

### 💡 Enhancements 💡

- `signalfx` exporter: Drop datapoints with more than 36 dimensions [#14625](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/14625)
- Security updates for third-party dependencies

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Reduce severity of logged unsupported config fields warning [#2072](https://github.com/signalfx/splunk-otel-collector/pull/2072)

## v0.60.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.60.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.60.0) and the [opentelemetry-collector-contrib v0.60.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.60.0) releases.

### 💡 Enhancements 💡

- Update auto instrumentation java agent to [v1.16.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.16.0)
- Replace usage of Map.Insert* and Map.Update* with Map.Upsert (#1957)
- Refactor main flags as settings.Settings (#1952)
- Support installing with ansible and skipping restart of services (#1930)

## v0.59.1

### 💡 Enhancements 💡

- Upgrade to include changes from [SignalFx Smart Agent v5.23.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.23.0)
- Add `processlist` and `resourcedetection` to default config

## v0.59.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.59.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.59.0) and the [opentelemetry-collector-contrib v0.59.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.59.0) releases.

### 💡 Enhancements 💡

- Upgrade Golang to 1.19
- debug/configz: Address multiple confmap.Providers for service config and index debug/configz/initial by provider scheme.
- Add tar.gz distribution of Splunk Collector
- Update default gateway config to sync host metadata by default

## v0.58.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.58.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.58.0) and the [opentelemetry-collector-contrib v0.58.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.58.0) releases.

### 💡 Enhancements 💡

- Update auto instrumentation java agent to [v1.14.2](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.14.2)

## v0.57.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.57.2](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.57.2) and the [opentelemetry-collector-contrib v0.57.2](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.57.2) releases.

### 💡 Enhancements 💡

- Include [`sqlquery` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.57.2/receiver/sqlqueryreceiver/README.md)(#1833)
- Security updates for third-party dependencies

## v0.56.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.56.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.56.0) and the [opentelemetry-collector-contrib v0.56.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.56.0) releases.

### 💡 Enhancements 💡

- Add the `--collector-config` option to the Linux installer script to allow a custom config file path (#1806)
- Update auto instrumentation java agent to [v1.14.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.14.0)
- Update bundled Smart Agent to [v5.22.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.22.0)

### 🧰 Bug fixes 🧰

- `signalfx` exporter: Fix invalid error response message [#12654](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12654)

## v0.55.1

### 🧰 Bug fixes 🧰

- `pulsar` exporter: Removed pulsar producer name from config to avoid producer name conflict (#1782)

## v0.55.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.55.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.55.0) and the [opentelemetry-collector-contrib v0.55.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.55.0) releases.

### 💡 Enhancements 💡

- Update default `td-agent` version to 4.3.2 in the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) to support log collection with fluentd on Ubuntu 22.04
- Include [tail_sampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor) and [span_metrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/spanmetricsprocessor) in our distribution

### 🧰 Bug fixes 🧰

- Correct invalid environment variable expansion for ECS task metadata endpoints on EC2 (#1764)
- Adopt [metricstransformprocessor empty metrics fix](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12211)

## v0.54.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.54.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.54.0) and the [opentelemetry-collector-contrib v0.54.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.54.0) releases.

### 💡 Enhancements 💡

- Only use config server if env var unset (#1728)
- Update bundled Smart Agent to [v5.21.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.21.0)

### 🧰 Bug fixes 🧰

- Wrap log messages for windows support bundle (#1725)

## v0.53.1

### 🧰 Bug fixes 🧰

- Upgrade [`metricstransform`
  processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor)
  to pick up [migration from OpenCensus data model to
  OTLP](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10817) that fixes a few issues with
  the processor.

## v0.53.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.53.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.53.0) and the [opentelemetry-collector-contrib v0.53.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.53.0) releases.

### 🚀 New components 🚀

- [`k8sevents` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/k8seventsreceiver)
  to collect Kubernetes events in OpenTelemetry semantics (#1641)
- **Experimental**: [`pulsar` exporter](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/exporter/pulsarexporter) to export metrics to Pulsar (#1683)

## v0.52.2

### 💡 Enhancements 💡

- Upgrade Golang to 1.18.3 (#1633)
- Support multiple `--config` command-line arguments (#1576)

### 🧰 Bug fixes 🧰

- [`kubeletstats` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kubeletstatsreceiver) introduced a regression in version 52.0 that can break metrics for Kubernetes pods and containers, pinning this receiver's version to v0.51.0 until the regression is resolved (#1638)

## v0.52.1

### 🚀 New components 🚀

- [`transform` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor) to modify telemetry based on configuration using the [Telemetry Query Language](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/processing.md#telemetry-query-language) (Alpha)

### 💡 Enhancements 💡

- Initial release of [Chef cookbook](https://supermarket.chef.io/cookbooks/splunk_otel_collector) for Linux and Windows

## v0.52.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.52.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.52.0) and the [opentelemetry-collector-contrib v0.52.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.52.0) releases.

### 💡 Enhancements 💡

- Add Ubuntu 22.04 support to the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible), [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet), and [Salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) (collector only; log collection with Fluentd [not currently supported](https://www.fluentd.org/blog/td-agent-v4.3.1-has-been-released))

## v0.51.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.51.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.51.0) and the [opentelemetry-collector-contrib v0.51.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.51.0) releases.

Additionally, this release includes [an update to the `resourcedetection` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/10015) to support "cname" and "lookup" hostname sources.

### 🛑 Breaking changes 🛑

- Removed Debian 8 (jessie) support from the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (#1354), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible) (#1547), and [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet) (#1545)

### 💡 Enhancements 💡

- Added Debian 11 (bullseye) support to the [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (#1354), [Ansible playbook](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible) (#1547), [Puppet module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/puppet) (#1545), and [Salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) (#1546)
- Upgrade Golang to 1.18.2 (#1551)

## v0.50.1

### 💡 Enhancements 💡

- Security updates for third-party dependencies
- Update bundled Smart Agent to [v5.20.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.20.1)

## v0.50.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.50.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.50.0) and the [opentelemetry-collector-contrib v0.50.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.50.0) releases.

Additionally, this release includes [an update to `k8scluster` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/9523) that allows it to run on older k8s clusters (1.20-).

## v0.49.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.49.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.49.0) and the [opentelemetry-collector-contrib v0.49.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.49.0) releases.

### 🚀 New components 🚀

- [`syslog` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/syslogreceiver) parses Syslogs from tcp/udp using the [opentelemetry-log-collection](https://github.com/open-telemetry/opentelemetry-log-collection) library

### 💡 Enhancements 💡

- Updated the [Migrating from SignalFx Smart Agent to Splunk Distribution of OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/signalfx-smart-agent-migration.md) documentation (#1489)
- Upgrade to Go 1.18.1 (#1464)
- Initial support for [Cloud Foundry Buildpack](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/buildpack) (#1404)
- Initial support for [BOSH Release](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/bosh) (#1480)
- Update bundled Smart Agent to [v5.20.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.20.0)

## v0.48.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.48.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.48.0) and the [opentelemetry-collector-contrib v0.48.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.48.0) releases.

### 🚀 New components 🚀

- [`cloudfoundry` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/cloudfoundryreceiver)
  to receive metrics from Cloud Foundry deployments and services.

## v0.47.1

### 🧰 Bug fixes 🧰

- Remove `signalfx` exporter from traces pipeline in default gateway config (#1393)
- Update `github.com/open-telemetry/opentelemetry-log-collection` to [v0.27.1](https://github.com/open-telemetry/opentelemetry-log-collection/releases/tag/v0.27.1) to fix logging pipeline issues after upgrading to Go 1.18 (#1418)

## v0.47.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.47.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.47.0) and the [opentelemetry-collector-contrib v0.47.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.47.0) releases.

### 🚀 New components 🚀

- [`tcplog` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/tcplogreceiver) to receive logs from tcp using the [opentelemetry-log-collection](https://github.com/open-telemetry/opentelemetry-log-collection) library

### 💡 Enhancements 💡

- Upgrade to Go 1.18 (#1380)

### 🧰 Bug fixes 🧰

- Update core version during build (#1379)
- Update SA event type to fix processlist (#1385)

## v0.46.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.46.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.46.0) and the [opentelemetry-collector-contrib v0.46.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.46.0) releases.

### 💡 Enhancements 💡

- Upgrade `hostmetrics` receiver dependency (#1341)
- Update Linux installer script to fail immediately if running on an unsupported Linux distribution (#1351)
- Update bundled Smart Agent to [v5.19.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.19.1)

### 🧰 Bug fixes 🧰

- As a bug fix for hosts number miscalculation in Splunk Observability Cloud, Splunk OpenTelemetry Collector running in
  agent mode now is configured to override `host.name` attribute of all signals sent from instrumentation libraries by
  default (#1307)

## v0.45.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.45.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.45.0) and the [opentelemetry-collector-contrib v0.45.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.45.1) releases.

### 💡 Enhancements 💡

- Upgrade golang to 1.17.7 (#1294)

### 🧰 Bug fixes 🧰

- Correct collectd/hadoopjmx monitor type in windows Smart Agent receiver config validation [#1254](https://github.com/signalfx/splunk-otel-collector/pull/1254)

## v0.44.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.44.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.44.0) and the [opentelemetry-collector-contrib v0.44.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.44.0) releases.

### 🚀 New components 🚀

- [`databricks` receiver](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/receiver/databricksreceiver) to generate metrics about the operation of a Databricks instance (Alpha)

### 💡 Enhancements 💡

- Bump default `td-agent` version to 4.3.0 in installer scripts (#1205)
- Enable shared pipeline for profiling by default (#1181)
- Update bundled Smart Agent to [v5.19.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.19.0)

## v0.43.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.43.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.43.1) and the [opentelemetry-collector-contrib v0.43.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.43.0) releases.

### 🧰 Bug fixes 🧰

- Provide informative unsupported monitor error on Windows for Smart Agent receiver [#1150](https://github.com/signalfx/splunk-otel-collector/pull/1150)
- Fix Windows support bundle script if fluentd is not installed (#1162)

## v0.42.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.42.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.42.0) and the [opentelemetry-collector-contrib v0.42.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.42.0) releases.

### 🛑 Breaking changes 🛑

- This version adopts OpenTelemetry Core version 0.42.0, and in doing so the configuration parsing process has changed slightly. The Splunk OpenTelemetry Collector used to [evaluate user configuration twice](https://github.com/signalfx/splunk-otel-collector/issues/628) and this required escaping desired `$` literals with an additional `$` character to prevent unwanted environment variable expansion. This version no longer doubly evaluates configuration so any `$$` instances in your configuration as a workaround should be updated to `$`.  [Config source directives](./internal/configsource) that include an additional `$` are provided with a temporary, backward-compatible `$${config_source:value}` and `$$config_source:value` parsing rule controlled by `SPLUNK_DOUBLE_DOLLAR_CONFIG_SOURCE_COMPATIBLE` environment variable (default `"true"`) to migrate them to single `$` usage to continue supporting the updating configs from [#930](https://github.com/signalfx/splunk-otel-collector/pull/930) and [#935](https://github.com/signalfx/splunk-otel-collector/pull/935). This functionality will be removed in a future release (#1099)

### 🚀 New components 🚀

- [`docker_observer`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/dockerobserver) to detect and create container endpoints, to be used with the [`receiver_creator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator) (#1044)
- [`ecs_task_observer`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecstaskobserver) to detect and create ECS task container endpoints, to be used with the [`receiver_creator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/receivercreator) (#1125)

### 💡 Enhancements 💡

- Initial [salt module](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) for Linux (#1050)
- Update bundled Smart Agent to [v5.18.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.18.0)

### 🧰 Bug fixes 🧰

- [`smartagent` receiver](https://github.com/signalfx/splunk-otel-collector/tree/v0.42.0/internal/receiver/smartagentreceiver) will now attempt to create _any_ monitor from a Receiver Creator instance, disregarding its provided `endpoint`. Previously would error out if a monitor did not accept endpoints ([#1107](https://github.com/signalfx/splunk-otel-collector/pull/1107))
- Remove `$$`-escaped `env` config source usage in ECS configs ([#1139](https://github.com/signalfx/splunk-otel-collector/pull/1139)).

## v0.41.1

- Upgrade golang to 1.17.6 (#1088)

## v0.41.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.41.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.41.0) and the [opentelemetry-collector-contrib v0.41.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.41.0) releases.

### 🚀 New components 🚀

- [`journald` receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/journaldreceiver) to parse journald events from systemd journal

### 💡 Enhancements 💡

- Update bundled Smart Agent to [v5.17.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.17.1)
- Update OTLP HTTP receiver endpoint to use port 4318 in default configuration files (#1017)

## v0.40.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.40.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.40.0) and the [opentelemetry-collector-contrib v0.40.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.40.0) releases.

### 🚀 New components 🚀

- [mongodbatlas](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mongodbatlasreceiver) receiver to receive metrics from MongoDB Atlas via their monitoring APIs (#997)
- [routing](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/routingprocessor) processor to route logs, metrics or traces to specific exporters (#982)

## v0.39.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.39.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.39.0) and the [opentelemetry-collector-contrib v0.39.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.39.0) releases.

### 💡 Enhancements 💡

- Initial [Chocolatey package](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/windows-manual.md#chocolatey-installation) release
- Update bundled Smart Agent to [v5.16.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.16.0)

### 🧰 Bug fixes 🧰

- Fix token passthrough for splunkhec receiver/exporter ([#5435](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5435))
- Fix --set command line flag functionality (#939)

## v0.38.1

### 🧰 Bug fixes 🧰

- Fix evaluating env variables in ecs ec2 configs (#930)
- Correct certifi CA bundle removal from Smart Agent bundle (#933)
- Fix evaluating env variables in fargate config (#935)

## v0.38.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.38.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.38.0) and the [opentelemetry-collector-contrib v0.38.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.38.0) releases.

### 💡 Enhancements 💡

- Initial release of multi-arch manifest for amd64 and arm64 linux docker images (#866)
  - **Note:** The Smart Agent and Collectd bundle is only included with the amd64 image
- Enable otlp receiver in the gateway logs pipeline (#903)
- Update bundled Smart Agent to [v5.15.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.15.0)

## v0.37.1

### 💡 Enhancements 💡

- Initial release of [`migratecheckpoint`](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/migratecheckpoint) to migrate Fluentd's position file to Otel checkpoints
- Upgrade golang to v1.17.2 for CVE-2021-38297
- Upgrade `github.com/hashicorp/consul/api` to v1.11.0 for CVE-2021-37219
- Upgrade `github.com/hashicorp/vault` to v1.7.2 for CVE-2021-27400, CVE-2021-29653, and CVE-2021-32923
- Upgrade `github.com/jackc/pgproto3/v2` to v2.1.1
- Upgrade `go.etcd.io/etcd` to `go.etcd.io/etcd/client/v2` for CVE-2020-15114
- Remove test certs from the smart agent bundle (#861)
- Run the `otelcol` container process as non-root user in provided docker image (#864)

### 🧰 Bug fixes 🧰

- Temporarily downgrade `gopsutil` dep to avoid errors in k8s deployment (#877)

## v0.37.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.37.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.37.0) and the [opentelemetry-collector-contrib v0.37.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.37.1) releases. Due to go modules dep issues, the Collector Contrib release 0.37.0 has been retracted in favor of 0.37.1.

### 💡 Enhancements 💡

- `signalfx` exporter: Add support for per cpu metrics [#5756](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5756)
- Add [Hashicorp Nomad](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/nomad) support (#819)
- Add config converter function to unsquash Splunk HEC exporter tls fields (#832)
- Rename `k8s_tagger` processor config entries to [`k8sattributes`](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5384) (#848)
- Update bundled Smart Agent to [v5.14.2](https://github.com/signalfx/signalfx-agent/releases/tag/v5.14.2)

## v0.36.1

### 🚀 New components 🚀

- [`httpsink` exporter](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/exporter/httpsinkexporter) to make span data available via a HTTP endpoint
- Initial release of [`translatesfx`](https://github.com/signalfx/splunk-otel-collector/tree/main/cmd/translatesfx) to translate a SignalFx Smart Agent configuration file into a configuration that can be used by an OpenTelemetry Collector

### 🛑 Breaking changes 🛑

- Reorder detectors in default configs, moving the `system` detector to the
  end of the list. Applying this change to a pre-existing config in an EC2
  or Azure deployment will change both the `host.name` dimension and the
  resource ID dimension on some MTSes, possibly causing detectors to fire.
  (#822)

### 💡 Enhancements 💡

- Add `--skip-collector-repo` and `--skip-fluentd-repo` options to the Linux installer script to skip apt/yum/zypper repo config (#801)
- Add `collector_msi_url` and `fluentd_msi_url` options to the Windows installer script to allow custom URLs for downloading MSIs (#803)
- Start collector service after deb/rpm install or upgrade if env file exists (#805)

### 🧰 Bug fixes 🧰

- Allow the version flag without environment variables (#800)
- Fix Linux installer to set `SPLUNK_MEMORY_TOTAL_MIB` in the environment file if `--ballast` option is specified (#807)

## v0.36.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.36.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.36.0) and the [opentelemetry-collector-contrib v0.36.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.36.0) releases.

### 💡 Enhancements 💡

- Handle backwards compatibility of internal ballast removal (#759)
- Update bundled Smart Agent to [v5.14.1](https://github.com/signalfx/signalfx-agent/releases/tag/v5.14.1)
- Automatically relocate removed OTLP exporter "insecure" field (#783)

### 🧰 Bug fixes 🧰

- Move Heroku buildpack to [https://github.com/signalfx/splunk-otel-collector-heroku](https://github.com/signalfx/splunk-otel-collector-heroku) (#755)
- Fix rpm installation conflicts with the Smart Agent rpm (#773)

## v0.35.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.35.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.35.0) and the [opentelemetry-collector-contrib v0.35.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.35.0) releases.

### 🚀 New components 🚀

- [`groupbyattrs` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor)

### 💡 Enhancements 💡

- Update bundled Smart Agent to [v5.13.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.13.0) (#738)
- Add SUSE support to [Linux installer script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md) (collector only, log collection with Fluentd not yet supported) (#720)
- Add SUSE support to [puppet module](https://forge.puppet.com/modules/signalfx/splunk_otel_collector) (collector only, log collection with Fluentd not yet supported) (#737)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Properly parse receiver creator endpoints (#718)

## v0.34.1

### 💡 Enhancements 💡

- Automatically add `system.type` dimension to all `smartagent` receiver datapoints (#702)
- Include ECS EC2 config in docker images (#713)

## v0.34.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.34.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.34.0) and the [opentelemetry-collector-contrib v0.34.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.34.0) releases.

### 💡 Enhancements 💡

- Add [Amazon ECS EC2](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ecs/ec2) deployment support (#642)
- Enable `memory_ballast` extension in Fargate config (#675)
- Initial [support bundle PowerShell script](https://github.com/signalfx/splunk-otel-collector/blob/main/internal/buildscripts/packaging/msi/splunk-support-bundle.ps1); included in the Windows MSI (#654)
- Remove strict `libcap` dependency from the collector RPM (#676)
  - Allows installation on Linux distros without the `libcap` package.
  - If installing the collector RPM manually, `libcap` will now need to be installed separately as a prerequisite.  See [linux-manual.md](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-manual.md#deb-and-rpm-packages) for details.

### 🧰 Bug fixes 🧰

- Use system env vars for default paths in the Windows installer script (#667)

## v0.33.1

### 💡 Enhancements 💡

- Initial release of the `quay.io/signalfx/splunk-otel-collector-windows` [docker image for Windows](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/windows-manual.md#docker)
- Upgrade to Go 1.17 (#650)
- Update bundled Smart Agent to [v5.12.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.12.0)

## v0.33.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.33.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.33.0) and the [opentelemetry-collector-contrib v0.33.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.33.0) releases.

### 💡 Enhancements 💡

- `smartagent` receiver: `signalfx-forwarder` now works with `k8s_tagger` processor. (#590)
- Add [Fargate](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/fargate/README.md) deployment support
- Update bundled Smart Agent to [v5.11.4](https://github.com/signalfx/signalfx-agent/releases/tag/v5.11.4)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Set redirected logrus logger level (#593)

## v0.31.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.31.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.31.0) and the [opentelemetry-collector-contrib v0.31.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.31.0) releases.

### 🚀 New components 🚀

- [`file_storage` extension](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage)

### 🛑 Breaking changes 🛑

- Change default config server port to 55554 (#579)

### 💡 Enhancements 💡

- Add support for non-persisted journald in the default fluentd config (#516)
- Add `SPLUNK_CONFIG_YAML` env var support for storing configuration YAML (#462)
- Initial puppet support for windows (#524)
- Update to use the `memory_ballast` extension instead of the `--mem-ballast-size-mib` flag (#567)
- Add Heroku buildpack (#571)
- Set required URL and TOKEN env vars for agent config (#572)

### 🧰 Bug fixes 🧰

- Remove SAPM receiver from default configuration (#517)
- `zookeeper` config source: Remove config validation for zk endpoints (#533)
- Fix memory limit calculation for deployments with 20Gi+ of total memory (#558)
- Set path ownership on deb/rpm postinstall (#582)

## v0.29.0

This Splunk OpenTelemetry Connector release includes changes from the [opentelemetry-collector v0.29.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.29.0) and the [opentelemetry-collector-contrib v0.29.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.29.0) releases.

### 💡 Enhancements 💡

- Add OTLP to logs pipeline for agent (#495)
- Enable collecting in memory config locally by default (#497)
- Enable host metadata updates by default (#513)

## v0.28.1

- Update bundled Smart Agent to [v5.11.0](https://github.com/signalfx/signalfx-agent/releases/tag/v5.11.0) (#487)
- Document APM infra correlation (#458)
- Alpha translatesfx feature additions.

## v0.28.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.28.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.28.0) and the [opentelemetry-collector-contrib v0.28.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.28.0) releases.

### 💡 Enhancements 💡

- Initial puppet module for linux (#405)
- Add `include` config source (#419, #402, #397)
- Allow setting both `SPLUNK_CONFIG` and `--config` with priority given to `--config` (#450)
- Use internal pipelines for collector prometheus metrics (#469)

### 🧰 Bug fixes 🧰

- Correctly handle nil value on the config provider (#434)

## v0.26.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.26.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.26.0) and the [opentelemetry-collector-contrib v0.26.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.26.0) releases.

### 🚀 New components 🚀

- [kafkametrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/kafkametricsreceiver) receiver

### 💡 Enhancements 💡

- zookeeper config source (#318)
- etcd2 config source (#317)
- Enable primary cloud resource detection in the default agent config (#344)
- Unset exclusion and translations by default in gateway config (#350)
- Update bundled Smart Agent to [v5.10.2](https://github.com/signalfx/signalfx-agent/releases/tag/v5.10.2) (#354)
- Set PATH in the docker image to include Smart Agent bundled utilities (#313)
- Remove 55680 exposed port from the docker image (#371)
- Expose initial and effective config for debugging purposes (#325)
- Add a config source for env vars (#348)

### 🧰 Bug fixes 🧰

- `smartagent` receiver: Remove premature protection for Start/Stop, trust Service to start/stop once (#342)
- `smartagent` receiver and extension: Fix config parsing for structs and pointers to structs (#345)

## v0.25.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.25.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.25.0) and the [opentelemetry-collector-contrib v0.25.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.25.0) releases.

### 🚀 New components 🚀

- [filelog](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) receiver (#289)
- [probabilisticsampler](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) processor (#300)

### 💡 Enhancements 💡

- Add the config source manager (#295, #303)

### 🧰 Bug fixes 🧰

- Correct Jaeger Thrift HTTP Receiver URL to /api/traces (#288)

## v0.24.3

### 💡 Enhancements 💡

- Add AKS resource detector (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3035)

### 🧰 Bug fixes 🧰

- Fallback to `os.Hostname` when FQDN is not available (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3099)

## v0.24.2

### 💡 Enhancements 💡

- Include smart agent bundle in docker image (#241)
- Use agent bundle-relative Collectd ConfigDir default (#263, #268)

### 🧰 Bug fixes 🧰

- Sanitize monitor IDs in SA receiver (#266, #269)

## v0.24.1

### 🧰 Bug fixes 🧰

- Fix HEC Exporter throwing 400s (https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3032)

### 💡 Enhancements 💡
- Remove unnecessary hostname mapping in fluentd configs (#250)
- Add OTLP HTTP exporter (#252)
- Env variable NO_WINDOWS_SERVICE to force interactive mode on Windows (#254)

## v0.24.0

### 🛑 Breaking changes 🛑

- Remove opencensus receiver (#230)
- Don't override system resource attrs in default config (#239)
  - Detectors run as part of the `resourcedetection` processor no longer overwrite resource attributes already present.

### 💡 Enhancements 💡

- Support gateway mode for Linux installer (#187)
- Support gateway mode for windows installer (#231)
- Add SignalFx forwarder to default configs (#218)
- Include Smart Agent bundle in msi (#222)
- Add Linux support bundle script (#208)
- Add Kafka receiver/exporter (#201)

### 🧰 Bug fixes 🧰

## v0.23.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.23.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.23.0) and the [opentelemetry-collector-contrib v0.23.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.23.0) releases.

### 🛑 Breaking changes 🛑

- Renamed default config from `splunk_config_linux.yaml` to `gateway_config.yaml` (#170)

### 💡 Enhancements 💡

- Include smart agent bundle in amd64 deb/rpm packages (#177)
- `smartagent` receiver: Add support for logs (#161) and traces (#192)

### 🧰 Bug fixes 🧰

- `smartagent` extension: Ensure propagation of collectd bundle dir (#180)
- `smartagent` receiver: Fix logrus logger hook data race condition (#181)
