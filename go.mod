module github.com/signalfx/splunk-otel-collector

go 1.21

require (
	github.com/alecthomas/participle/v2 v2.1.1
	github.com/antonmedv/expr v1.15.5
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-zookeeper/zk v1.0.3
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/vault v1.15.6
	github.com/hashicorp/vault-plugin-auth-gcp v0.16.2
	github.com/hashicorp/vault/api v1.12.0
	github.com/jaegertracing/jaeger v1.55.0
	github.com/knadh/koanf v1.5.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/pulsarexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.96.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.96.0
	github.com/prometheus/client_model v0.6.0
	github.com/prometheus/common v0.50.0
	github.com/prometheus/prometheus v0.48.1
	github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension v0.83.0
	github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor v0.83.0
	github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver v0.83.0
	github.com/signalfx/splunk-otel-collector/tests v0.83.0
	github.com/spf13/cast v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	go.etcd.io/bbolt v1.3.9
	go.etcd.io/etcd/client/v2 v2.305.12
	go.opentelemetry.io/collector/config/confighttp v0.96.0
	go.opentelemetry.io/collector/config/configtelemetry v0.96.0
	go.opentelemetry.io/collector/confmap v0.96.0
	go.opentelemetry.io/collector/confmap/converter/expandconverter v0.96.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v0.96.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v0.96.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v0.96.0
	go.opentelemetry.io/collector/connector v0.96.0
	go.opentelemetry.io/collector/connector/forwardconnector v0.96.0
	go.opentelemetry.io/collector/exporter v0.96.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.96.0
	go.opentelemetry.io/collector/exporter/loggingexporter v0.96.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.96.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.96.0
	go.opentelemetry.io/collector/extension v0.96.0
	go.opentelemetry.io/collector/extension/ballastextension v0.96.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.96.0
	go.opentelemetry.io/collector/otelcol v0.96.0
	go.opentelemetry.io/collector/pdata v1.3.0
	go.opentelemetry.io/collector/processor v0.96.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.96.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.96.0
	go.opentelemetry.io/collector/receiver v0.96.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.96.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sys v0.18.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/cloudsqlconn v1.7.0 // indirect
	cloud.google.com/go/kms v1.15.7 // indirect
	cloud.google.com/go/monitoring v1.18.0 // indirect
	dario.cat/mergo v1.0.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4 v4.2.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2 v2.2.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/GehirnInc/crypt v0.0.0-20200316065508-bb7000b8a962 // indirect
	github.com/IBM/sarama v1.43.0 // indirect
	github.com/Jeffail/gabs/v2 v2.7.0 // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.62.690 // indirect
	github.com/apache/arrow/go/v14 v14.0.2 // indirect
	github.com/apache/pulsar-client-go v0.11.0 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20240124082744-24bca3a5b39b // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/circonus-labs/circonusllhist v0.1.5 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/denverdino/aliyungo v0.0.0-20230411124812-ab98a9173ace // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/duosecurity/duo_api_golang v0.0.0-20240205144049-bb361ad4ae1c // indirect
	github.com/expr-lang/expr v1.16.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/go-openapi/errors v0.21.1 // indirect
	github.com/go-openapi/strfmt v0.22.2 // indirect
	github.com/go-openapi/validate v0.23.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-metrics-stackdriver v0.6.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/haimrubinstein/go-syslog/v3 v3.0.0 // indirect
	github.com/hashicorp/eventlogger v0.2.9 // indirect
	github.com/hashicorp/go-bexpr v0.1.14 // indirect
	github.com/hashicorp/go-discover v0.0.0-20230724184603-e89ebd1b2f65 // indirect
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.16 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/aead/v2 v2.0.9 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/alicloudkms/v2 v2.0.3 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2 v2.0.9 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2 v2.0.11 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2 v2.0.11 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/ocikms/v2 v2.0.8 // indirect
	github.com/hashicorp/go-kms-wrapping/wrappers/transit/v2 v2.0.11 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.2 // indirect
	github.com/hashicorp/go-raftchunking v0.7.0 // indirect
	github.com/hashicorp/go-secure-stdlib/plugincontainer v0.3.0 // indirect
	github.com/hashicorp/go-secure-stdlib/tlsutil v0.1.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcp-sdk-go v0.86.0 // indirect
	github.com/hashicorp/mdns v1.0.5 // indirect
	github.com/hashicorp/raft v1.6.1 // indirect
	github.com/hashicorp/raft-boltdb/v2 v2.3.0 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.4.0 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jefferai/jsonx v1.0.1 // indirect
	github.com/joshlf/go-acl v0.0.0-20200411065538-eae00ae38531 // indirect
	github.com/joyent/triton-go v1.8.5 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/knadh/koanf/v2 v2.1.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/microsoft/go-mssqldb v1.7.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/okta/okta-sdk-golang/v2 v2.20.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/collectd v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azure v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.96.0 // indirect
	github.com/ovh/go-ovh v1.4.3 // indirect
	github.com/packethost/packngo v0.31.0 // indirect
	github.com/petermattis/goid v0.0.0-20231207134359-e60b3f734c67 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.6 // indirect
	github.com/pquerna/otp v1.4.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	github.com/rboyer/safeio v0.2.3 // indirect
	github.com/redis/go-redis/v9 v9.4.0 // indirect
	github.com/relvacode/iso8601 v1.4.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/signalfx/golib/v3 v3.3.53 // indirect
	github.com/signalfx/signalfx-agent v1.0.1-0.20230222185249-54e5d1064c5b // indirect
	github.com/softlayer/softlayer-go v1.1.3 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible // indirect
	github.com/testcontainers/testcontainers-go v0.29.1 // indirect
	github.com/tg123/go-htpasswd v1.2.2 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/twmb/murmur3 v1.1.7 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
	go.opentelemetry.io/collector v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.96.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.3.0 // indirect
	go.opentelemetry.io/collector/config/configretry v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.96.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.96.0 // indirect
	go.opentelemetry.io/collector/confmap/provider/httpprovider v0.96.0 // indirect
	go.opentelemetry.io/collector/confmap/provider/httpsprovider v0.96.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.96.0 // indirect
	go.opentelemetry.io/collector/service v0.96.0 // indirect
	go.opentelemetry.io/contrib/config v0.4.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.46.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240308144416-29370a3891b7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240308144416-29370a3891b7 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
	sigs.k8s.io/controller-runtime v0.17.2 // indirect
)

