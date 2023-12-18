# Changelog

## Unreleased

## v0.91.1

### 💡 Enhancements 💡

- (Splunk) Remove the project beta label ([#4070](https://github.com/signalfx/splunk-otel-collector/pull/4070))
- (Splunk) Source SPLUNK_LISTEN_INTERFACE on all host endpoints([#4065](https://github.com/signalfx/splunk-otel-collector/pull/4065))
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
- (Core) `exporterhelper`: fix missed metric aggregations ([#9048](https://github.com/open-telemetry/opentelemetry-collector/issues/9048))
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
- (Core) `exporterhelper`: Fix shutdown logic in persistent queue to not require consumers to be closed first ([#8899](https://github.com/open-telemetry/opentelemetry-collector/issues/8899))
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
- (Core) `exporterhelper`: Fix invalid write index updates in the persistent queue ([#8115](https://github.com/open-telemetry/opentelemetry-collector/issues/8115))
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
- (Core) `exporterhelper`: Log export errors when retry is not used by the component. ([#8791](https://github.com/open-telemetry/opentelemetry-collector/issues/8791))

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
- (Core) `exporterhelper`: fix bug with queue size and capacity metrics ([#8682](https://github.com/open-telemetry/opentelemetry-collector/issues/8682))


## v0.88.0

This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector v0.88.0](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/v0.88.0) and the [opentelemetry-collector-contrib v0.88.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.88.0) releases where appropriate.

### 🛑 Breaking changes 🛑

- (Splunk) `smartagent`: Respect `JAVA_HOME` environment variable instead of enforcing bundle-relative value ([#3877](https://github.com/signalfx/splunk-otel-collector/pull/3877))
- (Contrib) `k8sclusterreceiver`: Remove opencensus.resourcetype resource attribute ([#26487](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/26487))
- (Contrib) `splunkhecexporter`: Remove `max_connections` configuration setting. ([#27610](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27610))
  - use `max_idle_conns` or `max_idle_conns_per_host` instead.
- (Contrib) `signalfxexporter`: Remove `max_connections` configuration setting. ([#27610](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/27610))
  - use `max_idle_conns` or `max_idle_conns_per_host` instead.
- (Core) `exporterhelper`: make enqueue failures available for otel metrics ([#8673](https://github.com/open-telemetry/opentelemetry-collector/issues/8673)). This will prevent internal Collector `otelcol_exporter_enqueue_failed_<telemetry_type>` metrics from being reported unless greater than 0.


### 💡 Enhancements 💡
- (Splunk) Add an option, `-msi_public_properties`, to allow passing MSI public properties when installing the Splunk OpenTelemetry Collector using the Windows installer script ([#3921](https://github.com/signalfx/splunk-otel-collector/pull/3921))
- (Splunk) Add support for config map providers in discovery configuration. ([#3874](https://github.com/signalfx/splunk-otel-collector/pull/3874))
- (Splunk) Add zero config support for chef deployments ([#3903](https://github.com/signalfx/splunk-otel-collector/pull/3903))
- (Splunk) Add zero config support for puppet deployments ([#3922](https://github.com/signalfx/splunk-otel-collector/pull/3922))
- (Contrib) `receiver/prometheus`: Warn instead of failing when users rename using metric_relabel_configs in the prometheus receiver ([#5001](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/5001))
- (Contrib) `k8sobjectsreceiver`: Move k8sobjectsreceiver from Alpha stability to Beta stability for logs. ([#27635](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/27635))
- (Contrib) `doubleconverter`: Adding a double converter into pkg/ottl ([#22056](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/22056))
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
- (Core) `exporterhelper`: Fix nil pointer dereference when stopping persistent queue after a start encountered an error ([#8718](https://github.com/open-telemetry/opentelemetry-collector/pull/8718))


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
- (Core) `service/telemetry exporter/exporterhelper`: Enable sampling logging by default and apply it to all components. ([#8134](https://github.com/open-telemetry/opentelemetry-collector/pull/8134))
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
- (Contrib) `exporter/prometheusremotewrite`: Improve the latency and memory utilisation of the conversion from OpenTelemetry to Prometheus remote write ([#24288](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/24288))
- (Contrib) `exporter/prometheusremotewrite`, `exporter/prometheus`: Add `add_metric_suffixes` configuration option, which can disable the addition of type and unit suffixes. ([#21743](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21743), [#8950](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/8950))
- (Contrib) `exporter/prometheusremotewrite`: Downscale exponential histograms to fit prometheus native histograms if necessary ([#17565](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/17565))
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
