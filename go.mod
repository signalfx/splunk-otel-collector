module github.com/signalfx/splunk-otel-collector

go 1.17

require (
	cloud.google.com/go/kms v0.1.0 // indirect
	cloud.google.com/go/monitoring v0.1.0 // indirect
	github.com/antonmedv/expr v1.9.0
	github.com/aws/aws-sdk-go-v2/internal/ini v1.1.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-zookeeper/zk v1.0.2
	github.com/hashicorp/vault v1.7.0
	github.com/hashicorp/vault-plugin-auth-gcp v0.9.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/hashicorp/vault/sdk v0.2.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.33.0
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/golib/v3 v3.3.36
	github.com/signalfx/signalfx-agent v1.0.1-0.20210825181409-f66c88d96525
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200425165423-262c93980547
	go.opentelemetry.io/collector v0.33.0
	go.opentelemetry.io/collector/model v0.33.0
	go.uber.org/zap v1.19.0
	golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.92.3 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20180905200951-72629b5276e3 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	collectd.org v0.5.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.3.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go v52.5.0+incompatible // indirect
	github.com/Azure/azure-storage-blob-go v0.13.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Shopify/sarama v1.29.1 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/apache/arrow/go/arrow v0.0.0-20200601151325-b2287a20f230 // indirect
	github.com/apache/thrift v0.14.2 // indirect
	github.com/armon/go-metrics v0.3.4 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.40.24 // indirect
	github.com/aws/aws-sdk-go-v2 v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.1.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.0.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.5.0 // indirect
	github.com/aws/smithy-go v1.5.0 // indirect
	github.com/beevik/ntp v0.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisenkom/go-mssqldb v0.10.0 // indirect
	github.com/digitalocean/godo v1.58.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.10.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa // indirect
	github.com/go-kit/kit v0.11.0 // indirect
	github.com/go-kit/log v0.1.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/cadvisor v0.26.1 // indirect
	github.com/google/flatbuffers v1.12.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gophercloud/gophercloud v0.16.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/consul/api v1.8.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.7.0 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy v0.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-plugin v1.4.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.7 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.1-vault // indirect
	github.com/hashicorp/serf v0.9.5 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/hetznercloud/hcloud-go v1.24.0 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/influxdata/tail v1.0.0 // indirect
	github.com/influxdata/telegraf v0.0.0-00010101000000-000000000000 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.5.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200307190119-3430c5407db8 // indirect
	github.com/jackc/pgtype v1.3.0 // indirect
	github.com/jackc/pgx/v4 v4.6.0 // indirect
	github.com/jaegertracing/jaeger v1.25.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/karrick/godirwalk v1.10.3 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.12.2 // indirect
	github.com/knadh/koanf v1.2.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leoluk/perflib_exporter v0.1.0 // indirect
	github.com/lib/pq v1.9.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mauricelam/genny v0.0.0-20190320071652-0800202903e5 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/observiq/ctimefmt v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/interval v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.24.1-0.20210408210148-736647af91e1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.0.0-00010101000000-000000000000 // indirect
	github.com/open-telemetry/opentelemetry-log-collection v0.20.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/prometheus/prometheus v2.5.0+incompatible // indirect
	github.com/prometheus/statsd_exporter v0.21.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/cors v1.8.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20200724154423-2164a8ac840e // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.7.0.20210223165440-c65ae3540d44 // indirect
	github.com/shirou/gopsutil v3.21.7+incompatible // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.2 // indirect
	github.com/signalfx/gateway v1.2.19-0.20191125135538-2c417b7ae0bd // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/ingest-protocols v0.1.3 // indirect
	github.com/signalfx/sapm-proto v0.7.0 // indirect
	github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20201202163743-65b4fa925fc8 // indirect
	github.com/signalfx/signalfx-go v1.6.38-0.20200518153434-ceee8d2570d5 // indirect
	github.com/snowflakedb/gosnowflake v1.4.3 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cobra v1.2.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.8.1 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tidwall/gjson v1.8.1 // indirect
	github.com/tidwall/match v1.0.3 // indirect
	github.com/tidwall/pretty v1.1.0 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.6 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.0 // indirect
	github.com/vmware/govmomi v0.23.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd/api/v3 v3.5.0 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.0 // indirect
	go.etcd.io/etcd/client/v2 v2.305.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib v0.22.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.22.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.22.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.22.0 // indirect
	go.opentelemetry.io/otel v1.0.0-RC2 // indirect
	go.opentelemetry.io/otel/internal/metric v0.22.0 // indirect
	go.opentelemetry.io/otel/metric v0.22.0 // indirect
	go.opentelemetry.io/otel/sdk v1.0.0-RC2 // indirect
	go.opentelemetry.io/otel/trace v1.0.0-RC2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gonum.org/v1/gonum v0.9.3 // indirect
	google.golang.org/api v0.54.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210813162853-db860fec028c // indirect
	google.golang.org/grpc v1.40.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/fatih/set.v0 v0.1.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.22.0 // indirect
	k8s.io/apimachinery v0.22.0 // indirect
	k8s.io/client-go v0.22.0 // indirect
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	k8s.io/kubelet v0.22.0 // indirect
	k8s.io/utils v0.0.0-20210707171843-4b05e18ac7d9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage => github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/interval => github.com/open-telemetry/opentelemetry-collector-contrib/internal/interval v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet => github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk => github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza => github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.33.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.33.0
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20201211214327-200738592ced
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20210430082741-2a4b8e12bbf2
	github.com/signalfx/signalfx-agent/pkg/apm => github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20210617144543-be16994eef6b
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep
	google.golang.org/grpc => google.golang.org/grpc v1.29.1 // required for smartagentreceiver's go.etcd.io/etcd dep
)