require (
	cloud.google.com/go/compute v1.25.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20211115184647-b584dd5df32c // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	collectd.org v0.5.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/AthenZ/athenz v1.10.39 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.6.2 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-amqp v1.0.4 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.21.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/SAP/go-hdb v1.8.9 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/apache/thrift v0.19.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.50.35 // indirect
	github.com/aws/aws-sdk-go-v2 v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.16.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.51.4 // indirect
	github.com/aws/smithy-go v1.20.1 // indirect
	github.com/beevik/ntp v1.3.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/cncf/xds/go v0.0.0-20240306133729-91a88dc4e959 // indirect
	github.com/containerd/containerd v1.7.13 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/digitalocean/godo v1.109.0 // indirect
	github.com/docker/docker v25.0.4+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/eapache/go-resiliency v1.6.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.3 // indirect
	github.com/envoyproxy/go-control-plane v0.12.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.3 // indirect
	github.com/go-openapi/jsonreference v0.20.5 // indirect
	github.com/go-openapi/swag v0.22.10 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-resty/resty/v2 v2.11.0 // indirect
	github.com/go-sql-driver/mysql v1.8.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4
	github.com/google/cadvisor v0.49.1 // indirect
	github.com/google/flatbuffers v24.3.7+incompatible // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.2 // indirect
	github.com/gophercloud/gophercloud v1.11.0 // indirect
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/consul/api v1.28.2 // indirect
	github.com/hashicorp/cronexpr v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.8.0 // indirect
	github.com/hashicorp/go-hclog v1.6.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy/v2 v2.0.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.6.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.3.0 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.3 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.8 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.6 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20240308195635-9c3bbd191a76 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hashicorp/vault/sdk v0.11.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/tail v1.0.0 // indirect
	github.com/influxdata/telegraf v0.0.0-00010101000000-000000000000 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.1.9 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgtype v1.14.2 // indirect
	github.com/jackc/pgx/v4 v4.18.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/leoluk/perflib_exporter v0.2.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.9.8 // indirect
	github.com/linode/linodego v1.29.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20240226150601-1dcf7310316a // indirect
	github.com/magiconair/properties v1.8.7
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/miekg/dns v1.1.58 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mongodb-forks/digest v1.0.5 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.96.1-0.20240308205510-045e32991077 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.96.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.96.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20230419131419-497c7032c581 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/openzipkin/zipkin-go v0.4.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.21 // indirect
	github.com/shirou/gopsutil/v3 v3.24.2 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/ingest-protocols v0.2.0 // indirect
	github.com/signalfx/sapm-proto v0.14.0 // indirect
	github.com/sijms/go-ora/v2 v2.8.8 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/snowflakedb/gosnowflake v1.8.0 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.9 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.0 // indirect
	github.com/vmware/govmomi v0.36.0 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.12 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.12 // indirect
	go.mongodb.org/atlas v0.36.0 // indirect
	go.opencensus.io v0.24.0
	go.opentelemetry.io/collector/component v0.96.0
	go.opentelemetry.io/collector/consumer v0.96.0
	go.opentelemetry.io/collector/featuregate v1.3.0 // indirect
	go.opentelemetry.io/collector/semconv v0.96.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.24.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.24.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gonum.org/v1/gonum v0.14.0 // indirect
	google.golang.org/api v0.169.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240308144416-29370a3891b7 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.29.2 // indirect
	k8s.io/apimachinery v0.29.2 // indirect
	k8s.io/client-go v0.29.2 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	k8s.io/kubelet v0.29.2 // indirect
	k8s.io/utils v0.0.0-20240102154912-e7106e64919e // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension => ./pkg/extension/smartagentextension
	github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor => ./pkg/processor/timestampprocessor
	github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver => ./pkg/receiver/smartagentreceiver
	github.com/signalfx/splunk-otel-collector/tests => ./tests
)

