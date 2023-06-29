module github.com/signalfx/splunk-otel-collector

go 1.19

require (
	github.com/alecthomas/participle/v2 v2.0.0
	github.com/antonmedv/expr v1.12.5
	github.com/apache/pulsar-client-go v0.10.0
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-zookeeper/zk v1.0.3
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/vault v1.13.3
	github.com/hashicorp/vault-plugin-auth-gcp v0.16.0
	github.com/hashicorp/vault/api v1.9.2
	github.com/jaegertracing/jaeger v1.41.0
	github.com/knadh/koanf v1.5.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.80.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.80.0
	github.com/prometheus/client_model v0.4.0
	github.com/prometheus/common v0.44.0
	github.com/prometheus/prometheus v0.45.0
	github.com/signalfx/golib/v3 v3.3.50
	github.com/signalfx/signalfx-agent v1.0.1-0.20230222185249-54e5d1064c5b
	github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension v0.72.0
	github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor v0.72.0
	github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver v0.72.0
	github.com/signalfx/splunk-otel-collector/tests v0.72.0
	github.com/spf13/cast v1.5.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	go.etcd.io/bbolt v1.3.7
	go.etcd.io/etcd/client/v2 v2.305.9
	go.opentelemetry.io/collector v0.80.0
	go.opentelemetry.io/collector/config/confighttp v0.80.0
	go.opentelemetry.io/collector/config/configtelemetry v0.80.0
	go.opentelemetry.io/collector/config/configtls v0.80.0
	go.opentelemetry.io/collector/confmap v0.80.0
	go.opentelemetry.io/collector/connector v0.80.0
	go.opentelemetry.io/collector/connector/forwardconnector v0.80.0
	go.opentelemetry.io/collector/exporter v0.80.0
	go.opentelemetry.io/collector/exporter/loggingexporter v0.80.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.80.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.80.0
	go.opentelemetry.io/collector/extension v0.80.0
	go.opentelemetry.io/collector/extension/ballastextension v0.80.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.80.0
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0013
	go.opentelemetry.io/collector/processor v0.80.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.80.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.80.0
	go.opentelemetry.io/collector/receiver v0.80.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.80.0
	go.opentelemetry.io/otel/metric v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.24.0
	golang.org/x/sys v0.9.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.0.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.8.1 // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/arrow/go/v12 v12.0.0 // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.8 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.80.0 // indirect
	github.com/ovh/go-ovh v1.4.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	github.com/relvacode/iso8601 v1.3.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector/config/configauth v0.80.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.80.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.80.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.80.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.80.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.80.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.80.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.39.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
)

require (
	cloud.google.com/go/compute v1.20.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20211115184647-b584dd5df32c // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	collectd.org v0.5.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/AthenZ/athenz v1.10.39 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.6.0 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-amqp v1.0.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/DataDog/zstd v1.5.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.15.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/SAP/go-hdb v1.3.6 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Shopify/sarama v1.38.1 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/apache/thrift v0.18.1 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.282 // indirect
	github.com/aws/aws-sdk-go-v2 v1.18.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.24 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.59 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.31.0 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/beevik/ntp v1.0.0 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/cncf/xds/go v0.0.0-20230607035331-e9ce68804cb4 // indirect
	github.com/containerd/containerd v1.6.19 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/denisenkom/go-mssqldb v0.12.3 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/digitalocean/godo v1.99.0 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.2+incompatible // indirect
	github.com/docker/go-connections v0.4.1-0.20210727194412-58542c764a11 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/envoyproxy/go-control-plane v0.11.1-0.20230524094728-9239064ad72f // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.1 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-test/deep v1.1.0 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4
	github.com/google/cadvisor v0.47.1 // indirect
	github.com/google/flatbuffers v23.1.21+incompatible // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.4 // indirect
	github.com/googleapis/gax-go/v2 v2.10.0 // indirect
	github.com/gophercloud/gophercloud v1.4.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/consul/api v1.22.0 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.8.0 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy/v2 v2.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.8 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.7 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20230605233119-67e39d5d248f // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hashicorp/vault/sdk v0.9.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87 // indirect
	github.com/hetznercloud/hcloud-go v1.45.1 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/go-syslog/v3 v3.0.1-0.20210608084020-ac565dc76ba6 // indirect
	github.com/influxdata/tail v1.0.0 // indirect
	github.com/influxdata/telegraf v0.0.0-00010101000000-000000000000 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.1.7 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/pgx/v4 v4.18.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/leoluk/perflib_exporter v0.2.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.9.8 // indirect
	github.com/linode/linodego v1.17.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mauricelam/genny v0.0.0-20190320071652-0800202903e5 // indirect
	github.com/miekg/dns v1.1.54 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mongodb-forks/digest v1.0.4 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.1.19 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/observiq/ctimefmt v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.80.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.80.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opencontainers/runc v1.1.6 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20230419131419-497c7032c581 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/openzipkin/zipkin-go v0.4.1 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.17 // indirect
	github.com/shirou/gopsutil/v3 v3.23.5 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/ingest-protocols v0.2.0 // indirect
	github.com/signalfx/sapm-proto v0.13.0 // indirect
	github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20230222185249-54e5d1064c5b // indirect
	github.com/signalfx/signalfx-go v1.32.0 // indirect
	github.com/sijms/go-ora/v2 v2.7.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/snowflakedb/gosnowflake v1.6.21 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/testcontainers/testcontainers-go v0.20.1 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.0 // indirect
	github.com/vmware/govmomi v0.30.4 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.etcd.io/etcd/api/v3 v3.5.9 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.9 // indirect
	go.mongodb.org/atlas v0.29.0 // indirect
	go.opencensus.io v0.24.0
	go.opentelemetry.io/collector/component v0.80.0
	go.opentelemetry.io/collector/consumer v0.80.0
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0013 // indirect
	go.opentelemetry.io/collector/semconv v0.80.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.42.1-0.20230612162650-64be7e574a17 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.17.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.42.0 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.39.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	golang.org/x/crypto v0.10.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/oauth2 v0.9.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/term v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gonum.org/v1/gonum v0.13.0 // indirect
	google.golang.org/api v0.127.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.56.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.27.3 // indirect
	k8s.io/apimachinery v0.27.3 // indirect
	k8s.io/client-go v0.27.3 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/kubelet v0.27.2 // indirect
	k8s.io/utils v0.0.0-20230313181309-38a27ef9d749 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
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
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20210820123244-82265917ca87

	github.com/signalfx/signalfx-agent => ./internal/signalfx-agent
	github.com/signalfx/signalfx-agent/pkg/apm => ./internal/signalfx-agent/pkg/apm

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
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.8.1
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.14.0
	github.com/spf13/viper => github.com/spf13/viper v1.11.0 // required to drop dependency on deprecated github.com/coreos/etcd and github.com/coreos/go-etcd
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.36.0
	golang.org/x/crypto => golang.org/x/crypto v0.5.0
	golang.org/x/net => golang.org/x/net v0.7.0
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
