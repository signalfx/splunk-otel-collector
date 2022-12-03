module github.com/signalfx/splunk-otel-collector

go 1.19

require (
	github.com/antonmedv/expr v1.9.0
	github.com/apache/pulsar-client-go v0.9.0
	github.com/cenkalti/backoff/v4 v4.2.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-zookeeper/zk v1.0.3
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/vault v1.12.1
	github.com/hashicorp/vault-plugin-auth-gcp v0.14.0
	github.com/hashicorp/vault/api v1.8.2
	github.com/jaegertracing/jaeger v1.39.1-0.20221110195127-14c11365a856
	github.com/knadh/koanf v1.4.4
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.66.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.66.0
	github.com/signalfx/golib/v3 v3.3.47
	github.com/signalfx/signalfx-agent v1.0.1-0.20221111170137-60843d376e91
	github.com/signalfx/splunk-otel-collector/extension/smartagentextension v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/processor/timestampprocessor v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/receiver/smartagentreceiver v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/tests v0.0.0-00010101000000-000000000000
	github.com/spf13/cast v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.1
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/client/v2 v2.305.5
	go.opentelemetry.io/collector v0.66.0
	go.opentelemetry.io/collector/exporter/loggingexporter v0.66.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.66.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.66.0
	go.opentelemetry.io/collector/extension/ballastextension v0.66.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.66.0
	go.opentelemetry.io/collector/pdata v0.66.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.66.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.66.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.66.0
	go.opentelemetry.io/otel/metric v0.33.0
	go.opentelemetry.io/otel/trace v1.11.1
	go.uber.org/atomic v1.10.0
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.24.0
	golang.org/x/sys v0.2.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20211115184647-b584dd5df32c // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	collectd.org v0.5.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/AthenZ/athenz v1.10.39 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v65.0.0+incompatible // indirect
	github.com/Azure/azure-storage-blob-go v0.15.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.28 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.21 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/DataDog/zstd v1.5.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v0.34.1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.4 // indirect
	github.com/SAP/go-hdb v0.109.1 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Shopify/sarama v1.37.2 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/participle/v2 v2.0.0-beta.5 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40 // indirect
	github.com/apache/thrift v0.17.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/armon/go-metrics v0.4.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.142 // indirect
	github.com/aws/aws-sdk-go-v2 v1.17.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.8 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.11 // indirect
	github.com/aws/smithy-go v1.13.4 // indirect
	github.com/beevik/ntp v0.3.0 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/cncf/xds/go v0.0.0-20220314180256-7f1daf1720fc // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/containerd/containerd v1.6.8 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/denisenkom/go-mssqldb v0.12.2 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/godo v1.82.0 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.21+incompatible // indirect
	github.com/docker/go-connections v0.4.1-0.20210727194412-58542c764a11 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.3 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.7 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.1 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-resty/resty/v2 v2.1.1-0.20191201195748-d7b97669fe48 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-test/deep v1.0.8 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang-jwt/jwt v3.2.1+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/cadvisor v0.45.0 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/safetext v0.0.0-20220905092116-b49f7bc46da2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/gophercloud/gophercloud v0.25.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grafana/regexp v0.0.0-20220304095617-2e8d9baf4ac2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.13.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/consul/api v1.15.3 // indirect
	github.com/hashicorp/cronexpr v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.8.0 // indirect
	github.com/hashicorp/go-hclog v1.3.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy/v2 v2.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.6 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20220809212729-939d643fec2c // indirect
	github.com/hashicorp/serf v0.9.8 // indirect
	github.com/hashicorp/vault/sdk v0.6.1-0.20221012172514-e12bdc5ce648 // indirect
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87 // indirect
	github.com/hetznercloud/hcloud-go v1.35.2 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/influxdata/go-syslog/v3 v3.0.1-0.20210608084020-ac565dc76ba6 // indirect
	github.com/influxdata/tail v1.0.0 // indirect
	github.com/influxdata/telegraf v0.0.0-00010101000000-000000000000 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.1.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/kolo/xmlrpc v0.0.0-20201022064351-38db28db192b // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20181214104525-299bdde78165 // indirect
	github.com/leoluk/perflib_exporter v0.2.0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/linkedin/goavro/v2 v2.9.8 // indirect
	github.com/linode/linodego v1.8.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-ieproxy v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/mauricelam/genny v0.0.0-20190320071652-0800202903e5 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/mount v0.3.3 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mongodb-forks/digest v1.0.4 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.1.17 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/observiq/ctimefmt v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.66.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.66.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/openlyinc/pointy v1.2.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.1 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.9 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.22.10 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657 // indirect
	github.com/signalfx/gateway v1.2.23 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/golib v2.5.1+incompatible // indirect
	github.com/signalfx/ingest-protocols v0.1.14 // indirect
	github.com/signalfx/sapm-proto v0.12.0 // indirect
	github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20221111170137-60843d376e91 // indirect
	github.com/signalfx/signalfx-go v1.25.0 // indirect
	github.com/sijms/go-ora/v2 v2.5.8 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/snowflakedb/gosnowflake v1.6.15 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/testcontainers/testcontainers-go v0.15.0 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.0 // indirect
	github.com/vmware/govmomi v0.29.0 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/etcd/api/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.5 // indirect
	go.mongodb.org/atlas v0.19.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector/component v0.66.0
	go.opentelemetry.io/collector/consumer v0.66.0
	go.opentelemetry.io/collector/featuregate v0.66.0 // indirect
	go.opentelemetry.io/collector/semconv v0.66.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.4 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.4 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.11.1 // indirect
	go.opentelemetry.io/contrib/zpages v0.36.4 // indirect
	go.opentelemetry.io/otel v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.11.1 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.33.0 // indirect
	go.uber.org/goleak v1.2.0 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20221019170559-20944726eadf // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/term v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	golang.org/x/tools v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gonum.org/v1/gonum v0.12.0 // indirect
	google.golang.org/api v0.103.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/fatih/set.v0 v0.1.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.25.4 // indirect
	k8s.io/apimachinery v0.25.4 // indirect
	k8s.io/client-go v0.25.4 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1 // indirect
	k8s.io/kubelet v0.25.2 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/kind v0.17.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/signalfx/splunk-otel-collector/extension/smartagentextension => ./pkg/extension/smartagentextension
	github.com/signalfx/splunk-otel-collector/processor/timestampprocessor => ./pkg/processor/timestamp
	github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver => ./pkg/receiver/oracledb
	github.com/signalfx/splunk-otel-collector/receiver/smartagentreceiver => ./pkg/receiver/smartagent
	github.com/signalfx/splunk-otel-collector/tests => ./tests
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20210820123244-82265917ca87
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.38.0

	github.com/signalfx/signalfx-agent/pkg/apm => github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20220810191306-a41fb5c94d53
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12322#issuecomment-1185029670
// https://github.com/docker/go-connections/issues/99
replace github.com/docker/go-connections => github.com/docker/go-connections v0.4.0

// security updates
replace (
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.1
	github.com/apache/thrift => github.com/apache/thrift v0.16.0
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.13
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v1.1.1
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.7
	github.com/go-kit/kit => github.com/go-kit/kit v0.12.0 // required to drop dependency on deprecated go.etcd.io/etcd
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.2.0
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.8.1
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.14.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
	github.com/spf13/viper => github.com/spf13/viper v1.11.0 // required to drop dependency on deprecated github.com/coreos/etcd and github.com/coreos/go-etcd
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.36.0
	golang.org/x/crypto => golang.org/x/crypto v0.3.0
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