replace (
	// there's an old v3.9.0 tag that shouldn't be used for openshift deps
	github.com/openshift/api => github.com/openshift/api v0.0.0-20230417092139-1b2161d23365
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20230419131419-497c7032c581
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20240111190717-3494050f2933

	github.com/signalfx/signalfx-agent => ./internal/signalfx-agent

	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12322#issuecomment-1185029670
// https://github.com/docker/go-connections/issues/99
replace github.com/docker/go-connections => github.com/docker/go-connections v0.4.0

// security updates
replace (
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.1
	github.com/apache/thrift => github.com/apache/thrift v0.16.0
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.18
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v1.1.1
	github.com/go-kit/kit => github.com/go-kit/kit v0.12.0 // required to drop dependency on deprecated go.etcd.io/etcd
	github.com/hashicorp/consul/sdk => github.com/hashicorp/consul/sdk v0.14.0
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.2.0
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.9.23
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.14.0
	github.com/spf13/viper => github.com/spf13/viper v1.11.0 // required to drop dependency on deprecated github.com/coreos/etcd and github.com/coreos/go-etcd
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.36.0
	golang.org/x/crypto => golang.org/x/crypto v0.17.0
	golang.org/x/net => golang.org/x/net v0.17.0
	k8s.io/apiserver => k8s.io/apiserver v0.24.1 // required to drop dependency on deprecated go.etcd.io/etcd
)

// this is the version that doesn't suffer from https://github.com/mattn/go-ieproxy/issues/45
replace github.com/mattn/go-ieproxy => github.com/mattn/go-ieproxy v0.0.1

// vault has invalid requirements https://github.com/hashicorp/vault/pull/13321
replace (
	github.com/hashicorp/vault/api/auth/approle => github.com/hashicorp/vault/api/auth/approle v0.1.2-0.20211223174530-3688d63348b3
	github.com/hashicorp/vault/api/auth/userpass => github.com/hashicorp/vault/api/auth/userpass v0.1.1-0.20211223174530-3688d63348b3
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/8081
replace github.com/googleapis/gnostic v0.5.6 => github.com/googleapis/gnostic v0.5.5

// required to drop dependency on deprecated git.apache.org/thrift.git
exclude go.opencensus.io v0.19.1
