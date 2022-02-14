module github.com/signalfx/splunk-otel-collector

go 1.17

require (
	cloud.google.com/go/kms v0.1.0 // indirect
	cloud.google.com/go/monitoring v0.1.0 // indirect
	github.com/antonmedv/expr v1.9.0
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/fsnotify/fsnotify v1.5.1
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-zookeeper/zk v1.0.2
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/vault v1.9.3
	github.com/hashicorp/vault-plugin-auth-gcp v0.11.3
	github.com/hashicorp/vault/api v1.3.1
	github.com/jaegertracing/jaeger v1.30.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/routingprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.40.0 // ignore-update-deps
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.44.0
	github.com/openzipkin/zipkin-go v0.4.0
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/golib/v3 v3.3.43
	github.com/signalfx/signalfx-agent v1.0.1-0.20220124192246-00e4f3c7b0c0
	github.com/signalfx/splunk-otel-collector/tests v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/client/v2 v2.305.2
	go.opentelemetry.io/collector v0.44.0
	go.opentelemetry.io/collector/model v0.44.0
	go.opentelemetry.io/otel/trace v1.4.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.21.0
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/compute v1.1.0 // indirect
	cloud.google.com/go/iam v0.1.1 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20180905200951-72629b5276e3 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	collectd.org v0.5.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v61.1.0+incompatible // indirect
	github.com/Azure/azure-storage-blob-go v0.14.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.23 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/Microsoft/hcsshim v0.9.0 // indirect
	github.com/SAP/go-hdb v0.105.2 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Shopify/sarama v1.31.0 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20210818145353-234c94e4ce64 // indirect
	github.com/apache/thrift v0.15.0 // indirect
	github.com/armon/go-metrics v0.3.10 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.42.44 // indirect
	github.com/aws/aws-sdk-go-v2 v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.12.0 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/beevik/ntp v0.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/cncf/xds/go v0.0.0-20211130200136-a8f946100490 // indirect
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.5.9 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/godo v1.73.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.12+incompatible // indirect
	github.com/docker/go-connections v0.4.1-0.20210727194412-58542c764a11 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/envoyproxy/go-control-plane v0.10.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.2 // indirect
	github.com/evanphx/json-patch/v5 v5.5.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.3.1 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-resty/resty/v2 v2.1.1-0.20191201195748-d7b97669fe48 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-test/deep v1.0.8 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.0.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/cadvisor v0.26.1 // indirect
	github.com/google/flatbuffers v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gophercloud/gophercloud v0.24.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/consul/api v1.12.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.7.0 // indirect
	github.com/hashicorp/go-hclog v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy v0.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.1.5 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-3 // indirect
	github.com/hashicorp/serf v0.9.6 // indirect
	github.com/hashicorp/vault/sdk v0.3.1-0.20220103172553-29ded54520a4 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/hetznercloud/hcloud-go v1.33.1 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/influxdata/tail v1.0.0 // indirect
	github.com/influxdata/telegraf v0.0.0-00010101000000-000000000000 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.10.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.9.1 // indirect
	github.com/jackc/pgx/v4 v4.14.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.10.3 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/knadh/koanf v1.4.0 // indirect
	github.com/kolo/xmlrpc v0.0.0-20201022064351-38db28db192b // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leoluk/perflib_exporter v0.1.0 // indirect
	github.com/lib/pq v1.10.3 // indirect
	github.com/linode/linodego v1.2.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mauricelam/genny v0.0.0-20190320071652-0800202903e5 // indirect
	github.com/miekg/dns v1.1.45 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.4.1 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mongodb-forks/digest v1.0.3 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.1.16 // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/observiq/ctimefmt v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-log-collection v0.24.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/openlyinc/pointy v1.1.2 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/prometheus/statsd_exporter v0.21.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.7.0.20210223165440-c65ae3540d44 // indirect
	github.com/shirou/gopsutil v3.21.10+incompatible // indirect
	github.com/shirou/gopsutil/v3 v3.21.12 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.2 // indirect
	github.com/signalfx/gateway v1.2.19-0.20191125135538-2c417b7ae0bd // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/ingest-protocols v0.1.3 // indirect
	github.com/signalfx/sapm-proto v0.7.2 // indirect
	github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20211117172757-84f64129e93d // indirect
	github.com/signalfx/signalfx-go v1.8.1 // indirect
	github.com/snowflakedb/gosnowflake v1.6.3 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.10.1 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/testcontainers/testcontainers-go v0.11.1 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.0 // indirect
	github.com/vmware/govmomi v0.23.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.0 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.2 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.2 // indirect
	go.mongodb.org/atlas v0.14.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.28.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.28.0 // indirect
	go.opentelemetry.io/otel v1.4.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.26.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/sdk v1.3.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.26.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	golang.org/x/crypto v0.0.0-20211202192323-5770296d904e // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/net v0.0.0-20220105145211-5b0dc2dfae98 // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	golang.org/x/tools v0.1.9 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gonum.org/v1/gonum v0.9.3 // indirect
	google.golang.org/api v0.66.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220114231437-d2e6a121cae0 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/fatih/set.v0 v0.1.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.23.3 // indirect
	k8s.io/apimachinery v0.23.3 // indirect
	k8s.io/client-go v0.23.3 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/kubelet v0.23.3 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage => github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal => github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet => github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk => github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza => github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.44.0
	github.com/signalfx/splunk-otel-collector/tests => ./tests
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20210820123244-82265917ca87
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20220111145625-076109fa1910

	github.com/signalfx/signalfx-agent/pkg/apm => github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20220124192246-00e4f3c7b0c0
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep

	// cvss scan requires this to be >= 3.3.23
	go.etcd.io/etcd => go.etcd.io/etcd v3.3.27+incompatible
)

// downgrade github.com/shirou/gopsutil until https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/5879 is resolved
replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.40.0 // ignore-update-deps
	github.com/shirou/gopsutil => github.com/shirou/gopsutil v3.21.8+incompatible
	github.com/shirou/gopsutil/v3 => github.com/shirou/gopsutil/v3 v3.21.8
)

// security updates
replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3
)

// vault has invalid requirements https://github.com/hashicorp/vault/pull/13321
replace (
	github.com/hashicorp/vault/api/auth/approle => github.com/hashicorp/vault/api/auth/approle v0.1.2-0.20211223174530-3688d63348b3
	github.com/hashicorp/vault/api/auth/userpass => github.com/hashicorp/vault/api/auth/userpass v0.1.1-0.20211223174530-3688d63348b3
)
