# Changelog

## Unreleased

- (Splunk) `receiver/discovery`: Reduce amount of attributes sent with the entities to the required set ([#6419](https://github.com/signalfx/splunk-otel-collector/pull/6419))

## v0.129.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.129.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.129.0)
and the [opentelemetry-collector-contrib v0.129.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.129.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `exporterhelper`: Remove deprecated sending_queue::blocking options, use sending_queue::block_on_overflow. ([#13211](https://github.com/open-telemetry/opentelemetry-collector/pull/13211))
- (Contrib) `kafka`: remove broken SASL/AWS_MSK_IAM auth mechanism from Kafka components ([#40554](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40554))
- (Contrib) `kafkareceiver`: Improve kafkareceiver internal metrics telemetry ([#40816](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40816))
  This change adds the following internal telemetry metrics to the receiver:
    - kafka_broker_connects
    - kafka_broker_closed
    - kafka_broker_throttling_duration
    - kafka_receiver_bytes
    - kafka_receiver_bytes_uncompressed
    - kafka_receiver_latency

  In addition it removes explicit component "name" metrics attribute which should be enabled with "telemetry.newPipelineTelemetry"
  collector feature gate instead. And changes "partition" metrics attribute to more direct "int64" type instead of "string".
- (Contrib) `receiver/httpcheck`: Always close resp.Body and add timeout ([#40552](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40552))
- (Contrib) `sqlserverreceiver`: Only empty strings for query texts and query plans when obfuscation failed. ([#40527](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40527))
  This change applies only to top query collection and query sample collection.

### 🚩 Deprecations 🚩

- (Splunk) `packaging`: Fluentd support has been deprecated in packages and will be removed in a future release.
  Please refer to [deprecation documentation](./docs/deprecations/fluentd-support.md) for more information ([#6366](https://github.com/signalfx/splunk-otel-collector/pull/6366))
- (Contrib) `resourcedetectionprocessor`: Promote the processor.resourcedetection.removeGCPFaasID feature gate to beta. ([#40601](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40601))
  The faas.id attribute is replaced by the faas.instance attribute. | This disables detection of the faas.id resource attribute by default. | Re-enable by disabling the processor.resourcedetection.removeGCPFaasID feature gate.
- (Contrib) `resourcedetectionprocessor`: Remove the stable processor.resourcedetection.hostCPUSteppingAsString featuregate ([#40569](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40569))
  It was supposed to be removed in v0.110.0.
- (Contrib) `prometheusreceiver`: Promote the receiver.prometheusreceiver.RemoveLegacyResourceAttributes featuregate to stable ([#40572](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40572))
  It has been beta since v0.126.0

### 💡 Enhancements 💡

- (Splunk) Update `splunk-otel-javaagent` to `v2.17.0` ([#6397](https://github.com/signalfx/splunk-otel-collector/pull/6397))
- (Splunk) Update `splunk-otel-js` to `v3.3.0` ([#6396](https://github.com/signalfx/splunk-otel-collector/pull/6396))
- (Splunk) `receiver/discovery`: Propagate logs reported by dynamic receivers ([#6388](https://github.com/signalfx/splunk-otel-collector/pull/6388))
  Sample and propagate all logs from the dynamically created receivers to the stdout. This brings visibility to discovery issues that are not covered by the matching discovery rules
- (Splunk) `receiver/discovery`: Set service.type as part of a discovery rule ([#6400](https://github.com/signalfx/splunk-otel-collector/pull/6400))
  Deducing service.type from the existing data isn't reliable. For example, discovered kafka service gets "kafkametrics" name.
  This changes adds another required field to a service discovery rule to signify what type of a service is being discovered.
- (Core) `service`: Support setting `sampler` and `limits` under `service::telemetry::traces` ([#13201](https://github.com/open-telemetry/opentelemetry-collector/pull/13201))
  This allows users to enable sampling and set span limits on internal Collector traces using the
  OpenTelemetry SDK declarative configuration.
- (Contrib) `windowseventlogreceiver`: Added option to prevent collector shutdown due to errors while opening channels ([#36237](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36237))
- (Contrib) `azuremonitorreceiver`: Add support for azureauth when batch api is enabled. ([#40872](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40872))
- (Contrib) `azuremonitorreceiver`: fix unnecessarily exported struct ([#40662](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40662))
- (Contrib) `azuremonitorreceiver`: Added new config that sets the number of unique resource IDs to fetch per Batch API call ([#40112](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40112))
- (Contrib) `hostmetricsreceiver`: Skeleton nfsscraper -- adds Linux nfs and nfsd metrics from /proc ([#39978](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39978))
- (Contrib) `receiver/hostmetrics`: Add `system.paging.faults` metrics on Windows ([#40468](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40468))
- (Contrib) `elasticsearchreceiver`: Add support for ES version 9 ([#39929](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39929))
- (Contrib) `kafkareceiver`: Add an Alpha feature gate `receiver.kafkareceiver.UseFranzGo` to use franz-go in the Kafka receiver for better performance. ([#40628](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40628))
  Adds an experimental opt-in support to use the franz-go client in the Kafka receiver.
  The franz-go client is a high-performance Kafka client that can improve the performance of the Kafka receiver.
  The default client remains sarama, which is used by the Kafka exporter and other components.
  Enable the franz-go client by setting the `receiver.kafkareceiver.UseFranzGo` feature gate.
- (Contrib) `kafkareceiver`: Refactor Sarama client to pave the way for franz-go `featuregate` opt-in ([#40628](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40628))
- (Contrib) `kafkaexporter`: Improve kafkaexporter internal metrics telemetry ([#40815](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40815))
- (Contrib) `kafkareceiver`: Create the consumer group asynchronously, and report component status ([#40516](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40516))
  This change ensures that the collector can start up if the Kafka cluster is not available
  or the consumer group creation fails due to some other transient, retryable error.
- (Contrib) `sqlserverreceiver`: add feature gate for removing metrics resource attributes `server.address`, and `server.port` ([#40579](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40579))
  These are not resource attributes according to the semantic conventions.
  To enable this feature gate, when running the otel collector, pass the following argument `--feature-gates=receiver.sqlserver.RemoveServerResourceAttribute`.
- (Contrib) `sqlqueryreceiver`: Add SQL connection fields `host`, `port`, `database`, `username`, `password`, and `additional_params`. ([#39760](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39760))
  These options can be used instead of the existing `datasource` configuration option.
- (Contrib) `sqlserverreceiver`: Update obfuscation library to use `mssql` mode. ([#40525](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40525))
  Now it can obfuscate more SQL Server specific queries which may not work before.
- (Contrib) `processor/transform`: Introduce optional metric name suffix setup for metric conversion functions ([#33850](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33850))
  The affected functions are: convert_summary_count_val_to_sum(), convert_summary_sum_val_to_sum(), extract_count_metric(), extract_sum_metric()

### 🧰 Bug fixes 🧰

- (Core) `pdata`: Fix copying of optional fields when the source is unset. ([#13268](https://github.com/open-telemetry/opentelemetry-collector/pull/13268))
- (Core) `service`: Only allocate one set of internal log sampling counters ([#13014](https://github.com/open-telemetry/opentelemetry-collector/pull/13014))
  The case where logs are only exported to stdout was fixed in v0.126.0;
  this new fix also covers the case where logs are exported through OTLP.
- (Contrib) `collectdreceiver`: Ensure the receiver closes its port when shutting down quickly after starting. ([#40406](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40406))
- (Contrib) `hostmetricsreceiver`: Minor fix to the handling of conntrack errors ([#40175](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40175))
- (Contrib) `k8sattributesprocessor`: Make sure getIdentifiersFromAssoc() can handle container.id ([#40745](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40745))
- (Contrib) `kafka`: Fixes a defect introduced in [#39115](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39115) that prevents MSK IAM auth from working ([#40720](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40720))
- (Contrib) `kafkareceiver`: Protect partition lost/revoke callback against unassigned groups ([#40920](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40920))
  Note that due to the nature of the timing issue this is extremely unlikely to affect a real user,
  and really only likely to occur in unit tests.
  in the metadata file, the `host.name` resource attribute is defined for metrics, but it is not exported in reality. this fixed the issue and export it as a resource attribute.
- (Contrib) `opampextension`: Set the default value of `ReportsAvailableComponents` to `true` ([#40485](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40485))
- (Contrib) `receivercreator`: Properly handle default endpoint for annotation discovery ([#40712](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40712))
  Annotation discovery should not add the default endpoint explicitly. Configuration
  should be left empty if users do not set anything. Main flow of the receiver_creator
  already adds the default endpoint, if not provided, at a later stage by also checking
  if target receiver supports it. This patch ensures this logic is unified for both
  annotation discovery and templated based discovery.
  IAM auth requires TLS, but the config translation was enabling SASL when it intended to enable TLS
- (Contrib) `receiver/httpcheck`: Fix resource leak caused by failure to close request body ([#40695](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40695))
- (Contrib) `receiver/simpleprometheus`: Fix bug that caused the labels configuration option to be ignored ([#40722](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40722))
- (Contrib) `routingconnector`: Copy IsMonotonic and AggregationTemporality when routing metrics on datapoint context. ([#37273](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37273))
- (Contrib) `snowflakereceiver`: adds a pre-scrape check to see if metrics are enabled. ([#40418](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40418))
- (Contrib) `snowflakereceiver`: Correct handling of datatypes of Snowflake receiver ([#40418](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40418))
- (Contrib) `snowflakereceiver`: Rectified datatypes of storagemetrics of Snowflake receiver ([#40857](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40857))
- (Contrib) `sqlserverreceiver`: export `host.name` as resource attributes for metrics ([#40576](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40576))

## v0.128.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.128.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.128.0)
and the [opentelemetry-collector-contrib v0.128.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.128.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `service/telemetry`: Mark "telemetry.disableAddressFieldForInternalTelemetry" as stable ([#13152](https://github.com/open-telemetry/opentelemetry-collector/pull/13152))
- (Contrib) `pkg`: Remove the fields from category FrontDoorWebApplicationFirewallLog from the body log record and place them as log record attributes. ([#39993](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39993))
- (Contrib) `countconnector`: Upgrade profiles proto to 1.7.0 ([#40285](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40285))
- (Contrib) `pkg/ottl`: Upgrade profiles proto to 1.7.0 ([#40285](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40285))
- (Contrib) `pkg/ottl`: Remove access to the profile lookup tables ([#40227](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40227))
  the mapping_table, location_table, function_table, attribute_table, attribute_units, link_table, string_stable have been moved to a root dictionary attribute and are not part of profile anymore.
- (Contrib) `sqlserverreceiver`: Rename flags for enabling top query collection and query sample collection. ([#40416](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40416))
  Renamed `top_query_collection.enabled` to `events."db.server.top_query".enabled` in top query collection.
  Renamed `query_sample_collection.enabled` to `events."db.server.query_sample".enabled` in query sample collection.

### 🚩 Deprecations 🚩

- (Splunk) `deployments`: Fluentd support has been deprecated in all mass deployments and will be removed in a future release.
  Please refer to [deprecation documentation](./docs/deprecations/fluentd-support.md) for more information ([#6339](https://github.com/signalfx/splunk-otel-collector/pull/6339))
- (Splunk) `Windows installer script`: Fluentd support has been deprecated and will be removed in a future release. ([#6362](https://github.com/signalfx/splunk-otel-collector/pull/6362))
  Please refer to [deprecation documentation](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md) for more information.
- (Contrib) `receiver/hostmetrics`: Mark `hostmetrics.process.onWindowsUseNewGetProcesses` feature gate as stable ([#32947](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32947)

### 💡 Enhancements 💡

- (Splunk) `gateway`: Support discovery entities for gateway agent config ([#6311](https://github.com/signalfx/splunk-otel-collector/pull/6311))
- (Core) `confighttp`: Update the HTTP server span naming to use the HTTP method and route pattern instead of the path. ([#12468](https://github.com/open-telemetry/opentelemetry-collector/pull/12468))
  The HTTP server span name will now be formatted as `<http.request.method> <http.route>`.
  If a route pattern is not available, it will fall back to `<http.request.method>`.
- (Core) `service`: Use configured loggers to log errors as soon as it is available ([#13081](https://github.com/open-telemetry/opentelemetry-collector/pull/13081))
- (Core) `service`: Remove stabilized `featuregate` useOtelWithSDKConfigurationForInternalTelemetry ([#13152](https://github.com/open-telemetry/opentelemetry-collector/pull/13152))
- (Contrib) `resourcedetectionprocessor`: Add the option to retrieve resource attributes from the K8s API server and EC2 api when the IMDS service is not available. ([#39503](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39503))
- (Contrib) `windowseventlogreceiver`: Add a boolean option to include the `log.record.original` attribute of each event record. ([#40365](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40365))
- (Contrib) `resourcedetectionprocessor`: Add additional OS properties to resource detection: `os.build.id` and `os.name` ([#39941](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39941))
- (Contrib) `receivercreator`: Add an option to set default annotations to override the discovery logic ([#37436](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37436))
- (Contrib) `kafkaexporter`: Add an Alpha feature gate `exporter.kafkaexporter.UseFranzGoClient` to use franz-go in the Kafka exporter for better performance. ([#40364](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40364))
  This change adds an experimental opt-in support to use the franz-go client in the Kafka exporter.
  The franz-go client is a high-performance Kafka client that can improve the performance of the Kafka exporter.
  The default client remains sarama, which is used by the Kafka receiver and other components.
  Enable the franz-go client by setting the `exporter.kafkaexporter.UseFranzGo` feature gate.
- (Contrib) `filelogreceiver`: The fingerprint of gzip compressed files is created by decompressing and reading the first `fingerprint_size` bytes. ([#37772](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37772))
  This feature can be enabled via the following feature gate `--feature-gates=filelog.decompressFingerprint`. This can cause existing gzip files to be re-ingested because of changes in how fingerprints are computed.
- (Contrib) `sqlserverreceiver`: add one attribute for performance_counter.object_name to deduplicate data ([#40359](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40359))
- (Contrib) `processor/k8sattributes`: Support extracting labels and annotations from k8s Deployments ([#37957](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37957))
- (Contrib) `receiver/k8s_cluster`: Add onDelete handler to emit the experimental entity delete events ([#40278](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40278))
- (Contrib) `processor/resourcedetection`: add `host.interface` resource attribute to `system` detector ([#39419](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39419))
- (Contrib) `pkg/ottl`: Add `event_name` path to the OTTL Log context ([#40230](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40230))
- (Contrib) `pkg/ottl`: Add metric.metadata as a valid OTTL path ([#40214](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40214))
- (Contrib) `pkg/ottl`: Add ability to compare slices/pcommon.Slice in OTTL Boolean Expressions ([#40370](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40370))
- (Contrib) `postgresqlreceiver`: add the ability to obtain query plan for top n queries ([#39995](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39995))
- (Contrib) `splunkenterprisereceiver`: Add search artifact size metrics to the existing search dispatch collection, which uses the /services/server/status/dispatch-artifacts API. ([#40383](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40383))
- (Contrib) `pkg/ottl`: Improve time parsing error messages by including the ctime directive instead of the go time layout ([#35176](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35176))

### 🧰 Bug fixes 🧰

- (Core) `telemetry`: Add generated resource attributes to the printed log messages. ([#13110](https://github.com/open-telemetry/opentelemetry-collector/pull/13110))
  If service.name, service.version, or service.instance.id are not specified in the config, they will be generated automatically.
  This change ensures that these attributes are also included in the printed log messages.
- (Core) `confmap`: Do not panic on assigning nil maps to non-nil maps ([#13117](https://github.com/open-telemetry/opentelemetry-collector/pull/13117))
- (Core) `pdata`: Fix event_name skipped when unmarshalling LogRecord from JSON ([#13127](https://github.com/open-telemetry/opentelemetry-collector/pull/13127))
- (Contrib) `kafkareceiver`: enforce a backoff mechanism on non-permanent errors, such as when the queue is full ([#39580](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39580))
- (Contrib) `kafkaexporter`: Fix Snappy compression codec support for the Kafka exporter ([#40288](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40288))
- (Contrib) `kafkareceiver`: Don't restart the Kafka consumer on failed errors when message marking is enabled for them. ([#39909](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39909))
  The Kafka consumer will now only be restarted due to failed messages if
  both `message_marking::after` is true, and `message_marking::on_error` is false.
  In this case the message will not be marked as consumed, and the consumer will
  restart with an error.
- (Contrib) `signalfxexporter`: Fix invalid and missing host metadata when the collector is deployed in a containerized environment and | the host root filesystem is mounted to non-standard location. ([#40218](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40218))
  Use the newly introduced `root_path` configuration option to specify the root path of the host filesystem.| This is required when the host root filesystem is mounted to a non-standard location.
- (Contrib) `sqlserverreceiver`: Fix incorrect attribute name for database system name in the top query event collection. ([#40361](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40361))
  Rename `db.server.name` to `db.system.name` in SQL Server receiver to align with semantic conventions.
- (Contrib) `sqlserverreceiver`: Fix scraping failure on SQL obfuscation on top query collection and query sample collection. ([#40347](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40347))
  Collector will now report original SQL statement if the obfuscation is failed.

## v0.127.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.127.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.127.0)
and the [opentelemetry-collector-contrib v0.127.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.127.0) releases where appropriate.

### ❗ Known Issues ❗

- **`simpleprometheusreceiver`: Metrics from this receiver may be missing labels or exhibit unexpected label behavior**  
  An issue has been identified in this receiver where labels are not being passed and processed correctly.
  This means Prometheus metrics may be missing expected labels or exhibit unexpected label behavior, 
  which can affect internal Collector processing as well as downstream dashboards, alerts, and metric analysis.
  See contrib issue [#40722](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/40722) for more details on this bug.

### 🛑 Breaking changes 🛑

- (Contrib) `sqlserverreceiver`: Zero values in delta attributes will be reported in top query collection. ([#40041](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40041))
  This change includes the following attributes:
  - `sqlserver.total_worker_time`
  - `sqlserver.execution_count`
  - `sqlserver.total_logical_reads`
  - `sqlserver.total_logical_writes`
  - `sqlserver.total_physical_reads`
  - `sqlserver.total_rows`
  - `sqlserver.total_grant_kb`
- (Contrib) `receiver/sqlserver`: Enable `sqlserver.page.life_expectancy` metric for all configurations ([#39940](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39940))
  This metric was originally only available when running on Windows, but is
  now available for all configuration options, including direct connection.
  Since this metric is enabled by default, users who have direct connection
  configured will now have this metric emitted by default.
  - NOTE: Marked as breaking only because new metrics are enabled by default, potentially impacting resource usage and billing.
- (Core) `service`: Add size metrics defined in Pipeline Component Telemetry RFC ([#13032](https://github.com/open-telemetry/opentelemetry-collector/pull/13032))
  - See [Pipeline Component Telemetry RFC](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/component-universal-telemetry.md) for more details:
    - `otelcol.receiver.produced.size`
    - `otelcol.processor.consumed.size`
    - `otelcol.processor.produced.size`
    - `otelcol.connector.consumed.size`
    - `otelcol.connector.produced.size`
    - `otelcol.exporter.consumed.size`
  - NOTE: Marked as breaking only because new metrics are enabled by default, potentially impacting resource usage and billing.

### 🚩 Deprecations 🚩

- (Splunk) `Linux installer script`: Fluentd support has been deprecated and will be removed in a future release. ([#6264](https://github.com/signalfx/splunk-otel-collector/pull/6264))
  - Please use native OTel Collector receivers instead.
  - A common replacement for Fluentd's functionality is the [filelog receiver](https://help.splunk.com/en/splunk-observability-cloud/manage-data/available-data-sources/supported-integrations-in-splunk-observability-cloud/opentelemetry-receivers/filelog-receiver).
    Many common configuration examples of the `filelog` receiver can be found in the [logs_config_linux.yaml](https://github.com/signalfx/splunk-otel-collector/blob/87bee7ae45b08be8d143a758d0f7004fd92d8f60/cmd/otelcol/config/collector/logs_config_linux.yaml) file.

### 💡 Enhancements 💡

- (Splunk) Add an install property, `COLLECTOR_SVC_ARGS`, to the Windows MSI to configure the command-line arguments used to launch the collector service on Windows. ([#6268](https://github.com/signalfx/splunk-otel-collector/pull/6268))
- (Splunk) `discoveryreceiver` Add more metrics that are enabled by default for the SQL Server receiver ([#6259](https://github.com/signalfx/splunk-otel-collector/pull/6259))
- (Splunk) `discoveryreceiver` Send delete entity events for discovered services ([#6260](https://github.com/signalfx/splunk-otel-collector/pull/6260))
- (Splunk) Add version change support to Windows install script ([#6304](https://github.com/signalfx/splunk-otel-collector/pull/6304))
- (Core) `exporter/debug`: Display resource and scope in `normal` verbosity ([#10515](https://github.com/open-telemetry/opentelemetry-collector/pull/10515))
- (Contrib) `azuremonitorreceiver`: Add support for azureauthextension as a token provider for azuremonitorreceiver. ([#39048](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39048))
- (Contrib) `kafkaexporter`: Allow Kafka exporter to produce to topics based on metadata key values ([#39208](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39208))
  Allows the Kafka exporter to dynamically use a signal's export target topic based
  on the value of the pipeline's metadata, allowing dynamic signal routing.
- (Contrib) `processor/tailsampling`: Add first policy match decision to tailsampling processor ([#36795](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36795))
- (Contrib) `receiver/k8sclusterreceiver`: Added new resource attributes `k8s.hpa.scaletargetref.kind`, `k8s.hpa.scaletargetref.name`, and `k8s.hpa.scaletargetref.apiversion` to the `k8s.hpa` resource. These attributes are disabled by default. ([#38768](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38768))
- (Contrib) `k8sobserver`: Add namespaces setting for scoping k8s client to specific namespaces ([#39677](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39677))
- (Contrib) `awss3exporter`: Add the retry mode, max attempts and max backoff to the settings ([#36264](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36264))
- (Contrib) `k8sattributesprocessor`: Add option to configure automatic service resource attributes ([#37114](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37114))
  Implements [Service Attributes](https://opentelemetry.io/docs/specs/semconv/non-normative/k8s-attributes/#service-attributes).
  If you are using the file log receiver, you can now create the same resource attributes as traces (via OTLP) received
  from an application instrumented with the OpenTelemetry Operator -
  simply by adding the
  `extract: { metadata: ["service.namespace", "service.name", "service.version", "service.instance.id"] }`
  configuration to the `k8sattributesprocessor` processor.
  See the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/k8sattributesprocessor/README.md#configuring-recommended-resource-attributes) for more details.
- (Contrib) `receiver/sqlserver`: Add new metric for disk IO rate on a resource pool ([#39977](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39977))
  The new metric `sqlserver.resource_pool.disk.operations` is disabled by default.
- (Contrib) `receiver/sqlserver`: Add new metric `sqlserver.lock.wait.count` ([#39892](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39892))
  This metric is disabled by default.
- (Contrib) `receiver/sqlserver`: Add new metric to track OS wait times ([#39977](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39977))
  The new metric is named `sqlserver.os.wait.duration` and disabled by default.
- (Contrib) `sqlserverreceiver`: Add configuration option `top_query_collection.collection_interval` for top query collection to make the collection less frequent. ([#40002](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40002))
  - This change only applies to the `top_query_collection` feature.
  - The default value is `60s`

### 🧰 Bug fixes 🧰

- (Core) `confmap`: Do not panic on assigning nil maps to non-nil maps ([#13117](https://github.com/open-telemetry/opentelemetry-collector/pull/13117))
  - This fix was backported from [opentelemetry-collector v0.128.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.128.0) to ensure the bug does not affect splunk-otel-collector v0.127.0.
- (Contrib) `azuremonitorreceiver`: Use ``metrics`` aggregation filter when ``use_batch_api: true`` ([#40079](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40079))
- (Contrib) `postgresqlreceiver`: Fix too many top query got reported. Top query should only report those queries were executed during the query interval ([#39942](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39942))
- (Contrib) `filelogreceiver`: Introduce `utf8-raw` encoding to avoid replacing invalid bytes with \uFFFD when reading UTF-8 input. ([#39653](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39653))
- (Contrib) `internal/splunk`: Treat HTTP 403 Forbidden as a permanent error. ([#39037](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39037))
  - Splunk responses with a 403 typically indicate an authentication or authorization issue that is not likely to be resolved by retrying.
  - This change ensures that the error is treated as permanent to avoid unnecessary retries.
  - This change is applicable to the `splunkhecexporter` and `signalfxexporter` components.
- (Contrib) `spanmetricsconnector`: Fix bug causing span metrics calls count to be always 0 when using delta temporality ([#40139](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40139))
- (Contrib) `kafkareceiver, kafkaexporter`: Add support for named encoding extensions in kafkareceiver and kafkaexporter ([#40142](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/40142))

## v0.126.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.126.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.126.0)
and the [opentelemetry-collector-contrib v0.126.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.126.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `azuremonitorreceiver`: Renamed the `auth` field to `credentials` in the azuremonitorreceiver configuration. ([#39738](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39738))
- (Contrib) `googlecloudpubsubreceiver`: Add `ignore_encoding_error` configuration to ignore decoding failures from the configured encoder ([#38164](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38164))
  Introduce a setting to ignore errors when the configured encoder. It's advised to set this to `true` when using
  a custom encoder, and use the new `receiver.googlecloudpubsub.encoding_error` metric to monitor the number of errors.
  Ignoring the error will cause the receiver to drop the message.

- (Contrib) `sqlserverreceiver`: Update event properties ([#39868](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39868))
  This change only applies to top query events and query sample events:
  * Update event name of top query event from `top query` to `db.server.top_query`
  * Update event name of query sample event from `query sample` to `db.server.query_sample`
  * Remove body of query sample event
  * Rename `sqlserver.username` to `user.name` in query sample event

- (Contrib) `receiver/prometheus`: Upgrade `receiver.prometheusreceiver.RemoveLegacyResourceAttributes` feature gate to `beta` stability (enabled by default) ([#32814](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32814))
  Disable the `receiver.prometheusreceiver.RemoveLegacyResourceAttributes` feature gate to restore previous behavior. This feature gate will be removed in a future release.
  The feature gate is used to replace the following legacy resource attributes:
  `net.host.name` -> `server.address`
  `net.host.port` -> `server.port`
  `http.scheme` -> `url.scheme`

### 🚩 Deprecations 🚩

- (Splunk) `receiver/lightprometheus`: The following configurable `resource_attributes` are being renamed to match semantic conventions. ([#6257](https://github.com/signalfx/splunk-otel-collector/pull/6257))
  - `net.host.name` -> `server.address`
  - `net.host.port` -> `server.port`
  - `http.scheme`   -> `url.scheme`

  `net.host.name`, `net.host.port`, and `http.scheme` are now considered to be deprecated and will be removed in a future release.

- (Contrib) `googlecloudpubsubreceiver`: Add deprecation warning for the build-in encoders ([#39371](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39371))
  The build-in encoders `cloud_logging` and `raw_text` both have encoding extension alternatives and will be removed
  in version v0.132.0 of the collector.

- (Contrib) `processor/tailsampling`: The invert decisions (InvertSampled and InvertNotSampled) have been deprecated, please make use of drop policy to explicitly not sample select traces. ([#39833](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39833))

### 💡 Enhancements 💡

- (Splunk) `deployments/nomad`: Add official support for `v1.9.7` ([#6248](https://github.com/signalfx/splunk-otel-collector/pull/6248))

- (Core) `confmap`: Update the behavior of the confmap.enableMergeAppendOption feature gate to merge only component lists. ([#12926](https://github.com/open-telemetry/opentelemetry-collector/pull/12926))
- (Core) `service`: Add item count metrics defined in Pipeline Component Telemetry RFC ([#12812](https://github.com/open-telemetry/opentelemetry-collector/pull/12812))
  See [Pipeline Component Telemetry RFC](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/component-universal-telemetry.md) for more details:
  - `otelcol.receiver.produced.items`
  - `otelcol.processor.consumed.items`
  - `otelcol.processor.produced.items`
  - `otelcol.connector.consumed.items`
  - `otelcol.connector.produced.items`
  - `otelcol.exporter.consumed.items`

- (Core) `tls`: Add trusted platform module (TPM) support to TLS authentication. ([#12801](https://github.com/open-telemetry/opentelemetry-collector/pull/12801))
  Now the TLS allows the use of TPM for loading private keys (e.g. in TSS2 format).

- (Contrib) `fileconsumer`: Support "`compression:auto`" which automatically detects file compression type and ingests data. Currently, only gzip compressed files are auto detected. This allows for mix of compressed and uncompressed files to be ingested with the same filelogreceiver. ([#37772](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37772))
- (Contrib) `kafkaexporter`: Add compression level in kafka producer. ([#39772](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39772))
- (Contrib) `splunkenterprisereceiver`: adds `splunk.indexer.rollingrestart.status` metric ([#39023](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39023))
- (Contrib) `carbonreceiver`: Support unnamed groups in carbon receiver regex parser ([#39137](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39137))
- (Contrib) `splunkenterprisereceiver`: Added a new `splunk.search.initiation`, `splunk.search.duration`, `splunk.search.status`, and `splunk.search.success` metrics. ([#39566](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39566))
- (Contrib) `eventlogreceiver`: add raw XML query filtering option ([#39055](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39055))
- (Contrib) `mongodbatlasreceiver`: Add support for setting custom base URL for the MongoDB Atlas API ([#39345](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39345))
- (Contrib) `postgresqlreceiver`: Reduce component footprint by removing the loading of unnecessary SQL drivers ([#39918](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39918))
  Custom builds that are using the postgresql receiver and no other SQL related receivers
  will see a reduction in the output binary size. A similar effect is expected also for
  the sqlserverreceiver.

- (Contrib) `spanmetricsconnector`: Separate Dimensions for calls and duration metrics ([#36805](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36805))
  Add two new fields to the settings: `histogram.dimensions` and `calls_dimensions`.
  Use them to add independent dimensions to the duration and calls metrics.

- (Contrib) `pkg/ottl`: Add support for HasPrefix and HasSuffix functions ([#39825](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39825))
- (Contrib) `k8sobjectsreceiver`: k8sobject receiver support leader election ([#39054](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39054))
- (Contrib) `receiver/k8scluster`: Add node name to pod metadata ([#37454](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37454))
- (Contrib) `receiver/hostmetricsreceiver`: Migrated `hostmetricsreceiver` to use `pkg/winperfcounters` when retrieving Windows performance counters. ([#38858](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38858))
- (Contrib) `pkg/ottl`: Add the OTTL function ProfileID() ([#39587](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39587))
- (Contrib) `pkg/ottl`: OTTL is now officially considered Beta stability ([#39910](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39910))
- (Contrib) `postgresqlreceiver`: we introduced the top query collection to postgresql receiver. this will collect the queries from `pg_stat_statements` and report the related metrics. this will help the end user to identify which queries are most executed, and resources most used. ([#39573](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39573))
- (Contrib) `processor/redaction`: Apply redaction to log.body ([#37239](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37239))
- (Contrib) `processor/tailsampling`: New policy type to explicitly drop traces regardless of other policy decisions. ([#39668](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39668))
- (Contrib) `processor/transform`: The transformprocessor is now officially considered Beta stability ([#39928](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39928))

### 🧰 Bug fixes 🧰

- (Splunk) `receiver/smartagent`: Use the system certificate pool as the default pool on Windows, keeping behavior of monitors consistent with other OSes. ([#6240](https://github.com/signalfx/splunk-otel-collector/pull/6240))
- (Splunk) `discovery`: Fix the prometheus config for envoy discovery receiver ([#6243](https://github.com/signalfx/splunk-otel-collector/pull/6243))

- (Core) `exporterhelper`: Add validation error for batch config if min_size is greater than queue_size. ([#12948](https://github.com/open-telemetry/opentelemetry-collector/pull/12948))
- (Core) `telemetry`: Allocate less memory per component when OTLP exporting of logs is disabled ([#13014](https://github.com/open-telemetry/opentelemetry-collector/pull/13014))
- (Core) `confmap`: Use reflect.DeepEqual to avoid panic when confmap.enableMergeAppendOption feature gate is enabled. ([#12932](https://github.com/open-telemetry/opentelemetry-collector/pull/12932))
- (Core) `internal telemetry`: Add resource attributes from telemetry.resource to the logger ([#12582](https://github.com/open-telemetry/opentelemetry-collector/pull/12582))
  Resource attributes from telemetry.resource were not added to the internal
  console logs. Now, they are added to the logger as part of the "resource" field.

- (Core) `confighttp and configcompression`: Fix handling of `snappy` content-encoding in a backwards-compatible way ([#10584](https://github.com/open-telemetry/opentelemetry-collector/pull/10584), [#12825](https://github.com/open-telemetry/opentelemetry-collector/pull/12825))
  The collector used the Snappy compression type of "framed" to handle the HTTP
  content-encoding "snappy".  However, this encoding is typically used to indicate
  the "block" compression variant of "snappy".  This change allows the collector to:
- When receiving a request with encoding 'snappy', the server endpoints will peek
  at the first bytes of the payload to determine if it is "framed" or "block" snappy,
  and will decompress accordingly.  This is a backwards-compatible change.

If the feature-gate "confighttp.framedSnappy" is enabled, you'll see new behavior for both client and server:
- Client compression type "snappy" will now compress to the "block" variant of snappy
  instead of "framed". Client compression type "x-snappy-framed" will now compress to the "framed" variant of snappy.
- Servers will accept both "snappy" and "x-snappy-framed" as valid content-encodings.

- (Contrib) `kafkaexporter`: Make Sarama's ConfigurationError as permanent to prevent retries ([#38604](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38604))
- (Contrib) `kafkaexporter`: fix the default topic configuration, and default message partitioning (#39810, #39816)
- (Contrib) `kafkareceiver`: Fix support for built-in text encodings with hyphens in the encoding name ([#39793](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39793))
- (Contrib) `tailsamplingprocessor`: Fix the misleading decision timer latency metric ([#38502](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38502))
- (Contrib) `pkg/stanza/fileconsumer`: Mute new reader log when the noStateTracker is used ([#39855](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39855))
- (Contrib) `prometheusreceiver`: Add 'fallback_scrape_protocol' to entries obtained through Target Allocator, unless already defined. ([#39672](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39672))
  This applies the same configuration change that is already applied on static configuration.
- (Contrib) `prometheusreceiver`: Convert ms to s when setting start time on the metric ([#39912](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39912))

- (Contrib) `receiver/sqlserver`: Properly parse numbers stored in scientific notation to integers ([#39124](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/39124))

## v0.125.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.125.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.125.0)
and the [opentelemetry-collector-contrib v0.125.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.125.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `service`: Lowercase values for 'otelcol.component.kind' attributes. ([#12865](https://github.com/open-telemetry/opentelemetry-collector/issues/12865))
- (Core) `service`: Restrict the `telemetry.newPipelineTelemetry` feature gate to metrics. ([#12856](https://github.com/open-telemetry/opentelemetry-collector/issues/12856), [#12933](https://github.com/open-telemetry/opentelemetry-collector/issues/12933))
  The "off" state of this feature gate introduced a regression, where the Collector's internal logs were missing component attributes. See issue [#12870](https://github.com/open-telemetry/opentelemetry-collector/issues/12870) for more details on this bug.

  On the other hand, the "on" state introduced an issue with the Collector's default internal metrics, because the Prometheus exporter does not currently support instrumentation scope attributes.

  To solve both of these issues, this change turns on the new scope attributes for logs and traces by default regardless of the feature gate.
  However, the new scope attributes for metrics stay locked behind the feature gate, and will remain off by default until the Prometheus exporter is updated to support scope attributes.

  Please understand that enabling the `telemetry.newPipelineTelemetry` feature gate may break the export of Collector metrics through, depending on your configuration.
  Having a `batch` processor in multiple pipelines is a known trigger for this.

  This comes with a breaking change, where internal logs exported through OTLP will now use instrumentation scope attributes to identify the source component instead of log attributes.
  This does not affect the Collector's stderr output. See the changelog for v0.123.0 for a more detailed description of the gate's effects.
- (Contrib) `receiver/k8sobjects`: Check for K8s API objects existence on receiver startup and not during config validation. ([#38803](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38803))
- (Contrib) `receiver/kubeletstats`: Move receiver.kubeletstats.enableCPUUsageMetrics feature gate to beta ([#39487](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39487))
  Deprecated metrics `container.cpu.utilization`, `k8s.pod.cpu.utilization` and `k8s.node.cpu.utilization` are being replaced
  by `container.cpu.usage`, `k8s.pod.cpu.usage` and `k8s.node.cpu.usage`.
  To be able to use the deprecated metrics, switch `receiver.kubeletstats.enableCPUUsageMetrics` feature gate to `false`.

- (Contrib) `processor/k8sattributes`: Change processor/k8sattributes to return error if unset envvar is used for `node_from_env_var` ([#39447](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39447))
  Before this was a valid configuration, but had an unexpected behavior to monitor the entire cluster. | To keep the same behavior simply do not set the `node_from_env_var` value or use empty string.
- (Contrib) `receiver/sqlserver`: `host.name`, `sqlserver.computer.name`, and `sqlserver.instance.name` are now resource attributes instead of log attributes. We used to report `computer_name` and `instance_name` in the log attributes for top query collection and they are now deprecated. Now we report the three resources attributes in both top query collection and sample query collection. ([#39449](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39449))
  This change is only relevant for logs.

### 🚩 Deprecations 🚩

- (Contrib) `connector/spanmetrics`: Deprecate the unused configuration `dimensions_cache_size` ([#39646](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39646))
  Deprecated configuration `dimensions_cache_size`, please use `aggregation_cardinality_limit` instead

### 💡 Enhancements 💡

- (Splunk) `auto-instrumentation`: Improve error message in case of failed .NET instrumentation installer download ([#6165](https://github.com/signalfx/splunk-otel-collector/pull/6165))
- (Splunk) `auto-instrumentation`: Update NodeJS instrumentation library to v3.1.2 ([#6162](https://github.com/signalfx/splunk-otel-collector/pull/6162))
- (Splunk) Update bundled OpenJDK to 11.0.27_6 ([#6139](https://github.com/signalfx/splunk-otel-collector/pull/6139))
- (Core) `mdatagen`: Add support for attributes for telemetry configuration in metadata. ([#12919](https://github.com/open-telemetry/opentelemetry-collector/issues/12919))
- (Core) `configmiddleware`: Add extensionmiddleware interface. ([#12603](https://github.com/open-telemetry/opentelemetry-collector/issues/12603), [#9591](https://github.com/open-telemetry/opentelemetry-collector/issues/9591))
- (Core) `configgrpc`: Add gRPC middleware support. ([#12603](https://github.com/open-telemetry/opentelemetry-collector/issues/12603), [#9591](https://github.com/open-telemetry/opentelemetry-collector/issues/9591))
- (Core) `confighttp`: Add HTTP middleware support. ([#12603](https://github.com/open-telemetry/opentelemetry-collector/issues/12603), [#9591](https://github.com/open-telemetry/opentelemetry-collector/issues/9591), [#7441](https://github.com/open-telemetry/opentelemetry-collector/issues/7441))
- (Core) `configmiddleware`: Add configmiddleware struct. ([#12603](https://github.com/open-telemetry/opentelemetry-collector/issues/12603), [#9591](https://github.com/open-telemetry/opentelemetry-collector/issues/9591))
- (Contrib) `connector/spanmetrics`: Add new `aggregation_cardinality_limit` configuration option to limit the number of unique combinations of dimensions that will be tracked for metrics aggregation. ([#38990](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38990))
- (Contrib) `connector/count`: Add profiles support ([#39577](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39577))
- (Contrib) `extension/opamp`: Introduces a new config field `include_resource_attributes` which allows the extension to copy the agent's resource attributes to the non-identifying attributes in the agent description. ([#37487](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37487))
- (Contrib) `receiver/statsd`: Add new config to customize socket permissions when transport is set to `unixgram`. ([#37807](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37807))
- (Contrib) `receiver/kubeletstats`: Adds support for collecting Node and Pod network IO/error metrics for all network interfaces ([#30196](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30196))
- (Contrib) `exporter/awss3`: add configuration field `resource_attrs_to_s3/s3_prefix` to support mapping s3 bucket prefix to OTel resource attributes ([#37858](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37858))
  If `resource_attrs_to_s3/s3_prefix` is configured, s3 prefix will be determined based on the specified resource attribute and `s3uploader/s3_prefix` will serve as a fallback.
- (Contrib) `exporter/splunkhec`: Use ObservedTimestamp if Timestamp is empty. ([#39221](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39221))
- (Contrib) `receiver/hostmetrics`: Possible to enable the process scraper under FreeBSD in the hostmetrics receiver. ([#39622](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39622))
- (Contrib) `receiver/kafka`: Add support for configuring Kafka consumer rebalance strategy and group instance ID ([#39513](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39513))
  This enhancement introduces two optional settings: group_rebalance_strategy and group_instance_id.
  These allow users to override the default Range-based rebalance strategy and optionally provide a static instance ID (as per KIP-345) for cooperative sticky balancing.
  This is particularly useful when handling high-cardinality metric workloads, as it reduces rebalance impact, improves cache reuse, and boosts CPU efficiency.
  Both settings are optional to maintain full backward compatibility.

- (Contrib) `receiver/kafka`: Added a new 'topic' attribute to all existing internal consume-claim metrics. ([#35336](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35336))
- (Contrib) `receiver/kafka`: Add `max_fetch_wait` config setting ([#39360](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39360))
  This setting allows you to specify the maximum time that the broker will wait for
  min_fetch_size bytes of data to be available before sending a response to the client.
  Defaults to 250ms.

- (Contrib) `receiver/sqlserver`: Accept `context_info` with `traceparent` format in query sample collection, setting log record with correct traceId and spanId. ([#39539](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39539))

### 🧰 Bug fixes 🧰

- (Splunk) `receiver/smartagent`: Fix the receiver failing to start by receiver_creator since 0.124.0 ([#6187](https://github.com/signalfx/splunk-otel-collector/pull/6187))
- (Core) `exporterhelper`: Do not ignore the `num_consumers` setting when batching is enabled. ([#12244](https://github.com/open-telemetry/opentelemetry-collector/issues/12244))
- (Core) `exporterhelper`: Reject elements larger than the queue capacity ([#12847](https://github.com/open-telemetry/opentelemetry-collector/issues/12847))
- (Core) `mdatagen`: Add time and plog package imports ([#12907](https://github.com/open-telemetry/opentelemetry-collector/issues/12907))
- (Core) `confmap`: Maintain nil values when marshaling or unmarshaling nil slices ([#11882](https://github.com/open-telemetry/opentelemetry-collector/issues/11882))
  Previously, nil slices were converted to empty lists, which are semantically different
  than a nil slice. This change makes this conversion more consistent when encoding
  or decoding config, and these values are now maintained.
- (Contrib) `receiver/sqlquery`: respect `max_open_conn` configuration for multiple queries ([#39270](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39270))
- (Contrib) `processor/resourcedetection`: change the EKS cluster identifier and check the cluster version instead of the existence of aws-auth configmap ([#39479](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39479))
- (Contrib) `receiver/filelog`: Fix frozen receiver when max_concurrent_files is 1 ([#39598](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39598))
- (Contrib) `processor/transform`: Fix the context inferrer to also take into consideration the global OTTL conditions configuration. ([#39455](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39455))
- (Contrib) `receiver/prometheus`: When a histogram metric has both classic and native histogram buckets, keep both, instead of throwing away the native histogram buckets. ([#26555](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26555))
- (Contrib) `receiver_creator`: Fix how escaped backticks are handled in receiver creator templates ([#39163](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39163))
- (Contrib) `receiver/k8sobjects`: Introduces `error_mode`, so users can choose between propagating, ignoring, or silencing missing objects. ([#38803](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38803))
- (Contrib) `extension/opamp`: Skips loading TLS config for insecure endpoints ([#39515](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39515))
- (Contrib) `receiver/kafka`: `name` label will be correctly set in receiver metrics ([#39483](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39483))
- (Contrib) `receiver/tcplog`: Ignore SplitFuncBuilder field to prevent panic during config marshaling ([#39474](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39474))

## v0.124.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.124.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.124.0)
and the [opentelemetry-collector-contrib v0.124.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.124.1) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `splunkenterprisereceiver`: added new attributes to the receiver and modified config ([#36330](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36330))
- (Contrib) `extension/headerssetter`: Change `DefaultValue` to use `configopaque.String` type. ([#39127](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39127))
- (Contrib) `splunkenterprisereceiver`: disabled default metrics except for splunkHealth to ensure scrapes run on Splunk instance are opt-in ([#39068](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39068))
- (Contrib) `processor/transform`: Fix Basic Config style to properly handle `cache` access. ([#38926](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38926))
  The Transform processor now requires only one configuration style per processor's configuration, which means Advanced Config and Basic Config cannot be used together anymore.
- (Contrib) `sqlserverreceiver`: update the unit of `db.lock_timeout` attribute from millisecond to second. this attribute is part of the emitted query sample collection. ([#39042](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39042))

### 🚩 Deprecations 🚩

- (Contrib) `kafkaexporter`: Deprecate `auth::tls` and introduce `tls` config ([#37776](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37776))
- (Contrib) `kafkametricsreceiver`: Deprecate `auth::tls` and introduce `tls` config ([#37776](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37776))
- (Contrib) `kafkareceiver`: Deprecate `auth::tls` and introduce `tls` config ([#37776](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37776))
- (Contrib) `kafkaexporter`: deprecate `topic` and `encoding`, introduce signal-specific configuration ([#35432](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35432))
- (Contrib) `kafkareceiver`: Add signal-specific topic and encoding config, deprecate existing topic/encoding config. ([#32735](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32735))

### 💡 Enhancements 💡

- (Splunk) Update `splunk-otel-javaagent` to v2.15.0 ([#6131](https://github.com/signalfx/splunk-otel-collector/pull/6131))
- (Splunk) Update `jmx-metric-gatherer` to v1.46.0 ([#6127](https://github.com/signalfx/splunk-otel-collector/pull/6127))
- (Core) `exporterhelper`: Add support for bytes-based batching for profiles in the exporterhelper package. ([#3262](https://github.com/open-telemetry/opentelemetry-collector/issues/3262))
- (Core) `otelcol`: Enhance config validation using <validate> command to capture all validation errors that prevents the collector from starting. ([#8721](https://github.com/open-telemetry/opentelemetry-collector/issues/8721))
- (Core) `exporterhelper`: Link batcher context to all batched request's span contexts. ([#12212](https://github.com/open-telemetry/opentelemetry-collector/issues/12212), [#8122](https://github.com/open-telemetry/opentelemetry-collector/issues/8122))
- (Contrib) `azuremonitorreceiver`: Add subscription name resource attribute ([#39029](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39029))
- (Contrib) `azuremonitorreceiver`: Allow to use metrics:getBatch API (Azure Monitor Metrics Data Plane) ([#38651](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38651))
- (Contrib) `kafkareceiver`: Propagate Kafka headers as metadata ([#39129](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39129))
  Allows the Kafka receiver to propagate Kafka headers as client.Info (metadata). Allowing downstream processors and exporters to access the values via the enriched context.
- (Contrib) `kafkaexporter`: Propagate metadata keys as headers ([#39130](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39130))
  Adds a new config option specifying a list of metadata keys that should be propagated as Kafka message headers.
- (Contrib) `receivercreator`: Add kafkatopicsobserver to the receivercreator configuration ([#37665](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37665))
- (Contrib) `kafkaexporter`: enable partitioning for all encodings ([#39001](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39001), [#38999](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38999))
  With the exception of Jaeger encodings which have their own partitioning logic,
  partitioning is now independent of the encoding used. This means that all encodings
  now support partitioning.

- (Contrib) `signalfxexporter`: Errors will now include the URL that it was trying to access ([#39026](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39026))
- (Contrib) `splunkhecexporter`: Errors will now include the URL that it was trying to access ([#39026](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39026))
- (Contrib) `k8sattributesprocessor`: Add option to configure automatic resource attributes - with annotation prefix ([#37114](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37114))
  Implements [Specify resource attributes using Kubernetes annotations](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/non-normative/k8s-attributes.md#specify-resource-attributes-using-kubernetes-annotations).

  If you are using the file log receiver, you can now create the same resource attributes as traces (via OTLP) received
  from an application instrumented with the OpenTelemetry Operator -
  simply by adding the `extract: { otel_annotations: true }` configuration to the `k8sattributesprocessor` processor.
  See the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/k8sattributesprocessor/README.md#config-example) for more details.

- (Contrib) `oracledbreceiver`: Add support for parallel operations metrics ([#39215](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39215))
  The following metrics are now available, all disabled by default:
  - 'DDL statements parallelized'
  - 'DML statements parallelized'
  - 'Parallel operations not downgraded'
  - 'Parallel operations downgraded to serial'
  - 'Parallel operations downgraded (1-25%)'
  - 'Parallel operations downgraded (25-50%)'
  - 'Parallel operations downgraded (50-75%)'
  - 'Parallel operations downgraded (75-99%)'

- (Contrib) `k8sclusterreceiver`: Add missing attributes to entities in experimental entity feature ([#39038](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39038))
- (Contrib) `pkg/stanza`: Use buffer pool for the read buffers to limit allocations ([#39373](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39373))
- (Contrib) `postgresqlreceiver`: add top query collection to help end user identify which query were executed in the postgresql database. ([#39311](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39311))
- (Contrib) `sqlserverreceiver`: Allow full control of the "connection string" via the `datasource` configuration option ([#39235](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39235))
- (Contrib) `spanmetricsconnector`: Initialise new calls_total metrics at 0 ([#38537](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38537))
- (Contrib) `pkg/stanza`: Remove unnecessary slice allocation to track errors (even nil) ([#39367](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39367))

### 🧰 Bug fixes 🧰

- (Core) `confighttp`: Ensure http authentication server failures are handled by the provided error handler ([#12666](https://github.com/open-telemetry/opentelemetry-collector/issues/12666))
- (Contrib) `metricstransformprocessor`: Fix aggregation of exponential histograms in metricstransform processor. ([#39143](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39143))
  Fix a panic when the number of populated buckets varies, and fix summing of counts for the Zero bucket.

- (Contrib) `pkg/ottl`: Fix OTTL context inference order to prioritize the `scope` context over `resource`. ([#39155](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39155))
- (Contrib) `pkg/ottl`: Fix so replace_all_patterns can replace keys using optional function ([#32896](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32896))
  When using the `replace_all_patterns` with `key` and `optional` function on the replacement, the value was being replaced with the key. This change fixes that and now the key is replaced as intended.
- (Contrib) `awss3exporter`: Fixes an issue where the AWS S3 Exporter was forcing an ACL to be set, leading to unexpected behavior in S3 bucket permissions ([#39346](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39346))
  Current behavior of the AWS S3 Exporter is to set the ACL to 'private' by default, this removes that behavior and sets no ACL if not specified.
- (Contrib) `connector/spanmetrics`: This change proposes moving the start timestamp (and last seen timestamp) from the resourceMetrics level to the individual metrics level. This will ensure that each metric has its own accurate start and last seen timestamps, regardless of its relationship to other spans. ([#35994](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35994))
- (Contrib) `receiver/kubeletstats`: support user defined CA path for service account using the configtls option `ca_file` ([#39291](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39291))
- (Contrib) `splunkenterprisereceiver`: Fixes `otelcol_scraper_errored_metric_points` metric, which was not incrementing properly ([#38691](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38691))
- (Contrib) `receivercreator`: Fix automatic discovery of kafka endpoints ([#39313](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39313)).

## v0.123.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.123.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.123.0)
and the [opentelemetry-collector-contrib v0.123.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.123.0)
releases where appropriate.

### ❗ Known Issues ❗

- This version won't collect kafka metrics with the discovery mode enabled. Will be fixed in 0.124.0.
  See https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/39313 for more details.

###  🛑 Breaking changes 🛑

- (Splunk) Support for OpenSUSE 12 and CentOS 7 (tar package) has been dropped ([#6073](https://github.com/signalfx/splunk-otel-collector/pull/6073))
- (Core) `service/telemetry`: Mark `telemetry.disableAddressFieldForInternalTelemetry` as beta, usage of deprecated `service::telemetry::address` are ignored ([#12756](https://github.com/open-telemetry/opentelemetry-collector/issues/12756))
  IMPORTANT: If you have `service::telemetry::address` configured manually, ensure to change it to the new format:
  ```yaml
  telemetry:
    metrics:
      readers:
        - pull:
            exporter:
              prometheus:
                host: <HOST>
                port: <PORT>
                without_scope_info: true
                without_type_suffix: true
                without_units: true
  ```
  Otherwise, the Collector will not emit any internal telemetry metrics.
  To restore the previous behavior disable `telemetry.disableAddressFieldForInternalTelemetry` feature gate.

- (Core) `exporterbatch`: Remove deprecated fields `min_size_items` and `max_size_items` from batch config. ([#12684](https://github.com/open-telemetry/opentelemetry-collector/issues/12684))
- (Contrib) `receiver/azuremonitor`: multi subscriptions support and automatic discovery ([#36612](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36612))
- (Contrib) `exporter/kafka`: change default client_id to "otel-collector" ([#38411](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38411))
  The exporter now uses the "configkafka" package which consolidates
  common configuration structures and default values. As a result of
  this change, we update the default client_id value to "otel-collector".

- (Contrib) `receiver/kafkametrics`: change default client_id to "otel-collector", deprecate "refresh_frequency" ([#38411](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38411))
  The receiver now uses the "configkafka" package which consolidates
  common configuration structures and default values. As a result of
  this change, we update the default client_id value to "otel-collector",
  and deprecate "refresh_frequency" in favour of "metadata.refresh_interval".

- (Contrib) `processor/k8sattributes`: Remove stable feature gate `k8sattr.rfc3339` ([#38810](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38810))

### 🚩 Deprecations 🚩

- (Core) `exporters`: Deprecate `blocking` in favor of `block_on_overflow`. ([#12710](https://github.com/open-telemetry/opentelemetry-collector/issues/12710))
- (Core) `exporters`: Deprecate configuring exporter batching separately. Use `sending_queue::batch` instead. ([#12772](https://github.com/open-telemetry/opentelemetry-collector/issues/12772))
  Moving the batching configuration to `sending_queue::batch` requires setting `sending_queue::sizer` to `items`
  which means that `sending_queue::queue_size` needs to be also increased by the average batch size number (roughly
  x5000 for the default batching configuration).
  See https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/exporterhelper#configuration
- (Contrib) `exporter/kafka`: Deprecate `auth.plain_text` configuration. Use `auth.sasl` with mechanism set to PLAIN instead. ([#38883](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38883))
- (Contrib) `receiver/kafkametrics`: Deprecate `auth.plain_text` configuration. Use `auth.sasl` with mechanism set to PLAIN instead. ([#38883](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38883))
- (Contrib) `receiver/kafka`: Deprecate `auth.plain_text` configuration. Use `auth.sasl` with mechanism set to PLAIN instead. ([#38883](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38883))

### 💡 Enhancements 💡

- (Splunk) Add discovery rule (disabled by default) for istio ([#5854](https://github.com/signalfx/splunk-otel-collector/pull/5854))
- (Core) `exporters`: Add support to configure batching in the sending queue. ([#12746](https://github.com/open-telemetry/opentelemetry-collector/issues/12746))
- (Core) `exporters`: Add support for `wait_for_result`, remove `disabled_queue` ([#12742](https://github.com/open-telemetry/opentelemetry-collector/issues/12742))
- (Core) `exporters`: Allow exporter memory queue to use different type of sizers. ([#12708](https://github.com/open-telemetry/opentelemetry-collector/issues/12708))
- (Core) `service`: Add `telemetry.newPipelineTelemetry` feature gate to inject component-identifying attributes in internal telemetry ([#12217](https://github.com/open-telemetry/opentelemetry-collector/issues/12217))
  With the feature gate enabled, all internal telemetry (metrics/traces/logs) will include some of
  the following instrumentation scope attributes:
  - `otelcol.component.kind`
  - `otelcol.component.id`
  - `otelcol.pipeline.id`
  - `otelcol.signal`
  - `otelcol.signal.output`

  These attributes are defined in the [Pipeline Component Telemetry RFC](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/component-universal-telemetry.md#attributes),
  and identify the component instance from which the telemetry originates.
  They are added automatically without changes to component code.

  These attributes were already included in internal logs as regular log attributes, starting from
  v0.120.0. For consistency with other signals, they have been switched to scope attributes (with
  the exception of logs emitted to standard output), and are now enabled by the feature gate.

  Please make sure that the exporter / backend endpoint you use has support for instrumentation
  scope attributes before using this feature. If the internal telemetry is exported to another
  Collector, a transform processor could be used to turn them into other kinds of attributes if
  necessary.

- (Core) `exporters`: Enable support to do batching using `bytes` sizer ([#12751](https://github.com/open-telemetry/opentelemetry-collector/issues/12751))
- (Core) `service`: Add config key to set metric views used for internal telemetry ([#10769](https://github.com/open-telemetry/opentelemetry-collector/issues/10769))
  The `service::telemetry::metrics::views` config key can now be used to explicitly set the list of
  metric views used for internal telemetry, mirroring `meter_provider::views` in the SDK config.
  This can be used to disable specific internal metrics, among other uses.

  This key will cause an error if used alongside other features which would normally implicitly create views, such as:
  - not setting `service::telemetry::metrics::level` to `detailed`;
  - enabling the `telemetry.disableHighCardinalityMetrics` feature flag.
- (Contrib) `connector/spanmetrics`: Add instrumentation scope to span metrics connector. ([#23662](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23662))
  This change adds the instrumentation scope to the span metrics connector, which allows users to specify the instrumentation scope for the connector.
  Now, the connector has a new configuration option:
  - `include_instrumentation_scope`: A list of instrumentation scope names to include from the traces.

  The instrumentation scope name is the name of the instrumentation library that collected the span.

- (Contrib) `receiver/splunkhec`: Splunk HEC timestamp accepts nanoseconds, microseconds, milliseconds, and seconds epoch. ([#36571](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36571))
- (Contrib) `receiver/oracledb`: Add support for additional read/write metrics ([#37814](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37814))
  The following metrics are now available, all disabled by default:
  'physical reads direct', 'physical writes', 'physical writes direct',
  'physical read io requests' and 'physical write io requests' metrics'

- (Contrib) `exporter/awss3`: Implement timeout for S3 exporter ([#36264](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36264))
- (Contrib) `extension/bearertokenauth`: Allow the header name to be customized in the bearerauthtoken extension ([#38793](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38793))
- (Contrib) `receiver/hostmetrics`: Reduced the cost of retrieving number of threads and parent process ID on Windows.
  Disable the featuregate `hostmetrics.process.onWindowsUseNewGetProcesses` to fallback to the previous[] implementation.
  ([#32947](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32947), [#38589](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38589))
- (Contrib) `receiver/hostmetrics`: Reduced the CPU cost of collecting the `process.handles` metric on Windows. ([#38886](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38886))
  Instead of using WMI to retrieve the number of opened handles by each process
  the scraper now uses the GetProcessHandleCount Win32 API which results in
  reduced CPU usage when the metric `process.handles` is enabled.

- (Contrib) `pkg/ottl`: Enhance the Decode OTTL function to support all flavors of Base64 ([#38854](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38854))
- (Contrib) `receiver/rabbitmq`: Enhance the RabbitMQ receiver to collect and report additional node-level metrics: `rabbitmq.node.disk_free`, `rabbitmq.node.disk_free_limit`, `rabbitmq.node.disk_free_alarm`, `rabbitmq.node.disk_free_details.rate`, `rabbitmq.node.mem_used`, `rabbitmq.node.mem_limit`, `rabbitmq.node.mem_alarm`, `rabbitmq.node.mem_used_details.rate`, `rabbitmq.node.fd_used`, `rabbitmq.node.fd_total`, `rabbitmq.node.fd_used_details.rate`, `rabbitmq.node.sockets_used`, `rabbitmq.node.sockets_total`, `rabbitmq.node.sockets_used_details.rate`, `rabbitmq.node.proc_used`, `rabbitmq.node.proc_total`, `rabbitmq.node.proc_used_details.rate`. These provide additional observability into the state and resource usage of RabbitMQ nodes. ([#38976](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38976))
- (Contrib) `receiver/rabbitmq`: Enhance the RabbitMQ receiver to collect and report additional node-level metrics across multiple categories. These include metrics related to memory, file descriptors, sockets, processes, disk, uptime, scheduling, garbage collection (GC), I/O, message store, connections, clustering, configuration, application info, and context switches. This significantly improves visibility into the performance, state, and resource usage of RabbitMQ nodes. ([#38997](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38997))
- (Contrib) `processor/resourcedetection`: Adding the `os.version` resource attribute to system resourcedetection processor ([#38087](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38087))
- (Contrib) `receiver/sqlserver`: support sqlserverreceiver to record every executing query ([#36462](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36462))
  We introduced Query Sample collection in this PR. The scraper will record all the currently
  executing queries once (in most case) and report related metrics. With this, user will be
  able to see what queries got executed and can combine with the Top Query to get more insights
  on troubleshooting and fine tuning.

- (Contrib) `receiver/sqlserver`: Support query-level log collection ([#36462](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36462))
  Added top query (most CPU time consumed) collection. The query will gather the queries took most of the time during the last
  query interval and report related metrics. The number of queries can be configured. This will enable user to have better
  understanding on what is going on with the database. This enhancement empowers users to not only monitor but also actively
  manage and optimize their MSSQL database performance based on real usage patterns.

### 🧰 Bug fixes 🧰

- (Splunk) Add headers_setter extension to OTLP HTTP exporters in the default configs to facilitate token passthrough ([#6083](https://github.com/signalfx/splunk-otel-collector/pull/6083), [#6075](https://github.com/signalfx/splunk-otel-collector/pull/6075))
- (Core) `exporters`: Fix order of starting between queue and batch. ([#12705](https://github.com/open-telemetry/opentelemetry-collector/issues/12705))
- (Contrib) `receiver/awsecscontainermetrics`: Ensure that the storage.read_bytes and storage.write_bytes metrics include i/o counts from all devices ([#38301](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38301))
- (Contrib) `receiver/sqlserver`: the current metric scraper would report error when parsing value, the value was parsed as int but actually it should be a float ([#38823](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38823))
- (Contrib) `receiver/sqlserver`: Ensure all enabled metrics are emitted ([#38839](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38839))
- (Contrib) `receiver/kafka`: make calls to error backoff thread-safe and add logging ([#38941](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38941))
- (Contrib) `receiver/sqlserver`: Remove warning that `server.address` and `server.port` resource attributes will be enabled ([#38831](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38831))
  There is no intention of enabling these resource attributes by default,
  this change is to simply remove the warning.
  The `server.address` and `server.port` resource attributes were, and still are,
  disabled by default.

- (Contrib) `pkg/stanza`: Add retries when calls to retrieve Windows event via `EvtNext` fail with error RPC_S_INVALID_BOUND (1734). ([#38149](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38149))
  Whenever large events were read in by the Windows event log receiver, via the stanza input operator,
  the collector would fail with error RPC_S_INVALID_BOUND (1734). Now the operator tries to workaround
  this issue by reducing the number of events read on each attempt.

## v0.122.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.122.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.122.1)
and the [opentelemetry-collector-contrib v0.122.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.122.0)
releases where appropriate.

###  🛑 Breaking changes 🛑

- (Core) `service`: Batch processor telemetry is no longer emitted at "basic" verbosity level ([#7890](https://github.com/open-telemetry/opentelemetry-collector/issues/7890))
  According to the guidelines, basic-level telemetry should be reserved for core Collector APIs.
  Components such as the batch processor should emit telemetry starting from the "normal" level
  (which is also the default level). Migration: If your Collector telemetry was set to level:
  basic and you want to keep seeing batch processor-related metrics, consider switching to level:
  normal instead.

- (Contrib) `deltatocumulative`: removes legacy and clean up existing metrics ([#38079](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38079))
- (Contrib) `sqlserverreceiver`: SQL Server receiver now requires a positive serial number for X509 certificates.
  See known issues in README of `receiver/sqlserverreceiver` for more information. ([#38099](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38099))

### 🚀 New components 🚀

- (Splunk) Add `opamp` extension ([#5999](https://github.com/signalfx/splunk-otel-collector/pull/5999))

### 💡 Enhancements 💡

- (Splunk) Add a new discovery bundle for Istio metrics which includes proxy, gateway, and pilot/istiod  ([#5854](https://github.com/signalfx/splunk-otel-collector/pull/5854))
  - This discovery receiver, named prometheus/istio, is disabled by default. Users can enable it by setting the discovery config `splunk.discovery.receivers.prometheus/istio.enabled=true`.
- (Splunk) Update `splunk-otel-javaagent` to `v2.14.0` ([#6000](https://github.com/signalfx/splunk-otel-collector/pull/6000))
- (Splunk) Update `jmx-metric-gatherer` to `v1.45.0` ([#5995](https://github.com/signalfx/splunk-otel-collector/pull/5995))
- (Splunk) Use direct connection for MongoDB discovery ([#6042](https://github.com/signalfx/splunk-otel-collector/pull/6042))

- (Core) `service`: Add `service.AllowNoPipelines` feature gate to allow starting
  the Collector without pipelines. This can be used to start with only extensions.([#12613](https://github.com/open-telemetry/opentelemetry-collector/pull/12613))
- (Core) `service`: include component id/type in start error ([#10426](https://github.com/open-telemetry/opentelemetry-collector/issues/10426))
- (Core) `confmap`: Introduce a new feature flag to allow for merging lists instead of discarding the existing ones.
  You can enable this option via the command line by running following command:
  `otelcol --config=main.yaml --config=extra_config.yaml --feature-gates=-confmap.enableMergeAppendOption`
  ([#8394](https://github.com/open-telemetry/opentelemetry-collector/issues/8394))
  ([#8754](https://github.com/open-telemetry/opentelemetry-collector/issues/8754))
  ([#10370](https://github.com/open-telemetry/opentelemetry-collector/issues/10370))

- (Contrib) `splunkhecreceiver`: splunk hec receiver accepts metrics with empty string Event field ([#38464](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38464))
- (Contrib) `sqlserverreceiver`: Add new performance-related metrics ([#37884](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37884))
- (Contrib) `hostmetricsreceiver`: Added the system.linux.memory.dirty and system.memory.page_size metrics. ([#38672](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38672))
- (Contrib) `mongodbreceiver`: Added new mongodb metrics to achieve parity with Telegraf ([#37227](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37227))
- (Contrib) `routingconnector`: Adds Standard Converter functions to routing connector. ([#38282](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38282))
- (Contrib) `azuremonitorreceiver`: Adds filtering by metric and/or aggregation ([#37420](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37420))
- (Contrib) `headerssetterextension`: Add support for setting headers based on authentication data ([#38441](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38441))
- (Contrib) `collectdreceiver`: apply fieldalignment to collectd records to reduce memory allocation ([#37321](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37321))
- (Contrib) `pkg/ottl`: Enhance flatten() editor to resolve attribute key conflicts by adding a number suffix to the conflicting keys. ([#35793](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35793))
- (Contrib) `loadbalancingexporter`: Add support for route with composite keys ([#35320](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35320))
- (Contrib) `kafkaexporter`: do not ask for user and password if auth mechanism is set to AWS IAM ([#37417](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37417))
- (Contrib) `bearertokenauthextension`: Add the ability to configure multiple bearer tokens for the same endpoint. ([#38148](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38148))
- (Contrib) `pkg/ottl`: Add `IsValidLuhn()` checksum verification function ([#37880](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37880))
- (Contrib) `pkg/ottl`: Add ability to compare maps in Boolean Expressions ([#38611](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38611))
- (Contrib) `pkg/ottl`: Add `Murmur3Hash`and `Murmur3Hash128` functions to convert the `target` string to Murmur3 hash in hexadecimal string format ([#34077](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34077))
- (Contrib) `prometheusreceiver`: Adds the Prometheus API server to more easily debug the Prometheus config, service discovery, and targets. ([#32646](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32646))
- (Contrib) `rabbitmqreceiver`: Enhance the RabbitMQ receiver to collect and report node-level metrics (`rabbitmq.node.disk_free`, `rabbitmq.node.fd_used`, `rabbitmq.node.mem_limit`, and `rabbitmq.node.mem_used`). This provides additional observability into the state and resource usage of RabbitMQ nodes. ([#36925](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36925))
- (Contrib) `redactionprocessor`: Support hashing instead of masking values via 'hash_function' parameter ([#35830](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35830))
- (Contrib) `pkg/stanza`: Add 'regex_replace' operator ([#37443](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37443))
- (Contrib) `jaegerreceiver`: Remove dependency on jaeger/cmd/agent ([#38655](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38655))
- (Contrib) `resourceprocessor`: Add support for profiles signal type ([#37513](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37513))
- (Contrib) `azuremonitorreceiver`: Add dimensions.enabled and dimensions.overrides which allows to opt out from automatically split by all the dimensions of the resource type ([#36611](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36611))
- (Contrib) `sqlqueryreceiver`: Allow users to specify the maximum number of concurrent open connections to DB server using `max_open_conn` config parameter ([#36752](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36752))
- (Contrib) `pkg/stanza`: Prevent data loss in Stanza-based receivers on ungraceful shutdown of the collector ([#35456](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35456))
  Enable the `stanza.synchronousLogEmitter` feature gate to unlock this feature.
  See the [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/README.md) for more information.

### 🧰 Bug fixes 🧰

- (Splunk) `deployment/databricks`: Ensure `curl` commands are properly redirected ([#6047](https://github.com/signalfx/splunk-otel-collector/pull/6047))

- (Core) `service`: do not attempt to register process metrics if they are disabled ([#12098](https://github.com/open-telemetry/opentelemetry-collector/issues/12098))

- (Contrib) `filelogreceiver`: Extend container parser log file path pattern to include rotated files. ([#35137](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35137))
- (Contrib) `mysqlreceiver`: Fixed issue where the system attempted to convert string value '0.0000' (stored as []uint8) to int64 type, which was causing an invalid syntax error. ([#38276](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38276))
- (Contrib) `attributesprocessor`: Validate metrics configuration parameters before processing ([#36077](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36077))
- (Contrib) `probabilisticsampler`: Logs priority sampling behavior applies only when the priority attribute is present. ([#38468](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/38468))
- (Contrib) `pkg/ottl`: Fix limitation of map literals within slice literals not being handled correctly ([#37405](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37405))

## v0.121.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.121.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.121.0) and the [opentelemetry-collector-contrib v0.121.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.121.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Support for Oracle Linux 7 has been dropped ([#5974](https://github.com/signalfx/splunk-otel-collector/pull/5974))

- (Core) `confighttp`: Make the client config options `max_idle_conns`, `max_idle_conns_per_host`, `max_conns_per_host`, and `idle_conn_timeout` integers ([#9478](https://github.com/open-telemetry/opentelemetry-collector/issues/9478))
  All four options can be set to `0` where they were previously set to `null`

- (Contrib) `awss3exporter`: Replaced the `s3_partition` option with `s3_partition_format` to provide more flexibility to users. ([#37915](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37915), [#37503](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37503))
  Users can provide custom file partitions using [strftime](https://www.man7.org/linux/man-pages/man3/strftime.3.html) formatting.
  The default value of `year=%Y/month=%m/day=%d/hour=%H/minute=%M` matches the older pattern (with `s3_partition: minute`)

  If users do not provide a value for `s3_prefix`, the exporter will not create a `/` folder in the bucket.

- (Contrib) `processor/k8sattributes`: Move k8sattr.fieldExtractConfigRegex.disallow feature gate to stable ([#25128](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25128))
- (Contrib) `signalfxexporter`: Remove the deprecated configuration option `translation_rules` ([#35332](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35332))
  Please use processors to handle desired metric transformations instead. Find migration guidance in the
  [translation rules migration guide](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/signalfxexporter/docs/translation_rules_migration_guide.md).

### 🚩 Deprecations 🚩

- (Core) `exporterhelper`: Deprecate `min_size_items` and `max_size_items` in favor of `min_size` and `max_size`. ([#12486](https://github.com/open-telemetry/opentelemetry-collector/pull/12486))

- (Contrib) `prometheusreceiver`: Deprecate metric start time adjustment in the prometheus receiver. It is being replaced by the metricstarttime processor. ([#37186](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37186))
  Start time adjustment is still enabled by default. To disable it, enable the | receiver.prometheusreceiver.RemoveStartTimeAdjustment feature gate.

### 💡 Enhancements 💡

- (Core) `pipeline`: output pipeline name with signal as signal[/name] format in logs. ([#12410](https://github.com/open-telemetry/opentelemetry-collector/pull/12410))
- (Core) `memorylimiter`: Add support to configure min GC intervals for soft and hard limits. ([#12450](https://github.com/open-telemetry/opentelemetry-collector/pull/12450))
- (Core) `otlpexporter`: Update the stability level for logs, it has been as stable as traces and metrics for some time. ([#12423](https://github.com/open-telemetry/opentelemetry-collector/pull/12423))
- (Core) `service`: Create a new subcommand to dump the initial configuration after resolving/merging. ([#11479](https://github.com/open-telemetry/opentelemetry-collector/pull/11479))
  To use the `print-initial-config` subcommand, invoke the Collector with the subcommand and corresponding feature gate: `otelcol print-initial-config --feature-gates=otelcol.printInitialConfig --config=config.yaml`.
  Note that the feature gate enabling this flag is currently in alpha stability, and the subcommand may
  be changed in the future.
- (Core) `memorylimiterprocessor`: Add support for profiles. ([#12453](https://github.com/open-telemetry/opentelemetry-collector/pull/12453))
- (Core) `otelcol`: Converters are now available in the `components` command. ([#11900](https://github.com/open-telemetry/opentelemetry-collector/pull/11900), [#12385](https://github.com/open-telemetry/opentelemetry-collector/pull/12385))
- (Core) `confmap`: Surface YAML parsing errors when they happen at the top-level. ([#12180](https://github.com/open-telemetry/opentelemetry-collector/pull/12180))
  This adds context to some instances of the error "retrieved value (type=string) cannot be used as a Conf", which typically happens because of invalid YAML documents
- (Core) `pprofile`: Add LinkIndex attribute to the generated Sample type ([#12485](https://github.com/open-telemetry/opentelemetry-collector/pull/12485))
- (Core) `exporterhelper`: Stabilize exporter.UsePullingBasedExporterQueueBatcher and remove old batch sender ([#12425](https://github.com/open-telemetry/opentelemetry-collector/pull/12425))

- (Contrib) `processor/resourcedetection`: Introduce retry logic for failed resource detection. ([#34761](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34761))
- (Contrib) `pkg/ottl`: Support dynamic indexing of math expressions in maps and slices ([#37644](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37644))
- (Contrib) `receiver/sqlquery`: Add support for SapASE (sybase) database connections ([#36328](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36328))

- (Contrib) `iisreceiver`: Added state and uptime metrics for application pools ([#34924](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34924))
- (Contrib) `pkg/stanza`: Add entry's timestamp and attributes to errors logs from log transformers processors ([#37285](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37285))
  When a log transformer processor fails to process an log entry it will include entry's timestamp and attributes in its own logs.
  With this information the user can more easily identify the log file and find the entry that's having issues.

- (Contrib) `kafkareceiver`: Add error backoff configuration to kafka receiver which allows to wait and retry a failed message when the next consumer returns some errors. ([#37009](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37009))
- (Contrib) `receiver/kafkametricsreceiver`: Add `refresh_frequency` config to `kafkametricsreceiver`, to configure custom duration for cluster metadata refresh ([#37896](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37896))
  - Helps fine tuning the refresh_frequency, and enables custom cluster metadata refresh intervals
  - Default refresh_frequency is set 10 minutes from Sarama library defaults

- (Contrib) `processor/resourcedetection`: Add k8s.cluster.uid to kubeadm detector ([#38207](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38207))
- (Contrib) `mongodbreceiver`: Added mongodb replica metrics and routing logic for multiple mongodb instances ([#37517](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37517))
- (Contrib) `pkg/ottl`: Add `event_index` to the available paths of the span event context ([#35778](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35778))
- (Contrib) `pkg/ottl`: Introduce Weekday() converter function ([#38126](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38126))
- (Contrib) `prometheusreceiver`: Make use of creation timestamp from prometheus ([#36473](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36473))
- (Contrib) `processor/redaction`: Introduce 'blocked_key_patterns' parameter ([#35830](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35830))

- (Contrib) `awss3exporter`: Added `acl` option ([#37935](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37935))
- (Contrib) `receiver/sqlserverreceiver`: Add `server.address` and `server.port` resource attributes to SQL server receiver. ([#35183](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35183))
  The new resource attributes are added to the SQL server receiver to distinguish metrics coming from different SQL server instances.
  - (Contrib) `server.address`: The address of the SQL server host, disabled by default.
  - (Contrib) `server.port`: The port of the SQL server host, disabled by default.

### 🧰 Bug fixes 🧰

- (Core) `service`: Fix crash at startup when converting from v0.2.0 to v0.3.0 ([#12438](https://github.com/open-telemetry/opentelemetry-collector/pull/12438))
- (Core) `service`: fix bug in parsing service::telemetry configuration ([#12437](https://github.com/open-telemetry/opentelemetry-collector/pull/12437))
- (Core) `exporterhelper`: Fix bug where the error logged when conversion of data fails is always nil ([#12510](https://github.com/open-telemetry/opentelemetry-collector/pull/12510))

- (Contrib) `ecsobserver`: Fixed ecs task tags not being included in metadata labels ([#38278](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38278))
- (Contrib) `redactionprocessor`: Fix redaction processor to redact span event attributes ([#36633](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36633))
- (Contrib) `azuremonitorreceiver`: Fix bug where the time grain wasn't honored ([#37337](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37337))
- (Contrib) `exporter/splunk_hec`: Do not pass errors from draining the response body to the pipeline as a export failure. ([#38118](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38118))
- (Contrib) `kafkametricsreceiver`: Fix incorrect cluster admin initialization in consumer scraper ([#36818](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36818))
- (Contrib) `pkg/ottl`: Change the `ottlmetric` context to properly display the `TransformContext` value in debug logs ([#38103](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38103))
- (Contrib) `redisreceiver`: Collect keyspace metrics even if reported dbs are nonsequential ([#38135](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/38135))
  If a redis instance has no activity on a db, the db number is not reported in the keyspace metrics.
  This change ensures that the keyspace metrics are collected even if the reported dbs have gaps.

## v0.120.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.120.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.120.0) and the [opentelemetry-collector-contrib v0.120.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.120.1) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `receiver/prometheus`: Prometheus receiver now uses scrapers in Prometheus 3.0. ([#36873](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36873))
  There are a number of breaking changes in Prometheus 3.0. Learn more about those changes and migration guide on https://prometheus.io/docs/prometheus/latest/migration/.
  As a result of [adding support for UTF-8 names](https://prometheus.io/docs/prometheus/latest/migration/#utf-8-names),
  the metrics and labels containing UTF-8 characters are no longer escaped. Consequently, the dots (.) in internal
  collector metrics and resource attributes scraped by Prometheus are no longer replaced with underscores (_).
  - The `service_name`, `service_instance_id`, and `service_version` resource attributes are now scraped as
    `service.name`, `service.instance.id`, and `service.version`, respectively.
  - The following metrics containing dots reported by several components are no longer escaped:
    - `filter` processor:
      - `processor_filter_datapoints_filtered` -> `processor_filter_datapoints.filtered`
      - `processor_filter_logs_filtered` -> `processor_filter_logs.filtered`
      - `processor_filter_spans_filtered` -> `processor_filter_spans.filtered`
    - `deltatocumulative` processor:
      - `deltatocumulative_streams_tracked` -> `deltatocumulative.streams.tracked`
      - `deltatocumulative_streams_tracked_linear` -> `deltatocumulative.streams.tracked.linear`
      - `deltatocumulative_streams_limit` -> `deltatocumulative.streams.limit`
      - `deltatocumulative_streams_evicted` -> `deltatocumulative.streams.evicted`
      - `deltatocumulative_streams_max_stale` -> `deltatocumulative.streams.max_stale`
      - `deltatocumulative_datapoints_processed` -> `deltatocumulative.datapoints.processed`
      - `deltatocumulative_datapoints_dropped` -> `deltatocumulative.datapoints.dropped`
      - `deltatocumulative_datapoints_linear` -> `deltatocumulative.datapoints.linear`
      - `deltatocumulative_gaps_length` -> `deltatocumulative.gaps.length`
    - `googlecloudpubsub` receiver:
      - `receiver_googlecloudpubsub_stream_restarts` -> `receiver.googlecloudpubsub.stream_restarts`

- (Contrib) `activedirectorydsreceiver`: Fixed typo in the attribute `distingushed_names`, renaming it to `distinguished_names`. ([#37606](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37606))
- (Contrib) `receiver/hostmetrics`: Remove receiver.hostmetrics.normalizeProcessCPUUtilization feature gate ([#34763](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34763))
- (Contrib) `tailsamplingprocessor`: Fix the decision timer metric to capture longer latencies beyond 50ms. ([#37722](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37722))
  This changes the unit of the decision timer metric from microseconds to milliseconds.
- (Contrib) `routingconnector`: Remove `match_once` configuration parameter. ([#36824](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36824))
- (Core) `service`: Align component logger attributes with those defined in RFC ([#12217](https://github.com/open-telemetry/opentelemetry-collector/pull/12217))
  See [Pipeline Component Telemetry RFC](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/component-universal-telemetry.md#attributes)

### 💡 Enhancements 💡

- (Splunk) Make Windows TA agnostic to Powershell ExecutionPolicy ([#5935](https://github.com/signalfx/splunk-otel-collector/pull/5935))
- (Contrib) `processor/transformprocessor`: Add support for global conditions and error mode overrides. ([#29017](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29017))
  Global conditions are now available for context-inferred structured configurations, allowing the use of fully
  qualified paths. Additionally, a new configuration key called `error_mode` has been added to the context statements group.
  This key determines how the processor reacts to errors that occur while processing that specific group of statements.
  When provided, it overrides the top-level error mode, offering more granular control over error handling.

- (Contrib) `pkg/stanza`: Allow users to configure initial buffer size ([#37786](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37786))
- (Contrib) `vcenterreceiver`: Adds three more vCenter virtual machine performance metrics ([#37488](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37488))
- (Contrib) `k8sclusterreceiver`: Adds new descriptive attributes/metadata to the k8s.namespace and the container entity emitted from k8sclusterreceiver. ([#37580](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37580))
  - Adds the following attributes to k8s.namespace entity:
    - k8s.namespace.phase: The phase of a namespace indicates where the namespace is in its lifecycle. E.g. 'active', 'terminating'
    - k8s.namespace.creation_timestamp: The time when the namespace object was created.
  - Adds the following attributes to container entity:
    - container.creation_timestamp: The time when the container was started. Only available if container is either in 'running' or 'terminated' state.

- (Contrib) `splunkenterprisereceiver`: Added a new `splunk.health` metric. ([#36695](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36695))
- (Contrib) `resourcedetectionprocessor`: add the Dynatrace detector to the resource detection processor ([#37577](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37577))
- (Contrib) `extension/oauth2clientauth`: Add `expiry_buffer` config to `oauth2client` extension, allowing token refresh before expiration with a default buffer of 5 minutes. ([#35148](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35148))
  - Prevents authentication failures by refreshing the token early.
  - The default expiry buffer is set to 5 minutes, and users can adjust it as needed.

- (Contrib) `googlecloudpubsubreceiver`: Turn noisy `warn` log about Pub/Sub servers into `debug`, and turn the reset count into a metric ([#37571](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37571))
  The receiver uses the Google Cloud Pub/Sub StreamingPull API and keeps a open connection. The Pub/Sub servers
  recurrently close the connection after a time period to avoid a long-running sticky connection. Before the
  receiver logged `warn` log lines everytime this happened. These log lines are moved to debug so that fleets with
  lots of collectors with the receiver don't span logs at warn level. To keep track of the resets, whenever a
  connection reset happens a `otelcol_receiver_googlecloudpubsub_stream_restarts` metric is increased by one.

- (Contrib) `processor/redaction`: Introduce 'allowed_values' parameter for allowed values of attributes ([#35840](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35840))
- (Contrib) `routingconnector`: Avoid unnecessary copy of the data in routing connector ([#37946](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37946))
- (Contrib) `awscontainerinsightreceiver`: Add support for HOST_PROC environment variable in AWS Container Insight Receiver. ([#35862](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35862))
- (Contrib) `syslogreceiver`: Support setting `on_error` config for syslog receiver. ([#36906](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36906))
- (Contrib) `processor/tailsampling`: Adds support for optionally recording the policy (and any composite policy) associated with an inclusive tail processor sampling decision.
  This functionality is disabled by default, you can enable it by passing the following feature flag to the collector: `+processor.tailsamplingprocessor.recordpolicy`
  ([#35180](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35180))
- (Contrib) `tailsamplingprocessor`: makes the `numeric_attribute` more flexible and allows to set only `min_value` or `max_value`, without the need to set both ([#37328](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37328))
  This is useful to have simple configurations like these:
  ```
  {
    type: numeric_attribute,
    numeric_attribute: {
      key: http.status_code,
      min_value: 400
    }
  }
  ```

- (Core) `otlpreceiver`: Update stability for logs ([#12335](https://github.com/open-telemetry/opentelemetry-collector/pull/12335))
- (Core) `exporterhelper`: Implement sync disabled queue used when batching is enabled. ([#12245](https://github.com/open-telemetry/opentelemetry-collector/pull/12245))
- (Core) `exporterhelper`: Enable the new pull-based batcher in exporterhelper ([#12291](https://github.com/open-telemetry/opentelemetry-collector/pull/12291))
- (Core) `exporterhelper`: Update queue size after the element is done exported ([#12399](https://github.com/open-telemetry/opentelemetry-collector/pull/12399))
  After this change the active queue size will include elements in the process of being exported.
- (Core) `otelcol`: Add featuregate command to display information about available features ([#11998](https://github.com/open-telemetry/opentelemetry-collector/pull/11998))
  The featuregate command allows users to view detailed information about feature gates
  including their status, stage, and description.

### 🧰 Bug fixes 🧰
- (Contrib) `azureeventhubreceiver`: Fix bug where persisted offset would be ignored after restart ([#37157](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37157))
- (Contrib) `bearertokenauthextension`: Load token lazily for gRPC AUTH to fix token refresh issue ([#36749](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36749))
- (Contrib) `k8sattributes`: Fix bug where `Filters.Labels` failed with when the `exists` or `not-exists` operations were used. ([#37913](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37913))
- (Contrib) `prometheusreceiver`: Start time metric adjuster now handles reset points correctly ([#37717](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37717))
- (Contrib) `awscontainerinsightreceiver`: Fix race condition in shutdown of AWS Container Insight receiver ([#37695](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/37695))
- (Core) `memorylimiter`: Logger no longer attributes to single signal, pipeline, or component. ([#12217](https://github.com/open-telemetry/opentelemetry-collector/pull/12217))
- (Core) `otlpreceiver`: Logger no longer attributes to random signal when receiving multiple signals. ([#12217](https://github.com/open-telemetry/opentelemetry-collector/pull/12217))
- (Core) `exporterhelper`: Fix undefined behavior access to request after send to next component. This causes random memory access. ([#12281](https://github.com/open-telemetry/opentelemetry-collector/pull/12281))
- (Core) `exporterhelper`: Fix default batcher to correctly call all done callbacks exactly once ([#12247](https://github.com/open-telemetry/opentelemetry-collector/pull/12247))
- (Core) `otlpreceiver`: Fix OTLP http receiver to correctly set Retry-After ([#12367](https://github.com/open-telemetry/opentelemetry-collector/pull/12367))
- (Core) `otlphttpexporter`: Fix parsing logic for Retry-After in OTLP http protocol. ([#12366](https://github.com/open-telemetry/opentelemetry-collector/pull/12366))
  The value of Retry-After field can be either an HTTP-date or delay-seconds and the current logic only parsed delay-seconds.

### 🚀 New components 🚀

- (Splunk) Add `influxdb` receiver ([#5925](https://github.com/signalfx/splunk-otel-collector/pull/5925))

## v0.119.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.119.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.119.0) and the [opentelemetry-collector-contrib v0.119.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.119.0) releases where appropriate.

### 🚩 Deprecations 🚩

- (Contrib) `signalfxreceiver`: `access_token_passthrough` is deprecated ([#37575](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37575))
  Please use `include_metadata` in the receiver instead and add the following config to the batch processor:
    ```yaml
    batch:
      metadata_keys: [X-Sf-Token]
    ```

### 🛑 Breaking changes 🛑

- (Splunk) Deprecated ASP.NET and .NET SignalFx monitors are removed ([#5868](https://github.com/signalfx/splunk-otel-collector/pull/5868))
- (Core) `exporters`: Rename exporter span signal specific attributes (e.g. "sent_spans" / "send_failed_span") to "items.sent" / "items.failed". ([#12165](https://github.com/open-telemetry/opentelemetry-collector/issues/12165))
- (Core) `exporters`: Change exporter ID to be a Span level attribute instead on each event. ([#12164](https://github.com/open-telemetry/opentelemetry-collector/issues/12164))

### 💡 Enhancements 💡

- (Core) `configtls`: Allow users to mention their preferred curve types for ECDHE handshake ([#12174](https://github.com/open-telemetry/opentelemetry-collector/issues/12174))
- (Contrib) `processor/transformprocessor`: Add support for flat configuration style. ([#29017](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29017))

  The flat configuration style allows users to configure statements by providing a list of statements instead of a
  structured configuration map. The statement's context is expressed by adding the context's name prefix to path names,
  which are used to infer and to select the appropriate context for the statement.

- (Contrib) `receiver/httpcheck`: Added support for specifying multiple endpoints in the `httpcheckreceiver` using the `endpoints` field. Users can now monitor multiple URLs with a single configuration block, improving flexibility and reducing redundancy. ([#37121](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37121))
- (Contrib) `processor/resourcedetection`: Expose additional configuration parameters for the AWS metadata client used by the EC2 detector ([#35936](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35936))

  In some cases, you might need to change the behavior of the AWS metadata client from the [standard retryer](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-retries-timeouts.html)
  By default, the client retries 3 times with a max backoff delay of 20s.
  We offer a limited set of options to override those defaults specifically, such that you can set the client to retry 10 times, for up to 5 minutes, for example:
  ```yaml
  processors:
    resourcedetection/ec2:
      detectors: ["ec2"]
      ec2:
        max_attempts: 10
        max_backoff: 5m
  ```

- (Contrib) `processor/cumulativetodelta`: Add metric type filter for cumulativetodelta processor ([#33673](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33673))
- (Contrib) `processor/resourcedetection`: Add `fail_on_missing_metadata` option on EC2 detector ([#35936](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35936))

  If the EC2 metadata endpoint is unavailable, the EC2 detector by default ignores the error.
  By setting `fail_on_missing_metadata` to true on the detector, the user will now trigger an error explicitly,
  which will stop the collector from starting.

- (Contrib) `processor/resourcedetection`: The `gcp` resource detector will now detect resource attributes identifying a GCE instance's managed instance group. ([#36142](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36142))
- (Contrib) `receiver/jaeger`: Log the endpoints of different servers started by jaegerreceiver ([#36961](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36961))

  This change logs the endpoints of different servers started by jaegerreceiver. It simplifies debugging by ensuring log messages match configuration settings.

- (Contrib) `receiver/hostmetrics/process`: Added support for tracking process.uptime ([#36667](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36667))
- (Contrib) `receiver/googlecloudpubsub`: Added support for encoding extensions. ([#37109](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37109))
- (Contrib) `processor/transform`: Replace parser collection implementations with `ottl.ParserCollection` and add initial support for expressing statement's context via path names. ([#29017](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29017))
- (Contrib) `receiver/prometheus`: Add `receiver.prometheusreceiver.UseCollectorStartTimeFallback` featuregate for the start time metric adjuster to use the collector start time as an approximation of process start time as a fallback. ([#36364](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36364))
- (Contrib) `processor/tailsampling`: Reworked the consume traces, sampling decision, and policy loading paths to improve performance and readability ([#37560](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37560))

### 🧰 Bug fixes 🧰

- (Core) `exporters`: Fix bug that the exporter with new batcher may have been marked as non mutation. ([#12239](https://github.com/open-telemetry/opentelemetry-collector/issues/12239))
  Only affects users that manually turned on `exporter.UsePullingBasedExporterQueueBatcher` featuregate.
- (Core) `exporters`: Fix MergeSplit issue that ignores the initial message size. ([#12257](https://github.com/open-telemetry/opentelemetry-collector/issues/12257))
- (Core) `service-telemetry`: pass the missing async error channel into service telemetry settings ([#11417](https://github.com/open-telemetry/opentelemetry-collector/issues/11417))
- (Contrib) `receiver/filelog`: Fix issue where flushed tokens could be truncated. ([#35042](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35042))
- (Contrib) `connector/routing`: Fix config validation with context other than `resource` ([#37410](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37410))
- (Contrib) `processor/k8sattributes`: Wait for the other informers to complete their initial sync before starting the pod informers ([#37056](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37056))
- (Contrib) `processor/metricsgeneration`: Generated metric name may not match metric being scaled ([#37474](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37474))
- (Contrib) `connector/routing`: The connector splits the original payload so that it may be emitted in parts to each route. ([#37390](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37390))
- (Contrib) `pkg/stanza`: Fix default source identifier in recombine operator ([#37210](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37210))
  Its default value is now aligned with the semantic conventions: `attributes["log.file.path"]`
- (Contrib) `processor/tailsampling`: Fixed sampling decision metrics `otelcol_processor_tail_sampling_sampling_trace_dropped_too_early` and `otelcol_processor_tail_sampling_sampling_policy_evaluation_error_total`, these were sometimes overcounted. ([#37212](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37212))

  As a result of this change non-zero values of `otelcol_processor_tail_sampling_sampling_trace_dropped_too_early`
  and `otelcol_processor_tail_sampling_sampling_policy_evaluation_error_total` metrics will be lower.
  Before this fix, errors got counted several times depending on the amount of traces being processed
  that tick and where in the batch the error happened.
  Zero values are unaffected.

- (Contrib) `exporter/signalfx`: Warn on dropping metric data points when they have more than allowed dimension count ([#37484](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37484))

  The SignalFx exporter drops metric data points if they have more than 36 dimensions.
  Currently, the exporter logs at debug level when this occurs.
  With this change, the exporter will log at the warning level.

## v0.118.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.118.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.118.0) and the [opentelemetry-collector-contrib v0.118.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.118.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Stop supporting deprecated syntax for config source expansion ([#5832](https://github.com/signalfx/splunk-otel-collector/pull/5832))
  Use the following guidelines to update your configuration:
  - `$ENV` must be replaced with `${env:ENV}`
  - `$include:file_path` must be replaced with `${include:file_path}`. The same applied for any other config source.
    More information can be found in ([the upgrade guidelines](https://github.com/signalfx/splunk-otel-collector?tab=readme-ov-file#from-01170-to-01180)).
- (Contrib) `pkg/stanza`: Move `filelog.container.removeOriginalTimeField` feature gate to stable ([#33389](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33389))
- (Contrib) `pkg/ottl`: Support dynamic indexing of maps and slices. ([#36644](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36644))
- (Contrib) `routingconnector`: Disconnect `match_once` parameter from functionality. ([#29882](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29882))
  The parameter will be ignored, except to trigger a warning log about its upcoming removal in v0.120.0.

### 💡 Enhancements 💡

- (Splunk) Turn on `splunk.continuousDiscovery` feature gate by default ([#5833](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/5832))
  This enables new way to discover third-party components and collect metrics from them:
  - It allows discovering services that were not available at the time of the collector startup but were started later.
  - Information about discovered services will be sent to Splunk Observability Cloud that can be found in "Discovered services" page.
  If you want to disable this feature and bring back the old behavior, add `--feature-gates=-splunk.continuousDiscovery` command line flag.
- (Contrib) `mysqlreceiver`: Add delete_multi, update_multi metrics to mysql.commands for mysqlreceiver ([#37301](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37301))
- (Contrib) `pkg/ottl`: Add the `Nanosecond` converter to return the nanosecond component from the specified time.Time ([#37042](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37042))
- (Contrib) `pkg/ottl`: Add the `Second` converter to return the second component from the specified time.Time ([#37042](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37042))
- (Contrib) `vcenterreceiver`: Adds vCenter memory capacity for hosts and memory granted metric for VMs. ([#37257](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37257))
- (Contrib) `awss3exporter`: Implement sending queue for S3 exporter ([#37274](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37274), [#36264](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36264))
- (Contrib) `resourcedetectionprocessor`: This enhancement detects AWS EKS cloud account ID ([#37179](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37179))
- (Contrib) `k8sattributesprocessor`: For pods with only one container, the `container.id` and `k8s.container.name` are not longer required in the resource attributes to add the container attributes ([#34189](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34189))
- (Contrib) `pkg/ottl`: Enhanced error messages for invalid cache access and introduced options to configure their values within the OTTL contexts. ([#29017](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29017))
- (Contrib) `pkg/ottl`: Add the `FormatTime` function to convert `time.Time` values to human-readable strings ([#36870](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36870))
- (Contrib) `resourcedetectionprocessor`: Introduce kubeadm detector to retrieve local cluster name. ([#35116](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35116))
- (Contrib) `signalfxexporter`: Prioritize retrieving token from context when accesstokenpassthrough is enabled ([#37102](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37102))
- (Core) `exporters`: Add blocking option to control queue behavior when full ([#12090](https://github.com/open-telemetry/opentelemetry-collector/issues/12090))
- (Core) `debugexporter`: Add EventName to debug exporter for Logs. EventName was added as top-level field in the LogRecord from 1.5.0 of proto definition. ([#11966](https://github.com/open-telemetry/opentelemetry-collector/issues/11966))
- (Core) `confighttp`: Added support for configuring compression levels. ([#10467](https://github.com/open-telemetry/opentelemetry-collector/issues/10467))
  A new configuration option called CompressionParams has been added to confighttp. | This allows users to configure the compression levels for the confighttp client.
- (Core) `exporters`: Change the memory queue implementation to not pre-allocate capacity objects. ([#12070](https://github.com/open-telemetry/opentelemetry-collector/issues/12070))
  This change improves memory usage of the collector under low utilization and is a prerequisite for supporting different other size limitations (number of items, bytes).

### 🧰 Bug fixes 🧰

- (Contrib) `googlecloudpubsubreceiver`: Fix a goroutine leak during shutdown. ([#30438](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30438))
  A goroutine leak was found in the googlecloudpubsubreceiver.
  The goroutine leak was caused by the receiver not closing the underlying created gRPC client when using an insecure custom endpoint.

- (Contrib) `signalfxexporter`: Honor access_token_passthrough config option for sending events ([#37102](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37102))
- (Contrib) `k8sattributesprocessor`: Ensure the pods gathered by the processor contain the information about their related replica sets and deployments after the initial sync ([#37056](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37056))
- (Contrib) `pkg/ottl`: Fix bug with `replace_all_matches` and `replace_all_patterns` that caused non-string values to be changed to empty string when matching against empty string. ([#37071](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37071))
- (Contrib) `tailsamplingprocessor`: Fixed sampling policy evaluation debug logging batch metrics (e.g. sampled). ([#37040](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37040))
- (Contrib) `tailsamplingprocessor`: Late span age histogram should include sampled traces. ([#37180](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37180))

## v0.117.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.117.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.117.0) and the [opentelemetry-collector-contrib v0.117.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.117.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `cloudfoundryreceiver`: Introduce a feature gate enable copying envelope tags to the metrics as resource attributes instead of datapoint attributes. ([#34824](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34824))
- (Contrib) `pkg/ottl`: removed the ability to reference entire parent objects. ([#36872](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36872))
  Statements like `set(cache["resource"], resource)` in non-resource contexts will no longer work.

- (Contrib) `routingconnector`: Change default value of `match_once` parameter to `true`. ([#29882](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29882))
  This field was deprecated in v0.116.0 and will be removed in v0.120.0.

- (Core) `otelcol`: Remove warnings when 0.0.0.0 is used ([#11713](https://github.com/open-telemetry/opentelemetry-collector/issues/11713), [#8510](https://github.com/open-telemetry/opentelemetry-collector/issues/8510))


### 🚩 Deprecations 🚩

- (Contrib) `sapmexporter`: Deprecate SAPM exporter ([#36028](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36028))
  The SAPM exporter is being marked as deprecated. Please use the `otlphttp` exporter with the configuration shown
  below. Also update your pipeline configuration for Traces accordingly.
  ```yaml
    exporters:
        otlphttp:
            traces_endpoint: "${SPLUNK_INGEST_URL}/v2/trace/otlp"
            headers:
                "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"
    ```

### 🚀 New components 🚀

- (Splunk) Add `metricsgeneration` processor ([#5769](https://github.com/signalfx/splunk-otel-collector/pull/5769))

### 💡 Enhancements 💡

- (Splunk) Add a new discovery bundle for Envoy proxy metrics ([#5780](https://github.com/signalfx/splunk-otel-collector/pull/5780))
- (Contrib) `k8sclusterreceiver`: Add additional attributes to node and pod entities ([#35879](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35879))
  Adds the following attributes to node and pod metadata/entities:
    - `k8s.pod.phase`: The phase of a Pod indicates where the Pod is in its lifecycle. E.g. 'Pending', 'Running'
    - `k8s.pod.status_reason`: A brief message indicating details about why the pod is in this state. E.g. 'Evicted'
    - `k8s.node.condition_*`: The condition of a node. e.g. `k8s.node.condition_ready`. The value can be `true`, `false`, `unknown`.

- (Contrib) `receivercreator`: Add support for starting logs' collection based on provided k8s annotations' hints ([#34427](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34427))
- (Contrib) `ottl`: Add a new ottl trim function that trims leading and trailing characters from a string (default- whitespace). ([#34100](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34100))
- (Contrib) `sqlqueryreceiver`: Add instrumentation scope to SQL query receiver metrics and logs ([#31028](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31028))
- (Contrib) `statsdreceiver`: Add UDS support to statsdreceiver ([#21385](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21385))
- (Contrib) `tailsamplingprocessor`: Support hot sampling policy loading ([#37014](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/37014))

### 🧰 Bug fixes 🧰

- (Splunk) Fix deprecation warning for multiline config source calls ([#5829](https://github.com/signalfx/splunk-otel-collector/pull/5829))
- (Contrib) `receiver/azureeventhub`: Ensure that observed timestamp is set when unmarshaling logs. ([#36861](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36861))
- (Contrib) `internal/docker`: Fix image matching regular expression to properly match SHA256 strings. ([#36239](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36239))
  This affects the `docker_observer` extension.
- (Contrib) `k8sobjectsreceiver`: ensure the `k8s.namespace.name` attribute is set for objects retrieved using the `watch` mode ([#36352](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36352))
- (Contrib) `mongodbatlasreceiver`: Update the mongoDB Atlas receiver to use the Default HTTP Transport that supports default proxy configuration ([#36412](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36412))
- (Contrib) `mysqlreceiver`: Avoid recording a value for the MysqlBufferPoolPages metric when out-of-bounds. ([#35495](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35495))
  When using compressed tables, Innodb_buffer_pool_pages_misc may report an out-of-bounds value.
  See https://bugs.mysql.com/bug.php?id=59550 for context.

- (Contrib) `pkg/ottl`: fix handling of nested maps within slices in the `flatten` function ([#36162](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36162))
- (Contrib) `mysqlreceiver`: Divide large values directly in SQL queries to avoid int overflows ([#35495](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35495))
- (Core) `internal/sharedcomponent`: Fixed bug where sharedcomponent would use too much memory remembering all the previously reported statuses ([#11826](https://github.com/open-telemetry/opentelemetry-collector/issues/11826))

## v0.116.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.116.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.116.0) and the [opentelemetry-collector-contrib v0.116.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.116.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `processor/tailsampling`: Reverts [#33671](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33671), allowing for composite policies to specify inverted clauses in conjunction with other policies. This is a change bringing the previous state into place, breaking users who rely on what was introduced as part of [#33671](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33671). ([#34085](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34085))

### 🚀 New components 🚀

- (Splunk) Add `bearertokenauth` extension ([#5727](https://github.com/signalfx/splunk-otel-collector/pull/5727))
- (Splunk) Add `snowflake` receiver ([#5724](https://github.com/signalfx/splunk-otel-collector/pull/5724))
- (Splunk) Add `purefa` receiver ([#5731](https://github.com/signalfx/splunk-otel-collector/pull/5731))
- (Splunk) Add `sum` connector ([#5730 ](https://github.com/signalfx/splunk-otel-collector/pull/5730))

### 🚩 Deprecations 🚩

- (Contrib) `connector/routing`: Deprecate `match_once` parameter. ([#29882](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29882))
- (Contrib) `routingprocessor`: Deprecated in favor of the routing connector. ([#36616](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36616))
- (Contrib) `splunkhecexporter`: Add `otel_attrs_to_hec_metadata/*` config fields to replace `hec_metadata_to_otel_attrs/*` fields. ([#35092](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35092))
  `otel_attrs_to_hec_metadata/*` config fields will replace the `hec_metadata_to_otel_attrs/*` fields in a later release.

### 💡 Enhancements 💡

- (Contrib) `purefareceiver`: Implements support for scraping Pure Storage FlashArray with Purity version 6.6.11+ ([#36251](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36251))
- (Contrib) `azureeventhubreceiver`: support providing one or more time formats for timestamp parsing ([#36650](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36650))
- (Contrib) `fileexporter`: Add support for profiles signal. ([#35978](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35978))
- (Contrib) `pkg/ottl`: GetXML Converter now supports selecting text, CDATA, and attribute (value) content. ([#36821](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36821))
- (Contrib) `kafkaexporter, kafkareceiver`: Add a new mechanism "AWS_MSK_IAM_OAUTHBEARER" for kafka exporter and kafka receiver. This mechanism use the AWS MSK IAM SASL Signer for Go https://github.com/aws/aws-msk-iam-sasl-signer-go. ([#19747](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19747))
- (Contrib) `loadbalancingexporter`: Adds a an optional configuration to the k8s resolver which returns hostnames instead of IPs for headless services pointing at statefulsets ([#18412](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18412))
- (Contrib) `mongodbatlasreceiver`: Adds additional metrics to the MongoDB Atlas receiver ([#36525](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36525))
  Adds a number of new default disabled metrics to the MongoDB Atlas receiver. These metrics are:
  - mongodbatlas.disk.partition.queue.depth
  - mongodbatlas.disk.partition.throughput
  - mongodbatlas.process.cache.ratio
- (Contrib) `s3exporter`: Upgrading to adopt aws sdk v2 ([#36699](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36699))

### 🧰 Bug fixes 🧰

- (Core) `exporters`: Fix memory leak at exporter shutdown ([#11401](https://github.com/open-telemetry/opentelemetry-collector/issues/11401))
- (Core) `sharedcomponent`: Remove race-condition and cleanup locking ([#11819](https://github.com/open-telemetry/opentelemetry-collector/issues/11819))
- (Contrib) `receiver/prometheusreceiver`: Metric adjuster no longer assumes that all metrics from a scrape come from the same resource ([#36477](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36477))
- (Contrib) `k8sattributesprocessor`: Override extracted k8s attributes if original value has been empty ([#36373](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36373))
- (Contrib) `k8sattributesprocessor`: Log any errors encountered during kube client initialisation ([#35879](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35879))
  This addresses an issue where the collector, due to an error encountered during the kubernetes client initialisation,
  was reporting an 'unavailable' status via the health check extension without any further information to be found in the logs.
- (Contrib) `postgresqlreceiver`: Update the postgresqlreceiver to handle new table schema for the bgwriter metrics in pg17+ ([#36784](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/36784))
- (Contrib) `jaegerreceiver`: Remove ineffectual warning from jaegerreceiver ([#35894](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35894))

## v0.115.0

### 🛑 Breaking changes 🛑

- (Splunk) Change default traces exporter to otlphttp exporter. ([#5588](https://github.com/signalfx/splunk-otel-collector/pull/5588)).
  See [the upgrade guidelines](https://github.com/signalfx/splunk-otel-collector?tab=readme-ov-file#from-01140-to-01150) on how to change existing configs to replace SAPM exporter.
- (Core) Change all logged timestamps to ISO8601 ([#10543](https://github.com/open-telemetry/opentelemetry-collector/pull/10543)).
  This makes log timestamps human-readable (as opposed to epoch seconds in
  scientific notation), but may break users trying to parse logged lines in the
  old format.
- (Contrib) `k8sattributesprocessor`: Move k8sattr.fieldExtractConfigRegex.disallow feature gate to Beta. ([#25128](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25128))
  Disable the `k8sattr.fieldExtractConfigRegex.disallow` feature gate to get the old behavior.

### 🚀 New components 🚀

- (Splunk) Add `filestats` receiver ([#5229](https://github.com/signalfx/splunk-otel-collector/pull/5229))
- (Splunk) Add `iis` receiver ([#5717](https://github.com/signalfx/splunk-otel-collector/pull/5717))
- (Splunk) Add `bearertokenauth` extension ([#5727](https://github.com/signalfx/splunk-otel-collector/pull/5727))

### 💡 Enhancements 💡

- (Splunk) Automatic Discovery:
  - Switch bundled NGINX discovery to create [OpenTelemetry NGINX receiver](https://docs.splunk.com/observability/en/gdi/opentelemetry/components/nginx-receiver.html#nginx-receiver) instead of the Smart Agent NGINX monitor ([#5689](https://github.com/signalfx/splunk-otel-collector/pull/5689))
- (Splunk) Expose internal metrics at default `localhost:8888` address instead of `${SPLUNK_LISTEN_INTERFACE}:8888` ([#5706](https://github.com/signalfx/splunk-otel-collector/pull/5706))
  This can be changed in `service::telemetry::metrics` section:
  ```yaml
  service:
    telemetry:
      metrics:
        readers:
          - pull:
              exporter:
                prometheus:
                  host: localhost
                  port: 8888
  ```
  This also removes a warning about deprecated `service::telemetry::metrics::address`.
- (Splunk) Publish a FIPS-140 compliant Docker [images](https://quay.io/repository/signalfx/splunk-otel-collector-fips?tab=tags) and binaries for Linux and Windows. ([#5725](https://github.com/signalfx/splunk-otel-collector/pull/5725))
- (Core) `exporterqueue`: Introduce a feature gate exporter.UsePullingBasedExporterQueueBatcher to use the new pulling model in exporter queue batching. ([#8122](https://github.com/open-telemetry/opentelemetry-collector/pull/8122), [#10368](https://github.com/open-telemetry/opentelemetry-collector/pull/10368))
  If both queuing and batching is enabled for exporter, we now use a pulling model instead of a
  pushing model. num_consumer in queue configuration is now used to specify the maximum number of
  concurrent workers that are sending out the request.
- (Core) `service`: Add support for ca certificates in telemetry metrics otlp grpc exporter ([#11633](https://github.com/open-telemetry/opentelemetry-collector/pull/11633))
  Before this change the Certificate value in config was silently ignored.
- (Contrib) `postgresqlreceiver`: Added new postgresql metrics to achieve parity with Telegraf ([#36528](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36528))
- (Contrib) `loadbalancingexporter`: Adding sending_queue, retry_on_failure and timeout settings to loadbalancing exporter configuration ([#35378](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35378))
  When switching to top-level sending_queue configuration - users should carefully review queue size
  In some rare cases setting top-level queue size to n*queueSize might be not enough to prevent data loss
- (Contrib) `routingconnector`: Add abiilty to route by 'datapoint' context ([#36523](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36523))
- (Contrib) `signalfxreceiver`: Follow receiver contract based on type of error ([#5909](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/5909))
  Use 503 error code for retryable and 400 error code for not-retryable errors instead of responding with a 500 unconditionally.

### 🧰 Bug fixes 🧰

- (Core) `service`: ensure OTLP emitted logs respect severity ([#11718](https://github.com/open-telemetry/opentelemetry-collector/pull/11718))
- (Core) `featuregate`: Fix an unfriendly display message `runtime error` when feature gate is used to display command line usage. ([#11651](https://github.com/open-telemetry/opentelemetry-collector/pull/11651))
- (Contrib) `exporter/pulsarexporter`: Change configuration option `map_connections_per_broker`, rename to `max_connections_per_broker`. ([#36579](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36579))
- (Contrib) `processor/k8sattribute`: fixes parsing of k8s image names to support images with tags and digests. ([#36131](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36131))
- (Contrib) `loadbalancingexporter`: The k8sresolver in loadbalancingexporter was triggering exporter churn in the way the change event was handled. ([#35658](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35658))
- (Contrib) `vcenterreceiver`: The existing code did not honor TLS settings beyond 'insecure'. All TLS client config should now be honored. ([#36482](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36482))

### 🚩Deprecations 🚩

- (Splunk) Deprecate the collectd/genericjmx monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/activemq monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the activemq target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/cassandra monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the cassandra target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/hadoop monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the hadoop target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/kafka monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the kafka target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/kafka-consumer monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the kafka-consumer target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/kafka-producer monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the kafka-producer target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/solr monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the solr target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))
- (Splunk) Deprecate the collectd/tomcat monitor. Please use the [jmxreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/jmxreceiver) with the tomcat target system instead. ([#5539](https://github.com/signalfx/splunk-otel-collector/pull/5539))

## v0.114.0

### 💡 Enhancements 💡

- (Contrib) `processor/k8sattributes`: Add support for profiles signal ([#35983](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35983))
- (Contrib) `receiver/k8scluster`: Add support for limiting observed resources to a specific namespace. ([#9401](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/9401))
  This change allows to make use of this receiver with `Roles`/`RoleBindings`, as opposed to giving the collector cluster-wide read access.
- (Contrib) `processor/resourcedetection`: Introduce support for Profiles signal type. ([#35980](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35980))
- (Contrib) `connector/routing`: Add ability to route by metric context ([#36236](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36236))
- (Contrib) `connector/routing`: Add ability to route by span context ([#36276](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36276))
- (Contrib) `processor/spanprocessor`: Add a new configuration option to keep the original span name when extracting attributes from the span name. ([#36120](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36120))
- (Contrib) `receiver/splunkenterprise`: Add new metrics for Splunk Enterprise dispatch artifacts caches ([#36181](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36181))

### 🚩Deprecations 🚩

- (Splunk) `SPLUNK_TRACE_URL` environment variable is deprecated. It's replaced with `${SPLUNK_INGEST_URL}/v2/trace`
  in the default configs. Default value for `SPLUNK_TRACE_URL` is still set in the binary from `SPLUNK_REALM` or
  `SPLUNK_INGEST_URL` environment variables to not break existing configurations. However, it is recommended to
  update the configurations to use `${SPLUNK_INGEST_URL}/v2/trace` instead. ([#5672](https://github.com/signalfx/splunk-otel-collector/pull/5672)).

### 🛑 Breaking changes 🛑

- (Splunk) Given that `SPLUNK_TRACE_URL` environment variable is deprecated and replaced with
  `${SPLUNK_INGEST_URL}/v2/trace` in the default configurations, the option to set the Trace URL has been removed from
  all packaging and mass deployment solutions to an avoid confusion. ([#5672](https://github.com/signalfx/splunk-otel-collector/pull/5672)).

### 🧰 Bug fixes 🧰

- (Splunk) `receiver/journald`: Upgrade journald client libraries in the Collector docker image by taking them from latest Debian image.
  This fixes journald receiver on kubernetes nodes with recent versions of systemd ([#5664](https://github.com/signalfx/splunk-otel-collector/pull/5664)).
- (Core) scraperhelper: If the scraper shuts down, do not scrape first. ([#11632](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/11632))
  When the scraper is shutting down, it currently will scrape at least once. With this change, upon receiving a shutdown order, the receiver's scraperhelper will exit immediately.
- (Contrib) `pkg/stanza`: Ensure that time parsing happens before entry is sent to downstream operators ([#36213](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36213))
- (Contrib) `processor/k8sattributes`: Block when starting until the metadata have been synced, to fix that some data couldn't be associated with metadata when the agent was just started. ([#32556](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32556))
- (Contrib) `exporter/loadbalancing`: Shutdown exporters during collector shutdown. This fixes a memory leak. ([#36024](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36024))
- (Contrib) `pkg/ottl`: Respect the `depth` option when flattening slices using `flatten` ([#36161](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36161))
  The `depth` option is also now required to be at least `1`.
- (Contrib) `pkg/stanza`: Synchronous handling of entries passed from the log emitter to the receiver adapter ([#35453](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35453))
- (Contrib) `receiver/prometheus`: Fix prometheus receiver to support static scrape config with Target Allocator ([#36062](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36062))

## v0.113.0

This Splunk OpenTelemetry Collector release includes changes from the opentelemetry-collector v0.113.0 and the opentelemetry-collector-contrib v0.113.0 releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `sapmreceiver`: Remove the deprecated access_token_passthrough from SAPM receiver. ([#35972](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35972))
  Please use `include_metadata` instead with the following config option applied to the batch processor:
  batch:
    metadata_keys: [X-Sf-Token]

- (Contrib) `pkg/ottl`: Promote `processor.transform.ConvertBetweenSumAndGaugeMetricContext` feature gate to Stable ([#36216](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36216))
  This gate can no longer be disabled. The `convert_sum_to_gauge` and `convert_gauge_to_sum` may now only be used with the `metric` context.


### 💡 Enhancements 💡

- (Contrib) `splunkenterprisereceiver`: Add telemetry around the Splunk Enterprise kv-store. ([#35445](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35445))
- (Contrib) `journaldreceiver`: adds ability to parse journald's MESSAGE field as a string if desired ([#36005](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36005))
- (Contrib) `journaldreceiver`: allows querying a journald namespace ([#36031](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36031))
- (Contrib) `hostmetricsreceiver`: Add the system.uptime metric in the hostmetrics receiver ([#31627](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31627))
  This metric is provided by the new `system` scraper.

- (Contrib) `hostmetrics`: Adjust scraper creation to make it so the scraper name is reported with hostmetrics scraper errors. ([#35814](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35814))
- (Contrib) `pkg/ottl`: Add SliceToMap function ([#35256](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35256))
- (Contrib) `journaldreceiver`: Restart journalctl if it exits unexpectedly ([#35635](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35635))
- (Contrib) `routingconnector`: Add ability to route by request metadata. ([#19738](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/19738))
- (Contrib) `exporter/signalfx`: Enabling retrying for dimension properties update without tags in case of 400 response error. ([#36044](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36044))
  Property and tag updates are done using the same API call. After this change, the exporter will retry once to sync
  properties in case of 400 response error.

- (Contrib) `signalfxexporter`: Add more default metrics related to Kubernetes cronjobs, jobs, statefulset, and hpa ([#36026](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36026))
- (Contrib) `simpleprometheusreceiver`: Support to set `job_name` in config ([#31502](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31502))
- (Contrib) `solacereceiver`: Add support to the Solace Receiver to convert the new `Move to Dead Message Queue` and new `Delete` spans generated by Solace Event Broker to OTLP. ([#36071](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36071))
- (Contrib) `routingconnector`: Add ability to route log records individually using OTTL log record context. ([#35939](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35939))
- (Contrib) `splunkenterprisereceiver`: Add new metrics for Splunk Enterprise dispatch artifacts ([#35950](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35950))
- (Core) `batchprocessor`: Move single shard batcher creation to the constructor ([#11594](https://github.com/open-telemetry/opentelemetry-collector/issues/11594))
- (Core) `service`: add support for using the otelzap bridge and emit logs using the OTel Go SDK ([#10544](https://github.com/open-telemetry/opentelemetry-collector/issues/10544))

### 🧰 Bug fixes 🧰

- (Contrib) `receiver/windowseventlog`: Fix panic when rendering long event messages. ([#36179](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/36179))
- (Contrib) `hostmetricsreceiver`: Do not set the default value of HOST_PROC_MOUNTINFO to respect root_path ([#35990](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35990))
- (Contrib) `prometheusexporter`: Fixes an issue where the prometheus exporter would not shut down the server when the collector was stopped. ([#35464](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35464))
- (Contrib) `k8sobserver`: Enable observation of ingress objects if the `ObserveIngresses` config option is set to true ([#35324](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35324))
- (Contrib) `pkg/stanza`: Fixed bug causing Operators with DropOnErrorQuiet to send log entries to the next operator. ([#35010](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35010))
  This issue was introduced by a bug fix meant to ensure Silent Operators are not logging errors ([#35010](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35010)). With this fix,
  this side effect bug has been resolved.

- (Contrib) `splunkhecreceiver`: Avoid a memory leak by changing how we record obsreports for logs and metrics. ([#35294](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35294))
- (Contrib) `receiver/filelog`: fix record counting with header ([#35869](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35869))
- (Contrib) `connector/routing`: Fix detection of duplicate conditions in routing table. ([#35962](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35962))
- (Contrib) `solacereceiver`: The Solace receiver may unexpectedly terminate on reporting traces when used with a memory limiter processor and under high load ([#35958](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35958))
- (Contrib) `pkg/stanza/operator`: Retain Operator should propagate the severity field ([#35832](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35832))
  The retain operator should propagate the severity field like it does with timestamps.

- (Contrib) `pkg/stanza`: Handle error of callback function of `ParserOperator.ProcessWithCallback` ([#35769](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35769))
  `ProcessWithCallback` of `ParserOperator` first calls the `ParseWith` method
  which properly handles errors with `HandleEntryError`.
  Then the callback function is called and its returned error should also
  be handled by the `HandleEntryError` ensuring a consistent experience.

- (Core) `service`: ensure traces and logs emitted by the otel go SDK use the same resource information ([#11578](https://github.com/open-telemetry/opentelemetry-collector/issues/11578))
- (Core) `config/configgrpc`: Patch for bug in the grpc-go NewClient that makes the way the hostname is resolved incompatible with the way proxy setting are applied. ([#11537](https://github.com/open-telemetry/opentelemetry-collector/issues/11537))

## v0.112.0

This Splunk OpenTelemetry Collector release includes changes from the opentelemetry-collector v0.112.0 and the opentelemetry-collector-contrib v0.112.0 releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Remove httpsink exporter ([#5503](https://github.com/signalfx/splunk-otel-collector/pull/5503))
- (Splunk) Remove signalfx-metadata and collectd/metadata monitors ([#5508](https://github.com/signalfx/splunk-otel-collector/pull/5508))
  Both monitors are deprecated and replaced by the hostmetricsreceiver and processlist monitor.
- (Splunk) Remove deprecated collectd/etcd monitor. [Please use the etcd prometheus endpoint to scrape metrics.](https://etcd.io/docs/v3.5/metrics/) ([#5520](https://github.com/signalfx/splunk-otel-collector/pull/5520))
- (Splunk) Remove deprecated collectd/health-checker monitor. ([#5522](https://github.com/signalfx/splunk-otel-collector/pull/5522))
- (Splunk) Remove deprecated loggingexporter from the distribution ([#5551](https://github.com/signalfx/splunk-otel-collector/pull/5551))
- (Core) `service`: Remove stable gate component.UseLocalHostAsDefaultHost ([#11412](https://github.com/open-telemetry/opentelemetry-collector/pull/11412))

### 🚩Deprecations 🚩

- (Splunk) Deprecate cloudfoundry monitor ([#5495](https://github.com/signalfx/splunk-otel-collector/pull/5495))
- (Splunk) Deprecate the heroku observer. Use the [resource detection observer with heroku detector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor#heroku) instead. ([#5496](https://github.com/signalfx/splunk-otel-collector/pull/5496))
- (Splunk) Deprecate mongodb atlas monitor. [Please use the mongodbatlasreceiver instead](https://docs.splunk.com/observability/en/gdi/opentelemetry/components/mongodb-atlas-receiver.html) ([#5500](https://github.com/signalfx/splunk-otel-collector/pull/5500))
- (Splunk) Deprecate python-monitor monitor ([#5501](https://github.com/signalfx/splunk-otel-collector/pull/5501))
- (Splunk) Deprecate windowslegacy monitor ([#5518](https://github.com/signalfx/splunk-otel-collector/pull/5518))
- (Splunk) Deprecate statsd monitor. Use the [statsd receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver) instead. ([#5513](https://github.com/signalfx/splunk-otel-collector/pull/5513))
- (Splunk) Deprecate the collectd/consul monitor. Please use the statsd or prometheus receiver instead. See https://developer.hashicorp.com/consul/docs/agent/monitor/telemetry for more information. ([#5521](https://github.com/signalfx/splunk-otel-collector/pull/5521))
- (Splunk) Deprecate collectd/mysql monitor. Use the [mysql receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/mysqlreceiver) instead. ([#5538](https://github.com/signalfx/splunk-otel-collector/pull/5538))
- (Splunk) Deprecate the collectd/nginx monitor. Please use the [nginx receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/nginxreceiver/) instead. ([#5537](https://github.com/signalfx/splunk-otel-collector/pull/5537))
- (Splunk) Deprecate the collectd/chrony monitor. Please use the [chronyreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/chronyreceiver) instead. ([#5536](https://github.com/signalfx/splunk-otel-collector/pull/5536))
- (Splunk) Deprecate the collectd/statsd monitor. Please use the [statsdreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver) instead. ([#5542](https://github.com/signalfx/splunk-otel-collector/pull/5542))
- (Splunk) Deprecate the ecs-metadata monitor ([#5541](https://github.com/signalfx/splunk-otel-collector/pull/5541))
- (Splunk) Deprecate the collectd/statsd monitor. Please use the [statsdreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/statsdreceiver) instead. ([#](https://github.com/signalfx/splunk-otel-collector/pull/))
- (Splunk) Deprecate the haproxy monitor. Please use the [haproxyreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/haproxyreceiver) instead. ([#5543](https://github.com/signalfx/splunk-otel-collector/pull/5543))
- (Contrib) `sapmreceiver`: Deprecate SAPM receiver ([#32125](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32125))

### 🚀 New components 🚀

- (Splunk) Add [chrony receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/chronyreceiver) ([#5535](https://github.com/signalfx/splunk-otel-collector/pull/5535))

### 💡 Enhancements 💡

- (Splunk) Update Python to 3.13.0 ([5552](https://github.com/signalfx/splunk-otel-collector/pull/5552))
- (Core) `confighttp`: Adding support for lz4 compression into the project ([#9128](https://github.com/open-telemetry/opentelemetry-collector/pull/9128))
- (Core) `service`: Hide profiles support behind a feature gate while it remains alpha. ([#11477](https://github.com/open-telemetry/opentelemetry-collector/pull/11477))
- (Core) `exporters`: Retry sender will fail fast when the context timeout is shorter than the next retry interval. ([#11183](https://github.com/open-telemetry/opentelemetry-collector/pull/11183))
- (Contrib) `azureeventshubreceiver`: Updates the Azure Event Hub receiver to use the new Resource Logs translator. ([#35357](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35357))
- (Contrib) `pkg/ottl`: Add ConvertAttributesToElementsXML Converter ([#35328](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35328))
- (Contrib) `azureblobreceiver`: adds support for using azidentity default auth, enabling the use of Azure Managed Identities, e.g. Workload Identities on AKS ([#35636](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35636))
  This change allows to use authentication type "default", which makes the receiver use azidentity default Credentials,
  which automatically picks up, identities assigned to e.g. a container or a VirtualMachine
- (Contrib) `k8sobserver`: Emit endpoint per Pod's container ([#35491](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35491))
- (Contrib) `mongodbreceiver`: Add support for MongoDB direct connection ([#35427](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35427))
- (Contrib) `chronyreceiver`: Move chronyreceiver to beta ([#35913](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35913))
- (Contrib) `pkg/ottl`: Parsing invalid statements and conditions now prints all errors instead of just the first one found. ([#35728](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35728))
- (Contrib) `pkg/ottl`: Add ParseSimplifiedXML Converter ([#35421](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35421))
- (Contrib) `routingconnector`: Allow routing based on OTTL Conditions ([#35731](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35731))
  Each route must contain either a statement or a condition.
- (Contrib) `sapmreceiver`: Respond 503 on non-permanent and 400 on permanent errors ([#35300](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35300))
- (Contrib) `hostmetricsreceiver`: Use HOST_PROC_MOUNTINFO as part of configuration instead of environment variable ([#35504](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35504))
- (Contrib) `pkg/ottl`: Add ConvertTextToElements Converter ([#35364](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35364))

### 🧰 Bug fixes 🧰

- (Core) `processorhelper`: Fix issue where in/out parameters were not recorded when error was returned from consumer. ([#11351](https://github.com/open-telemetry/opentelemetry-collector/pull/11351))
- (Contrib) `metricstransform`: The previously removed functionality of aggregating against an empty label set is restored. ([#34430](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34430))
- (Contrib) `filelogreceiver`: Supports `add_metadata_from_filepath` for Windows filepaths ([#35558](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35558))
- (Contrib) `filelogreceiver`: Suppress errors on EBADF when unlocking files. ([#35706](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35706))
  This error is harmless and happens regularly when delete_after_read is set. This is because we acquire the lock right at the start of the ReadToEnd function and then defer the unlock, but that function also performs the delete. So, by the time it returns and the defer runs the file descriptor is no longer valid.
- (Contrib) `kafkareceiver`: Fixes issue causing kafkareceiver to block during Shutdown(). ([#30789](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30789))
- (Contrib) `hostmetrics receiver`: Fix duplicate filesystem metrics ([#34635](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34635), [#34512](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34512))
  The hostmetrics exposes duplicate metrics of identical mounts exposed in namespaces. The duplication causes errors in exporters that are sensitive to duplicate metrics. We can safely drop the duplicates as the metrics should be exactly the same.
- (Contrib) `pkg/ottl`: Allow indexing string slice type ([#29441](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29441))
- (Contrib) `mysqlreceiver`: Add replica metric support for versions of MySQL earlier than 8.0.22. ([#35217](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35217))
- (Contrib) `stanza/input/windows`: Close remote session while resubscribing ([#35577](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35577))
- (Contrib) `receiver/windowseventlog`: Errors returned when passing data downstream will now be propagated correctly. ([#35461](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35461))

## v0.111.0


This Splunk OpenTelemetry Collector release includes changes from the opentelemetry-collector v0.111.0 and the opentelemetry-collector-contrib v0.111.0 releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) signalfxexporter: Do not exclude the metric container.memory.working_set ([#35475](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35475))
- (Contrib) sqlqueryreceiver: Fail if value for log column in result set is missing, collect errors ([#35068](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35068))
- (Contrib) windowseventlogreceiver: The 'raw' flag no longer suppresses rendering info. ([#34720](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34720))
- (Splunk) Remove deprecated memory ballast extension ([#5429](https://github.com/open-telemetry/opentelemetry-collector/pull/5429))

### 🚩Deprecations 🚩

- (Contrib) sapmreceiver: access_token_passthrough is deprecated ([#35330](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35330))
- (Splunk) Remove ballast extension entirely from components  ([#5429](https://github.com/signalfx/splunk-otel-collector/pull/5429))
- (Splunk) Deprecate jaegergrpc monitor ([#5428](https://github.com/signalfx/splunk-otel-collector/pull/5428))
- (Splunk) Deprecate the jaegergrpc monitor ([#5428](https://github.com/signalfx/splunk-otel-collector/pull/5428))

### 💡 Enhancements 💡

- (Splunk) Initial release of standalone collector binaries for Linux (amd64/arm64) and Windows (amd64) with FIPS 140-2 support. These are experimental (alpha) binaries, and it is not suitable to use them in production environments. ([#5378](https://github.com/signalfx/splunk-otel-collector/pull/5378)):
  - `otelcol-fips_linux_<amd64|arm64>`: Built with [`GOEXPERIMENT=boringcrypto`](https://go.dev/src/crypto/internal/boring/README) and [`crypto/tls/fipsonly`](https://go.dev/src/crypto/tls/fipsonly/fipsonly.go).
  - `otelcol-fips_windows_amd64.exe`: Built with [`GOEXPERIMENT=cngcrypto`](https://github.com/microsoft/go/blob/microsoft/main/eng/doc/fips/README.md) and [`requirefips`](https://github.com/microsoft/go/blob/microsoft/main/eng/doc/fips/README.md#build-option-to-require-fips-mode) (the collector will panic if FIPS is not enabled on the Windows host).
  - Smart Agent components are not currently supported.
  - Download the binaries from the list of assets below.
- (Core) `confignet:` Add Profiles Marshaler to otlptext. ([#11161](https://github.com/open-telemetry/opentelemetry-collector/pull/11161))
- (Contrib) `receivercreator:` Validate endpoint's configuration before starting receivers ([#33145](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33145))
- (Contrib) `receiver/statsd:` Add support for aggregating on Host/IP ([#23809](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23809))
- (Contrib) `hostmetricsreceiver:` Add ability to mute all errors (mainly due to access rights) coming from process scraper of the hostmetricsreceiver ([#20435](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20435))
- (Contrib) `kubeletstats:` Introduce feature gate for deprecation of container.cpu.utilization, k8s.pod.cpu.utilization and k8s.node.cpu.utilization metrics ([#35139](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35139))
- (Contrib) `pkg/ottl:` Add InsertXML Converter ([#35436](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35436))
- (Contrib) `pkg/ottl`: Add GetXML Converter ([#35462](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35462))
- (Contrib) `pkg/ottl`: Add ToKeyValueString Converter ([#35334](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35334))
- (Contrib) `pkg/ottl`: Add RemoveXML Converter ([#35301](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35301))
- (Contrib) `sqlserverreceiver:` Add computer name resource attribute to relevant metrics ([#35040](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35040))
- (Contrib) `windowseventlogreceiver:` Add 'suppress_rendering_info' option. ([#34720](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34720))
- (Contrib) `receiver/awss3receiver:` Add ingest progress notifications via OpAMP ([#33980](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33980))
- (Contrib) `receiver/azureblobreceiver:` support for default auth ([#35636](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35636))
- (Contrib) update sapm-proto to 0.16.0 ([#35630](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35630))


### 🧰 Bug fixes 🧰

- (Contrib) `windowseventlogreceiver:` While collecting from a remote windows host, the stanza operator will no longer log "subscription handle is already open" constantly during successful collection. ([#35520](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35520))
- (Contrib) `windowseventlogreceiver:` If collecting from a remote host, the receiver will stop collecting if the host restarts. This change resubscribes when the host restarts. ([#35175](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35175))
- (Contrib) `sqlqueryreceiver:` Fix reprocessing of logs when tracking_column type is timestamp ([#35194](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35194))
- (Core) `processorhelper`: Fix bug where record in/out metrics were skipped ([#11360](https://github.com/open-telemetry/opentelemetry-collector/pull/11360))

## v0.110.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.110.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.110.0) and the [opentelemetry-collector-contrib v0.110.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.110.0) releases where appropriate.

Additionally, updates `splunk-otel-javaagent` to [`v2.8.1`](https://github.com/signalfx/splunk-otel-java/releases/tag/v2.8.1) and `jmx-metric-gatherer` to [`v1.39.0`](https://github.com/open-telemetry/opentelemetry-java-contrib/releases/tag/v1.39.0)

### 🛑 Breaking changes 🛑

- (Core) `processorhelper`: Update incoming/outgoing metrics to a single metric with `otel.signal` attributes. ([#11144](https://github.com/open-telemetry/opentelemetry-collector/pull/11144))
- (Core) processorhelper: Remove deprecated [Traces|Metrics|Logs]Inserted funcs ([#11151](https://github.com/open-telemetry/opentelemetry-collector/pull/11151))
- (Core) config: Mark UseLocalHostAsDefaultHostfeatureGate as stable  ([#11235](https://github.com/open-telemetry/opentelemetry-collector/pull/11235))
- (Contrib) `pkg/stanza`: Move filelog.container.removeOriginalTimeField feature gate to beta. Disable the filelog.container.removeOriginalTimeField feature gate to get the old behavior. ([#33389](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33389))
- (Contrib) `resourcedetectionprocessor`: Move processor.resourcedetection.hostCPUSteppingAsString feature gate to stable. ([#31136](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31136))
- (Contrib) `resourcedetectionprocessor`: Remove processor.resourcedetection.hostCPUModelAndFamilyAsString feature gate. ([#29025](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29025))


### 🚩 Deprecations 🚩

- (Core) `processorhelper`: deprecate accepted/refused/dropped metrics ([#11201](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/11201))
- (Contrib) `hostmetricsreceiver`: Set the receiver.hostmetrics.normalizeProcessCPUUtilization feature gate to stable. ([#34763](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34763))

### 💡 Enhancements 💡

- (Core) `confignet`: Mark module as Stable. ([#9801](https://github.com/open-telemetry/opentelemetry-collector/pull/9801))
- (Core) `confmap/provider/envprovider`: Support default values when env var is empty ([#5228](https://github.com/open-telemetry/opentelemetry-collector/pull/5228))
- (Core) `service/telemetry`: Mark useOtelWithSDKConfigurationForInternalTelemetry as stable ([#7532](https://github.com/open-telemetry/opentelemetry-collector/pull/7532))
- (Contrib) `processor/transform`: Add custom function to the transform processor to convert exponential histograms to explicit histograms. ([#33827](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33827))
- (Contrib) `file_storage`: provide a new option to the user to create a directory on start ([#34939](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34939))
- (Contrib) `headersetterextension`: adding default_value config. `default_value` config item applied in case context value is empty. ([#34412](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34412))
- (Contrib) `kafkaexporter`: Add support for encoding extensions in the Kafka exporter. This change adds support for encoding extensions in the Kafka exporter. Loading extensions takes precedence over the internally supported encodings. ([#34384](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34384))
- (Contrib) `kafkareceiver`: Add support for otlp_json encoding to Kafka receiver. The payload is deserialized into OpenTelemetry traces using JSON format. This encoding allows the Kafka receiver to handle trace data in JSON format, enabling integration with systems that export traces as JSON-encoded data. ([#33627](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33627))
- (Contrib) `pkg/ottl`: Improved JSON unmarshaling performance by 10-20% by switching dependencies. ([#35130](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35130))
- (Contrib) `pkg/ottl`: Added support for locale in the Time converter ([#32978](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32978))
- (Contrib) `remotetapprocessor`: Origin header is no longer required for websocket connections ([#34925](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34925))
- (Contrib) `transformprocessor`: Remove unnecessary data copy when transform sum to/from gauge ([#35177](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35177))
- (Contrib) `sapmexporter`: Prioritize token in context when accesstokenpassthrough is enabled ([#35123](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35123))
- (Contrib) `tailsamplingprocessor`: Fix the behavior for numeric tag filters with inverse_match set to true. ([#34296](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34296))
- (Splunk) Update instruction for partial discovery ([#5402](https://github.com/signalfx/splunk-otel-collector/pull/5402))

### 🧰 Bug fixes 🧰

- (Core) `service`: Ensure process telemetry is registered when internal telemetry is configured with readers instead of an address. ([#11093](https://github.com/open-telemetry/opentelemetry-collector/pull/11093))
- (Contrib) `splunkenterprise`: Fix a flaky search related to iops metrics. ([#35081](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35081))
- (Contrib) `azuremonitorexporter`: fix issue for property endpoint is ignored when using instrumentation_key ([#33971](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33971))
- (Contrib) `pkg/stanza`: Do not get formatted message for Windows events without an event provider. Attempting to get the formatted message for Windows events without an event provider can result in an error being logged. This change ensures that the formatted message is not retrieved for such events. ([#35135](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35135))
- (Contrib) `signalfxexporter`: Ensure token is not sent through for event data ([#35154](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35154))
- (Contrib) `prometheusreceiver`: Fix the retrieval of scrape configurations by also considering scrape config files ([#34786](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34786))
- (Contrib) `redactionprocessor`: Fix panic when using the redaction processor in a logs pipeline ([#35331](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35331))
- (Contrib) `exporter/splunkhec`: Fix incorrect claim that the exporter doesn't mutate data when batching is enabled. The bug lead to runtime panics when the exporter was used with the batcher enabled in a fanout scenario. ([#35306](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/35306))
- (Splunk) Choco upgrade should preserve collector service custom env vars ([#5386](https://github.com/signalfx/splunk-otel-collector/pull/5386))
- (Splunk) `discoveryreceiver` with `splunk.continuousDiscovery` feature gate enabled: Remove redundant discovery.receiver.rule attribute ([#5403](https://github.com/signalfx/splunk-otel-collector/pull/5403))
- (Splunk) `discoveryreceiver` with `splunk.continuousDiscovery` feature gate enabled: Remove redundant resource attributes ([#5409](https://github.com/signalfx/splunk-otel-collector/pull/5409))

## v0.109.0

### 🛑 Breaking changes 🛑

- (Splunk) Update Python to 3.12.5 in the Smart Agent bundle for Linux and Windows. Check [What’s New In Python 3.12](https://docs.python.org/3/whatsnew/3.12.html) for details. ([#5298](https://github.com/signalfx/splunk-otel-collector/pull/5298))
- (Contrib) `spanmetricsconnector`: Improve consistency between metrics generated by spanmetricsconnector. Added traces.span.metrics as default namespace ([#33227](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33227)
  Default namespace for the generated metrics is traces.span.metrics now. | The deprecated metrics are: calls, duration and events. | The feature flag connector.spanmetrics.legacyLatencyMetricNames was added to revert the behavior.
- (Contrib) `ottl`: Remove tracing from OTTL due to performance concerns ([#34910](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34910)

### 🚀 New components 🚀

- (Splunk) Add `apachespark` receiver ([#5318](https://github.com/signalfx/splunk-otel-collector/pull/5318))
- (Splunk) Add `nop` receiver and exporter ([#5355](https://github.com/signalfx/splunk-otel-collector/pull/5355))

### 💡 Enhancements 💡

- (Splunk) Apache Web Server Auto Discovery: set the default discovered endpoint to match the OpenTelemetry `apachereceiver` default: `http://`endpoint`/server-status?auto` ([#5353](https://github.com/signalfx/splunk-otel-collector/pull/5353))
  If the collector is running as a process on the host OS and the Apache Web Server is in a Docker container add `--set=splunk.discovery.extensions.docker_observer.config.use_host_bindings=true` to the command-line arguments for the discovery to create the correct endpoint.
- (Splunk) Introduce continuous service discovery mode. This mode can be enabled with a feature gate by adding `--feature-gates=splunk.continuousDiscovery` command line argument. ([#5363](https://github.com/signalfx/splunk-otel-collector/pull/5363))
  The new mode does the following:
  - It allows discovering new services that were not available at the time of the collector startup. If discovery is
    successful, the metrics collection will be started.
  - Information about discovered services is being sent to Splunk Observability Cloud. The information will include
    instructions to complete discovery for particular services if the discovery was not successful out of the box.
- (Core) `service`: move `useOtelWithSDKConfigurationForInternalTelemetry` gate to beta ([#11091](https://github.com/open-telemetry/opentelemetry-collector/issues/11091))
- (Core) `service`: implement a no-op tracer provider that doesn't propagate the context ([#11026](https://github.com/open-telemetry/opentelemetry-collector/issues/11026))
  The no-op tracer provider supported by the SDK incurs a memory cost of propagating the context no matter
  what. This is not needed if tracing is not enabled in the Collector. This implementation of the no-op tracer
  provider removes the need to allocate memory when tracing is disabled.
- (Core) `processor`: Add incoming and outgoing counts for processors using processorhelper. ([#10910](https://github.com/open-telemetry/opentelemetry-collector/issues/10910))
  Any processor using the processorhelper package (this is most processors) will automatically report
  incoming and outgoing item counts. The new metrics are:
  - otelcol_processor_incoming_spans
  - otelcol_processor_outgoing_spans
  - otelcol_processor_incoming_metric_points
  - otelcol_processor_outgoing_metric_points
  - otelcol_processor_incoming_log_records
  - otelcol_processor_outgoing_log_records
- (Contrib) `pkg/ottl`: Added Decode() converter function ([#32493](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32493)
- (Contrib) `filestorage`: Add directory validation for compaction on-rebound ([#35114](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35114)
- (Contrib) `windowseventlogreceiver`: Avoid rendering the whole event to obtain the provider name ([#34755](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34755)
- (Contrib) `splunkhecexporter`: Drop empty log events ([#34871](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34871)
  Log records with no body are dropped by Splunk on reception as they contain no log message, albeit they may have attributes.
  This is in tune with the behavior of splunkhecreceiver, which refuses HEC events with no event ([#19769](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19769)
- (Contrib) `transformprocessor`: Support aggregating metrics based on their attribute values and substituting the values with a new value. ([#16224](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/16224)
- (Contrib) `kafkareceiver`: Adds tunable fetch sizes to Kafka Receiver ([#22741](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22741)
  Adds the ability to tune the minumum, default and maximum fetch sizes for the Kafka Receiver
- (Contrib) `kafkareceiver`: Add support for encoding extensions in the Kafka receiver. ([#33888](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33888)
  This change adds support for encoding extensions in the Kafka receiver. Loading extensions takes precedence over the internally supported encodings.
- (Contrib) `pkg/ottl`: Add `Sort` function to sort array to ascending order or descending order ([#34200](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34200)
- (Contrib) `redactionprocessor`: Add support for logs and metrics ([#34479](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34479)
- (Contrib) `spanmetricsconnector`: Extract the `getDimensionValue` function as a common function. ([#34627](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34627)
- (Contrib) `sqlqueryreceiver`: Support populating log attributes from sql query ([#24459](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24459)

### 🧰 Bug fixes 🧰

- (Core) `configgrpc`: Change the value of max_recv_msg_size_mib from uint64 to int to avoid a case where misconfiguration caused an integer overflow. ([#10948](https://github.com/open-telemetry/opentelemetry-collector/issues/10948))
- (Core) `exporterqueue`: Fix a bug in persistent queue that Offer can becomes deadlocked when queue is almost full ([#11015](https://github.com/open-telemetry/opentelemetry-collector/issues/11015))
- (Contrib) `apachereceiver`: Fix panic on invalid endpoint configuration ([#34992](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34992)
- (Contrib) `fileconsumer`: Fix bug where max_concurrent_files could not be set to 1. ([#35080](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35080)
- (Contrib) `hostmetricsreceiver`: In filesystem scraper, do not prefix partitions when using the environment variable HOST_PROC_MOUNTINFO ([#35043](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35043)
- (Contrib) `splunkhecreceiver`: Fix memory leak when the receiver is used for both metrics and logs at the same time ([#34886](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34886)
- (Contrib) `pkg/stanza`: Synchronize shutdown in stanza adapter ([#31074](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31074)
  Stanza-based receivers should now flush all data before shutting down
- (Contrib) `sqlserverreceiver`: Fix bug where metrics were being emitted with the wrong database name resource attribute ([#35036](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35036)
- (Contrib) `signalfxexporter`: Fix memory leak by re-organizing the exporter's functionality lifecycle ([#32781](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32781)
- (Contrib) `otlpjsonconnector`: Handle OTLPJSON unmarshal error ([#34782](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34782)
- (Contrib) `mysqlreceiver`: mysql client raise error when the TABLE_ROWS column is NULL, convert NULL to int64 ([#34195](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34195)
- (Contrib) `pkg/stanza`: An operator configured with silent errors shouldn't log errors while processing log entries. ([#35008](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35008)

## v0.108.1

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.108.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.108.1) and the [opentelemetry-collector-contrib v0.108.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.108.0) releases where appropriate.

### 🚩Deprecations 🚩

- (Splunk) Deprecate the nagios monitor ([#5172](https://github.com/signalfx/splunk-otel-collector/pull/5172))

### 🧰 Bug fixes 🧰

- (Splunk) Discovery observers start failures should not stop the collector ([#5299](https://github.com/signalfx/splunk-otel-collector/pull/5299))

## v0.108.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.108.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.108.1) and the [opentelemetry-collector-contrib v0.108.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.108.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `confmap`: Mark `confmap.strictlyTypedInput` as stable ([#10552](https://github.com/open-telemetry/opentelemetry-collector/issues/10552))
- (Contrib) `splunkhecexporter`: The scope name has been updated from `otelcol/splunkhec` to `github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter` ([#34710](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34710))
- (Contrib) `transformprocessor`: Promote processor.transform.ConvertBetweenSumAndGaugeMetricContext feature flag from alpha to beta ([#34567](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34567))
- (Contrib) `vcenterreceiver`: Several host performance metrics now return 1 data point per time series instead of 5. ([#34708](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34708))
  The 5 data points previously sent represented consecutive 20s sampling periods. Depending on the collection interval
  these could easily overlap. Sending just the latest of these data points is more in line with other performance metrics.

  This change also fixes an issue with the googlecloud exporter seeing these datapoints as duplicates.

  Following is the list of affected metrics which will now only report a single datapoint per set of unique attribute values.
  - vcenter.host.cpu.reserved
  - vcenter.host.disk.latency.avg
  - vcenter.host.disk.latency.max
  - vcenter.host.disk.throughput
  - vcenter.host.network.packet.drop.rate
  - vcenter.host.network.packet.error.rate
  - vcenter.host.network.packet.rate
  - vcenter.host.network.throughput
  - vcenter.host.network.usage

### 🚀 New components 🚀

- (Splunk) Add headersetterextension ([#5276](https://github.com/signalfx/splunk-otel-collector/pull/5276))
- (Splunk) Add `nginx` receiver ([5229](https://github.com/signalfx/splunk-otel-collector/pull/5229))

### 💡 Enhancements 💡

- (Core) `exporter/otlp`: Add batching option to otlp exporter ([#8122](https://github.com/open-telemetry/opentelemetry-collector/issues/8122))
- (Core) `service`: Adds `level` configuration option to `service::telemetry::trace` to allow users to disable the default TracerProvider ([#10892](https://github.com/open-telemetry/opentelemetry-collector/issues/10892))
  This replaces the feature gate `service.noopTracerProvider` introduced in v0.107.0
- (Contrib) `awss3receiver`: Enhance the logging of the AWS S3 Receiver in normal operation to make it easier for user to debug what is happening. ([#30750](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30750))
- (Contrib) `filelogreceiver`: If acquire_fs_lock is true, attempt to acquire a shared lock before reading a file. ([#34801](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34801))
  Unix only. If a lock cannot be acquired then the file will be ignored until the next poll cycle.
- (Contrib) `solacereceiver`: Updated the format for generated metrics. Included a `receiver_name` attribute that identifies the Solace receiver that generated the metrics ([#34541](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34541))
- (Contrib) `prometheusreceiver`: Ensure Target Allocator's confighttp is used in the receiver's service discovery ([#33370](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33370))
- (Contrib) `metricstransformprocessor`: Add scaling exponential histogram support ([#29803](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29803))
- (Contrib) `pkg/ottl`: Introduce `UserAgent` converter to parse UserAgent strings ([#32434](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32434))
- (Splunk) Update JMX Metric Gatherer to [v1.38.0](https://github.com/open-telemetry/opentelemetry-java-contrib/releases/tag/v1.37.0) ([#5287](https://github.com/signalfx/splunk-otel-collector/pull/5287))
- (Splunk) Auto Instrumentation for Linux ([#5243](https://github.com/signalfx/splunk-otel-collector/pull/5243))
  - Add support for the `OTEL_LOGS_EXPORTER` environment variable to `libsplunk.so` for system-wide auto instrumentation.
  - Linux installer script: Add the `--logs-exporter <value>` option:
    - Set the exporter for collected logs by all activated SDKs, for example `otlp`.
    - Set the value to `none` to disable collection and export of logs.
    - The value will be set to the `OTEL_LOGS_EXPORTER` environment variable.
    - Defaults to `''` (empty), i.e. defer to the default `OTEL_LOGS_EXPORTER` value for each activated SDK.

### 🧰 Bug fixes 🧰

- (Core) `batchprocessor`: Update units for internal telemetry ([#10652](https://github.com/open-telemetry/opentelemetry-collector/issues/10652))
- (Core) `confmap`: Fix bug where an unset env var used with a non-string field resulted in a panic ([#10950](https://github.com/open-telemetry/opentelemetry-collector/issues/10950))
- (Core) `service`: Fix memory leaks during service package shutdown ([#9165](https://github.com/open-telemetry/opentelemetry-collector/issues/9165))
- (Core) `confmap`: Use string representation for field types where all primitive types are strings. ([#10937](https://github.com/open-telemetry/opentelemetry-collector/issues/10937))
- (Core) `otelcol`: Preserve internal representation when unmarshaling component configs ([#10552](https://github.com/open-telemetry/opentelemetry-collector/issues/10552))
- (Contrib) `tailsamplingprocessor`: Update the `policy` value in metrics dimension value to be unique across multiple tail sampling components with the same policy name. ([#34192](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34192))
  This change ensures that the `policy` value in the metrics exported by the tail sampling processor is unique across multiple tail sampling processors with the same policy name.
- (Contrib) `prometheusreceiver`: Group scraped metrics into resources created from `job` and `instance` label pairs ([#34237](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34237))
  The receiver will now create a resource for each distinct job/instance label combination.
  In addition to the label/instance pairs detected from the scraped metrics, a resource representing the overall
  scrape configuration will be created. This additional resource will contain the scrape metrics, such as the number of scraped metrics, the scrape duration, etc.
- (Contrib) `tailsamplingprocessor`: Fix the behavior for numeric tag filters with `inverse_match` set to `true`. ([#34296](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34296))
- (Contrib) `pkg/stanza`: fix nil value conversion ([#34672](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34672))
- (Contrib) `k8sclusterreceiver`: Lower the log level of a message indicating a cache miss from WARN to DEBUG. ([#34817](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/34817))

## v0.107.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.107.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.107.0) and the [opentelemetry-collector-contrib v0.107.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.107.0) releases where appropriate.

This release fixes CVE-2024-42368 on the `bearerauthtokenextension` ([#34516](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34516)). The Splunk distribution was not impacted by this CVE.

### 🛑 Breaking changes 🛑

- (Splunk) `confmap`: Do not expand special shell variable such as `$*` in configuration files. ([#5206](https://github.com/signalfx/splunk-otel-collector/pull/5206))
- (Splunk) Upgrade golang to 1.22 ([#5248](https://github.com/signalfx/splunk-otel-collector/pull/5248))

- (Core) `service`: Remove OpenCensus bridge completely, mark feature gate as stable. ([#10414](https://github.com/open-telemetry/opentelemetry-collector/pull/10414))

- (Contrib) Update the scope name for telemetry produce by components. The following table summarizes the changes:

| Component name | Previous scope | New scope |  PR number |
|----------------|----------------|-----------|------------|
| `azureeventhubreceiver` | `otelcol/azureeventhubreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver` |  ([#34611](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34611)) |
| `cloudfoundryreceiver` | `otelcol/cloudfoundry` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver` |  ([#34612](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34612)) |
| `azuremonitorreceiver` | `otelcol/azuremonitorreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver` |  ([#34618](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34618)) |
| `fileconsumer` | `otelcol/fileconsumer` | `github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer` |  ([#34619](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34619)) |
| `loadbalancingexporter` | `otelcol/loadbalancing` | `github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter` |  ([#34429](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34429)) |
| `apachereceiver` | `otelcol/apachereceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver` |  ([#34517](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34517)) |
| `countconnector` | `otelcol/countconnector` | `github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector` |  ([#34583](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34583)) |
| `elasticsearchreceiver` | `otelcol/elasticsearchreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver` |  ([#34529](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34529)) |
| `filterprocessor` | `otelcol/filter` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `fluentforwardreceiver` | `otelcol/fluentforwardreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver` |  ([#34534](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34534)) |
| `groupbyattrsprocessor` | `otelcol/groupbyattrs` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `haproxyreceiver` | `otelcol/haproxyreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver` |  ([#34498](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34498)) |
| `hostmetricsreceiver` receiver's scrapers | `otelcol/hostmetricsreceiver/*` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/*` |  ([#34526](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34526)) |
| `httpcheckreceiver` | `otelcol/httpcheckreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver` |  ([#34497](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34497)) |
| `k8sattributesprocessor` | `otelcol/k8sattributes` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `k8sclusterreceiver` | `otelcol/k8sclusterreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver` |  ([#34536](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34536)) |
| `kafkametricsreceiver` | `otelcol/kafkametricsreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver` |  ([#34538](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34538)) |
| `kafkareceiver` | `otelcol/kafkareceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver` |  ([#34539](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34539)) |
| `kubeletstatsreceiver` | `otelcol/kubeletstatsreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver` |  ([#34537](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34537)) |
| `mongodbatlasreceiver` | `otelcol/mongodbatlasreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver` |  ([#34543](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34543)) |
| `mongodbreceiver` | `otelcol/mongodbreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver` |  ([#34544](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34544)) |
| `mysqlreceiver` | `otelcol/mysqlreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver` |  ([#34545](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34545)) |
| `nginxreceiver` | `otelcol/nginxreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver` |  ([#34493](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34493)) |
| `oracledbreceiver` | `otelcol/oracledbreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver` |  ([#34491](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34491)) |
| `postgresqlreceiver` | `otelcol/postgresqlreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver` |  ([#34476](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34476)) |
| `probabilisticsamplerprocessor` | `otelcol/probabilisticsampler` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `prometheusreceiver` | `otelcol/prometheusreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver` |  ([#34589](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34589)) |
| `rabbitmqreceiver` | `otelcol/rabbitmqreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver` |  ([#34475](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34475)) |
| `sshcheckreceiver` | `otelcol/sshcheckreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver` |  ([#34448](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34448)) |
| `vcenterreceiver` | `otelcol/vcenter` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver` |  ([#34449](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34449)) |
| `redisreceiver` | `otelcol/redisreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver` |  ([#34470](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34470)) |
| `routingprocessor` | `otelcol/routing` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `solacereceiver` | `otelcol/solacereceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver` |  ([#34466](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34466)) |
| `splunkenterprisereceiver` | `otelcol/splunkenterprisereceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver` |  ([#34452](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34452)) |
| `statsdreceiver` | `otelcol/statsdreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver` |  ([#34547](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34547)) |
| `tailsamplingprocessor` | `otelcol/tailsampling` | `github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor` |  ([#34550](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34550)) |
| `sqlserverreceiver` | `otelcol/sqlserverreceiver` | `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver` |  ([#34451](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34451)) |

- (Contrib) `elasticsearchreceiver`: Enable more index metrics by default ([#34396](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34396))
  This enables the following metrics by default:
  `elasticsearch.index.documents`
  `elasticsearch.index.operations.merge.current`
  `elasticsearch.index.segments.count`
  To preserve previous behavior, update your Elasticsearch receiver configuration to disable these metrics.
- (Contrib) `vcenterreceiver`: Enables all of the vSAN metrics by default. ([#34409](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34409))
  The following metrics will be enabled by default now:
  - vcenter.cluster.vsan.throughput
  - vcenter.cluster.vsan.operations
  - vcenter.cluster.vsan.latency.avg
  - vcenter.cluster.vsan.congestions
  - vcenter.host.vsan.throughput
  - vcenter.host.vsan.operations
  - vcenter.host.vsan.latency.avg
  - vcenter.host.vsan.congestions
  - vcenter.host.vsan.cache.hit_rate
  - vcenter.vm.vsan.throughput
  - vcenter.vm.vsan.operations
  - vcenter.vm.vsan.latency.avg
- (Contrib) `vcenterreceiver`: Several host performance metrics now return 1 data point per time series instead of 5. ([#34708](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34708))
  The 5 data points previously sent represented consecutive 20s sampling periods. Depending on the collection interval
  these could easily overlap. Sending just the latest of these data points is more in line with other performance metrics.

  This change also fixes an issue with the googlecloud exporter seeing these datapoints as duplicates.

  Following is the list of affected metrics which will now only report a single datapoint per set of unique attribute values.
  - vcenter.host.cpu.reserved
  - vcenter.host.disk.latency.avg
  - vcenter.host.disk.latency.max
  - vcenter.host.disk.throughput
  - vcenter.host.network.packet.drop.rate
  - vcenter.host.network.packet.error.rate
  - vcenter.host.network.packet.rate
  - vcenter.host.network.throughput
  - vcenter.host.network.usage

- (Splunk) Remove converters helping with old breaking changes. If those changes were not addressed, the collector will fail to start. ([#5267](https://github.com/signalfx/splunk-otel-collector/pull/5267))
  - Moving TLS config options in HEC exporter under tls group
  - Moving TLS insecure option in OTLP exporter under tls group
  - Renaming processor: k8s_tagger -> k8sattributes
  - Deprecation and removal of `ballast` extension
  - Debug exporter: `loglevel` -> `verbosity` renaming

### 🚀 New components 🚀

- (Splunk) Add Azure Blob receiver ([#5200](https://github.com/signalfx/splunk-otel-collector/pull/5200))
- (Splunk) Add Google Cloud PubSub receiver ([#5200](https://github.com/signalfx/splunk-otel-collector/pull/5200))

### 💡 Enhancements 💡

- (Core) `confmap`: Allow using any YAML structure as a string when loading configuration. ([#10800](https://github.com/open-telemetry/opentelemetry-collector/pull/10800))
  Previous to this change, slices could not be used as strings in configuration.
- (Core) `client`: Mark module as stable. ([#10775](https://github.com/open-telemetry/opentelemetry-collector/pull/10775))

- (Contrib) `azureeventhubreceiver`: Added traces support in azureeventhubreceiver ([#33583](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33583))
- (Contrib) `processor/k8sattributes`: Add support for `container.image.repo_digests` metadata ([#34029](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34029))
- (Contrib) `hostmetricsreceiver`: add reporting interval to entity event ([#34240](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34240))
- (Contrib) `elasticsearchreceiver`: Add metric for active index merges ([#34387](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34387))
- (Contrib) `kafkaexporter`: add an ability to partition logs based on resource attributes. ([#33229](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33229))
- (Contrib) `pkg/ottl`: Add support for map literals in OTTL ([#32388](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32388))
- (Contrib) `pkg/ottl`: Introduce ExtractGrokPatterns converter ([#32593](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32593))
- (Contrib) `pkg/ottl`: Add the `MD5` function to convert the `value` into a MD5 hash/digest ([#33792](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33792))
- (Contrib) `pkg/ottl`: Introduce `sha512` converter to generate SHA-512 hash/digest from given payload. ([#34007](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34007))
- (Contrib) `kafkametricsreceiver`: Add option to configure cluster alias name and add new metrics for kafka topic configurations ([#34148](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34148))
- (Contrib) `receiver/splunkhec`: Add a regex to enforce metrics naming for Splunk events fields based on metrics documentation. ([#34275](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34275))
- (Contrib) `filelogreceiver`: Check for unsupported fractional seconds directive when converting strptime time layout to native format ([#34390](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34390))
- (Contrib) `windowseventlogreceiver`: Add remote collection support to Stanza operator windows pkg to support remote log collect for the Windows Event Log receiver. ([#33100](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33100))
- (Contrib) `solacereceiver`: Updated the format for generated metrics. Included a `receiver_name` attribute that identifies the Solace receiver that generated the metrics ([#34541](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34541))
- (Contrib) `metricstransformprocessor`: Add scaling exponential histogram support ([#29803](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29803))

### 🧰 Bug fixes 🧰

- (Core) `configtelemetry`: Add 10s read header timeout on the configtelemetry Prometheus HTTP server. ([#5699](https://github.com/open-telemetry/opentelemetry-collector/pull/5699))
- (Core) `service`: Allow users to disable the tracer provider via the feature gate `service.noopTracerProvider` ([#10858](https://github.com/open-telemetry/opentelemetry-collector/pull/10858))
  The service is returning an instance of a SDK tracer provider regardless of whether there were any processors configured causing resources to be consumed unnecessarily.
- (Core) `processorhelper`: Fix processor metrics not being reported initially with 0 values. ([#10855](https://github.com/open-telemetry/opentelemetry-collector/pull/10855))
- (Core) `service`: Implement the `temporality_preference` setting for internal telemetry exported via OTLP ([#10745](https://github.com/open-telemetry/opentelemetry-collector/pull/10745))
- (Core) `configauth`: Fix unmarshaling of authentication in HTTP servers. ([#10750](https://github.com/open-telemetry/opentelemetry-collector/pull/10750))

- (Core) `component`: Allow component names of up to 1024 characters in length. ([#10816](https://github.com/open-telemetry/opentelemetry-collector/pull/10816))
- (Core) `service`: Fix memory leaks during service package shutdown ([#9241](https://github.com/open-telemetry/opentelemetry-collector/pull/9241))
- (Core) `batchprocessor`: Update units for internal telemetry ([#10652](https://github.com/open-telemetry/opentelemetry-collector/pull/10652))

- (Contrib) `configauth`: Fix unmarshaling of authentication in HTTP servers. ([#34325](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34325))
  This brings in a bug fix from the core collector. See https://github.com/open-telemetry/opentelemetry-collector/issues/10750.
- (Contrib) `docker_observer`: Change default endpoint for `docker_observer` on Windows to `npipe:////./pipe/docker_engine` ([#34358](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34358))
- (Contrib) `pkg/translator/jaeger`: Change the translation to jaeger spans to match semantic conventions. ([#34368](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34368))
  `otel.library.name` is deprecated and replaced by `otel.scope.name`
  `otel.library.version` is deprecated and replaced by `otel.scope.version`

- (Contrib) `pkg/stanza`: Ensure that errors from `Process` and `Write` do not break for loops ([#34295](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34295))
- (Contrib) `azuremonitorreceiver`: Add Azure China as a `cloud` option. ([#34315](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34315))
- (Contrib) `postgresqlreceiver`: Support unix socket based replication by handling null values in the client_addr field ([#33107](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33107))
- (Contrib) `splunkhecexporter`: Copy the bytes to be placed in the request body to avoid corruption on reuse ([#34357](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34357))
  This bug is a manifestation of https://github.com/golang/go/issues/51907.
  Under high load, the pool of buffers used to send requests is reused enough
  that the same buffer is used concurrently to process data and be sent as request body.
  The fix is to copy the payload into a new byte array before sending it.
- (Contrib) `pkg/stanza`: fix nil value conversion ([#34672](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34762))

## v0.106.1

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.106.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.106.1) and the [opentelemetry-collector-contrib v0.106.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.106.1) releases where appropriate.

### 🧰 Bug fixes 🧰

- (Splunk) Upgrade some `core` dependencies to proper `v0.106.1` version. ([#5203](https://github.com/signalfx/splunk-otel-collector/pull/5203))

## v0.106.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.106.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.106.0)-[v0.106.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.106.1) and the [opentelemetry-collector-contrib v0.106.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.106.0)-[v0.106.1](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.106.1) releases where appropriate.

Note: Some `core` dependencies were incorrectly still set to `v0.105.0` for this release.

### 🛑 Breaking changes 🛑

- (Core) `service`: Update all metrics to include `otelcol_` prefix to ensure consistency across OTLP and Prometheus metrics ([#9759](https://github.com/open-telemetry/opentelemetry-collector/pull/9759))
  This change is marked as a breaking change as anyone that was using OTLP for metrics will
  see the new prefix which was not present before. Prometheus generated metrics remain
  unchanged.
- (Core) `confighttp`: Delete `ClientConfig.CustomRoundTripper` ([#8627](https://github.com/open-telemetry/opentelemetry-collector/pull/8627))
  Set (*http.Client).Transport on the *http.Client returned from ToClient to configure this.
- (Core) `confmap`: When passing configuration for a string field using any provider, use the verbatim string representation as the value. ([#10605](https://github.com/open-telemetry/opentelemetry-collector/pull/10605), [#10405](https://github.com/open-telemetry/opentelemetry-collector/pull/10405))
  This matches the behavior of `${ENV}` syntax prior to the promotion of the `confmap.unifyEnvVarExpansion` feature gate
  to beta. It changes the behavior of the `${env:ENV}` syntax with escaped strings.
- (Core) `component`: Adds restrictions on the character set for component.ID name. ([#10673](https://github.com/open-telemetry/opentelemetry-collector/pull/10673))
- (Core) `processor/memorylimiter`: The memory limiter processor will no longer account for ballast size. ([#10696](https://github.com/open-telemetry/opentelemetry-collector/pull/10696))
  If you are already using GOMEMLIMIT instead of the ballast extension this does not affect you.
- (Core) `extension/memorylimiter`: The memory limiter extension will no longer account for ballast size. ([#10696](https://github.com/open-telemetry/opentelemetry-collector/pull/10696))
  If you are already using GOMEMLIMIT instead of the ballast extension this does not affect you.
- (Core) `service`: The service will no longer be able to get a ballast size from the deprecated ballast extension. ([#10696](https://github.com/open-telemetry/opentelemetry-collector/pull/10696))
  If you are already using GOMEMLIMIT instead of the ballast extension this does not affect you.
- (Contrib) `vcenterreceiver`: Enables various vCenter metrics that were disabled by default until v0.106.0 ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))
  The following metrics will be enabled by default "vcenter.datacenter.cluster.count", "vcenter.datacenter.vm.count", "vcenter.datacenter.datastore.count",
  "vcenter.datacenter.host.count", "vcenter.datacenter.disk.space", "vcenter.datacenter.cpu.limit", "vcenter.datacenter.memory.limit",
  "vcenter.resource_pool.memory.swapped", "vcenter.resource_pool.memory.ballooned", and "vcenter.resource_pool.memory.granted". The
  "resourcePoolMemoryUsageAttribute" has also been bumped up to release v.0.107.0
- (Contrib) `k8sattributesprocessor`: Deprecate `extract.annotations.regex` and `extract.labels.regex` config fields in favor of the `ExtractPatterns` function in the transform processor. The `FieldExtractConfig.Regex` parameter will be removed in version v0.111.0. ([#25128](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25128))
  Deprecating of FieldExtractConfig.Regex parameter means that it is recommended to use the `ExtractPatterns` function from the transform processor instead. To convert your current configuration please check the `ExtractPatterns` function [documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/ottl/ottlfuncs#extractpatterns). You should use the `pattern` parameter of `ExtractPatterns` instead of using the `FieldExtractConfig.Regex` parameter.

### 🚩Deprecations 🚩

- (Splunk) Deprecate the collectd/health-checker plugin ([#5167](https://github.com/signalfx/splunk-otel-collector/pull/5167))
- (Splunk) Deprecate the telegraf/exec monitor ([#5171](https://github.com/signalfx/splunk-otel-collector/pull/5171))

### 🚀 New components 🚀

- (Splunk) Add Elasticsearch receiver ([#5165](https://github.com/signalfx/splunk-otel-collector/pull/5165/))
- (Splunk) Add HAProxy receiver ([#5163](https://github.com/signalfx/splunk-otel-collector/pull/5163))

### 💡 Enhancements 💡

- (Splunk) Auto Discovery for Linux:
  - Bring Apache Web Server receiver into the discovery mode ([#5116](https://github.com/signalfx/splunk-otel-collector/pull/5116))
- (Splunk) linux installer script: decouple the endpoint and protocol options ([#5164](https://github.com/signalfx/splunk-otel-collector/pull/5164))
- (Splunk) Bump version of com.signalfx.public:signalfx-commons-protoc-java to 1.0.44 ([#5186](https://github.com/signalfx/splunk-otel-collector/pull/5186))
- (Splunk) Bump version of github.com/snowflakedb/gosnowflake from to 1.11.0 ([#5176](https://github.com/signalfx/splunk-otel-collector/pull/5176))
- (Core) `exporters`: Add data_type attribute to `otelcol_exporter_queue_size` metric to report the type of data being processed. ([#9943](https://github.com/open-telemetry/opentelemetry-collector/pull/9943))
- (Core) `confighttp`: Add option to include query params in auth context ([#4806](https://github.com/open-telemetry/opentelemetry-collector/pull/4806))
- (Core) `configgrpc`: gRPC auth errors now return gRPC status code UNAUTHENTICATED (16) ([#7646](https://github.com/open-telemetry/opentelemetry-collector/pull/7646))
- (Core) `httpprovider, httpsprovider`: Validate URIs in HTTP and HTTPS providers before fetching. ([#10468](https://github.com/open-telemetry/opentelemetry-collector/pull/10468))
- (Contrib) `processor/transform`: Add `scale_metric` function that scales all data points in a metric. ([#16214](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/16214))
- (Contrib) `vcenterreceiver`: Adds vCenter vSAN host metrics. ([#33556](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33556))
  Introduces the following vSAN host metrics to the vCenter receiver:
  - vcenter.host.vsan.throughput
  - vcenter.host.vsan.iops
  - vcenter.host.vsan.congestions
  - vcenter.host.vsan.cache.hit_rate
  - vcenter.host.vsan.latency.avg
- (Contrib) `transformprocessor`: Support aggregating metrics based on their attributes. ([#16224](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/16224))
- (Contrib) `metricstransformprocessor`: Adds the 'median' aggregation type to the Metrics Transform Processor. Also uses the refactored aggregation business logic from internal/core package. ([#16224](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/16224))
- (Contrib) `hostmetricsreceiver`: allow configuring log pipeline to send host EntityState event ([#33927](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33927))
- (Contrib) `windowsperfcountersreceiver`: Improve handling of non-existing instances for Windows Performance Counters ([#33815](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33815))
  It is an expected that when querying Windows Performance Counters the targeted instances may not be present.
  The receiver will no longer require the use of `recreate_query` to handle non-existing instances.
  As soon as the instances are available, the receiver will start collecting metrics for them.
  There won't be warning log messages when there are no matches for the configured instances.
- (Contrib) `kafkareceiver`: Add settings session_timeout and heartbeat_interval to Kafka Receiver for group management facilities ([#28630](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/28630))
- (Contrib) `vcenterreceiver`: Adds a number of default disabled vSAN metrics for Clusters. ([#33556](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33556))
- (Contrib) `vcenterreceiver`: Adds a number of default disabled vSAN metrics for Virtual Machines. ([#33556](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33556))

### 🧰 Bug fixes 🧰

- (Core) `processorhelper`: update units for internal telemetry ([#10647](https://github.com/open-telemetry/opentelemetry-collector/pull/10647))
- (Core) `confmap`: Increase the amount of recursion and URI expansions allowed in a single line ([#10712](https://github.com/open-telemetry/opentelemetry-collector/pull/10712))
- (Core) `exporters`: There is no guarantee that after the exporters sends the plog/pmetric/ptrace data downstream that the data won't be mutated in some way. (e.g by the batch_sender) This mutation could result in the proceeding call to req.ItemsCount() to provide inaccurate information to be logged. ([#10033](https://github.com/open-telemetry/opentelemetry-collector/pull/10033))
- (Core) `exporters`: Update units for internal telemetry ([#10648](https://github.com/open-telemetry/opentelemetry-collector/pull/10648))
- (Core) `receiverhelper`: Update units for internal telemetry ([#10650](https://github.com/open-telemetry/opentelemetry-collector/pull/10650))
- (Core) `scraperhelper`: Update units for internal telemetry ([#10649](https://github.com/open-telemetry/opentelemetry-collector/pull/10649))
- (Core) `service`: Use Command/Version to populate service name/version attributes ([#10644](https://github.com/open-telemetry/opentelemetry-collector/pull/10644))
- (Core) `configauth`: Fix unmarshaling of authentication in HTTP servers. ([#10750](https://github.com/open-telemetry/opentelemetry-collector/pull/10750))
- (Contrib) `opencensusreceiver`: Do not report an error into resource status during receiver shutdown when the listener connection was closed. ([#33865](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33865))
- (Contrib) `statsdeceiver`: Log only non-EOF errors when reading payload received via TCP. ([#33951](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33951))
- (Contrib) `vcenterreceiver`: Adds destroys to the ContainerViews in the client. ([#34254](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34254))
  This may not be necessary, but it should be better practice than not.

## v0.105.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.105.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.105.0) and the [opentelemetry-collector-contrib v0.105.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.105.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Don't expand environment variables starting with $$ in configuration files. This behavior was introduced
  in v0.42.0 to support a bug causing double expansion. $$ is treated as an escape sequence representing a literal
  $ character ([#5134](https://github.com/signalfx/splunk-otel-collector/pull/5134))
- (Core) `service`: add `service.disableOpenCensusBridge` feature gate which is enabled by default to remove the dependency on OpenCensus ([#10414](https://github.com/open-telemetry/opentelemetry-collector/pull/10414))
- (Core) `confmap`: Promote `confmap.strictlyTypedInput` feature gate to beta. ([#10552](https://github.com/open-telemetry/opentelemetry-collector/pull/10552))
  This feature gate changes the following:
  - Configurations relying on the implicit type casting behaviors listed on [#9532](https://github.com/open-telemetry/opentelemetry-collector/issues/9532) will start to fail.
  - Configurations using URI expansion (i.e. `field: ${env:ENV}`) for string-typed fields will use the value passed in `ENV` verbatim without intermediate type casting.
- (Contrib) `stanza`: errors from Operator.Process are returned instead of silently ignored. ([#33783](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33783))
  This public function is affected: https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza@v0.104.0/operator/helper#WriterOperator.Write
- (Contrib) `vcenterreceiver`: Enables various vCenter metrics that were disabled by default until v0.105 ([#34022](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34022))
  The following metrics will be enabled by default "vcenter.host.network.packet.drop.rate",
  "vcenter.vm.cpu.readiness", "vcenter.host.cpu.capacity", and "vcenter.host.cpu.reserved".

### 🚩Deprecations 🚩

- (Splunk) Deprecate usage of bare environment variables and config sources in configuration files ([#5153](https://github.com/signalfx/splunk-otel-collector/pull/5153))
  - Use `${env:VAR}` or `${VAR}` instead of `$VAR`.
  - Use `${uri:selector}` instead of `$uri:selector`, e.g. `${file:/path/to/file}` instead of `$file:/path/to/file`.

### 💡 Enhancements 💡
- (Splunk) Auto Discovery for Linux:
  - Bring SQL Server receiver into the discovery mode ([#5109](https://github.com/signalfx/splunk-otel-collector/pull/5109))
  - Bring Cassanda JMX receiver into the discovery mode ([#5112](https://github.com/signalfx/splunk-otel-collector/pull/5112))
  - Bring RabbitMQ receiver into the discovery mode ([#5051](https://github.com/signalfx/splunk-otel-collector/pull/5051))
- (Splunk) Update bundled OpenJDK to [11.0.24_8](https://github.com/adoptium/temurin11-binaries/releases/tag/jdk-11.0.24%2B8) ([#5113](https://github.com/signalfx/splunk-otel-collector/pull/5113), [#5119](https://github.com/signalfx/splunk-otel-collector/pull/5119))
- (Splunk) Upgrade github.com/hashicorp/vault to v1.17.2 ([#5089](https://github.com/signalfx/splunk-otel-collector/pull/5089))
- (Splunk) Upgrade github.com/go-zookeeper/zk to 1.0.4 ([#5146](https://github.com/signalfx/splunk-otel-collector/pull/5146))
- (Core) `configtls`: Mark module as stable. ([#9377](https://github.com/open-telemetry/opentelemetry-collector/pull/9377))
- (Core) `confmap`: Remove extra closing parenthesis in sub-config error ([#10480](https://github.com/open-telemetry/opentelemetry-collector/pull/10480))
- (Core) `configgrpc`: Update the default load balancer strategy to round_robin ([#10319](https://github.com/open-telemetry/opentelemetry-collector/pull/10319))
  To restore the behavior that was previously the default, set `balancer_name` to `pick_first`.
- (Core) `otelcol`: Add go module to components subcommand. ([#10570](https://github.com/open-telemetry/opentelemetry-collector/pull/10570))
- (Core) `confmap`: Add explanation to errors related to `confmap.strictlyTypedInput` feature gate. ([#9532](https://github.com/open-telemetry/opentelemetry-collector/pull/9532))
- (Core) `confmap`: Allow using `map[string]any` values in string interpolation ([#10605](https://github.com/open-telemetry/opentelemetry-collector/pull/10605))
- (Contrib) `pkg/ottl`: Added Hex() converter function ([#31929](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31929))
- (Contrib) `pkg/ottl`: Add IsRootSpan() converter function. ([#32918](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32918))
  Converter `IsRootSpan()` returns `true` if the span in the corresponding context is root, that means its `parent_span_id` equals to hexadecimal representation of zero. In all other scenarios function returns `false`.
- (Contrib) `vcenterreceiver`: Adds additional vCenter resource pool metrics and a memory_usage_type attribute for vcenter.resource_pool.memory.usage metric to use. ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))
  Added "vcenter.resource_pool.memory.swapped", "vcenter.resource_pool.memory.ballooned", and "vcenter.resource_pool.memory.granted"
  metrics. Also added an additional attribute, "memory_usage_type" for "vcenter.resource_pool.memory.usage" metric, which is currently under a feature gate.
- (Contrib) `kubeletstatsreceiver`: Add `k8s.pod.memory.node.utilization` and `k8s.container.memory.node.utilization` metrics ([#33591](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33591))
- (Contrib) `vcenterreceiver`: Adds vCenter metrics at the datacenter level. ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))
  Introduces various datacenter metrics which work by aggregating stats from datastores, clusters, hosts, and VM's.
- (Contrib) `processor/resource, processor/attributes`: Add an option to extract value from a client address by specifying `client.address` value in the `from_context` field. (#34051) ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))
- (Contrib) `receiver/azuremonitorreceiver`: Add support for Managed Identity and Default Credential auth ([#31268](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31268), [#33584](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33584))
- (Contrib) `azuremonitorreceiver`: Add `maximum_number_of_records_per_resource` config parameter in order to overwrite default ([#32165](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32165))
- (Contrib) `cloudfoundryreceiver`: Add support to receive CloudFoundry Logs ([#32671](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32671))
- (Contrib) `splunkhecexporter`: Increase the performance of JSON marshaling ([#34011](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34011))
- (Contrib) `loadbalancingexporter`: Adds a new streamID routingKey, which will route based on the datapoint ID. See updated README for details ([#32513](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32513))
- (Contrib) `dockerobserver`: Add hint to error when using float for `api_version` field ([#34043](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34043))
- (Contrib) `pkg/ottl`: Emit traces for statement sequence executions to troubleshoot OTTL statements/conditions ([#33433](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33433))
- (Contrib) `pkg/stanza`: Bump 'logs.jsonParserArray' and 'logs.assignKeys' feature gates to beta. ([#33948](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33948))
  - This enables the feature gates by default to allow use of the `json_array_parser` and `assign_keys` operations.
- (Contrib) `receiver/filelog`: Add filelog.container.removeOriginalTimeField feature-flag for removing original time field ([#33946](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33946))
- (Contrib) `statsdreceiver`: Allow configuring summary percentiles ([#33701](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33701))
- (Contrib) `pkg/stanza`: Switch to faster json parser lib for container operator ([#33929](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33929))
- (Contrib) `telemetrygen`: telemetrygen `--rate` flag changed from Int64 to Float64 ([#33984](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33984))
- (Contrib) `windowsperfcountersreceiver`: `windowsperfcountersreceiver` now appends an index number to additional instance names that share a name. An example of this is when scraping `process(*)` counters with multiple running instances of the same executable. ([#32319](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32319))
  **NOTES**
  - This change can expose cardinality issues where the counters were previously collapsed under the non-indexed instance name.
  - The change mimics Windows Performance Monitor behavior: The first instance name remains unchanged, additional instances are suffixed with `#<N>` where `N=1` and is increased for each duplicate.
    - e.g. Given 3 powershell instances, this will return `powershell`, `powershell#1` and `powershell#2`.

### 🧰 Bug fixes 🧰
- (Splunk) Auto Discovery for Linux:
  - Fix kafkametrics k8s issues for Auto Discovery ([#5103](https://github.com/signalfx/splunk-otel-collector/pull/5103))
  - Reuse discovery receiver's obsreport for receivercreator ([#5111](https://github.com/signalfx/splunk-otel-collector/pull/5111))
- (Core) `confmap`: Fixes issue where confmap could not escape `$$` when `confmap.unifyEnvVarExpansion` is enabled. ([#10560](https://github.com/open-telemetry/opentelemetry-collector/pull/10560))
- (Core) `otlpreceiver`: Fixes a bug where the otlp receiver's http response was not properly translating grpc error codes to http status codes. ([#10574](https://github.com/open-telemetry/opentelemetry-collector/pull/10444))
- (Core) `exporters`: Fix incorrect deduplication of otelcol_exporter_queue_size and otelcol_exporter_queue_capacity metrics if multiple exporters are used. ([#10444](https://github.com/open-telemetry/opentelemetry-collector/pull/10226))
- (Core) `service/telemetry`: Add ability to set service.name for spans emitted by the Collector ([#10489](https://github.com/open-telemetry/opentelemetry-collector/pull/10489))
- (Core) `internal/localhostgate`: Correctly log info message when `component.UseLocalHostAsDefaultHost` is enabled ([#8510](https://github.com/open-telemetry/opentelemetry-collector/pull/8510))
- (Contrib) `countconnector`: Updating the stability to reflect that the component is shipped as part of contrib. ([#33903](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33903))
- (Contrib) `httpcheckreceiver`: Updating the stability to reflect that the component is shipped as part of contrib. ([#33897](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33897))
- (Contrib) `probabilisticsamplerprocessor`: Fix bug where log sampling was being reported by the counter `otelcol_processor_probabilistic_sampler_count_traces_sampled` ([#33874](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33874))
- (Contrib) `processor/groupbyattrsprocessor`: Fix dropping of metadata fields when processing metrics. ([#33419](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33419))
- (Contrib) `prometheusreceiver`: Fix hash computation to include non exported fields like regex in scrape configuration for TargetAllocator ([#29313](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29313))
- (Contrib) `kafkametricsreceiver`: Fix issue with incorrect consumer offset ([#33309](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33309))
- (Contrib) `sqlserverreceiver`: Enable default metrics to properly trigger SQL Server scrape ([#34065](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/34065))
- (Contrib) `syslogreceiver`: Allow to define `max_octets` for octet counting RFC5424 syslog parser ([#33182](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33182))
- (Contrib) `windowsperfcountersreceiver`: Metric definitions with no matching performance counter are no longer included as metrics with zero datapoints in the scrape output. ([#4972](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/4972))

## v0.104.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.104.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.104.0) and the [opentelemetry-collector-contrib v0.104.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.104.0) releases where appropriate.

> :warning: In our efforts to align with the [goals](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/env-vars.md) defined upstream for environment variable resolution in the Collector's configuration, the Splunk OpenTelemetry Collector will be dropping support for expansion of BASH-style environment variables, such as `$FOO` in the configuration in an upcoming version. Users are advised to update their Collector's configuration to use `${env:FOO}` instead.

> 🚩 When setting properties for discovery receiver as environment variables (`SPLUNK_DISCOVERY_*`), the values cannot reference other environment variables without curly-braces. For example, user is trying to set discovery property `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED` to the value of another env var, `K8S_ENVIRONMENT`.
> For versions older than 0.104.0, setting it as `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=\$K8S_ENVIRONMENT` (note the escaped variable name does not have curly braces) was valid. But from v0.104.0, env var names need to be passed with braces. For this example, user should modify it to `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=\${K8S_ENVIRONMENT}`.

### ❗ Known Issues ❗

- A bug was discovered (and fixed in a future version) where expansion logic in confmaps wasn't correctly handling the escaping of $$ ([#10560](https://github.com/open-telemetry/opentelemetry-collector/pull/10560))
  - If you rely on the previous functionality, disable the `confmap.unifyEnvVarExpansion` feature gate. Note that this is a temporary workaround, and the root issue will be fixed in the next release by ([#10560](https://github.com/open-telemetry/opentelemetry-collector/pull/10560)).

### 🛑 Breaking changes 🛑

- (Splunk) Auto Discovery for Linux:
  - Update `splunk-otel-java` to v2.5.0 for the `splunk-otel-auto-instrumentation` deb/rpm packages. This is a major version bump that includes breaking changes. Check the [release notes](https://github.com/signalfx/splunk-otel-java/releases/tag/v2.5.0) for details about breaking changes.

- (Core) `filter`: Remove deprecated `filter.CombinedFilter` ([#10348](https://github.com/open-telemetry/opentelemetry-collector/pull/10348))

- (Core) `otelcol`: By default, `otelcol.NewCommand` and `otelcol.NewCommandMustSetProvider` will set the `DefaultScheme` to `env`. ([#10435](https://github.com/open-telemetry/opentelemetry-collector/pull/10435))

- (Core) `expandconverter`: By default expandconverter will now error if it is about to expand `$FOO` syntax. Update configuration to use `${env:FOO}` instead or disable the `confmap.unifyEnvVarExpansion` feature gate. ([#10435](https://github.com/open-telemetry/opentelemetry-collector/pull/10435))

- (Core) `otlpreceiver`: Switch to `localhost` as the default for all endpoints. ([#8510](https://github.com/open-telemetry/opentelemetry-collector/pull/8510))
  Disable the `component.UseLocalHostAsDefaultHost` feature gate to temporarily get the previous default.

- (Splunk) `discovery`: When setting properties for discovery receiver as environment variables (`SPLUNK_DISCOVERY_*`), the values cannot reference other escaped environment variables without braces. For example, when trying to set discovery property `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED` to the value of another env var, `K8S_ENVIRONMENT`. For versions older than 0.104.0, setting it as `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=\$K8S_ENVIRONMENT` (note the escaped variable name does not have braces) was valid. But from v0.104.0, env var names need to be passed with braces. For this example, user should modify it to `SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=\${K8S_ENVIRONMENT}`

- (Contrib) `vcenterreceiver`: Drops support for vCenter 6.7 ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))

- (Contrib) `all`: Promote `component.UseLocalHostAsDefaultHost` feature gate to beta. This changes default endpoints from 0.0.0.0 to localhost ([#30702](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30702))
  This change affects the following components:
    - extension/health_check
    - receiver/jaeger
    - receiver/sapm
    - receiver/signalfx
    - receiver/splunk_hec
    - receiver/zipkin

- (Contrib) `receiver/mongodb`: Graduate receiver.mongodb.removeDatabaseAttr feature gate to stable ([#24972](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24972))

### 💡 Enhancements 💡

- (Splunk) Auto Discovery for Linux:
  - Linux installer script:
    - The default for the `--otlp-endpoint` option is now empty, i.e. defers to the default `OTEL_EXPORTER_OTLP_ENDPOINT` value for each activated SDK
    - Add new `--otlp-endpoint-protocol <protocol>` option to set the `OTEL_EXPORTER_OTLP_PROTOCOL` environment variable for the configured endpoint. Only applicable if the `--otlp-endpoint` option is also specified.
    - Add new `--metrics-exporter <exporter>` option to configure the `OTEL_METRICS_EXPORTER` environment variable for instrumentation metrics. Specify `none` to disable metric collection and export.
- (Splunk) Set Go garbage collection target percentage to 400% ([#5034](https://github.com/signalfx/splunk-otel-collector/pull/5034))
  After removal of memory_ballast extension in v0.97.0, the Go garbage collection is running more aggressively, which
  increased CPU usage and leads to reduced throughput of the collector. This change reduces the frequency of garbage
  collection cycles to improves performance of the collector for typical workloads. As a result, the collector will
  report higher memory usage, but it will be bound to the same configured limits. If you want to revert to the previous
  behavior, set the `GOGC` environment variable to `100`.
- (Splunk) Upgrade to golang 1.21.12 ([#5074](https://github.com/signalfx/splunk-otel-collector/pull/5074))
- (Core) `confighttp`: Add support for cookies in HTTP clients with `cookies::enabled`. ([#10175](https://github.com/open-telemetry/opentelemetry-collector/pull/10175))
  The method `confighttp.ToClient` will return a client with a `cookiejar.Jar` which will reuse cookies from server responses in subsequent requests.
- (Core) `exporter/debug`: In `normal` verbosity, display one line of text for each telemetry record (log, data point, span) ([#7806](https://github.com/open-telemetry/opentelemetry-collector/pull/7806))
- (Core) `exporter/debug`: Add option `use_internal_logger` ([#10226](https://github.com/open-telemetry/opentelemetry-collector/pull/10226))
- (Core) `configretry`: Mark module as stable. ([#10279](https://github.com/open-telemetry/opentelemetry-collector/pull/10279))
- (Core) `exporter/debug`: Print Span.TraceState() when present. ([#10421](https://github.com/open-telemetry/opentelemetry-collector/pull/10421))
  Enables viewing sampling threshold information (as by OTEP 235 samplers).
- (Core) `processorhelper`: Add \"inserted\" metrics for processors. ([#10353](https://github.com/open-telemetry/opentelemetry-collector/pull/10353))
  This includes the following metrics for processors:
  - `processor_inserted_spans`
  - `processor_inserted_metric_points`
  - `processor_inserted_log_records`
- (Contrib) `k8sattributesprocessor`: Add support for exposing `k8s.pod.ip` as a resource attribute ([#32960](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32960))
- (Contrib) `vcenterreceiver`: Adds vCenter CPU readiness metric for VMs. ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))
- (Contrib) `receiver/mongodb`: Ensure support of 6.0 and 7.0 MongoDB versions with integration tests ([#32716](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32716))
- (Contrib) `pkg/stanza`: Switch JSON parser used by json_parser to github.com/goccy/go-json ([#33784](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33784))
- (Contrib) `k8sobserver`: Add support for k8s.ingress endpoint. ([#32971](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32971))
- (Contrib) `statsdreceiver`: Optimize statsdreceiver to reduce object allocations ([#33683](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33683))
- (Contrib) `routingprocessor`: Use mdatagen to define the component's telemetry ([#33526](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33526))
- (Contrib) `receiver/mongodbreceiver`: Add `server.address` and `server.port` resource attributes to MongoDB receiver. ([#32810](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32810),[#32350](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32350))
  The new resource attributes are added to the MongoDB receiver to distinguish metrics coming from different MongoDB instances.
    - `server.address`: The address of the MongoDB host, enabled by default.
    - `server.port`: The port of the MongoDB host, disabled by default.

- (Contrib) `observerextension`: Expose host and port in endpoint's environment ([#33571](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33571))
- (Contrib) `pkg/ottl`: Add a `schema_url` field to access the SchemaURL in resources and scopes on all signals ([#30229](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30229))
- (Contrib) `sqlserverreceiver`: Enable more perf counter metrics when directly connecting to SQL Server ([#33420](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33420))
  This enables the following metrics by default on non Windows-based systems:
  `sqlserver.batch.request.rate`
  `sqlserver.batch.sql_compilation.rate`
  `sqlserver.batch.sql_recompilation.rate`
  `sqlserver.page.buffer_cache.hit_ratio`
  `sqlserver.user.connection.count`
- (Contrib) `vcenterreceiver`: Adds vCenter CPU capacity and network drop rate metrics to hosts. ([#33607](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33607))

### 🧰 Bug fixes 🧰

- (Splunk) `receiver/discovery`: Do not emit entity events for discovered endpoints that are not evaluated yet
  to avoid showing "unknown" services on the Service Inventory page ([#5032](https://github.com/open-telemetry/opentelemetry-collector/pull/5032))
- (Core) `otlpexporter`: Update validation to support both dns:// and dns:/// ([#10449](https://github.com/open-telemetry/opentelemetry-collector/pull/10449))
- (Core) `service`: Fixed a bug that caused otel-collector to fail to start with ipv6 metrics endpoint service telemetry. ([#10011](https://github.com/open-telemetry/opentelemetry-collector/pull/10011))
- (Contrib) `resourcedetectionprocessor`: Fetch CPU info only if related attributes are enabled ([#33774](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33774))
- (Contrib) `tailsamplingprocessor`: Fix precedence of inverted match in and policy ([#33671](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33671))
  Previously if the decision from a policy evaluation was `NotSampled` or `InvertNotSampled` it would return a `NotSampled` decision regardless, effectively downgrading the result.
  This was breaking the documented behaviour that inverted decisions should take precedence over all others.
- (Contrib) `vcenterreceiver`: Fixes errors in some of the client calls for environments containing multiple datacenters. ([#33734](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33734))


## v0.103.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.103.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.103.0) and the [opentelemetry-collector-contrib v0.103.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.103.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `exporter/debug`: Disable sampling by default ([#9921](https://github.com/open-telemetry/opentelemetry-collector/pull/9921))
  To restore the behavior that was previously the default, set `sampling_thereafter` to `500`.
- (Contrib) `mongodbreceiver`: Now only supports `TCP` connections ([#32199](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32199))
  This fixes a bug where hosts had to explicitly set `tcp` as the transport type. The `transport` option has been removed.
- (Contrib) `sqlserverreceiver`: sqlserver.database.io.read_latency has been renamed to sqlserver.database.latency with a `direction` attribute. ([#29865](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29865))

### 🚀 New components 🚀

- (Splunk) Add Azure Monitor receiver ([#4971](https://github.com/signalfx/splunk-otel-collector/pull/4971))
- (Splunk) Add [upstream](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/rabbitmqreceiver) Opentelemetry Collector RabbitMQ receiver ([#4980](https://github.com/signalfx/splunk-otel-collector/pull/4980))
- (Splunk) Add Active Directory Domain Services receiver ([#4994](https://github.com/signalfx/splunk-otel-collector/pull/4994))
- (Splunk) Add Splunk Enterprise receiver ([#4998](https://github.com/signalfx/splunk-otel-collector/pull/4998))

### 💡 Enhancements 💡

- (Core) `otelcol/expandconverter`: Add `confmap.unifyEnvVarExpansion` feature gate to allow enabling Collector/Configuration SIG environment variable expansion rules. ([#10391](https://github.com/open-telemetry/opentelemetry-collector/pull/10391))
  When enabled, this feature gate will:
  - Disable expansion of BASH-style env vars (`$FOO`)
  - `${FOO}` will be expanded as if it was `${env:FOO}`
    See https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/env-vars.md for more details.

- (Core) `confmap`: Add `confmap.unifyEnvVarExpansion` feature gate to allow enabling Collector/Configuration SIG environment variable expansion rules. ([#10259](https://github.com/open-telemetry/opentelemetry-collector/pull/10259))
  When enabled, this feature gate will:
  - Disable expansion of BASH-style env vars (`$FOO`)
  - `${FOO}` will be expanded as if it was `${env:FOO}`
    See https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/rfcs/env-vars.md for more details.

- (Core) `confighttp`: Allow the compression list to be overridden ([#10295](https://github.com/open-telemetry/opentelemetry-collector/pull/10295))
  Allows Collector administrators to control which compression algorithms to enable for HTTP-based receivers.
- (Core) `configgrpc`: Revert the zstd compression for gRPC to the third-party library we were using previously. ([#10394](https://github.com/open-telemetry/opentelemetry-collector/pull/10394))
  We switched back to our compression logic for zstd when a CVE was found on the third-party library we were using. Now that the third-party library has been fixed, we can revert to that one. For end-users, this has no practical effect. The reproducers for the CVE were tested against this patch, confirming we are not reintroducing the bugs.
- (Core) `confmap`: Adds alpha `confmap.strictlyTypedInput` feature gate that enables strict type checks during configuration resolution ([#9532](https://github.com/open-telemetry/opentelemetry-collector/pull/9532))
  When enabled, the configuration resolution system will:
  - Stop doing most kinds of implicit type casting when resolving configuration values
  - Use the original string representation of configuration values if the ${} syntax is used in inline position

- (Core) `confighttp`: Use `confighttp.ServerConfig` as part of zpagesextension. See [server configuration](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/README.md#server-configuration) options. ([#9368](https://github.com/open-telemetry/opentelemetry-collector/pull/9368))

- (Contrib) `filelogreceiver`: If include_file_record_number is true, it will add the file record number as the attribute `log.file.record_number` ([#33530](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33530))
- (Contrib) `filelogreceiver`: Add support for gzip compressed log files ([#2328](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/2328))
- (Contrib) `kubeletstats`: Add k8s.pod.cpu.node.utilization metric ([#33390](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33390))
- (Contrib) `awss3exporter`: endpoint should contain the S3 bucket ([#32774](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32774))
- (Contrib) `statsdreceiver`: update statsd receiver to use mdatagen ([#33524](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33524))
- (Contrib) `statsdreceiver`: Added received/accepted/refused metrics ([#24278](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24278))
- (Contrib) `metricstransformprocessor`: Adds the 'count' aggregation type to the Metrics Transform Processor. ([#24978](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24978))
- (Contrib) `tailsamplingprocessor`: Simple LRU Decision Cache for "keep" decisions ([#31583](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31583))
- (Contrib) `tailsamplingprocessor`: Migrates internal telemetry to OpenTelemetry SDK via mdatagen ([#31581](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31581))
  The metric names and their properties, such as bucket boundaries for histograms, were kept like before, to keep backwards compatibility.
- (Contrib) `kafka`: Added `disable_fast_negotiation` configuration option for Kafka Kerberos authentication, allowing the disabling of PA-FX-FAST negotiation. ([#26345](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26345))
- (Contrib) `pkg/ottl`: Added `keep_matching_keys` function to allow dropping all keys from a map that don't match the pattern. ([#32989](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32989))
- (Contrib) `pkg/ottl`: Add debug logs to help troubleshoot OTTL statements/conditions ([#33274](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33274))
- (Contrib) `pkg/ottl`: Introducing `append` function for appending items into an existing array ([#32141](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32141))
- (Contrib) `pkg/ottl`: Introducing `Uri` converter parsing URI string into SemConv ([#32433](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32433))
- (Contrib) `probabilisticsamplerprocessor`: Add Proportional and Equalizing sampling modes ([#31918](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31918))
  Both the existing hash_seed mode and the two new modes use OTEP 235 semantic conventions to encode sampling probability.
- (Contrib) `prometheusreceiver`: Resource attributes produced by the prometheus receiver now include stable semantic conventions for `server` and `url`. ([#32814](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32814))
  To migrate from the legacy net.host.name, net.host.port, and http.scheme resource attributes, migrate to server.address, server.port, and url.scheme, and then set the receiver.prometheus.removeLegacyResourceAttributes feature gate.

- (Contrib) `spanmetricsconnector`: Produce delta temporality span metrics with StartTimeUnixNano and TimeUnixNano values representing an uninterrupted series ([#31671](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31671), [#30688](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30688))
  This allows producing delta span metrics instead of the more memory-intensive cumulative metrics, specifically when a downstream component can convert the delta metrics to cumulative.
- (Contrib) `sqlserverreceiver`: Add support for more Database IO metrics ([#29865](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29865))
  The following metrics have been added:
  - sqlserver.database.latency
  - sqlserver.database.io
  - sqlserver.database.operations

- (Contrib) `processor/transform`: Add `transform.flatten.logs` featuregate to give each log record a distinct resource and scope. ([#32080](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32080))
  This option is useful when applying transformations which alter the resource or scope. e.g. `set(resource.attributes["to"], attributes["from"])`, which may otherwise result in unexpected behavior. Using this option typically incurs a performance penalty as the processor must compute many hashes and create copies of resource and scope information for every log record.

- (Contrib) `receiver/windowsperfcounters`: Counter configuration now supports recreating the underlying performance query at scrape time. ([#32798](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32798))

### 🧰 Bug fixes 🧰

- (Core) `exporters`: Fix potential deadlock in the batch sender ([#10315](https://github.com/open-telemetry/opentelemetry-collector/pull/10315))
- (Core) `expandconverter`: Fix bug where an warning was logged incorrectly. ([#10392](https://github.com/open-telemetry/opentelemetry-collector/pull/10392))
- (Core) `exporters`: Fix a bug when the retry and timeout logic was not applied with enabled batching. ([#10166](https://github.com/open-telemetry/opentelemetry-collector/pull/10166))
- (Core) `exporters`: Fix a bug where an unstarted batch_sender exporter hangs on shutdown ([#10306](https://github.com/open-telemetry/opentelemetry-collector/pull/10306))
- (Core) `exporters`: Fix small batch due to unfavorable goroutine scheduling in batch sender ([#9952](https://github.com/open-telemetry/opentelemetry-collector/pull/9952))
- (Core) `confmap`: Fix issue where structs with only yaml tags were not marshaled correctly. ([#10282](https://github.com/open-telemetry/opentelemetry-collector/pull/10282))

- (Contrib) `filelogreceiver`: Container parser should add k8s metadata as resource attributes and not as log record attributes ([#33341](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33341))
- (Contrib) `postgresqlreceiver`: Fix bug where `postgresql.rows` always returning 0 for `state="dead"` ([#33489](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33489))
- (Contrib) `prometheusreceiver`: Fall back to scrape config job/instance labels for aggregated metrics without instance/job labels ([#32555](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32555))

## v0.102.1

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.102.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.102.1) and the [opentelemetry-collector-contrib v0.102.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.102.0) releases where appropriate.

### 🧰 Bug fixes 🧰

- (Core) `configrpc`: **This release addresses [GHSA-c74f-6mfw-mm4v](https://github.com/open-telemetry/opentelemetry-collector/security/advisories/GHSA-c74f-6mfw-mm4v) for `configgrpc`.** ([#10323](https://github.com/open-telemetry/opentelemetry-collector/issues/10323))
Before this change, the zstd compressor that was used didn't respect the max message size. This addresses `GHSA-c74f-6mfw-mm4v` on configgrpc.

### 💡 Enhancements 💡

- (Splunk) Upgrade golang to 1.21.11

## v0.102.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.102.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.102.0) and the [opentelemetry-collector-contrib v0.102.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.102.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) `receiver/discovery`: Replace `log_record` field with `message` in evaluation statements ([#4583](https://github.com/signalfx/splunk-otel-collector/pull/4583))
- (Core) `envprovider`: Restricts Environment Variable names.  Environment variable names must now be ASCII only and start with a letter or an underscore, and can only contain underscores, letters, or numbers. ([#9531](https://github.com/open-telemetry/opentelemetry-collector/issues/9531))
- (Core) `confighttp`: Apply MaxRequestBodySize to the result of a decompressed body [#10289](https://github.com/open-telemetry/opentelemetry-collector/pull/10289)
  When using compressed payloads, the Collector would verify only the size of the compressed payload.
  This change applies the same restriction to the decompressed content. As a security measure, a limit of 20 MiB was added, which makes this a breaking change.
  For most clients, this shouldn't be a problem, but if you often have payloads that decompress to more than 20 MiB, you might want to either configure your
  client to send smaller batches (recommended), or increase the limit using the MaxRequestBodySize option.
- (Contrib) `k8sattributesprocessor`: Move `k8sattr.rfc3339` feature gate to stable. ([#33304](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33304))
- (Contrib) `extension/filestorage`: Replace path-unsafe characters in component names ([#3148](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3148))
  The feature gate `extension.filestorage.replaceUnsafeCharacters` is now removed.
- (Contrib) `vcenterreceiver`: vcenterreceiver replaces deprecated packet metrics by removing them and enabling by default the newer ones. (([#32929](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32929)),([#32835](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32835))
  Removes the following metrics: `vcenter.host.network.packet.errors`, `vcenter.host.network.packet.count`, and
  `vcenter.vm.network.packet.count`.
  Also enables by default the following metrics: `vcenter.host.network.packet.error.rate`,
  `vcenter.host.network.packet.rate`, and `vcenter.vm.network.packet.rate`.

### 🧰 Bug fixes 🧰

- (Splunk) `discovery`: Fix crashing collector if discovered mongodb isn't reachable in Kubernetes ([#4911](https://github.com/signalfx/splunk-otel-collector/pull/4911))
- (Core) `batchprocessor`: ensure attributes are set on cardinality metadata metric [#9674](https://github.com/open-telemetry/opentelemetry-collector/pull/9674)
- (Core) `batchprocessor`: Fixing processor_batch_metadata_cardinality which was broken in v0.101.0 [#10231](https://github.com/open-telemetry/opentelemetry-collector/pull/10231)
- (Core) `batchprocessor`: respect telemetry level for all metrics [#10234](https://github.com/open-telemetry/opentelemetry-collector/pull/10234)
- (Core) `exporters`: Fix potential deadlocks in BatcherSender shutdown [#10255](https://github.com/open-telemetry/opentelemetry-collector/pull/10255)
- (Contrib) `receiver/mysql`: Remove the order by clause for the column that does not exist ([#33271](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33271))
- (Contrib) `kafkareceiver`: Fix bug that was blocking shutdown ([#30789](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30789))

### 🚩 Deprecations 🚩

- (Splunk) The following docker images/manifests are deprecated and may not be published in a future release:
  - `quay.io/signalfx/splunk-otel-collector:<version>-amd64`
  - `quay.io/signalfx/splunk-otel-collector:<version>-arm64`
  - `quay.io/signalfx/splunk-otel-collector:<version>-ppc64le`
  - `quay.io/signalfx/splunk-otel-collector-windows:<version>`
  - `quay.io/signalfx/splunk-otel-collector-windows:<version>-2019`
  - `quay.io/signalfx/splunk-otel-collector-windows:<version>-2022`

  Starting with this release, the `quay.io/signalfx/splunk-otel-collector:<version>` docker image manifest has been
  updated to support Windows (2019 amd64, 2022 amd64), in addition to Linux (amd64, arm64, ppc64le).

  Please update any configurations to use `quay.io/signalfx/splunk-otel-collector:<version>` for this and future releases.

### 💡 Enhancements 💡

- (Splunk) `discovery`: Update redis discovery instructions ([#4915](https://github.com/signalfx/splunk-otel-collector/pull/4915))
- (Splunk) `discovery`: Bring Kafkamatrics receiver into the discovery mode ([#4903](https://github.com/signalfx/splunk-otel-collector/pull/4903))
- (Contrib) `pkg/ottl`: Add the `Day` Converter to extract the int Day component from a time.Time ([#33106](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33106))
- (Contrib) `pkg/ottl`: Adds `Month` converter to extract the int Month component from a time.Time (#33106) ([#33106](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33106))
- (Contrib) `pkg/ottl`: Adds a `Year` converter for extracting the int year component from a time.Time ([#33106](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33106))
- (Contrib) `filelogreceiver`: Log when files are rotated/moved/truncated ([#33237](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33237))
- (Contrib) `stanza`: Add monitoring metrics for open and harvested files in fileconsumer ([#31256](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31256))
- (Contrib) `prometheusreceiver`: Allow to configure http client used by target allocator generated scrape targets ([#18054](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/18054))
- (Contrib) `pkg/stanza`: Expose recombine max log size option in the container parser configuration ([#33186](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33186))
- (Contrib) `processor/resourcedetectionprocessor`: Add support for Azure tags in ResourceDetectionProcessor. ([#32953](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32953))
- (Contrib) `kubeletstatsreceiver`: Add k8s.container.cpu.node.utilization metric ([#27885](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27885))
- (Contrib) `pkg/ottl`: Adds a `Minute` converter for extracting the int minute component from a time.Time ([#33106](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33106))

## v0.101.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.101.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.101.0) and the [opentelemetry-collector-contrib v0.101.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.101.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) `receiver/discovery`: Remove `append_pattern` option from log evaluation statements ([#4583](https://github.com/signalfx/splunk-otel-collector/pull/4583))
  - The matched log message is now set as `discovery.matched_log` entity attributes instead of being appended to
    the `discovery.message` attribute.
  - The matched log fields like `caller` and `stacktrace` are not sent as attributes anymore.
- (Contrib) `vcenterreceiver`: Removes vcenter.cluster.name attribute from vcenter.datastore metrics ([#32674](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32674))
  If there were multiple Clusters, Datastore metrics were being repeated under Resources differentiated with a
  vcenter.cluster.name resource attribute. In the same vein, if there were standalone Hosts, in addition to
  clusters the metrics would be repeated under a Resource without the vcenter.cluster.name attribute. Now there
  will only be a single set of metrics for one Datastore (as there should be, as Datastores don't belong to
  Clusters).
- (Contrib) `resourcedetectionprocessor`: Move `processor.resourcedetection.hostCPUModelAndFamilyAsString` feature gate to stable. ([#29025](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29025))
- (Contrib) `filelog`, `journald`, `tcp`, `udp`, `syslog`, `windowseventlog` receivers: The internal logger has been changed from `zap.SugaredLogger` to `zap.Logger`. ([#32177](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32177))
  This should not have any meaningful impact on most users but the logging format for some logs may have changed.


### 🚀 New components 🚀

- (Splunk) Add HTTP check receiver ([#4843](https://github.com/signalfx/splunk-otel-collector/pull/4843))
- (Splunk) Add OAuth2 Client extension ([#4843](https://github.com/signalfx/splunk-otel-collector/pull/4843))

### 💡 Enhancements 💡

- (Splunk) [`splunk-otel-collector` Salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt): Initial support for
  Splunk OpenTelemetry [Node.js](https://github.com/signalfx/splunk-otel-js) and [.NET](https://github.com/signalfx/splunk-otel-dotnet) Auto Instrumentation on Linux
  - Both are activated by default if the `install_auto_instrumentation` option is set to `True`.
  - To skip Node.js auto instrumentation, configure the `auto_instrumentation_sdks` option without `nodejs`.
  - To skip .NET auto instrumentation, configure the `auto_instrumentation_sdks` option without `dotnet`.
  - `npm` is required to be pre-installed on the node to install the Node.js SDK. Configure the `auto_instrumentation_npm_path` option to specify the path to `npm`.
  - .NET auto instrumentation is currently only supported on amd64/x64_64.
- (Core) `confmap`: Allow Converters to write logs during startup ([#10135](https://github.com/open-telemetry/opentelemetry-collector/pull/10135))
- (Core) `otelcol`: Enable logging during configuration resolution ([#10056](https://github.com/open-telemetry/opentelemetry-collector/pull/10056))
- (Contrib) `filelogreceiver`: Add container operator parser ([#31959](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31959))
- (Contrib) `extension/storage/filestorage`: New flag cleanup_on_start for the compaction section (default=false). ([#32863](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32863))
  It will remove all temporary files in the compaction directory (those which start with `tempdb`),
  temp files will be left if a previous run of the process is killed while compacting.
- (Contrib) `vcenterreceiver`: Refactors how and when client makes calls in order to provide for faster collection times. ([#31837](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31837))
- (Contrib) `resourcedetectionprocessor`: Support GCP Bare Metal Solution in resource detection processor. ([#32985](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32985))
- (Contrib) `splunkhecreceiver`: Make the channelID header check case-insensitive and allow hecreceiver endpoints able to extract channelID from query params ([#32995](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32995))
- (Contrib) `processor/transform`: Allow common where clause ([#27830](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27830))
- (Contrib) `pkg/ottl`: Added support for timezone in Time converter ([#32140](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32140))
- (Contrib) `probabilisticsamplerprocessor`: Adds the FailClosed flag to solidify current behavior when randomness source is missing. ([#31918](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31918))
- (Contrib) `vcenterreceiver`: Changing various default configurations for vcenterreceiver and removing warnings about future release. ([#32803](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32803), [#32805](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32805), [#32821](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32821), [#32531](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32531), [#32557](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32557))
  The resource attributes that will now be enabled by default are `vcenter.datacenter.name`, `vcenter.virtual_app.name`,
  `vcenter.virtual_app.inventory_path`, `vcenter.vm_template.name`, and `vcenter.vm_template.id`. The metric
  `vcenter.cluster.memory.used` will be removed.  The metrics `vcenter.cluster.vm_template.count` and
  `vcenter.vm.memory.utilization` will be enabled by default.

- (Contrib) `sqlserverreceiver`: Add metrics for database status ([#29865](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29865))
- (Contrib) `sqlserverreceiver`: Add more metrics ([#29865](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29865))
  Added metrics are:
  - sqlserver.resource_pool.disk.throttled.read.rate
  - sqlserver.resource_pool.disk.throttled.write.rate
  - sqlserver.processes.blocked
    These metrics are only available when directly connecting to the SQL server instance

### 🧰 Bug fixes 🧰

- `vcenterreceiver`: Adds inititially disabled packet drop rate metric for VMs. ([#32929](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32929))
- `splunkhecreceiver`: Fix single metric value parsing ([#33084](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33084))
- `vcenterreceiver`: vcenterreceiver client no longer returns error if no Virtual Apps are found. ([#33073](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/33073))
- `vcenterreceiver`: Adds inititially disabled new packet rate metrics to replace the existing ones for VMs & Hosts. ([#32835](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32835))
- `resourcedetectionprocessor`: Change type of `host.cpu.stepping` from int to string. ([#31136](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31136))
  - Disable the `processor.resourcedetection.hostCPUSteppingAsString` feature gate to get the old behavior.

- `pkg/ottl`: Fixes a bug where function name could be used in a condition, resulting in a cryptic error message. ([#33051](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33051))

## v0.100.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.100.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.100.0) and the [opentelemetry-collector-contrib v0.100.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.100.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Linux installer script:
  - Removed support for the deprecated `--[no-]generate-service-name` and `--[enable|disable]-telemetry` options.
  - The minimum supported version for the `--instrumentation-version` option is `0.87.0`.
- (Contrib) `receiver/hostmetrics`: Enable feature gate `receiver.hostmetrics.normalizeProcessCPUUtilization` ([#31368](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31368))

### 🚀 New components 🚀

- (Splunk) Add Redaction processor ([#4766](https://github.com/signalfx/splunk-otel-collector/pull/4766))

### 💡 Enhancements 💡

- (Splunk) Linux installer script: Initial support for [Splunk OpenTelemetry Auto Instrumentation for .NET](https://github.com/signalfx/splunk-otel-dotnet) (x86_64/amd64 only)
  - Activated by default when the `--with-instrumentation` or `--with-systemd-instrumentation` option is specified.
  - Use the `--without-instrumentation-sdk dotnet` option to skip activation.
- (Splunk) `receiver/discovery`: Update emitted entity events:
  - Record entity type ([#4761](https://github.com/signalfx/splunk-otel-collector/pull/4761))
  - Add service attributes ([#4760](https://github.com/signalfx/splunk-otel-collector/pull/4760))
  - Update entity events ID fields ([#4739](https://github.com/signalfx/splunk-otel-collector/pull/4739))
- (Contrib) `exporter/kafka`: Enable setting message topics using resource attributes. ([#31178](https://github.com/open-telemetry/)opentelemetry-collector-contrib/issues/31178)
- (Contrib) `exporter/kafka`: Add an ability to publish kafka messages with message key based on metric resource attributes - it will allow partitioning metrics in Kafka. ([#29433](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29433), [#30666](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30666), [#31675](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31675))
- (Contrib) `exporter/splunkhec`: Add experimental exporter batcher config ([#32545](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32545))
- (Contrib) `receiver/windowsperfcounters`: Returns partial errors for failures during scraping to prevent throwing out all successfully retrieved metrics ([#16712](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/16712))
- (Contrib) `receiver/prometheus`: Prometheus receivers and exporters now preserve 'unknown', 'info', and 'stateset' types. ([#16768](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/16768))
- (Contrib) `receiver/sqlserver`: Enable direct connection to SQL Server ([#30297](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30297))
- (Contrib) `receiver/sshcheck`: Add support for running this receiver on Windows ([#30650](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30650))

### 🧰 Bug fixes 🧰

- (Core) Fix `enabled` config option for batch sender ([#10076](https://github.com/open-telemetry/opentelemetry-collector/pull/10076))
- (Contrib) `receiver/k8scluster`: Fix container state metadata ([#32676](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32676))
- (Contrib) `receiver/filelog`: When a flush timed out make sure we are at EOF (can't read more) ([#31512](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31512), [#32170](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32170))
- (Contrib) `receiver/vcenter`:
  - Adds the `vcenter.cluster.name` resource attribute to resource pool with a ClusterComputeResource parent ([#32535](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32535))
  - Updates `vcenter.cluster.memory.effective` (primarily that the value was reporting MiB when it should have been bytes) ([#32782](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32782))
  - Adds warning to vcenter.cluster.memory.used metric if configured about its future removal ([#32805](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32805))
  - Updates the vcenter.cluster.vm.count metric to also report suspended VM counts ([#32803](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32803))
  - Adds `vcenter.datacenter.name` attributes to all resource types to help with resource identification ([#32531](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32531))
  - Adds `vcenter.cluster.name` attributes warning log related to Datastore resource ([#32674](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32674))
  - Adds new `vcenter.virtual_app.name` and `vcenter.virtual_app.inventory_path` resource attributes to appropriate VM Resources ([#32557](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32557))
  - Adds functionality for `vcenter.vm.disk.throughput` while also changing to a gauge. ([#32772](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32772))
  - Adds initially disabled functionality for VM Templates ([#32821](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32821))
- (Contrib) `connector/count`: Fix handling of non-string attributes in the count connector ([#30314](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30314))

## v0.99.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.99.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.99.0) and the [opentelemetry-collector-contrib v0.99.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.99.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) `receiver/discovery`: Update the component to emit entity events
  - The `log_endpoints` config option has been removed. Endpoints are now only reported if they match the configured receiver rules, and are now emitted as entity events.
    ([#4692](https://github.com/signalfx/splunk-otel-collector/pull/4692), [#4684](https://github.com/signalfx/splunk-otel-collector/pull/4684),
    [#4684](https://github.com/signalfx/splunk-otel-collector/pull/4684), and [#4691](https://github.com/signalfx/splunk-otel-collector/pull/4691))
- (Core) `telemetry`: Distributed internal metrics across different levels. ([#7890](https://github.com/open-telemetry/opentelemetry-collector/pull/7890))
  The internal metrics levels are updated along with reported metrics:
  - The default level is changed from `basic` to `normal`, which can be overridden with `service::telmetry::metrics::level` configuration.
  - Batch processor metrics are updated to be reported starting from `normal` level:
    - `processor_batch_batch_send_size`
    - `processor_batch_metadata_cardinality`
    - `processor_batch_timeout_trigger_send`
    - `processor_batch_size_trigger_send`
  - GRPC/HTTP server and client metrics are updated to be reported starting from `detailed` level:
    - http.client.* metrics
    - http.server.* metrics
    - rpc.server.* metrics
    - rpc.client.* metrics
  - Note: These metrics are all excluded by default in the Splunk distribution of the OpenTelemetry Collector.
    This change only affects users who have modified the default configuration's dropping rules (`metric_relabel_configs`)
    in the Prometheus receiver that scrapes internal metrics.
- (Contrib) `extension/filestorage`: Replace path-unsafe characters in component names ([#3148](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/3148))
  The feature gate `extension.filestorage.replaceUnsafeCharacters` is now stable and cannot be disabled.
  See the File Storage extension's README for details.
- (Contrib) `exporter/loadbalancing`: Change AWS Cloud map resolver config fields from camelCase to snake_case. ([#32331](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32331))
  The snake_case is required in OTel Collector config fields. It used to be enforced by tests in cmd/oteltestbedcol,
  but we had to disable them. Now, the tests are going to be enforced on every component independently.
  Hence, the camelCase config fields recently added with the new AWS Cloud Map resolver has to be fixed.

- (Splunk) `smartagent/collectd-mongodb`: Monitor has been removed to resolve CVE-2024-21506 ([#4731](https://github.com/signalfx/splunk-otel-collector/pull/4731))

### 🚀 New components 🚀

- (Splunk) Add ack extension ([#4724](https://github.com/signalfx/splunk-otel-collector/pull/4724))

### 💡 Enhancements 💡

- (Splunk) Include [`splunk-otel-dotnet`](https://github.com/signalfx/splunk-otel-dotnet) in the `splunk-otel-auto-instrumentation` deb/rpm packages (x86_64/amd64 only) ([#4679](https://github.com/signalfx/splunk-otel-collector/pull/4679))
  - **Note**: Only manual activation/configuration for .NET auto instrumentation is currently supported. See [README.md](https://github.com/signalfx/splunk-otel-collector/blob/main/instrumentation/README.md) for details.
- (Splunk) Update splunk-otel-javaagent to `v1.32.0` ([#4715](https://github.com/signalfx/splunk-otel-collector/pull/4715))
- (Splunk) Enable collecting MSI information on Windows in the support bundle ([#4710](https://github.com/signalfx/splunk-otel-collector/pull/4710))
- (Splunk) Bump version of bundled Python to 3.11.9 ([#4729](https://github.com/signalfx/splunk-otel-collector/pull/4729))
- (Splunk) `receiver/mongodb`: Enable auto-discovery when TLS is disabled ([#4722](https://github.com/signalfx/splunk-otel-collector/pull/4722))
- (Core) `confighttp`: Disable concurrency in zstd compression ([#8216](https://github.com/open-telemetry/opentelemetry-collector/pull/8216))
- (Core) `cmd/mdatagen`: support excluding some metrics based on string and regexes in resource_attributes ([#9661](https://github.com/open-telemetry/opentelemetry-collector/pull/9661))
- (Contrib) `vcenterreceiver`: Changes process for collecting VMs & VM perf metrics used by the `vccenterreceiver` to be more efficient (one call now for all VMs) ([#31837](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31837))
- (Contrib) `splunkhecreceiver`: adding support for ack in the splunkhecreceiver ([#26376](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26376))
- (Contrib) `hostmetricsreceiver`: The hostmetricsreceiver now caches the system boot time at receiver start and uses it for all subsequent calls. The featuregate `hostmetrics.process.bootTimeCache` can be disabled to restore previous behaviour. ([#28849](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/28849))
  This change was made because it greatly reduces the CPU usage of the process and processes scrapers.
- (Contrib) `filelogreceiver`: Add `send_quiet` and `drop_quiet` options for `on_error` setting of operators ([#32145](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32145))
- (Contrib) `pkg/ottl`: Add `IsList` OTTL Function ([#27870](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27870))
- (Contrib) `filelogreceiver`: Add `exclude_older_than` configuration setting ([#31053](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31053))
- (Contrib) `pkg/stanza/operator/transformer/recombine`: add a new "max_unmatched_batch_size" config parameter to configure the maximum number of consecutive entries that will be combined into a single entry before the match occurs ([#31653](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31653))

### 🧰 Bug fixes 🧰

- (Splunk) `receiver/discovery`: Fix locking mechanism on attributes ([#4712](https://github.com/signalfx/splunk-otel-collector/pull/4712))
- (Splunk) Fix MSI installs that required elevation. ([#4688](https://github.com/signalfx/splunk-otel-collector/pull/4688))
- (Core) `exporter/otlp`: Allow DNS scheme to be used in endpoint ([#4274](https://github.com/open-telemetry/opentelemetry-collector/pull/4274))
- (Core) `service`: fix record sampler configuration ([#9968](https://github.com/open-telemetry/opentelemetry-collector/pull/9968))
- (Core) `service`: ensure the tracer provider is configured via go.opentelemetry.io/contrib/config ([#9967](https://github.com/open-telemetry/opentelemetry-collector/pull/9967))
- (Core) `otlphttpexporter`: Fixes a bug that was preventing the otlp http exporter from propagating status. ([#9892](https://github.com/open-telemetry/opentelemetry-collector/pull/9892))
- (Core) `confmap`: Fix decoding negative configuration values into uints ([#9060](https://github.com/open-telemetry/opentelemetry-collector/pull/9060))
- (Contrib) `receiver/hostmetricsreceiver`: do not extract the cpu count if the metric is not enabled; this will prevent unnecessary overhead, especially on windows ([#32133](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32133))
- (Contrib) `pkg/stanza`: Fix race condition which prevented `jsonArrayParserFeatureGate` from working correctly. ([#32313](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32313))
- (Contrib) `vcenterreceiver`: Remove the `vcenter.cluster.name` resource attribute from Host resources if the Host is standalone (no cluster) ([#32548](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32548))
- (Contrib) `azureeventhubreceiver`: Fix memory leak on shutdown ([#32401](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32401))
- (Contrib) `fluentforwardreceiver`: Fix memory leak ([#32363](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32363))
- (Contrib) `processor/resourcedetection`: Fix memory leak on AKS ([#32574](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32574))
- (Contrib) `mongodbatlasreceiver`: Fix memory leak by closing idle connections on shutdown ([#32206](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32206))
- (Contrib) `spanmetricsconnector`: Discard counter span metric exemplars after each flush interval to avoid unbounded memory growth ([#31683](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31683))
  This aligns exemplar discarding for counter span metrics with the existing logic for histogram span metrics
- (Contrib) `pkg/stanza`: Unmarshaling now preserves the initial configuration. ([#32169](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32169))
- (Contrib) `resourcedetectionprocessor`: Update to ec2 scraper so that core attributes are not dropped if describeTags returns an error (likely due to permissions) ([#30672](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30672))

## v0.98.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.98.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.98.0) and the [opentelemetry-collector-contrib v0.98.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.98.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Remove the `bash`, `curl`, `nc`, and `tar` command-line utilities from the collector packages/images and smart agent bundle ([#4646](https://github.com/signalfx/splunk-otel-collector/pull/4646))
- (Splunk) `receiver/discovery`: Update metrics and logs evaluation statements schema:
  - Remove `severity_text` field from log evaluation statements ([#4583](https://github.com/signalfx/splunk-otel-collector/pull/4583))
  - Remove `first_only`  field from match struct. Events are always emitted only once for first matching metric or log statement ([#4593](https://github.com/signalfx/splunk-otel-collector/pull/4593))
  - Combine matching conditions with different statuses in one list ([#4588](https://github.com/signalfx/splunk-otel-collector/pull/4588))
  - Apply entity events schema to the logs emitted by the receiver ([#4638](https://github.com/signalfx/splunk-otel-collector/pull/4638))
  - Emit only one log record per matched endpoint ([#4586](https://github.com/signalfx/splunk-otel-collector/pull/4586))
- (Core) `service`: emit internal collector metrics with _ instead of / with OTLP export ([#9774](https://github.com/open-telemetry/opentelemetry-collector/issues/9774))
- (Contrib) `oracledbreceiver`: Fix incorrect values being set for oracledb.tablespace_size.limit and oracledb.tablespace_size.usage ([#31451](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31451))
- (Contrib) `pkg/stanza`: Revert recombine operator's 'overwrite_with' default value. ([#30783](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30783))
- (Contrib) `processor/attributes, processor/resource`: Remove stable coreinternal.attraction.hash.sha256 feature gate. ([#31997](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31997))

### 🚩 Deprecations 🚩

- (Contrib) `postgresqlreceiver`: Minimal supported PostgreSQL version will be updated from 9.6 to 12.0 in a future release. ([#30923](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30923))
  Aligning on the supported versions as can be seen [in the PostgreSQL releases section](https://www.postgresql.org/support/versioning)

### 🚀 New components 🚀

- (Splunk) Add SQL Server receiver ([#4649](https://github.com/signalfx/splunk-otel-collector/pull/4649))

### 💡 Enhancements 💡

- (Splunk) Automatically set `splunk_otlp_histograms: true` for collector telemetry exported via `signalfx` metrics exporter ([#4655](https://github.com/signalfx/splunk-otel-collector/pull/4655))
- (Splunk) Windows installer now removes the unused configuration files from the installation directory ([#4645](https://github.com/signalfx/splunk-otel-collector/pull/4645))
- (Core) `otlpexporter`: Checks for port in the config validation for the otlpexporter ([#9505](https://github.com/open-telemetry/opentelemetry-collector/issues/9505))
- (Core) `service`: Validate pipeline type against component types ([#8007](https://github.com/open-telemetry/opentelemetry-collector/issues/8007))
- (Contrib) `ottl`: Add new Unix function to convert from epoch timestamp to time.Time ([#27868](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27868))
- (Contrib) `filelogreceiver`: When reading a file on filelogreceiver not on windows, if include_file_owner_name is true, it will add the file owner name as the attribute `log.file.owner.name` and if include_file_owner_group_name is true, it will add the file owner group name as the attribute `log.file.owner.group.name`. ([#30775](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30775))
- (Contrib) - `prometheusreceiver`: Allows receiving prometheus native histograms ([#26555](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26555))
  - Native histograms are compatible with OTEL exponential histograms.
  - The feature can be enabled via the feature gate `receiver.prometheusreceiver.EnableNativeHistograms`.
    Run the collector with the command line option `--feature-gates=receiver.prometheusreceiver.EnableNativeHistograms`.
  - Currently the feature also requires that targets are scraped via the ProtoBuf format.
    To start scraping native histograms, set
    `config.global.scrape_protocols` to `[ PrometheusProto, OpenMetricsText1.0.0, OpenMetricsText0.0.1, PrometheusText0.0.4 ]` in the
    receiver configuration. This requirement will be lifted once Prometheus can scrape native histograms over text formats.
  - For more up to date information see the README.md file of the receiver at
    https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/prometheusreceiver/README.md#prometheus-native-histograms.
- (Contrib) `spanmetricsconnector`: Change default value of metrics_flush_interval from 15s to 60s ([#31776](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31776))
- (Contrib) `pkg/ottl`: Adding a string converter into pkg/ottl ([#27867](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27867))
- (Contrib) `loadbalancingexporter`: Support the timeout period of k8s resolver list watch can be configured. ([#31757](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31757))

### 🧰 Bug fixes 🧰

- (Splunk) `discovery`: Don't use component.MustNewIDWithName ([#4565](https://github.com/signalfx/splunk-otel-collector/pull/4565))
- (Core) `configtls`: Fix issue where `IncludeSystemCACertsPool` was not consistently used between `ServerConfig` and `ClientConfig`. ([#9835](https://github.com/open-telemetry/opentelemetry-collector/issues/9863))
- (Core) `component`: Fix issue where the `components` command wasn't properly printing the component type. ([#9856](https://github.com/open-telemetry/opentelemetry-collector/pull/9856))
- (Core) `otelcol`: Fix issue where the `validate` command wasn't properly printing valid component type. ([#9866](https://github.com/open-telemetry/opentelemetry-collector/pull/9866))
- (Core) `receiver/otlp`: Fix bug where the otlp receiver did not properly respond with a retryable error code when possible for http ([#9357](https://github.com/open-telemetry/opentelemetry-collector/pull/9357))
- (Contrib) `filelogreceiver`: Fix missing scope name and group logs based on scope ([#23387](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23387))
- (Contrib) `jmxreceiver`: Fix memory leak during component shutdown ([#32289](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/32289))
- (Contrib) `k8sobjectsreceiver`: Fix memory leak caused by the pull mode's interval ticker ([#31919](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31919))
- (Contrib) `kafkareceiver`: fix kafka receiver panic on shutdown ([#31926](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31926))
- (Contrib) `prometheusreceiver`: Fix a bug where a new prometheus receiver with the same name cannot be created after the previous receiver is Shutdown ([#32123](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/32123))
- (Contrib) `resourcedetectionprocessor`: Only attempt to detect Kubernetes node resource attributes when they're enabled. ([#31941](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31941))
- (Contrib) `syslogreceiver`: Fix issue where static resource and attributes were ignored ([#31849](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31849))

## v0.97.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.97.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.97.0) and the [opentelemetry-collector-contrib v0.97.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.97.0) releases where appropriate.

### 🚀 New components 🚀

- (Splunk) Add AWS container insights receiver ([#4125](https://github.com/signalfx/splunk-otel-collector/pull/4125))
- (Splunk) Add AWS ECS container metrics receiver ([#4125](https://github.com/signalfx/splunk-otel-collector/pull/4125))
- (Splunk) Add Apache metrics receiver ([#4505](https://github.com/signalfx/splunk-otel-collector/pull/4505))

### 💡 Enhancements 💡

- (Splunk) `memory_ballast` has been removed. If GOMEMLIMIT env var is not set, then 90% of the total available memory limit is set by default. ([#4404](https://github.com/signalfx/splunk-otel-collector/pull/4404))
- (Splunk) Support Windows offline installations ([#4471](https://github.com/signalfx/splunk-otel-collector/pull/4471))
- (Core) `configtls`: Validates TLS min_version and max_version ([#9475](https://github.com/open-telemetry/opentelemetry-collector/issues/9475))
  Introduces `Validate()` method in TLSSetting.
- (Contrib) `exporter/loadbalancingexporter`: Adding AWS Cloud Map for service discovery of Collectors backend. ([#27241](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27241))
- (Contrib) `awss3exporter`: add `compression` option to enable file compression on S3 ([#27872](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27872))
    Add `compression` option to compress files using `compress/gzip` library before uploading to S3.
- (Contrib) `awss3exporter`: Add support for encoding extension to awss3exporter ([#30554](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30554))
- (Contrib) `fileexporter`: Adopt the encoding extension with the file exporter. ([#31774](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31774))
- (Contrib) `pkg/ottl`: Add `ParseXML` function for parsing XML from a target string. ([#31133](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31133))
- (Contrib) `fileexporter`: Added the option to write telemetry data into multiple files, where the file path is based on a resource attribute. ([#24654](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24654))
- (Contrib) `fileexporter`: File write mode is configurable now (truncate or append) ([#31364](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31364))
- (Contrib) `k8sclusterreceiver`: add optional status_last_terminated_reason resource attribute ([#31282](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31282))
- (Contrib) `prometheusreceiver`: Use confighttp for target allocator client ([#31449](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31449))
- (Contrib) `spanmetricsconnector`: Add `metrics_expiration` option to enable expiration of metrics if spans are not received within a certain time frame. ([#30559](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30559))
    The feature can be configured by specifying the desired duration in the `metrics_expiration` option. By default, the expiration is disabled (set to 0).

### 🛑 Breaking changes 🛑

- (Splunk) `collectd/kong`: Remove `collectd/kong`. Please use the [Prometheus receiver](https://docs.splunk.com/observability/en/gdi/monitors-cloud/kong.html) instead. ([#4420](https://github.com/signalfx/splunk-otel-collector/pull/4420))
- (Splunk) `spanmetricsprocessor`: Remove `spanmetricsprocessor`. Please use `spanmetrics` connector instead. ([#4454](https://github.com/signalfx/splunk-otel-collector/pull/4454))
- (Core) `telemetry`: Remove telemetry.useOtelForInternalMetrics stable feature gate ([#9752](https://github.com/open-telemetry/opentelemetry-collector/pull/9752))
- (Contrib) `receiver/postgresql`: Bump postgresqlreceiver.preciselagmetrics gate to beta ([#31220](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31220))
- (Contrib) `receiver/vcenter`: Bump receiver.vcenter.emitPerfMetricsWithObjects feature gate to stable ([#31215](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31215))
- (Contrib) `prometheusreceiver`: Remove enable_protobuf_negotiation option on the prometheus receiver. Use config.global.scrape_protocols = [ PrometheusProto, OpenMetricsText1.0.0, OpenMetricsText0.0.1, PrometheusText0.0.4 ] instead. ([#30883](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30883))
  See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#configuration-file for details on setting scrape_protocols.
- (Contrib) `vcenterreceiver`: Fixed the resource attribute model to more accurately support multi-cluster deployments ([#30879](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30879))
  For more information on impacts please refer https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31113. The main impacts are that
  the `vcenter.resource_pool.name`, `vcenter.resource_pool.inventory_path`, and `vcenter.cluster.name` are reported with more accuracy on VM metrics.

### 🧰 Bug fixes 🧰

- (Splunk) `telemetry`: Simplify the config converter setting the `metric_relabel_configs` in the Prometheus receiver
  to remove the excessive internal metrics. Now, it only overrides the old default rule excluding `.*grpc_io.*` metrics.
  Any other custom setting is left untouched. Otherwise, customizing the `metric_relabel_configs` is very difficult.
  ([#4482](https://github.com/signalfx/splunk-otel-collector/pull/4482))
- (Core) `exporters`: Fix persistent queue size backup on reads.  ([#9740](https://github.com/open-telemetry/opentelemetry-collector/pull/9740))
- (Core) `processor/batch`: Prevent starting unnecessary goroutines.  ([#9739](https://github.com/open-telemetry/opentelemetry-collector/issues/9739))
- (Core) `otlphttpexporter`: prevent error on empty response body when content type is application/json  ([#9666](https://github.com/open-telemetry/opentelemetry-collector/issues/9666))
- (Core) `otelcol`: Respect telemetry configuration when running as a Windows service  ([#5300](https://github.com/open-telemetry/opentelemetry-collector/issues/5300))
- (Contrib) `carbonreceiver`: Do not report fatal error when closed normally ([#31913](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31913))
- (Contrib)`exporter/loadbalancing`: Fix panic when a sub-exporter is shut down while still handling requests. ([#31410](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31410))
- (Contrib) `hostmetricsreceiver`: Adds the receiver.hostmetrics.normalizeProcessCPUUtilization feature gate to optionally normalize process.cpu.utilization values. ([#31368](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31368))
    When enabled, the receiver.hostmetrics.normalizeProcessCPUUtilization feature gate will cause process.cpu.utilization values to be divided by the number of logical cores on the system. This is necessary to produce a value on the interval of [0-1], as the description of process.cpu.utilization the metric says.
- (Contrib) `transformprocessor`: Change metric unit for metrics extracted with `extract_count_metric()` to be the default unit (`1`) ([#31575](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31575))
  The original metric `unit` does not apply to extracted `count` metrics the same way it does to `sum`, `min` or `max`.
  Metrics extracted using `extract_count_metric()` now use the more appropriate default unit (`1`) instead.
- (Contrib) `loadbalancingexporter`: Fix memory leaks on shutdown ([#31050](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31050))
- (Contrib) `signalfxexporter`: Fix memory leak in shutdown ([#30864](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30864), [#30438](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30438))
- (Contrib) `processor/k8sattributes`: Allows k8sattributes processor to work with k8s role/rolebindings when filter::namespace is set. ([#14742](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14742))
- (Contrib) `sqlqueryreceiver`: Fix memory leak on shutdown for log telemetry ([#31782](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/31782))

## v0.96.1

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.96.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.96.0) and the [opentelemetry-collector-contrib v0.96.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.96.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `configgrpc`: Remove deprecated `GRPCClientSettings`, `GRPCServerSettings`, and `ServerConfig.ToListenerContext`. ([#9616](https://github.com/open-telemetry/opentelemetry-collector/pull/9616))
- (Core) `confighttp`: Remove deprecated `HTTPClientSettings`, `NewDefaultHTTPClientSettings`, and `CORSSettings`. ([#9625](https://github.com/open-telemetry/opentelemetry-collector/pull/9625))
- (Core) `confignet`: Removes deprecated `NetAddr` and `TCPAddr` ([#9614](https://github.com/open-telemetry/opentelemetry-collector/pull/9614))
- (Contrib) `spanmetricsprocessor`: Remove spanmetrics processor ([#29567](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29567))
  - You can use the spanmetrics connector as a replacement
- (Contrib) `httpforwarder`: Remove extension named httpforwarder, use httpforwarderextension instead. ([#24171](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24171))
- (Contrib) `k8sclusterreceiver`: Remove deprecated k8s.kubeproxy.version resource attribute ([#29748](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29748))

### 💡 Enhancements 💡

- (Core) `configtls`: Add `include_system_ca_certs_pool` to configtls, allowing to load system certs and additional custom certs. ([#7774](https://github.com/open-telemetry/opentelemetry-collector/pull/7774))
- (Core) `otelcol`: Add `ConfigProviderSettings` to `CollectorSettings` ([#4759](https://github.com/open-telemetry/opentelemetry-collector/pull/4759))
  This allows passing a custom list of `confmap.Provider`s to `otelcol.NewCommand`.
- (Core) `pdata`: Update to OTLP v1.1.0 ([#9587](https://github.com/open-telemetry/opentelemetry-collector/pull/9587))
  Introduces Span and SpanLink flags.
- (Core) `confmap`: Update mapstructure to use a maintained fork, github.com/go-viper/mapstructure/v2. ([#9634](https://github.com/open-telemetry/opentelemetry-collector/pull/9634))
  See https://github.com/mitchellh/mapstructure/issues/349 for context.
- (Contrib) `statsdreceiver`: Add support for the latest version of DogStatsD protocol (v1.3) ([#31295](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31295))
- (Contrib) `fileexporter`: Scope the behavior of the fileexporter to its lifecycle, so it is safe to shut it down or restart it. ([#27489](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27489))
- (Contrib) `processor/resourcedetection`: Add `processor.resourcedetection.hostCPUSteppingAsString` feature gate to change the type of `host.cpu.stepping` from `int` to `string`. ([#31136](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31136))
  This feature gate will graduate to beta in the next release.
- (Contrib) `routingconnector`: a warning is logged if there are two or more routing items with the same routing statement ([#30663](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30663))
- (Contrib) `pkg/ottl`: Add new IsInt function to facilitate type checking. ([#27894](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27894))
- (Contrib) `cmd/mdatagen`: Make lifecycle tests generated by default ([#31532](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31532))
- (Contrib) `pkg/stanza`: Improve timestamp parsing documentation ([#31490](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31490))
- (Contrib) `postgresqlreceiver`: Add `receiver.postgresql.connectionPool` feature gate to reuse database connections ([#30831](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30831))
  The default implementation recreates and closes connections on each scrape per database configured/discovered.
  This change offers a feature gated alternative to keep connections open. Also, it exposes connection configuration to control the behavior of the pool.

### 🧰 Bug fixes 🧰

- (Core) `configretry`: Allow max_elapsed_time to be set to 0 for indefinite retries ([#9641](https://github.com/open-telemetry/opentelemetry-collector/pull/9641))
- (Core) `client`: Make `Metadata.Get` thread safe ([#9595](https://github.com/open-telemetry/opentelemetry-collector/pull/9595))
- (Contrib) `carbonreceiver`: Accept carbon metrics with float timestamps ([#31312](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31312))
- (Contrib) `journaldreceiver`: Fix bug where failed startup could bury error message due to panic during shutdown ([#31476](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31476))
- (Contrib) `loadbalancingexporter`: Fixes a bug where the endpoint become required, despite not being used by the load balancing exporter. ([#31371](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31371))
- (Contrib) `oracledbreceiver`: Use metadata.Type for the scraper id to avoid invalid scraper IDs. ([#31457](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31457))
- (Contrib) `filelogreceiver`: Fix bug where delete_after_read would cause panic ([#31383](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31383))
- (Contrib) `receiver/filelog`: Fix issue where file fingerprint could be corrupted while reading. ([#22936](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/22936))

## v0.96.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.96.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.96.0) and the [opentelemetry-collector-contrib v0.96.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.96.0) releases where appropriate.

### 🚀 New components 🚀

- (Splunk) Add the `cumulativetodelta` processor ([#4401](https://github.com/signalfx/splunk-otel-collector/pull/4401))

### 💡 Enhancements 💡

- (Splunk) Bump github.com/prometheus/common from 0.46.0 to 0.49.0  ([#4353](https://github.com/signalfx/splunk-otel-collector/pull/4382))
- (Splunk) Bumps [aquasecurity/trivy-action](https://github.com/aquasecurity/trivy-action) from 0.17.0 to 0.18.0 ([#4382](https://github.com/signalfx/splunk-otel-collector/pull/4382))
- (Splunk) Update splunk-otel-javaagent to latest ([#4402](https://github.com/signalfx/splunk-otel-collector/pull/4402))
- (Splunk) Add X-SF-Token header to the configuration masked keys ([#4403](https://github.com/signalfx/splunk-otel-collector/pull/4403))
- (Splunk) Bump setuptools in /internal/signalfx-agent/bundle/script([#4330](https://github.com/signalfx/splunk-otel-collector/pull/4403))
- (Splunk) Rocky Linux installation support ([#4398](https://github.com/signalfx/splunk-otel-collector/pull/4398 ))
- (Splunk) Add a test to check what we choose to redact ([#4406](https://github.com/signalfx/splunk-otel-collector/pull/4406))
- (Splunk) Fixed high alert vulnerabity ([#4407](https://github.com/signalfx/splunk-otel-collector/pull/4407))
- (Splunk) Update pgproto to 2.3.3  ([#4409](https://github.com/signalfx/splunk-otel-collector/pull/4409))****

## v0.95.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.95.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.95.0) and the [opentelemetry-collector-contrib v0.95.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.95.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk/Core/Contrib) Bump minimum version to go 1.21 ([#4390](https://github.com/signalfx/splunk-otel-collector/pull/4390))
- (Core) `all`: scope name for all generated Meter/Tracer funcs now includes full package name ([#9494](https://github.com/open-telemetry/opentelemetry-collector/pull/9494))
- (Contrib) `receiver/mongodb`: Bump receiver.mongodb.removeDatabaseAttr feature gate to beta ([#31212](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31212))
- (Contrib) `extension/filestorage`: The `filestorage` extension is now a standalone module. ([#31040](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31040))

### 💡 Enhancements 💡

- (Splunk) MSI defaults to per machine install to avoid issues when different administrators install and update the collector on the same Windows machine ([#4352](https://github.com/signalfx/splunk-otel-collector/pull/4352))
- (Core) `confighttp`: Adds support for Snappy decompression of HTTP requests. ([#7632](https://github.com/open-telemetry/opentelemetry-collector/pull/7632))
- (Core) `configretry`: Validate `max_elapsed_time`, ensure it is larger than `max_interval` and `initial_interval` respectively. ([#9489](https://github.com/open-telemetry/opentelemetry-collector/pull/9489))
- (Core) `configopaque`: Mark module as stable ([#9167](https://github.com/open-telemetry/opentelemetry-collector/pull/9167))
- (Core) `otlphttpexporter`: Add support for json content encoding when exporting telemetry ([#6945](https://github.com/open-telemetry/opentelemetry-collector/pull/6945))
- (Core) `confmap/converter/expandconverter, confmap/provider/envprovider, confmap/provider/fileprovider, confmap/provider/httprovider, confmap/provider/httpsprovider, confmap/provider/yamlprovider`: Split confmap.Converter and confmap.Provider implementation packages out of confmap. ([#4759](https://github.com/open-telemetry/opentelemetry-collector/pull/4759), [#9460](https://github.com/open-telemetry/opentelemetry-collector/pull/9460))
- (Contrib) `hostmetricsreceiver`: Add a new optional resource attribute `process.cgroup` to the `process` scraper of the `hostmetrics` receiver. ([#29282](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29282))
- (Contrib) `awss3exporter`: Add a marshaler that stores the body of log records in s3. ([#30318](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30318))
- (Contrib) `pkg/ottl`: Adds a new ParseCSV converter that can be used to parse CSV strings. ([#30921](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30921))
- (Contrib) `loadbalancingexporter`: Add benchmarks for Metrics and Traces ([#30915](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30915))
- (Contrib) `pkg/ottl`: Add support to specify the format for a replacement string ([#27820](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27820))
- (Contrib) `pkg/ottl`: Add `ParseKeyValue` function for parsing key value pairs from a target string ([#30998](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30998))
- (Contrib) `receivercreator`: Remove use of `ReportFatalError` ([#30596](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30596))
- (Contrib) `processor/tail_sampling`: Add metrics that measure the number of sampled spans and the number of spans that are dropped due to sampling decisions. ([#30482](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30482))
- (Contrib) `exporter/signalfx`: Send histograms in otlp format with new config `send_otlp_histograms` option ([#26298](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26298))
- (Contrib) `receiver/signalfx`: Accept otlp protobuf requests when content-type is "application/x-protobuf;format=otlp" ([#26298](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26298))
- (Contrib) `signalfxreceiver`: Remove deprecated use of `host.ReportFatalError` ([#30598](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30598))

### 🧰 Bug fixes 🧰

- (Contrib) `pkg/stanza`: Add 'allow_skip_pri_header' flag to syslog setting. ([#30397](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30397))
  Allow parsing syslog records without PRI header. Currently pri header is beng enforced although it's not mandatory by the RFC standard. Since influxdata/go-syslog is not maintained we had to switch to haimrubinstein/go-syslog.

- (Contrib) `extension/storage`: Ensure fsync is turned on after compaction ([#20266](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/20266))
- (Contrib) `logstransformprocessor`: Fix potential panic on shutdown due to incorrect shutdown order ([#31139](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31139))
- (Contrib) `receiver/prometheusreceiver`: prometheusreceiver fix translation of metrics with _created suffix ([#30309](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30309))
- (Contrib) `pkg/stanza`: Fixed a bug in the keyvalue_parser where quoted values could be split if they contained a delimited. ([#31034](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31034))

## v0.94.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.94.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.94.1) and the [opentelemetry-collector-contrib v0.94.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.94.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) The Splunk OpenTelemetry Collector [Windows install script](https://docs.splunk.com/observability/en/gdi/opentelemetry/collector-windows/install-windows.html#install-the-collector-using-the-script)
  now installs the [Splunk Distribution of OpenTelemetry .NET](https://docs.splunk.com/observability/en/gdi/get-data-in/application/otel-dotnet/get-started.html#instrument-net-applications-for-splunk-observability-cloud-opentelemetry)
  instead of the `SignalFx Instrumentation for .NET`
  when the parameter `-with_dotnet_instrumentation` is set to `$true` ([#4343](https://github.com/signalfx/splunk-otel-collector/pull/4343))
- (Core) `receiver/otlp`: Update gRPC code from `codes.InvalidArgument` to `codes.Internal` when a permanent error doesn't contain a gRPC status ([#9415](https://github.com/open-telemetry/opentelemetry-collector/pull/#9415))
- (Contrib) `kafkareceiver`: standardizes the default topic name for metrics and logs receivers to the same topic name as the metrics and logs exporters of the kafkaexporter ([#27292](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27292))
  If you are using the Kafka receiver in a logs and/or a metrics pipeline
  and you are not customizing the name of the topic to read from with the `topic` property,
  the receiver will now read from `otlp_logs` or `otlp_metrics` topic instead of `otlp_spans` topic.
  To maintain previous behavior, set the `topic` property to `otlp_spans`.

- (Contrib) `pkg/stanza`: Entries are no longer logged during error conditions. ([#26670](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26670))
  This change is being made to ensure sensitive information contained in logs are never logged inadvertently.
  This change is a breaking change because it may change user expectations. However, it should require
  no action on the part of the user unless they are relying on logs from a few specific error cases.

- (Contrib) `pkg/stanza`: Invert recombine operator's 'overwrite_with' default value. ([#30783](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30783))
  Previously, the default value was `oldest`, meaning that the recombine operator _should_ emit the
  first entry from each batch (with the recombined field). However, the actual behavior was inverted.
  This fixes the bug but also inverts the default setting so as to effectively cancel out the bug fix
  for users who were not using this setting. For users who were explicitly setting `overwrite_with`,
  this corrects the intended behavior.


### 🚩 Deprecations 🚩

- (Core) `configgrpc`: Deprecate GRPCClientSettings, use ClientConfig instead ([#6767](https://github.com/open-telemetry/opentelemetry-collector/pull/6767))

### 💡 Enhancements 💡

- (Splunk) Add a resource attribute to internal metrics to track discovery usage ([#4323](https://github.com/signalfx/splunk-otel-collector/pull/4323))
- (Splunk) Create a multi-architecture Windows docker image for the collector ([#4296](https://github.com/signalfx/splunk-otel-collector/pull/4296))
- (Splunk) Bump `splunk-otel-javaagent` to `v1.30.2` ([#4300](https://github.com/signalfx/splunk-otel-collector/pull/4300))
- (Core) `mdatagen`: Add a generated test that checks the config struct using `componenttest.CheckConfigStruct` ([#9438](https://github.com/open-telemetry/opentelemetry-collector/pull/9438))
- (Core) `component`: Add `component.UseLocalHostAsDefaultHost` feature gate that changes default endpoints from 0.0.0.0 to localhost ([#8510](https://github.com/open-telemetry/opentelemetry-collector/pull/8510))
  The only component in this repository affected by this is the OTLP receiver.
- (Core) `confighttp`: Add support of Host header ([#9395](https://github.com/open-telemetry/opentelemetry-collector/pull/9395))
- (Core) `mdatagen`: Remove use of ReportFatalError in generated tests ([#9439](https://github.com/open-telemetry/opentelemetry-collector/pull/9439))
- (Contrib) `receiver/journald`: add a new config option "all" that turns on full output from journalctl, including lines that are too long. ([#30920](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30920))
- (Contrib) `pkg/stanza`: Add support in a header configuration for json array parser. ([#30321](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30321))
- (Contrib) `awss3exporter`: Add the ability to export trace/log/metrics in OTLP ProtoBuf format. ([#30682](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30682))
- (Contrib) `dockerobserver`: Upgrading Docker API version default from 1.22 to 1.24 ([#30900](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30900))
- (Contrib) `filterprocessor`: move metrics from OpenCensus to OpenTelemetry ([#30736](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30736))
- (Contrib) `groupbyattrsprocessor`: move metrics from OpenCensus to OpenTelemetry ([#30763](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30763))
- (Contrib) `loadbalancingexporter`: Optimize metrics and traces export ([#30141](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30141))
- (Contrib) `all`: Add `component.UseLocalHostAsDefaultHost` feature gate that changes default endpoints from 0.0.0.0 to localhost ([#30702](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30702))
  This change affects the following components:
  - extension/health_check
  - receiver/jaeger
  - receiver/sapm
  - receiver/signalfx
  - receiver/splunk_hec
  - receiver/zipkin

- (Contrib) `processor/resourcedetectionprocessor`: Detect Azure cluster name from IMDS metadata ([#26794](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26794))
- (Contrib) `processor/transform`: Add `copy_metric` function to allow duplicating a metric ([#30846](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30846))

### 🧰 Bug fixes 🧰

- (Splunk) Fixes the value of a default environment variable used by Windows msi. ([#4361](https://github.com/signalfx/splunk-otel-collector/pull/4361))
- (Core) `service`: fix opencensus bridge configuration in periodic readers ([#9361](https://github.com/open-telemetry/opentelemetry-collector/pull/9361))
- (Core) `otlpreceiver`: Fix goroutine leak when GRPC server is started but HTTP server is unsuccessful ([#9165](https://github.com/open-telemetry/opentelemetry-collector/pull/9165))
- (Core) `otlpexporter`: PartialSuccess is treated as success, logged as warning. ([#9243](https://github.com/open-telemetry/opentelemetry-collector/pull/9243))

- (Contrib) `basicauthextension`: Accept empty usernames. ([#30470](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30470))
  Per https://datatracker.ietf.org/doc/html/rfc2617#section-2, username and password may be empty strings ("").
  The validation used to enforce that usernames cannot be empty.

- (Contrib) `pkg/ottl`: Fix parsing of string escapes in OTTL ([#23238](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23238))
- (Contrib) `pkg/stanza`: Recombine operator should always recombine partial logs ([#30797](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30797))
  Previously, certain circumstances could result in partial logs being emitted without any
  recombiniation. This could occur when using `is_first_entry`, if the first partial log from
  a source was emitted before a matching "start of log" indicator was found. This could also
  occur when the collector was shutting down.

- (Contrib) `pkg/stanza`: Fix bug where recombine operator's 'overwrite_with' condition was inverted. ([#30783](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30783))
- (Contrib) `exporter/signalfx`: Use "unknown" value for the environment correlation calls as fallback. ([#31052](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/31052))
  This fixed the APM/IM correlation in the Splunk Observability UI for the users that send traces with no "deployment.environment" resource attribute value set.

## v0.93.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.93.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.93.0) and the [opentelemetry-collector-contrib v0.93.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.93.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) On Windows the `SPLUNK_*` environment variables were moved from the machine scope to the collector service scope This avoids collisions with other agents and instrumentation. If any of these environment variables are required by your apps, please adopt them directly. ([#3930](https://github.com/signalfx/splunk-otel-collector/pull/3930))
- (Splunk) `mysql` discovery now uses the OpenTelemetry Collector Contrib receiver by default instead of the smartagent receiver. ([#4231](https://github.com/signalfx/splunk-otel-collector/pull/4231))
- (Splunk) Stop sending internal Collector metrics from the batch processor. Drop them at the prometheus receiver level. ([#4273](https://github.com/signalfx/splunk-otel-collector/pull/4273))
- (Core) exporters: remove deprecated exporters.RetrySettings and exporters.NewDefaultRetrySettings ([#9256](https://github.com/open-telemetry/opentelemetry-collector/issues/9256))
- (Contrib) `vcenterreceiver`: "receiver.vcenter.emitPerfMetricsWithObjects" feature gate is beta and enabled by default ([#30615](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30615))
- (Contrib) `docker`: Adopt api_version as strings to correct invalid float truncation ([#24025](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24025))
- (Contrib) `extension/filestorage`: Replace path-unsafe characters in component names ([#3148](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/3148))
  The feature gate `extension.filestorage.replaceUnsafeCharacters` is now enabled by default.
  See the File Storage extension's README for details.
- (Contrib) `postgresqlreceiver`: add feature gate `receiver.postgresql.separateSchemaAttr` to include schema as separate attribute ([#29559](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29559))
  Enabling the featuregate adds a new resource attribute to store the schema of the table or index
  Existing table attributes are adjusted to not include the schema, which was inconsistently used

### 💡 Enhancements 💡
- (Splunk) Update opentelemetry-jmx-metrics version to 1.32.0 ([#4201](https://github.com/signalfx/splunk-otel-collector/pull/4201))
- (Core) `configtls`: add `cipher_suites` to configtls. ([#8105](https://github.com/open-telemetry/opentelemetry-collector/issues/8105))
  Users can specify a list of cipher suites to pick from. If left blank, a safe default list is used.
- (Core) `service`: mark `telemetry.useOtelForInternalMetrics` as stable ([#816](https://github.com/open-telemetry/opentelemetry-collector/issues/816))
  (Splunk) Remove disabled `telemetry.useOtelForInternalMetrics` feature gate from our distribution. Some new internal metrics are now dropped at scrape time.
- (Core) `exporters`: Cleanup log messages for export failures ([#9219]((https://github.com/open-telemetry/opentelemetry-collector/issues/9219)))
  1. Ensure an error message is logged every time and only once when data is dropped/rejected due to export failure.
  2. Update the wording. Specifically, don't use "dropped" term when an error is reported back to the pipeline.
     Keep the "dropped" wording for failures happened after the enabled queue.
  3. Properly report any error reported by a queue. For example, a persistent storage error must be reported as a storage error, not as "queue overflow".
- (Contrib) `pkg/stanza`: Add a json array parser operator and an assign keys transformer. ([#30321](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30321))
  Json array parser opreator can be used to parse a json array string input into a list of objects. |
  Assign keys transformer can be used to assigns keys from the configuration to an input list
- (Contrib) `splunkhecexporter`: Batch data according to access token and index, if present. ([#30404](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30404))
- (Contrib) `k8sattributesprocessor`: Apply lifecycle tests to k8sprocessor, change its behavior to report fatal error ([#30387](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30387))
- (Contrib) `k8sclusterreceiver`: add new disabled os.description, k8s.container_runtime.version resource attributes ([#30342](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30342))
- (Contrib) `k8sclusterreceiver`: add os.type resource attribute ([#30342](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30342))
- (Contrib) `kubeletstatsreceiver`: Add new `*.cpu.usage` metrics. ([#25901](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25901))
- (Contrib) `pkg/ottl`: Add `flatten` function for flattening maps ([#30455](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30455))
- (Contrib) `redisreciever`: adds metric for slave_repl_offset ([#6942](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6942))
  also adds a shell script to set up docker-compose integration test
- (Contrib) `receiver/sqlquery`: Add debug log when running SQL query ([#29672](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29672))

### 🧰 Bug fixes 🧰

- (Core) `otlpreceiver`: Ensure OTLP receiver handles consume errors correctly ([#4335]((https://github.com/open-telemetry/opentelemetry-collector/issues/4335)))
  Make sure OTLP receiver returns correct status code and follows the receiver contract (gRPC)
- (Core) `zpagesextension`: Remove mention of rpcz page from zpages extension ([#9328](https://github.com/open-telemetry/opentelemetry-collector/issues/9328))
- (Contrib) `kafkareceiver`: The Kafka receiver now exports some partition-specific metrics per-partition, with a `partition` tag ([#30177](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30177))
  The following metrics now render per partition:
    - kafka_receiver_messages
    - kafka_receiver_current_offset
    - kafka_receiver_offset_lag

## v0.92.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.92.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.92.0) and the [opentelemetry-collector-contrib v0.92.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.92.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Contrib) `httpforwarder`: Use confighttp.HTTPDefaultClientSettings when configuring the HTTPClientSettings for the httpforwarder extension. ([#6641](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6641))
  By default, the HTTP forwarder extension will now use the defaults set in the extension:
  * The idle connection timeout is set to 90s.
  * The max idle connection count is set to 100.
- (Contrib) `pkg/ottl`: Now validates against extraneous path segments that a context does not know how to use. ([#30042](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30042))
- (Contrib) `pkg/ottl`: Throw an error if keys are used on a path that does not allow them. ([#30162](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30162))
- (Core) `exporters/sending_queue`: Do not re-enqueue failed batches, rely on the retry_on_failure strategy instead. ([#8382](https://github.com/open-telemetry/opentelemetry-collector/issues/8382))
  The current re-enqueuing behavior is not obvious and cannot be configured. It takes place only for persistent queue
  and only if `retry_on_failure::enabled=true` even if `retry_on_failure` is a setting for a different backoff retry
  strategy. This change removes the re-enqueuing behavior. Consider increasing `retry_on_failure::max_elapsed_time`
  to reduce chances of data loss or set it to 0 to keep retrying until requests succeed.
- (Core) `confmap`: Make the option `WithErrorUnused` enabled by default when unmarshaling configuration ([#7102](https://github.com/open-telemetry/opentelemetry-collector/issues/7102))
  The option `WithErrorUnused` is now enabled by default, and a new option `WithIgnoreUnused` is introduced to ignore
  errors about unused fields.

### 🚩 Deprecations 🚩

- (Contrib) `k8sclusterreceiver`: deprecate optional k8s.kubeproxy.version resource attribute ([#29748](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29748))
- (Core) `exporters`: Deprecate exporters.RetrySettings in favor of configretry.BackOffConfig ([#9091](https://github.com/open-telemetry/opentelemetry-collector/pull/9091))
- (Core) `extension/ballast`: Deprecate `memory_ballast` extension. ([#8343](https://github.com/open-telemetry/opentelemetry-collector/issues/8343))
  Use `GOMEMLIMIT` environment variable instead.

### 💡 Enhancements 💡

- (Splunk) support core service validate command ([#4175](https://github.com/signalfx/splunk-otel-collector/pull/4175))
- (Splunk) Add routing connector to Splunk distribution ([#4167](https://github.com/signalfx/splunk-otel-collector/pull/4167))
- (Contrib) adopt splunkhec batch by token and index updates ([#4151](https://github.com/signalfx/splunk-otel-collector/pull/4151))
- (Contrib) `vcenterreceiver`: Add explicit statement of support for version 8 of ESXi and vCenter ([#30274](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30274))
- (Contrib) `routingconnector`: routingconnector supports matching the statement only once ([#26353](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26353))
- (Contrib) `filterprocessor`: Add telemetry for metrics, logs, and spans that were intentionally dropped via filterprocessor. ([#13169](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/13169))
- (Contrib) `pkg/ottl`: Add Hour OTTL Converter ([#29468](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29468))
- (Contrib) `kafkaexporter`: add ability to publish kafka messages with message key of TraceID - it will allow partitioning of the kafka Topic. ([#12318](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/12318))
- (Contrib) `kafkareceiver`: Add three new metrics to record unmarshal errors. ([#29302](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29302))
- (Contrib) `hostmetricsreceiver`: Add `system.memory.limit` metric reporting the total memory available. ([#30306](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30306))
  This metric is opt-in. To enable it, set `scrapers::memory::metrics::system.memory.limit::enabled` to `true` in the hostmetrics config.
- (Contrib) `kafkaexporter`: Adds the ability to configure the Kafka client's Client ID. ([#30144](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/30144))
- (Contrib) `pkg/stanza`: Remove sampling policy from logger ([#23801](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23801))
- (Contrib) `resourcedetectionprocessor`: Add "aws.ecs.task.id" attribute ([#8274](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/8274))
  Resourcedetectionprocessor now exports "aws.ecs.task.id" attribute, in addition to "aws.ecs.task.arn".
  This allows exporters like "awsemfexporter" to automatically pick up that attribute and make it available
  in templating (e.g. to use in CloudWatch log stream name).
- (Contrib) `spanmetricsconnector`: Fix OOM issue for spanmetrics by limiting the number of exemplars that can be added to a unique dimension set ([#27451](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27451))
- (Contrib) `connector/spanmetrics`: Configurable resource metrics key attributes, filter the resource attributes used to create the resource metrics key. ([#29711](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29711))
  This enhancement can be used to fix broken spanmetrics counters after a span producing service restart, when resource attributes contain dynamic/ephemeral values (e.g. process id).
- (Contrib) `splunkhecreceiver`: Returns json response in raw endpoint when it is successful ([#29875](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29875))
- (Contrib) `sqlqueryreceiver`: Swap MS SQL Server driver from legacy 'denisenkom' to official Microsoft fork ([#27200](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27200))
- (Core) `exporters`: Add RetrySettings validation function ([#9089](https://github.com/open-telemetry/opentelemetry-collector/pull/9089))
  Validate that time.Duration, multiplier values in configretry are non-negative, and randomization_factor is between 0 and 1
- (Core) `service`: Enable `telemetry.useOtelForInternalMetrics` by updating the flag to beta ([#7454](https://github.com/open-telemetry/opentelemetry-collector/issues/7454))
  The metrics generated should be consistent with the metrics generated
  previously with OpenCensus. Splunk note: this option is disabled in our distribution. Users can enable the behaviour
  by setting `--feature-gates +telemetry.useOtelForInternalMetrics` at collector start if the new histograms are desired.
- (Core) `confignet`: Add `dialer_timeout` config option. ([#9066](https://github.com/open-telemetry/opentelemetry-collector/pull/9066))
- (Core) `processor/memory_limiter`: Update config validation errors ([#9059](https://github.com/open-telemetry/opentelemetry-collector/pull/9059))
  - Fix names of the config fields that are validated in the error messages
  - Move the validation from start to the initialization phrase
- (Core) `exporters`: Add config Validate for TimeoutSettings ([#9104](https://github.com/open-telemetry/opentelemetry-collector/pull/9104))

### 🧰 Bug fixes 🧰

- (Contrib) `filterset`: Fix concurrency issue when enabling caching. ([#11829](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/11829))
- (Contrib) `pkg/ottl`: Fix issue with the hash value of a match subgroup in replace_pattern functions. ([#29409](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29409))
- (Contrib) `prometheusreceiver`: Fix configuration validation to allow specification of Target Allocator configuration without providing scrape configurations ([#30135](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30135))
- (Contrib) `wavefrontreceiver`: Return error if partially quoted ([#30315](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30315))
- (Contrib) `hosmetricsreceiver`: change cpu.load.average metrics from 1 to {thread} ([#29914](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29914))
- (Contrib) `pkg/ottl`: Fix bug where the Converter `IsBool` was not usable ([#30151](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/30151))
- (Contrib) `time`: The `%z` strptime format now correctly parses `Z` as a valid timezone ([#29929](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29929))
  `strptime(3)` says that `%z` is "an RFC-822/ISO 8601 standard
  timezone specification", but the previous code did not allow the
  string "Z" to signify UTC time, as required by ISO 8601. Now, both
  `+0000` and `Z` are recognized as UTC times in all components that
  handle `strptime` format strings.
- (Core) `memorylimiterprocessor`: Fixed leaking goroutines from memorylimiterprocessor ([#9099](https://github.com/open-telemetry/opentelemetry-collector/issues/9099))
- (Core) `cmd/otelcorecol`: Fix the code detecting if the collector is running as a service on Windows. ([#7350](https://github.com/open-telemetry/opentelemetry-collector/issues/7350))
  Removed the `NO_WINDOWS_SERVICE` environment variable given it is not needed anymore.
- (Core) `otlpexporter`: remove dependency of otlphttpreceiver on otlpexporter ([#6454](https://github.com/open-telemetry/opentelemetry-collector/issues/6454))

## v0.91.3

- (Splunk) Properly sign and associate changelog to release.  This should be otherwise identical to v0.91.2

## v0.91.2

### 🛑 Breaking changes 🛑
- (Splunk) - `ecs-metadata` sync the `known_status` property on the `container_id` dimension instead of lower cardinality `container_name`. This can be prevented by configuring `dimensionToUpdate` to `container_name` ([#4091](https://github.com/signalfx/splunk-otel-collector/pull/4091))
- (Splunk) Removes `collectd/disk` monitor ([#3998](https://github.com/signalfx/splunk-otel-collector/pull/3998))
   This monitor has been deprecated in favor of the `disk-io` monitor.
   Note that the `disk-io` monitor has a different dimension (`disk`
   instead of `plugin_instance`) to specify the disk.
- (Splunk) Removes `collectd/df` monitor ([#3996](https://github.com/signalfx/splunk-otel-collector/pull/3996))
   The monitor is deprecated and the filesystems monitor should be used instead.
- (Splunk) Removes `netinterface` monitor ([#3991](https://github.com/signalfx/splunk-otel-collector/pull/3991))
   This monitor is deprecated in favor of the `net-io` monitor.
- (Splunk) Removes `collectd/vmem` monitor ([#3993](https://github.com/signalfx/splunk-otel-collector/pull/3993))
   This monitor is deprecated in favor of the `vmem` monitor.  The metrics should be fully compatible with this monitor.
- (Splunk) Removes `collectd/load` monitor ([#3995](https://github.com/signalfx/splunk-otel-collector/pull/3995))
   This monitor has been deprecated in favor of the `load` monitor. That monitor emits the same metrics and is fully compatible.
- (Splunk) Removes `collectd/postgresql` monitor ([#3994](https://github.com/signalfx/splunk-otel-collector/pull/3994))
   This monitor is deprecated in favor of the postgresql monitor.

### 💡 Enhancements 💡
- (Splunk) Adopt `vcenter` receiver ([#4291](https://github.com/signalfx/splunk-otel-collector/pull/4121))
- (Splunk) Adopt `sshcheck` receiver ([#4099](https://github.com/signalfx/splunk-otel-collector/pull/4099))
- (Splunk) Adopt `awss3` exporter ([#4117](https://github.com/signalfx/splunk-otel-collector/pull/4117))
- (Splunk) Convert loglevel to verbosity on logging exporter ([#4097](https://github.com/signalfx/splunk-otel-collector/pull/4097))

## v0.91.1

### 💡 Enhancements 💡

- (Splunk) Remove the project beta label ([#4070](https://github.com/signalfx/splunk-otel-collector/pull/4070))
- (Splunk) Source SPLUNK_LISTEN_INTERFACE on all host endpoints([#4065](https://github.com/signalfx/splunk-otel-collector/pull/4065))
- (Splunk) Add support for start timestamps when using the light prometheus receiver ([#4037](https://github.com/signalfx/splunk-otel-collector/pull/4037))
- (Splunk) Node.js Auto Instrumentation:
  - Update splunk-otel-js to [v2.6.0](https://github.com/signalfx/splunk-otel-js/releases/tag/v2.6.0) ([#4064](https://github.com/signalfx/splunk-otel-collector/pull/4064))
  - Update linux installer script to use `--global=false` for local npm versions and configurations ([#4068](https://github.com/signalfx/splunk-otel-collector/pull/4068))

### 🛑 Breaking changes 🛑

- `postgresql` Discovery now uses the OpenTelemetry Collector Contrib receiver by default instead of the smartagent receiver ([#3957](https://github.com/signalfx/splunk-otel-collector/pull/3957))

## v0.91.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.91.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.91.0) and the [opentelemetry-collector-contrib v0.91.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.91.0) releases where appropriate.

### 🛑 Breaking changes 🛑
- (Splunk) Node.js Auto Instrumentation:
  - The `NODE_OPTIONS` environment variable in the default config file has been updated to load the Node.js SDK from an absolute path (`/usr/lib/splunk-instrumentation/splunk-otel-js/node_modules/@splunk/otel/instrument`).
  - The Linux installer script now installs the Node.js SDK to `/usr/lib/splunk-instrumentation/splunk-otel-js` instead of globally.
  - The `--npm-command` Linux installer script option is no longer supported. To specify a custom path to `npm`, use the `--npm-path <path>` option.
- (Splunk) `translatesfx`: Remove `translatesfx` ([#4028](https://github.com/signalfx/splunk-otel-collector/pull/4028))
- (Splunk) `collectd/elasticsearch`: Remove `collectd/elasticsearch` monitor ([#3997](https://github.com/signalfx/splunk-otel-collector/pull/3997))

### 🚩 Deprecations 🚩

- (Splunk) `collectd/cpu`: Deprecate `collectd/cpu` explicitly. Please migrate to the `cpu` monitor ([#4036](https://github.com/signalfx/splunk-otel-collector/pull/4036))

### 💡 Enhancements 💡

- (Contrib) `spanmetricsconnector`: Add exemplars to sum metric ([#27451](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27451))
- (Contrib) `jaegerreceiver`: mark featuregates to replace Thrift-gen with Proto-gen types for sampling strategies as stable ([#27636](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27636))
  The following featuregate is stable:
    `receiver.jaegerreceiver.replaceThriftWithProto`
- (Contrib) `kafkareceiver`: Add the ability to consume logs from Azure Diagnostic Settings streamed through Event Hubs using the Kafka API. ([#18210](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18210))
- (Contrib) `resourcedetectionprocessor`: Add detection of host.ip to system detector. ([#24450](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24450))
- (Contrib) `resourcedetectionprocessor`: Add detection of host.mac to system detector. ([#29587](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29587))
- (Contrib) `pkg/ottl`: Add silent ErrorMode to allow disabling logging of errors that are ignored. ([#29710](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29710))
- (Contrib) `postgresqlreceiver`: Add config property for excluding specific databases from scraping ([#29605](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29605))
- (Contrib) `redisreceiver`: Upgrade the redis library dependency to resolve security vulns in v7 ([#29600](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29600))
- (Contrib) `signalfxexporter`: Enable HTTP/2 health check by default ([#29716](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29716))
- (Contrib) `splunkhecexporter`: Enable HTTP/2 health check by default ([#29717](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29717))
- (Contrib) `statsdreceiver`: Add support for 'simple' tags that do not have a defined value, to accommodate DogStatsD metrics that may utilize these. ([#29012](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29012))
  This functionality is gated behind a new `enable_simple_tags` config boolean, as it is not part of the StatsD spec.
- (Core) `service`: add resource attributes as labels to otel metrics to ensures backwards compatibility with OpenCensus metrics. ([#9029](https://github.com/open-telemetry/opentelemetry-collector/issues/9029))
- (Core) `config/confighttp`: Exposes http/2 transport settings to enable health check and workaround golang http/2 issue https://github.com/golang/go/issues/59690 ([#9022](https://github.com/open-telemetry/opentelemetry-collector/issues/9022))

### 🧰 Bug fixes 🧰

- (Splunk) `migratecheckpoint`: Migrating offsets from SCK to SCK-Otel doesn't work. This is because of incorrect keys we use to populate the boltdb cache. ([#3879](https://github.com/signalfx/splunk-otel-collector/pull/3879))
- (Contrib) `connector/spanmetrics`: Fix memory leak when the cumulative temporality is used. ([#27654](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27654))
- (Contrib) `splunkhecexporter`: Do not send null event field values in HEC events. Replace null values with an empty string. ([#29551](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29551))
- (Contrib) `k8sobjectsreceiver`: fix k8sobjects receiver fails when some unrelated Kubernetes API is down ([#29706](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29706))
- (Contrib) `resourcedetectionprocessor`: Change type of `host.cpu.model.id` and `host.cpu.model.family` from int to string. ([#29025](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29025))
  Disable the `processor.resourcedetection.hostCPUModelAndFamilyAsString` feature gate to get the old behavior.
- (Contrib) `filelogreceiver`: Fix problem where checkpoints could be lost when collector is shutdown abruptly ([#29609](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29609), [#29491](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29491))
- (Contrib) `pkg/stanza`: Allow `key_value_parser` to parse values that contain the delimiter string. ([#29629](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29629)
- (Core) `exporters`: fix missed metric aggregations ([#9048](https://github.com/open-telemetry/opentelemetry-collector/issues/9048))
  This ensures that context cancellation in the exporter doesn't interfere with metric aggregation. The OTel
  SDK currently returns if there's an error in the context used in `Add`. This means that if there's a
  cancelled context in an export, the metrics are now recorded.

## v0.90.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.90.1](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.90.1) and the [opentelemetry-collector-contrib v0.90.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.90.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Core) `service`: To remain backwards compatible w/ the metrics generated today, otel generated metrics will be generated without the `_total` suffix ([#7454](https://github.com/open-telemetry/opentelemetry-collector/issues/7454))
- (Core) `service`: use WithNamespace instead of WrapRegistererWithPrefix ([#8988](https://github.com/open-telemetry/opentelemetry-collector/issues/8988))
  Using this functionality in the otel prom exporter fixes a bug where the
  target_info was prefixed as otelcol_target_info previously.

### 🚩 Deprecations 🚩

- (Splunk) Deprecate `collectd/marathon` ([#3992](https://github.com/signalfx/splunk-otel-collector/pull/3992))
- (Splunk) Add deprecation notice to `collectd/etcd` (use `etcd` instead) ([#3990](https://github.com/signalfx/splunk-otel-collector/pull/3990))
- (Splunk) Mark translatesfx as deprecated ([#3984](https://github.com/signalfx/splunk-otel-collector/pull/3984))

### 💡 Enhancements 💡

- (Splunk) `mysqlreceiver`: Add mysqlreceiver to the Splunk distribution ([#3989](https://github.com/signalfx/splunk-otel-collector/pull/3989))
- (Core) `exporter/debug`: Change default `verbosity` from `normal` to `basic` ([#8844](https://github.com/open-telemetry/opentelemetry-collector/issues/8844))
  This change has currently no effect, as `basic` and `normal` verbosity share the same behavior. This might change in the future though, with the `normal` verbosity being more verbose than it currently is (see https://github.com/open-telemetry/opentelemetry-collector/issues/7806). This is why we are changing the default to `basic`, which is expected to stay at the current level of verbosity (one line per batch).
- (Core) `exporters`: Fix shutdown logic in persistent queue to not require consumers to be closed first ([#8899](https://github.com/open-telemetry/opentelemetry-collector/issues/8899))
- (Core) `confighttp`: Support proxy configuration field in all exporters that support confighttp ([#5761](https://github.com/open-telemetry/opentelemetry-collector/issues/5761))
- (Contrib) `resourcedetectionprocessor`: Add k8s cluster name detection when running in EKS ([#26794](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26794))
- (Contrib) `pkg/ottl`: Add new IsDouble function to facilitate type checking. ([#27895](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27895))
- (Contrib) `mysqlreceiver`: expose tls in mysqlreceiver ([#29269](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29269))
  If tls is not set, the default is to disable TLS connections.
- (Contrib) `processor/transform`: Convert between sum and gauge in metric context when alpha feature gate `processor.transform.ConvertBetweenSumAndGaugeMetricContext` enabled ([#20773](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20773))
- (Contrib) `receiver/mongodbatlasreceiver`: adds project config to mongodbatlas metrics to filter by project name and clusters. ([#28865](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28865))
- (Contrib) `pkg/stanza`: Add "namedpipe" operator. ([#27234](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27234))
- (Contrib) `pkg/resourcetotelemetry`: Do not clone data in pkg/resourcetotelemetry by default ([#29327](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29327))
  The resulting consumer will be marked as MutatesData instead
- (Contrib) `pkg/stanza`: Improve performance by not calling decode when nop encoding is defined ([#28899](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28899))
- (Contrib) `receivercreator`: Added support for discovery of endpoints based on K8s services ([#29022](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29022))
  By discovering endpoints based on K8s services, a dynamic probing of K8s service leveraging for example the httpcheckreceiver get enabled
- (Contrib) `signalfxexporter`: change default timeout to 10 seconds ([#29436](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/29436))
- (Contrib) `hostmetricsreceiver`: Add optional Linux-only metric system.linux.memory.available ([#7417](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/7417))
  This is an alternative to `system.memory.usage` metric with state=free.
  Linux starting from 3.14 exports "available" memory. It takes "free" memory as a baseline, and then factors in kernel-specific values.
  This is supposed to be more accurate than just "free" memory.
  For reference, see the calculations [here](https://superuser.com/a/980821).
  See also `MemAvailable` in `/proc/meminfo`.

### 🧰 Bug fixes 🧰

- (Splunk) `cmd/otelcol`: Fix the code detecting if the collector is running as a service on Windows. The fix should make
  setting the `NO_WINDOWS_SERVICE` environment variable unnecessary. ([#4002](https://github.com/signalfx/splunk-otel-collector/pull/4002))
- (Core) `exporters`: Fix invalid write index updates in the persistent queue ([#8115](https://github.com/open-telemetry/opentelemetry-collector/issues/8115))
- (Contrib) `filelogreceiver`: Fix issue where files were unnecessarily kept open on Windows ([#29149](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29149))
- (Contrib) `mongodbreceiver`: add receiver.mongodb.removeDatabaseAttr Alpha feature gate to remove duplicate database name attribute ([#24972](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24972))
- (Contrib) `pkg/stanza`: Fix panic during stop for udp async mode only. ([#29120](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29120))

## v0.89.0

### 🛑 Breaking changes 🛑

- (Contrib) `pkg/stanza`/`receiver/windowseventlog`: Improve parsing of Windows Event XML by handling anonymous `Data` elements. ([#21491](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21491))
  This improves the contents of Windows log events for which the publisher manifest is unavailable. Previously, anonymous `Data` elements were ignored. This is a breaking change for users who were relying on the previous data format.

- (Contrib) `processor/k8sattributes`: Graduate "k8sattr.rfc3339" feature gate to Beta. ([#28817](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28817))
  Time format of `k8s.pod.start_time` attribute value migrated from RFC3339:
  Before: 2023-07-10 12:34:39.740638 -0700 PDT m=+0.020184946
  After: 2023-07-10T12:39:53.112485-07:00
  The feature gate can be temporary reverted back by adding `--feature-gate=-k8sattr.rfc3339` to the command line.

- (Contrib) `receiver/filelogreceiver`: Change "Started watching file" log behavior ([#28491](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28491))
  Previously, every unique file path which was found by the receiver would be remembered indefinitely.
  This list was kept independently of the uniqueness / checkpointing mechanism (which does not rely on the file path).
  The purpose of this list was to allow us to emit a log whenever a path was seen for the first time.
  This removes the separate list and relies instead on the same mechanism as checkpointing. Now, a similar log is emitted
  any time a file is found which is not currently checkpointed. Because the checkpointing mechanism does not maintain history
  indefinitely, it is now possible that a log will be emitted for the same file path. This will happen when no file exists at
  the path for a period of time.

### 🚩 Deprecations 🚩

- (Contrib) `postgresqlreceiver`: Deprecation of postgresql replication lag metrics `postgresql.wal.lag` in favor of more precise 'postgresql.wal.delay' ([#26714](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26714))

### 💡 Enhancements 💡

- (Splunk) `receiver/mongodbreceiver`: Adds mongobdreceiver in Splunk collector distro ([#3979](https://github.com/signalfx/splunk-otel-collector/pull/3979/))
- (Contrib) `processor/tailsampling`: adds optional upper bound duration for sampling ([#26115](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26115))
- (Contrib) `collectdreceiver`: Add support of confighttp.HTTPServerSettings ([#28811](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28811))
- (Contrib) `collectdreceiver`: Promote collectdreceiver as beta component ([#28658](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28658))
- (Contrib) `receiver/hostmetricsreceiver`: Added support for host's cpuinfo frequnecies. ([#27445](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27445))
  In Linux the current frequency is populated using the values from /proc/cpuinfo. An os specific implementation will be needed for Windows and others.
- (Contrib) `receiver/hostmetrics/scrapers/process`: add configuration option to mute `error reading username for process` ([#14311](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14311), [#17187](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/17187))
- (Contrib) `azureevenhubreceiver`: Allow the Consumer Group to be set in the Configuration. ([#28633](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28633))
- (Contrib) `spanmetricsconnector`: Add Events metric to span metrics connector that adds list of event attributes as dimensions ([#27451](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27451))
- (Contrib) `processor/k8sattribute`: support adding labels and annotations from node ([#22620](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22620))
- (Contrib) `windowseventlogreceiver`: Add parsing for Security and Execution event fields. ([#27810](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27810))
- (Contrib) `filelogreceiver`: Add the ability to order files by mtime, to only read the most recently modified files ([#27812](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27812))
- (Contrib) `wavefrontreceiver`: Wrap metrics receiver under carbon receiver instead of using export function ([#27248](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27248))
- (Contrib) `pkg/ottl`: Add IsBool function into OTTL ([#27897](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27897))
- (Contrib) `k8sclusterreceiver`: add k8s.node.condition metric ([#27617](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27617))
- (Contrib) `kafkaexporter`/`kafkametricsreceiver`/`kafkareceiver`: Expose resolve_canonical_bootstrap_servers_only config ([#26022](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26022))
- (Contrib) `receiver/mongodbatlasreceiver`: Enhanced collector logs to include more information about the MongoDB Atlas API calls being made during logs retrieval. ([#28851](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28851))
- (Contrib) `receiver/mongodbatlasreceiver`: emit resource attributes "`mongodb_atlas.region.name`" and "`mongodb_atlas.provider.name`" on metric scrape. ([#28833](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28833))
- (Contrib) `processor/resourcedetection`: Add `processor.resourcedetection.hostCPUModelAndFamilyAsString` feature gate to change the type of `host.cpu.family` and `host.cpu.model.id` attributes from `int` to `string`. ([#29025](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29025))
  This feature gate will graduate to beta in the next release.
- (Contrib) `processor/tailsampling`: Optimize performance of tailsamplingprocessor ([#27889](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27889))
- (Contrib) `redisreceiver`: include server.address and server.port resource attributes ([#22044](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22044))
- (Contrib) `spanmetricsconnector`: Add exemplars to sum metric ([#27451](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27451))
- (Core) `service/extensions`: Allow extensions to declare dependencies on other extensions and guarantee start/stop/notification order accordingly. ([#8732](https://github.com/open-telemetry/opentelemetry-collector/issues/8732))
- (Core) `exporters`: Log export errors when retry is not used by the component. ([#8791](https://github.com/open-telemetry/opentelemetry-collector/issues/8791))

### 🧰 Bug fixes 🧰

- (Splunk) `smartagent/processlist`: Reduce CPU usage when collecting process information on Windows ([#3980](https://github.com/signalfx/splunk-otel-collector/pull/3980))
- (Contrib) `filelogreceiver`: Fix issue where counting number of logs emitted could cause panic ([#27469](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27469), [#29107](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29107))
- (Contrib) `kafkareceiver`: Fix issue where counting number of logs emitted could cause panic ([#27469](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27469), [#29107](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29107))
- (Contrib) `k8sobjectsreceiver`: Fix issue where counting number of logs emitted could cause panic ([#27469](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27469), [#29107](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29107))
- (Contrib) `fluentforwardreceiver`: Fix issue where counting number of logs emitted could cause panic ([#27469](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27469), [#29107](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29107))
- (Contrib) `azureeventhubreceiver`: Updated documentation around Azure Metric to OTel mapping. ([#28622](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28622))
- (Contrib) `receiver/hostmetrics`: Fix panic on load_scraper_windows shutdown ([#28678](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28678))
- (Contrib) `splunkhecreceiver`: Do not encode JSON response objects as string. ([#27604](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27604))
- (Contrib) `processor/k8sattributes`: Set attributes from namespace/node labels or annotations even if node/namespaces name attribute are not set. ([#28837](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/28837))
- (Contrib) `pkg/stanza`: Fix data-corruption/race-condition issue in udp async (reuse of buffer); use buffer pool instead. (#27613)
- (Contrib) `zipkinreceiver`: Return BadRequest in case of permanent errors ([#4335](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4335))
- (Core) `exporters`: fix bug with queue size and capacity metrics ([#8682](https://github.com/open-telemetry/opentelemetry-collector/issues/8682))

## v0.88.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.88.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.88.0) and the [opentelemetry-collector-contrib v0.88.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.88.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) `smartagent`: Respect `JAVA_HOME` environment variable instead of enforcing bundle-relative value ([#3877](https://github.com/signalfx/splunk-otel-collector/pull/3877))
- (Contrib) `k8sclusterreceiver`: Remove opencensus.resourcetype resource attribute ([#26487](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26487))
- (Contrib) `splunkhecexporter`: Remove `max_connections` configuration setting. ([#27610](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27610))
  - use `max_idle_conns` or `max_idle_conns_per_host` instead.
- (Contrib) `signalfxexporter`: Remove `max_connections` configuration setting. ([#27610](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27610))
  - use `max_idle_conns` or `max_idle_conns_per_host` instead.
- (Core) `exporters`: make enqueue failures available for otel metrics ([#8673](https://github.com/open-telemetry/opentelemetry-collector/issues/8673)). This will prevent internal Collector `otelcol_exporter_enqueue_failed_<telemetry_type>` metrics from being reported unless greater than 0.


### 💡 Enhancements 💡
- (Splunk) Add an option, `-msi_public_properties`, to allow passing MSI public properties when installing the Splunk OpenTelemetry Collector using the Windows installer script ([#3921](https://github.com/signalfx/splunk-otel-collector/pull/3921))
- (Splunk) Add support for config map providers in discovery configuration. ([#3874](https://github.com/signalfx/splunk-otel-collector/pull/3874))
- (Splunk) Add zero config support for chef deployments ([#3903](https://github.com/signalfx/splunk-otel-collector/pull/3903))
- (Splunk) Add zero config support for puppet deployments ([#3922](https://github.com/signalfx/splunk-otel-collector/pull/3922))
- (Contrib) `receiver/prometheus`: Warn instead of failing when users rename using metric_relabel_configs in the prometheus receiver ([#5001](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/5001))
- (Contrib) `k8sobjectsreceiver`: Move k8sobjectsreceiver from Alpha stability to Beta stability for logs. ([#27635](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27635))
- (Contrib) `pkg/ottl`: Add a double converter ([#22056](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22056))
- (Contrib) `syslogreceiver`: validate protocol name ([#27581](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27581))
- (Contrib) `entension/storage/filestorage`: Add support for setting bbolt fsync option ([#20266](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20266))
- (Contrib) `filelogreceiver`: Add a new "top_n" option to specify the number of files to track when using ordering criteria ([#23788](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23788))
- (Contrib) `k8sclusterreceiver`: add optional k8s.pod.qos_class resource attribute ([#27483](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27483))
- (Contrib) `pkg/stanza`: Log warning, instead of error, when Windows Event Log publisher metadata is not available and cache the successfully retrieved ones. ([#27658](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27658))
- (Contrib) `pkg/ottl`: Add optional Converter parameters to replacement Editors ([#27235](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27235))
- (Contrib) `signalfxexporter`: Add an option to control the dimension client timeout ([#27815](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27815))
- (Contrib) `signalfxexporter`: Add the build version to the user agent of the SignalFx exporter ([#16841](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/16841))

### 🧰 Bug fixes 🧰

- (Splunk) Fix Tanzu Tile to properly set proxy exclusions. ([#3902](https://github.com/signalfx/splunk-otel-collector/pull/3902))
- (Contrib) `syslog`: add integration tests and fix related bugs ([#21245](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21245))
- (Contrib) `processor/resourcedetection`: Don't parse the field `cpuInfo.Model` if it's blank. ([#27678](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27678))
- (Contrib) `k8sclusterreceiver`: Change clusterquota and resourcequota metrics to use {resource} unit ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Contrib) `pkg/ottl`: Fix bug where named parameters needed a space after the equal sign (`=`). ([#28511](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/28511))
- (Contrib) `filelogreceiver`: Fix issue where batching of files could result in ignoring start_at setting. ([#27773](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27773))
- (Core) `exporters`: Fix nil pointer dereference when stopping persistent queue after a start encountered an error ([#8718](https://github.com/open-telemetry/opentelemetry-collector/pull/8718))


### 💡 Enhancements 💡

- (Splunk) Add an option, `-msi_public_properties`, to allow passing MSI public properties when installing the Splunk OpenTelemetry Collector using the Windows installer script ([#3921](https://github.com/signalfx/splunk-otel-collector/pull/3921))

## v0.87.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.87.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.87.0) and the [opentelemetry-collector-contrib v0.87.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.87.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Auto Instrumentation for Linux ([#3791](https://github.com/signalfx/splunk-otel-collector/pull/3791)):
  - The `/usr/lib/splunk-instrumentation/instrumentation.conf` config file is no longer
    supported, and is replaced by `/etc/splunk/zeroconfig/java.conf`. If the `splunk-otel-auto-instrumentation` deb/rpm
    package is manually upgraded, the options within `/usr/lib/splunk-instrumentation/instrumentation.conf` will need to
    be manually migrated to their corresponding environment variables within `/etc/splunk/zeroconfig/java.conf`.
  - Manual installation of the `splunk-otel-auto-instrumentation` deb/rpm package no longer automatically adds
    `/usr/lib/splunk-instrumentation/libsplunk.so` to `/etc/ld.so.preload`.
  - Manual upgrade of the `splunk-otel-auto-instrumentation` deb/rpm package will automatically remove
    `/usr/lib/splunk-instrumentation/libsplunk.so` from `/etc/ld.so.preload`.
  - The `splunk.linux-autoinstr.executions` metric is no longer generated by `libsplunk.so`.
  - See [Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Linux](https://github.com/signalfx/splunk-otel-collector/blob/main/instrumentation/README.md)
    for manual installation/configuration details.
  - For users of the [Ansible](https://galaxy.ansible.com/ui/repo/published/signalfx/splunk_otel_collector/), [Chef](https://supermarket.chef.io/cookbooks/splunk_otel_collector), [Puppet](https://forge.puppet.com/modules/signalfx/splunk_otel_collector), or [Salt](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) modules for Auto Instrumentation, it is recommended to update the following option in your configuration for version `0.86.0` or older until these modules are updated to manage these changes:
    - Ansible: `splunk_otel_auto_instrumentation_version`
    - Chef: `auto_instrumentation_version`
    - Puppet: `auto_instrumentation_version`
    - Salt: `auto_instrumentation_version`
- (Contrib) `kubeletstatsreceiver`: Fixes a bug where the "insecure_skip_verify" config was not being honored when "auth_type" is "serviceAccount" in kubelet client. ([#26319](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26319))
  - Before the fix, the kubelet client was not verifying kubelet's certificate. The default value of the config is false,
    so with the fix the client will start verifying tls cert unless the config is explicitly set to true.
- (Contrib) `tailsamplingprocessor`: Improve counting for the `count_traces_sampled` metric ([#25882](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25882))
- (Contrib) `extension/filestorage`: Replace path-unsafe characters in component names ([#3148](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/3148))
- (Core) `service/telemetry exporter/exporters`: Enable sampling logging by default and apply it to all components. ([#8134](https://github.com/open-telemetry/opentelemetry-collector/pull/8134))
  - The sampled logger configuration can be disabled easily by setting the `service::telemetry::logs::sampling::enabled` to `false`.

### 🚩 Deprecations 🚩

- (Splunk) The following Auto Instrumentation options for the Linux installer script are deprecated and will only apply if the `--instrumentation-version <version>`
  option is specified for version `0.86.0` or older:
  - `--[no-]generate-service-name`: `libsplunk.so` no longer generates service names for instrumented applications. The default behavior is for the activated Java
    and/or Node.js Auto Instrumentation agents to automatically generate service names. Use the `--service-name <name>` option to override the auto-generated service
    names for all instrumented applications.
  - `--[enable|disable]-telemetry`: `libsplunk.so` no longer generates the `splunk.linux-autoinstr.executions` telemetry metric.

### 🚀 New components 🚀

- (Splunk) Add the `loadbalancing` exporter ([#3825](https://github.com/signalfx/splunk-otel-collector/pull/3825))
- (Splunk) Add the `udplog` receiver ([#3826](https://github.com/signalfx/splunk-otel-collector/pull/3826))

### 💡 Enhancements 💡

- (Splunk) Update golang to 1.20.10 ([#3770](https://github.com/signalfx/splunk-otel-collector/pull/3770))
- (Splunk) Add debian 12 support to installer ([#3766](https://github.com/signalfx/splunk-otel-collector/pull/3766))
- (Splunk) Add new Auto Instrumentation options for the Linux installer script ([#3791](https://github.com/signalfx/splunk-otel-collector/pull/3791)):
  - `--with[out]-systemd-instrumentation`: Activate auto instrumentation for only `systemd` services without preloading
    the `libsplunk.so` shared object library (default: `--without-systemd-instrumentation`)
  - Initial support for [Splunk OpenTelemetry Auto Instrumentation for Node.js](https://github.com/signalfx/splunk-otel-js):
    - Activated by default if the `--with-instrumentation` or `--with-systemd-instrumentation` option is specified.
    - Use the `--without-instrumentation-sdk node` option to explicitly skip Node.js.
    - `npm` is required to install the Node.js Auto Instrumentation package. If `npm` is not installed, Node.js will
      be skipped automatically.
    - By default, the Node.js Auto Instrumentation package is installed with the `npm install --global` command. Use the
      `--npm-command "<command>"` option to specify a custom command.
    - Environment variables to activate and configure Node.js auto instrumentation are added to `/etc/splunk/zeroconfig/node.conf` (for `--with-instrumentation`) or
      `/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf` (for `--with-systemd-instrumentation`) based on defaults and specified installation options.
  - Auto Instrumentation for Java is also activated by default if the `--with-instrumentation` or
    `--with-systemd-instrumentation` option is specified. Use the `--without-instrumentation-sdk java` option to skip Java.
  - `--otlp-endpoint host:port`: Set the OTLP gRPC endpoint for captured traces (default: `http://LISTEN_INTERFACE:4317`
    where `LISTEN_INTERFACE` is the value from the `--listen-interface` option if specified, or `127.0.0.1` otherwise)
  - See [Linux Installer Script](https://github.com/signalfx/splunk-otel-collector/blob/main/docs/getting-started/linux-installer.md)
    for more details.
- (Splunk) Update splunk-otel-javaagent to [v1.29.0](https://github.com/signalfx/splunk-otel-java/releases/tag/v1.29.0) ([#3788](https://github.com/signalfx/splunk-otel-collector/pull/3788))
- (Splunk) Redis discovery ([#3731](https://github.com/signalfx/splunk-otel-collector/pull/3731))
- (Splunk) Update Bundled OpenJDK to [11.0.21+9](https://github.com/adoptium/temurin11-binaries/releases/tag/jdk-11.0.21%2B9) ([#3819](https://github.com/signalfx/splunk-otel-collector/pull/3819))
- (Splunk) Oracledb discovery tweaks (remove static endpoint) ([#3836](https://github.com/signalfx/splunk-otel-collector/pull/3836))
- (Contrib) `probabilisticsamplerprocessor`: Allow non-bytes values to be used as the source for the sampling decision ([#18222](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18222))
- (Contrib) `kafkareceiver`: Allow users to attach kafka header metadata with the log/metric/trace record in the pipeline. Introduce a new config param, 'header_extraction' and some examples. ([#24367](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24367))
- (Contrib) `kafkaexporter`: Adding Zipkin encoding option for traces to kafkaexporter ([#21102](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21102))
- (Contrib) `kubeletstatsreceiver`: Support specifying context for `kubeConfig` `auth_type` ([#26665](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26665))
- (Contrib) `kubeletstatsreceiver`: Adds new `k8s.pod.cpu_limit_utilization`, `k8s.pod.cpu_request_utilization`, `k8s.container.cpu_limit_utilization`, and `k8s.container.cpu_request_utilization` metrics that represent the ratio of cpu used vs set limits and requests. ([#27276](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27276))
- (Contrib) `kubeletstatsreceiver`: Adds new `k8s.pod.memory_limit_utilization`, `k8s.pod.memory_request_utilization`, `k8s.container.memory_limit_utilization`, and `k8s.container.memory_request_utilization` metrics that represent the ratio of memory used vs set limits and requests. ([#25894](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25894))

### 🧰 Bug fixes 🧰

- (Contrib) `spanmetricsprocessor`: Prune histograms when dimension cache is pruned. ([#27080](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27080))
  - Dimension cache was always pruned but histograms were not being pruned. This caused metric series created
    by processor/spanmetrics to grow unbounded.
- (Contrib) `splunkhecreceiver`: Fix receiver behavior when used for metrics and logs at the same time; metrics are no longer dropped. ([#27473](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27473))
- (Contrib) `metricstransformprocessor`: Fixes a nil pointer dereference when copying an exponential histogram ([#27409](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27409))
- (contrib) `k8sclusterreceiver`: change k8s.container.ready, k8s.pod.phase, k8s.pod.status_reason, k8s.namespace.phase units to empty ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Contrib) `k8sclusterreceiver`: Change k8s.node.condition* metric units to empty ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Contrib) `syslogreceiver`: Fix issue where long tokens would be truncated prematurely ([#27294](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27294))
- (Core) `telemetry`: remove workaround to ignore errors when an instrument includes a `/` ([#8346](https://github.com/open-telemetry/opentelemetry-collector/issues/8346))

## v0.86.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.86.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.86.0) and the [opentelemetry-collector-contrib v0.86.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.86.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) Set `SPLUNK_LISTEN_INTERFACE` environment variable value to 127.0.0.1 for [agent mode](https://docs.splunk.com/observability/en/gdi/opentelemetry/deployment-modes.html#host-monitoring-agent-mode) by default, as determined by config path. 0.0.0.0 will be set otherwise, with existing environment values respected. The installers have been updated to only set the environment variable for collector service if configured directly (e.g. via `--listen-interface <ip>` or `-network_interface "<ip>"` for Linux or Windows installer script options, respectively) ([#3732](https://github.com/signalfx/splunk-otel-collector/pull/3732))

### 🚩 Deprecations 🚩

- (Core) `loggingexporter`: Mark the logging exporter as deprecated, in favour of debug exporter ([#7769](https://github.com/open-telemetry/opentelemetry-collector/issues/7769))

### 🚀 New components 🚀

- (Splunk) enabling in-development `scriptedinputs` receiver in components ([#3627](https://github.com/signalfx/splunk-otel-collector/pull/3627))
- (Core) `debugexporter`: Add debug exporter, which replaces the logging exporter ([#7769](https://github.com/open-telemetry/opentelemetry-collector/issues/7769))

### 💡 Enhancements 💡

- (Splunk) Oracledb discovery ([#3633](https://github.com/signalfx/splunk-otel-collector/pull/3633))
- (Splunk) include debug exporter ([#3735](https://github.com/signalfx/splunk-otel-collector/pull/3735))
- (Splunk) Update bundled python to 3.11.6 ([#3727](https://github.com/signalfx/splunk-otel-collector/pull/3727))
- (Splunk) Switch pulsar exporter to contrib ([#3641](https://github.com/signalfx/splunk-otel-collector/pull/3641))
- (Splunk) demonstrate filelog receiver config equivalent to Splunk Addon for Unix and Linux File and Directory Inputs ([#3271](https://github.com/signalfx/splunk-otel-collector/pull/3271))
- (Splunk) remove unused Smart Agent package code (#3676, #3678, #3685, #3686, #3687, #3688, #3689, #3702, #3703, and #3706)
- (Contrib) `processor/tailsampling`: Allow sub-second decision wait time ([#26354](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26354))
- (Contrib) `processor/resourcedetection`: Added support for host's cpuinfo attributes. ([#26532](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26532))
  In Linux and Darwin all fields are populated. In Windows only family, vendor.id and model.name are populated.
- (Contrib) `pkg/stanza`: Add 'omit_pattern' setting to `split.Config`. ([#26381](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26381))
  This can be used omit the start or end pattern from a log entry.
- (Contrib) `statsdreceiver`: Add TCP support to statsdreceiver ([#23327](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23327))
- (Contrib) `statsdreceiver`: Allow for empty tag sets ([#27011](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27011))
- (Contrib) `pkg/ottl`: Update contexts to set and get time.Time ([#22010](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22010))
- (Contrib) `pkg/ottl`: Add a Now() function to ottl that returns the current system time ([#27038](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27038), [#26507](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26507))
- (Contrib) `filelogreceiver`: Log the globbing IO errors ([#23768](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23768))
- (Contrib) `pkg/ottl`: Allow named arguments in function invocations ([#20879](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20879))
  Arguments can now be specified by a snake-cased version of their name in the function's
  `Arguments` struct. Named arguments can be specified in any order, but must be specified
  after arguments without a name.
- (Contrib) `pkg/ottl`: Add new `TruncateTime` function to help with manipulation of timestamps ([#26696](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26696))
- (Contrib) `pkg/stanza`: Add 'overwrite_text' option to severity parser. ([#26671](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26671))
  Allows the user to overwrite the text of the severity parser with the official string representation of the severity level.
- (Contrib) `prometheusreceiver`: add a new flag, enable_protobuf_negotiation, which enables protobuf negotiation when scraping prometheus clients ([#27027](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27027))
- (Contrib) `redisreceiver`: Added `redis.cmd.latency` metric. ([#6942](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6942))
- (Contrib) `processor/resourcedetectionprocessor`: add k8snode detector to provide node metadata; currently the detector provides `k8d.node.uid` ([#26538](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26538))
- (Contrib) `splunkhecreceiver`: Update splunk hec receiver to extract time query parameter if it is provided ([#27006](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27006))
- (Contrib) `processor/k8sattributes`: allow metadata extractions to be set to empty list ([#14452](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14452))

### 🧰 Bug fixes 🧰

- (Contrib) `processor/tailsampling`: Prevent the tail-sampling processor from accepting duplicate policy names ([#27016](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27016))
- (Contrib) `k8sclusterreceiver`: Change k8s.deployment.available and k8s.deployment.desired metric units to {pod} ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Contrib) `k8sclusterreceiver`: Change k8scluster receiver metric units to follow otel semantic conventions ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Contrib) `pkg/stanza`: Fix bug where force_flush_period not applied ([#26691](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26691))
- (Contrib) `filelogreceiver`: Fix issue where truncated file could be read incorrectly. ([#27037](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27037))
- (Contrib) `receiver/hostmetricsreceiver`: Make sure the process scraper uses the gopsutil context, respecting the `root_path` configuration. ([#24777](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24777))
- (Contrib) `k8sclusterreceiver`: change k8s.container.restarts unit from 1 to {restart} ([#10553](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/10553))
- (Core) `configtls`: fix incorrect use of fsnotify ([#8438](https://github.com/open-telemetry/opentelemetry-collector/issues/8438))

## v0.85.0

***ADVANCED NOTICE - SPLUNK_LISTEN_INTERFACE DEFAULTS***

Starting with version 0.86.0 (next release), the collector installer will change the default value of the network listening interface option from `0.0.0.0` to `127.0.0.1`.

### 🛑 Breaking changes 🛑

- (Contrib) `k8sclusterreceiver`: Remove deprecated Kubernetes API resources ([#23612](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23612), [#26551](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26551))
Drop support of `HorizontalPodAutoscaler` `v2beta2` version and `CronJob` `v1beta1` version.
Note that metrics for those resources will not be emitted anymore on Kubernetes 1.22 and older.
- (Contrib) `prometheusexporters`: Append prometheus type and unit suffixes by default in prometheus exporters. ([#26488](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26488))
Suffixes can be disabled by setting add_metric_suffixes to false on the exporter.
- (Contrib) `attributesprocessor`, `resourceprocessor`: Transition featuregate `coreinternal.attraction.hash.sha256` to stable ([#4759](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/4759))

### 💡 Enhancements 💡

- (Splunk) `wavefrontreceiver`: Add wavefrontreceiver ([#3629](https://github.com/signalfx/splunk-otel-collector/pull/3629))
- (Splunk) Update `splunk-otel-javaagent` to 1.28.0 ([#3647](https://github.com/signalfx/splunk-otel-collector/pull/3647))
- (Contrib) `postgresqlreceiver`: Added postgresql.database.locks metric. ([#26317](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26317))
- (Contrib) `receiver/statsdreceiver`: Add support for distribution type metrics in the statsdreceiver. ([#24768](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24768))
- (Contrib) `pkg/ottl`: Add converters to convert time to unix nanoseconds, unix microseconds, unix milliseconds or unix seconds ([#24686](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24686))
- (Contrib) `receiver/hostmetrics`: Don't collect connections data from the host if system.network.connections metric is disabled to not waste CPU cycles. ([#25815](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25815))
- (Contrib) `jaegerreceiver`,`jaegerremotesamplingextension`: Add featuregates to replace Thrift-gen with Proto-gen types for sampling strategies ([#18401](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18401))

  Available featuregates are:
  * `extension.jaegerremotesampling.replaceThriftWithProto`
  *  `receiver.jaegerreceiver.replaceThriftWithProto`
- (Contrib) `k8sclusterreceiver`: Add optional `k8s.kubelet.version`, `k8s.kubeproxy.version` node resource attributes ([#24835](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24835))
- (Contrib) `k8sclusterreceiver`: Add `k8s.pod.status_reason` option metric ([#24034](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24034))
- (Contrib) `k8sobjectsreceiver`: Adds logic to properly handle 410 response codes when watching. This improves the reliability of the receiver. ([#26098](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26098))
- (Contrib) `k8sobjectreceiver`: Adds option to exclude event types (`MODIFIED`, `DELETED`, etc) in watch mode. ([#26042](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/26042))
- (Core) `confighttp`: Add option to disable HTTP keep-alives ([#8260](https://github.com/open-telemetry/opentelemetry-collector/issues/8260))

### 🧰 Bug fixes 🧰

- (Splunk) `fluentd`: Update fluentd url for windows ([#3635](https://github.com/signalfx/splunk-otel-collector/pull/3635))
- (Contrib) `processor/routing`: When using attributes instead of resource attributes, the routing processor would crash the collector. This does not affect the connector version of this component. ([#26462](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26462))
- (Contrib) `processor/tailsampling`: Added saving instrumentation library information for tail-sampling ([#13642](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/13642))
- (Contrib) `receiver/kubeletstats`: Fixes client to refresh service account token when authenticating with kubelet ([#26120](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26120))
- (Contrib) `filelogreceiver`: Fix the behavior of the add operator to continue to support `EXPR(env("MY_ENV_VAR"))` expressions ([#26373](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26373))
- (Contrib) `pkg/stanza`: Fix issue unsupported type 'syslog_parser' ([#26452](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26452))
- (Core) `confmap`: fix bugs of unmarshalling slice values ([#4001](https://github.com/open-telemetry/opentelemetry-collector/issues/4001))

## v0.84.0

### 🛑 Breaking changes 🛑

- (Contrib) `jaegerreceiver`: Deprecate remote_sampling config in the jaeger receiver ([#24186](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24186))
  The jaeger receiver will fail to start if remote_sampling config is specified in it.  The `receiver.jaeger.DisableRemoteSampling` feature gate can be set to let the receiver start and treat  remote_sampling config as no-op. In a future version this feature gate will be removed and the receiver will always  fail when remote_sampling config is specified.

### 💡 Enhancements 💡

- (Splunk) `jmxreceiver`: Bundle latest [JMX Metric Gatherer](https://github.com/open-telemetry/opentelemetry-java-contrib/tree/main/jmx-metrics) in installer packages and images for Windows and Linux ([#3262](https://github.com/signalfx/splunk-otel-collector/pull/3262))
- (Splunk) `solacereceiver`: Added solace receiver to the splunk otel collector ([#3590](https://github.com/signalfx/splunk-otel-collector/pull/3590))
- (Splunk) `receiver/smartagent`: Move to gopsutil 3.23.7 and remove the need to set environment variables ([#3509](https://github.com/signalfx/splunk-otel-collector/pull/3509))
- (Splunk) Update splunk-otel-javaagent to 1.27.0 ([#3537](https://github.com/signalfx/splunk-otel-collector/pull/3537))
- (Splunk) `receiver/smartagent`: Use `Leases` instead of `ConfigMapLeases` for leader-election in k8s. ([#3521](https://github.com/signalfx/splunk-otel-collector/pull/3521))
- (Splunk) Update bundled python to 3.11.5 ([#3543](https://github.com/signalfx/splunk-otel-collector/pull/3543))
- (Contrib) `redisreceiver`: Adding username parameter for connecting to redis ([#24408](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24408))
- (Contrib) `postgresqlreceiver`: Added `postgresql.temp_files` metric. ([#26080](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26080))
- (Contrib) `signalfxexporter`: Added a mechanism to drop histogram buckets ([#25845](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25845))
- (Contrib) `journaldreceiver`: add support for identifiers ([#20295](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20295))
- (Contrib) `journaldreceiver`: add support for dmesg ([#20295](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20295))
- (Contrib) `pkg/ottl`: Add converters to covert duration to nanoseconds, microseconds, milliseconds, seconds, minutes or hours ([#24686](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24686))
- (Contrib) `snmpreceiver`: Support scalar OID resource attributes ([#23373](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23373))
  Add column and scalar OID metrics to resources that have scalar OID attributes
- (Contrib) `kubeletstatsreceiver`: Add a new `uptime` metric for nodes, pods, and containers to track how many seconds have passed since the object started  ([#25867](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25867))
- (Contrib) `pkg/ottl`: Add new `ExtractPatterns` converter that extract regex pattern from string.  ([#25834](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25834), [#25856](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25856))
- (Contrib) `pkg/ottl`: Add support for Log, Metric and Trace Slices to `Len` converter ([#25868](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25868))
- (Contrib) `postgresqlreceiver`: Added `postgresql.deadlocks` metric. ([#25688](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/25688))
- (Contrib) `postgresqlreceiver`: Added `postgresql.sequential_scans` metric. ([#26096](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26096))
- (Contrib) `prometheusreceiver`: The otel_scope_name and otel_scope_version labels are used to populate scope name and version. otel_scope_info is used to populate scope attributes. ([#25870](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20295))
- (Contrib) `receiver/prometheus`: translate units from prometheus to UCUM ([#23208](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23208))
- (Core) `loggingexporter`: Adds exemplars logging to the logging exporter when `detailed` verbosity level is set. ([#7912](https://github.com/open-telemetry/opentelemetry-collector/issues/7912))
- (Core) `configgrpc`: Allow any registered gRPC load balancer name to be used. ([#8262](https://github.com/open-telemetry/opentelemetry-collector/issues/8262))
- (Core) `service`: add OTLP export for internal traces ([#8106](https://github.com/open-telemetry/opentelemetry-collector/issues/8106))
- (Core) `configgrpc`: Add support for :authority pseudo-header in grpc client ([#8228](https://github.com/open-telemetry/opentelemetry-collector/issues/8228))

### 🧰 Bug fixes 🧰

- (Core) `otlphttpexporter`: Fix the handling of the HTTP response to ignore responses not encoded as protobuf ([#8263](https://github.com/open-telemetry/opentelemetry-collector/issues/8263))
- (Contrib) `receiver_creator`: Update expr and relocate breaking `type` function to `typeOf` ([#26038](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26038))
- (Splunk) `deployment/cloudfoundry`: Add missing system resource detection ([#3541](https://github.com/signalfx/splunk-otel-collector/pull/3541))

## v0.83.0

### 🛑 Breaking changes 🛑

- (Splunk) Fluentd installation ***disabled*** by default for the [`splunk-otel-collector` salt formula](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/salt) ([#3448](https://github.com/signalfx/splunk-otel-collector/pull/3448))
  - Specify the `install_fluentd: True` attribute in your pillar to enable installation
- (Splunk/Contrib) Removes the deprecated `receiver/prometheus_exec` receiver. Please see [migration guide](docs/deprecations/migrating-from-prometheus-exec-to-prometheus.md) for further details. ([#24740](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24740)) ([#3512](https://github.com/signalfx/splunk-otel-collector/pull/3512))
- (Contrib) `receiver/k8scluster`: Unify predefined and custom node metrics. ([#24776](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24776))
  - Update metrics description and units to be consistent
  - Remove predefined metrics definitions from metadata.yaml because they are controlled by `node_conditions_to_report`
    and `allocatable_types_to_report` config options.

### 💡 Enhancements 💡

- (Splunk) Use `SPLUNK_LISTEN_INTERFACE` and associated installer option to configure the network interface used by the collector for default configurations ([#3421](https://github.com/signalfx/splunk-otel-collector/pull/3421))
  - Existing installations will rely on the default value of `SPLUNK_LISTEN_INTERFACE` set to `0.0.0.0`. Users must add `SPLUNK_LISTEN_INTERFACE` to their collector configuration to take advantage of this new option.
- (Contrib) `receiver/collectdreceiver`: Migrate from opencensus to pdata, change collectd, test to match pdata format. ([#20760](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20760))
- (Contrib) `pkg/ottl`: Add support for using addition and subtraction with time and duration ([#22009](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22009))
- (Contrib) `transformprocessor`: Add extract_count_metric OTTL function to transform processor ([#22853](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22853))
- (Contrib) `transformprocessor`: Add extract_sum_metric OTTL function to transform processor ([#22853](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22853))
- (Contrib) `prometheusreceiver`: Don't drop histograms without buckets ([#22070](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22070))
- (Contrib) `pkg/ottl`: Add a new Function Getter to the OTTL package, to allow passing Converters as literal parameters. ([#22961](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22961))
  Currently OTTL provides no way to use any defined Converter within another Editor/Converter.
  Although Converters can be passed as a parameter, they are always executed and the result is what is actually passed as the parameter.
  This allows OTTL to pass Converters themselves as a parameter so they can be executed within the function.
- (Contrib) `resourcedetectionprocessor`: GCP resource detection processor can automatically add `gcp.gce.instance.hostname` and `gcp.gce.instance.name` attributes. ([#24598](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24598))
- `splunkhecexporter`: Add heartbeat check while startup and new config param, heartbeat/startup (defaults to false). This is different than the healtcheck_startup, as the latter doesn't take token or index into account. ([#24411](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24411))
- (Contrib) `hostmetricsreceiver`: Report  logical and physical number of CPUs as metric. ([#22099](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22099))
  - Use the `system.cpu.logical.count::enabled` and `system.cpu.physical.count::enabled` flags to enable them
- (Contrib) `k8sclusterreceiver`: Allows disabling metrics and resource attributes ([#24568](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24568))
- (Contrib) `k8sclusterreceiver`: Reduce memory utilization ([#24769](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24769))
- (Contrib) `k8sattributes`: Added k8s.cluster.uid to k8sattributes processor to add cluster uid ([#21974](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21974))
- (Contrib) `resourcedetectionprocessor`: Collect heroku metadata available instead of exiting early. Log at debug level if metadata is missing to help troubleshooting. ([#25059](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25059))
- (Contrib) `hostmetricsreceiver`: Improved description of the system.cpu.utilization metrics. ([#25115](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25115))
- (Contrib) `cmd/mdatagen`: Avoid reusing the same ResourceBuilder instance for multiple ResourceMetrics ([#24762](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24762))
- (Contrib) `resourcedetectionprocessor`: Add detection of os.description to system detector ([#24541](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24541))
- (Contrib) `filelogreceiver`: Bump 'filelog.allowHeaderMetadataParsing' feature gate to beta ([#18198](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/18198))
- (Contrib) `receiver/prometheusreceiver`: Add config `report-extra-scrape-metrics` to report additional prometheus scraping metrics ([#21040](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21040))
  - Emits additional metrics - scrape_body_size_bytes, scrape_sample_limit, scrape_timeout_seconds. scrape_body_size_bytes metric can be used for checking failed scrapes due to body-size-limit.
- (Contrib) `receiver/sqlquery`: Set ObservedTimestamp on collected logs ([#23776](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23776))
- (Core) `extension`: Add optional `ConfigWatcher` interface ([#6596](https://github.com/open-telemetry/opentelemetry-collector/issues/6596))
  - Extensions implementing this interface will be notified of the Collector's effective config.
- (Core) `otelcol`: Add optional `ConfmapProvider` interface for Config Providers ([#6596](https://github.com/open-telemetry/opentelemetry-collector/issues/6596))
  - This allows providing the Collector's configuration as a marshaled confmap.Conf object from a ConfigProvider
- (Core) `service`: Add `CollectorConf` field to `service.Settings` ([#6596](https://github.com/open-telemetry/opentelemetry-collector/issues/6596))
  This field is intended to be used by the Collector to pass its effective configuration to the service.

### 🧰 Bug fixes 🧰

- (Contrib) `carbonreceiver`: Fix Carbon receiver obsrecv operations memory leak ([#24275](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24275))
  - The carbonreceiver has a memory leak where it will repeatedly open new obsrecv operations but not close them afterwards. Those operations eventually create a burden. The fix is to make sure the receiver only creates an operation per interaction over TCP.
- (Contrib) `pkg/stanza`: Create a new decoder for each TCP/UDP connection to prevent concurrent write to buffer. ([#24980](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24980))
- (Contrib) `exporter/kafkaexporter`: Fixes a panic when SASL configuration is not present ([#24797](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24797))
- (Contrib) `receiver/k8sobjects`: Fix bug where duplicate data would be ingested for watch mode if the client connection got reset. ([#24806](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24806))
- (Contrib) `zipkinreceiver`: Respects zipkin's serialised status tags to be converted to span status ([#14965](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/14965))
- (Contrib) `processor/resourcedetection`: Do not drop all system attributes if `host.id` cannot be fetched. ([#24669](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24669))
- (Contrib) `signalfxexporter`: convert vmpage_io* translated metrics to pages ([#25064](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/25064))
- (Contrib) `splunkhecreceiver`: aligns success resp body w/ splunk enterprise ([#19219](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/19219))
  - changes resp from plaintext "ok" to json {"text"："success", "code"：0}

## v0.82.0

### 🛑 Breaking changes 🛑

- (Splunk) Fluentd installation ***disabled*** by default for the Linux and Windows installer scripts ([#3369](https://github.com/signalfx/splunk-otel-collector/pull/3369))
  - Specify the `--with-fluentd` (Linux) or `with_fluentd = 1` (Windows) option to enable installation
- (Splunk) Fluentd installation ***disabled*** by default for the Windows Chocolatey package ([#3377](https://github.com/signalfx/splunk-otel-collector/pull/3377))
  - Specify the `/WITH_FLUENTD:true` parameter to enable installation
- (Contrib) `receiver/prometheus`: Remove unused `buffer_period` and `buffer_count` configuration options ([#24258](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24258))
- (Contrib) `receiver/prometheus`: Add the `trim_metric_suffixes` configuration option to allow enable metric suffix trimming.  ([#21743](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21743), [#8950](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/8950))
  - When enabled, suffixes for unit and type are trimmed from metric names. If you previously enabled the `pkg.translator.prometheus.NormalizeName` feature gate, you will need to enable this option to have suffixes trimmed.

### 💡 Enhancements 💡

- (Core) `service`: Add support for exporting internal metrics to the console ([#7641](https://github.com/open-telemetry/opentelemetry-collector/issues/7641))
  - Internal collector metrics can now be exported to the console using the otel-go stdout exporter.
- (Core) `service`: Add support for interval and timeout configuration in periodic reader ([#7641](https://github.com/open-telemetry/opentelemetry-collector/issues/7641))
- (Core) `service`: Add support for OTLP export for internal metrics ([#7641](https://github.com/open-telemetry/opentelemetry-collector/issues/7641))
  - Internal collector metrics can now be exported via OTLP using the otel-go otlpgrpc and otlphttp exporters.
- (Core) `scraperhelper`: Adding optional timeout field to scrapers ([#7951](https://github.com/open-telemetry/opentelemetry-collector/pull/7951))
- (Core) `receiver/otlp`: Add http url paths per signal config options to otlpreceiver ([#7511](https://github.com/open-telemetry/opentelemetry-collector/issues/7511))
- (Core) `exporter/otlphttp`: Add support for trailing slash in endpoint URL ([#8084](https://github.com/open-telemetry/opentelemetry-collector/issues/8084))
  - URLs like http://localhost:4318/ will now be treated as if they were http://localhost:4318
- (Contrib) `processor/resourcedetection`: Add an option to add `host.arch` resource attributio in `system` detector semantic convention ([#22939](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22939))
- (Contrib) `pkg/ottl`: Add new Len converter that computes the length of strings, slices, and maps. ([#23847](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/23847))
- (Contrib) `pkg/ottl`: Improve error reporting for errors during statement parsing ([#23840](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/23840))
  - Failures are now printed for all statements within a context, and the statements are printed next to the errors.
  - Erroneous values found during parsing are now quoted in error logs.
- (Contrib) `processor/routing`: Enables processor to extract metadata from client.Info ([#20913](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/20913))
- (Contrib) `processor/transform`: Report all errors from parsing OTTL statements ([#24245](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24245))

### 🧰 Bug fixes 🧰

- (Contrib) `receiver/prometheus`: Don't fail the whole scrape on invalid data ([#24030](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24030))
- (Contrib) `pkg/stanza`: Fix issue where nil body would be converted to string ([#24017](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24017))
- (Contrib) `pkg/stanza`: Fix issue where syslog input ignored enable_octet_counting setting ([#24073](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24073))
- (Contrib) `receiver/filelog`: Fix issue where files were deduplicated unnecessarily ([#24235](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24235))
- (Contrib) `processor/tailsamplingprocessor`: Fix data race when accessing spans during policies evaluation ([#24283](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24283))
- (Contrib) `zipkintranslator`: Stop dropping error tags from Zipkin spans. The old code removes all errors from those spans, rendering them useless if an actual error happened. In addition, no longer delete error tags if they contain useful information ([#16530](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/16530))

## v0.81.1

### 🧰 Bug fixes 🧰

- (Splunk) Discovery mode: Ensure all successful observers are used in resulting receiver creator instance ([#3391](https://github.com/signalfx/splunk-otel-collector/pull/3391))
- (Contrib) `processor/resourcedetection`: Fix panic when AKS detector is used. ([#24549](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24549))
- (Contrib) `processor/resourcedetection`: Avoid returning empty `host.id` by the `system` detector. ([#24230](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24230))
- (Contrib) `processor/resourcedetection`: Disable `host.id` by default on the `system` detector. This restores the behavior prior to v0.72.0 when using the `system` detector together with other detectors that set `host.id`. ([#21233](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21233))
  To re-enable `host.id` on the `system` detector set `system::resource_attributes::host.id::enabled` to `true`:
  ```
  resourcedetection:
    detectors: [system]
    system:
      resource_attributes:
        host.id:
          enabled: true
  ```
- (Contrib) `processor/resourcedetection`: Fix docker detector not setting any attributes. ([#24280](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/24280))
- (Contrib) `processor/resourcedetection`: Fix Heroku config option for the `service.name` and `service.version` attributes. ([#24355](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24355))

### 💡 Enhancements 💡

- (Splunk) Add support for basicauth extension. ([#3413](https://github.com/signalfx/splunk-otel-collector/pull/3413))
- (Splunk) `receiver/databricks`: Add retry/backoff on http 429s. ([#3374](https://github.com/signalfx/splunk-otel-collector/pull/3374))
- (Contrib) `processor/resourcedetection`: The system detector now can optionally set the `host.arch` resource attribute. ([#22939](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22939))

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
- (Core) `exporters`: Fix persistent storage behaviour with no available space on device ([#7198](https://github.com/open-telemetry/opentelemetry-collector/issues/7198))

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
- Include [tail_sampling](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor) and [span_metrics](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.95.0/processor/spanmetricsprocessor) in our distribution

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

- [`transform` processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor) to modify telemetry based on configuration using the [Telemetry Transformation Language](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/ottl) (Alpha)

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
- Initial [support bundle PowerShell script](https://github.com/signalfx/splunk-otel-collector/blob/main/packaging/msi/splunk-support-bundle.ps1); included in the Windows MSI (#654)
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
