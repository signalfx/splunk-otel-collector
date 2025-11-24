module github.com/signalfx/splunk-otel-collector

go 1.25.3

require (
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/antonmedv/expr v1.15.5
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/expr-lang/expr v1.17.6
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-zookeeper/zk v1.0.4
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/vault v1.21.0
	github.com/hashicorp/vault-plugin-auth-gcp v0.22.0
	github.com/hashicorp/vault/api v1.22.0
	github.com/knadh/koanf v1.5.0
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/connector/sumconnector v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/pulsarexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/ackextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureblobreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/icmpcheckreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ntpreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusremotewritereceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefareceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snowflakereceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tlscheckreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver v0.140.1
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.67.4
	github.com/prometheus/prometheus v0.307.3
	github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension v0.83.0
	github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-collector/tests v0.0.0-00010101000000-000000000000
	github.com/spf13/cast v1.10.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	go.etcd.io/etcd/client/v2 v2.305.24
	go.opentelemetry.io/collector/component/componentstatus v0.140.0
	go.opentelemetry.io/collector/component/componenttest v0.140.0
	go.opentelemetry.io/collector/config/confighttp v0.140.0
	go.opentelemetry.io/collector/config/configopaque v1.46.0
	go.opentelemetry.io/collector/confmap v1.46.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.46.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.46.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.46.0
	go.opentelemetry.io/collector/confmap/xconfmap v0.140.0
	go.opentelemetry.io/collector/connector v0.140.0
	go.opentelemetry.io/collector/connector/forwardconnector v0.140.0
	go.opentelemetry.io/collector/consumer/consumertest v0.140.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.140.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.140.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.140.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.140.0
	go.opentelemetry.io/collector/extension v1.46.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.140.0
	go.opentelemetry.io/collector/otelcol v0.140.0
	go.opentelemetry.io/collector/pdata v1.46.0
	go.opentelemetry.io/collector/pipeline v1.46.0
	go.opentelemetry.io/collector/processor v1.46.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.140.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.140.0
	go.opentelemetry.io/collector/receiver v1.46.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.140.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.140.0
	go.opentelemetry.io/collector/receiver/receiverhelper v0.140.0
	go.opentelemetry.io/collector/receiver/receivertest v0.140.0
	go.opentelemetry.io/collector/scraper v0.140.0
	go.opentelemetry.io/collector/scraper/scraperhelper v0.140.0
	go.opentelemetry.io/collector/service v0.140.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/atomic v1.11.0
	go.uber.org/goleak v1.3.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sys v0.38.0
	gopkg.in/yaml.v2 v2.4.0
)

require sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect

require (
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/cloudsqlconn v1.4.3 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/pubsub v1.50.1 // indirect
	cloud.google.com/go/pubsub/v2 v2.0.0 // indirect
	collectd.org v0.6.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2 v2.0.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.7.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 v4.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3 v3.0.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.3 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/Code-Hex/go-generics-cache v1.5.1 // indirect
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.73.0-rc.5 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/DataDog/datadog-go/v5 v5.8.1 // indirect
	github.com/DataDog/go-sqllexer v0.1.9 // indirect
	github.com/GehirnInc/crypt v0.0.0-20230320061759-8cc1b52080c5 // indirect
	github.com/IBM/sarama v1.46.3 // indirect
	github.com/Jeffail/gabs/v2 v2.7.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/antchfx/xmlquery v1.5.0 // indirect
	github.com/antchfx/xpath v1.3.5 // indirect
	github.com/apache/arrow-go/v18 v18.4.0 // indirect
	github.com/apache/pulsar-client-go v0.17.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.19 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.58.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.266.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecs v1.67.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.49.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.39.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.40.1 // indirect
	github.com/bboreham/go-loser v0.0.0-20230920113527-fcc2c21820a3 // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/circonus-labs/circonusllhist v0.1.5 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/containerd/containerd/api v1.9.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/coreos/go-systemd/v22 v22.6.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/denverdino/aliyungo v0.0.0-20230411124812-ab98a9173ace // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/digitalocean/go-metadata v0.0.0-20250129100319-e3650a3df44b // indirect
	github.com/digitalocean/godo v1.165.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/duosecurity/duo_api_golang v0.0.0-20240205144049-bb361ad4ae1c // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/edsrzf/mmap-go v1.2.0 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.35.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/facebook/time v0.0.0-20240510113249-fa89cc575891 // indirect
	github.com/facette/natsort v0.0.0-20181210072756-2cd4dd1e2dcb // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.2 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.3 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/strfmt v0.24.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/google/certificate-transparency-go v1.3.2 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-metrics-stackdriver v0.6.0 // indirect
	github.com/google/go-tpm v0.9.7 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gophercloud/gophercloud/v2 v2.8.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gosnmp/gosnmp v1.42.1 // indirect
	github.com/grafana/clusterurl v0.2.1 // indirect
	github.com/grafana/regexp v0.0.0-20250905093917-f7b3be9d1853 // indirect
	github.com/hamba/avro/v2 v2.29.0 // indirect
	github.com/hashicorp/consul/sdk v0.16.2 // indirect
	github.com/hashicorp/cronexpr v1.1.3 // indirect
	github.com/hashicorp/go-bexpr v0.1.14 // indirect
	github.com/hashicorp/go-hmac-drbg v0.0.0-20210916214228-a6e5a68489f6 // indirect
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.18 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-raftchunking v0.7.0 // indirect
	github.com/hashicorp/go-secure-stdlib/cryptoutil v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/permitpool v1.0.0 // indirect
	github.com/hashicorp/go-secure-stdlib/plugincontainer v0.4.2 // indirect
	github.com/hashicorp/go-secure-stdlib/regexp v1.0.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/mdns v1.0.5 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20250930071859-eaa0fe0e27af // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.30.0 // indirect
	github.com/influxdata/influxdb-observability/common v0.5.12 // indirect
	github.com/influxdata/influxdb-observability/influx2otel v0.5.12 // indirect
	github.com/influxdata/line-protocol/v2 v2.2.1 // indirect
	github.com/influxdata/telegraf v1.30.1 // indirect
	github.com/influxdata/wlog v0.0.0-20160411224016-7c63b0a71ef8 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.3.4 // indirect
	github.com/itchyny/timefmt-go v0.1.7 // indirect
	github.com/jaegertracing/jaeger-idl v0.6.0 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/joshlf/go-acl v0.0.0-20200411065538-eae00ae38531 // indirect
	github.com/joyent/triton-go v1.8.5 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.3.0 // indirect
	github.com/linode/go-metadata v0.2.2 // indirect
	github.com/linode/linodego v1.59.0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/mdlayher/vsock v1.2.1 // indirect
	github.com/michel-laterman/proxy-connect-dialer-go v0.1.0 // indirect
	github.com/microsoft/go-mssqldb v1.9.3 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/nginx/nginx-prometheus-exporter v1.4.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/open-telemetry/opamp-go v0.22.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampcustommessages v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/containerinsight v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/k8s v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/metrics v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/collectd v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/exp/metrics v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/gopsutilenv v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/core/xidutils v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/configkafka v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/status v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azure v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azurelogs v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatocumulativeprocessor v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/scraper/zookeeperscraper v0.140.1 // indirect
	github.com/opencontainers/cgroups v0.0.2 // indirect
	github.com/opencontainers/runtime-spec v1.2.1 // indirect
	github.com/orcaman/concurrent-map/v2 v2.0.1 // indirect
	github.com/outcaste-io/ristretto v0.2.3 // indirect
	github.com/ovh/go-ovh v1.9.0 // indirect
	github.com/packethost/packngo v0.31.0 // indirect
	github.com/petermattis/goid v0.0.0-20250721140440-ea1c0173183e // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pquerna/otp v1.4.0 // indirect
	github.com/prometheus-community/pro-bing v0.7.0 // indirect
	github.com/prometheus/alertmanager v0.28.1 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/common/assets v0.2.0 // indirect
	github.com/prometheus/exporter-toolkit v0.15.0 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/prometheus/sigv4 v0.2.1 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/redis/go-redis/v9 v9.16.0 // indirect
	github.com/relvacode/iso8601 v1.7.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.35 // indirect
	github.com/sethvargo/go-limiter v0.7.2 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shirou/gopsutil/v4 v4.25.10 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/shurcooL/httpfs v0.0.0-20230704072500-f1e31cf0ba5c // indirect
	github.com/signalfx/golib/v3 v3.4.1 // indirect
	github.com/signalfx/signalfx-agent v1.0.1-0.20230222185249-54e5d1064c5b // indirect
	github.com/softlayer/softlayer-go v1.1.3 // indirect
	github.com/soniah/gosnmp v0.0.0-20190220004421-68e8beac0db9 // indirect
	github.com/stackitcloud/stackit-sdk-go/core v0.17.3 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible // indirect
	github.com/testcontainers/testcontainers-go v0.40.0 // indirect
	github.com/tg123/go-htpasswd v1.2.4 // indirect
	github.com/thda/tds v0.1.7 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tidwall/wal v1.2.1 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/twmb/franz-go v1.20.4 // indirect
	github.com/twmb/franz-go/pkg/kadm v1.17.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.12.0 // indirect
	github.com/twmb/franz-go/pkg/sasl/kerberos v1.1.0 // indirect
	github.com/twmb/franz-go/plugin/kzap v1.1.2 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	go.mongodb.org/mongo-driver v1.17.4 // indirect
	go.mongodb.org/mongo-driver/v2 v2.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.140.0 // indirect
	go.opentelemetry.io/collector/client v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.140.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.46.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.46.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.140.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.46.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.140.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.140.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter v1.46.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.140.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.140.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.46.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.140.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.140.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.140.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.140.0 // indirect
	go.opentelemetry.io/collector/filter v0.140.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.140.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.140.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.140.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.140.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.140.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.140.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.140.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.140.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.140.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.140.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.140.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.63.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.18.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.60.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.14.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/telemetry v0.0.0-20251008203120-078029d740a8 // indirect
	google.golang.org/genproto v0.0.0-20251002232023-7c0ddcbb5797 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	sigs.k8s.io/controller-runtime v0.22.4 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20241007161556-ec30366c7912 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/AthenZ/athenz v1.12.13 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.6.2 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-amqp v1.5.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.29 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/SAP/go-hdb v1.14.12 // indirect
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/apache/thrift v0.22.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.6 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.23 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.20.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.90.1 // indirect
	github.com/aws/smithy-go v1.23.2 // indirect
	github.com/beevik/ntp v1.5.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/docker/docker v28.5.2+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dvsekhvalnov/jose2go v1.8.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-test/deep v1.1.1 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0
	github.com/google/cadvisor v0.53.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gophercloud/gophercloud v1.14.1 // indirect
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/consul/api v1.32.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-gcp-common v0.9.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-kms-wrapping/entropy/v2 v2.0.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.6.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/awsutil v0.3.0 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.3 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hashicorp/vault/sdk v0.20.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/influxdata/tail v1.0.1-0.20221130111531-19b97bffd978 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.3 // indirect
	github.com/jackc/pgx/v4 v4.18.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magiconair/properties v1.8.10
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-xmlrpc v0.0.3 // indirect
	github.com/miekg/dns v1.1.68 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mongodb-forks/digest v1.1.0 // indirect
	github.com/mongodb/go-client-mongodb-atlas v0.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwielbut/pointy v1.1.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryencodingextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/textencodingextension v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.140.1
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/signalfx v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.140.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.140.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20251015124057-db0dee36e235 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657 // indirect
	github.com/signalfx/gohistogram v0.0.0-20160107210732-1ccfd2ff5083 // indirect
	github.com/signalfx/ingest-protocols v0.4.1 // indirect
	github.com/signalfx/sapm-proto v0.18.0 // indirect
	github.com/sijms/go-ora/v2 v2.9.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/snowflakedb/gosnowflake v1.17.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tinylib/msgp v1.5.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/ulule/deepcopier v0.0.0-20171107155558-ca99b135e50f // indirect
	github.com/vjeantet/grok v1.0.1 // indirect
	github.com/vmware/govmomi v0.51.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/etcd/api/v3 v3.6.0 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.0 // indirect
	go.mongodb.org/atlas v0.38.0 // indirect
	go.opencensus.io v0.24.0
	go.opentelemetry.io/collector/component v1.46.0
	go.opentelemetry.io/collector/consumer v1.46.0
	go.opentelemetry.io/collector/featuregate v1.46.0
	go.opentelemetry.io/collector/semconv v0.128.1-0.20250610090210-188191247685 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.38.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20251002181428-27f1f14c8bb9
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/text v0.31.0
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/api v0.256.0 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.34.1 // indirect
	k8s.io/apimachinery v0.34.1 // indirect
	k8s.io/client-go v0.34.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	k8s.io/kubelet v0.34.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	github.com/signalfx/signalfx-agent => ./internal/signalfx-agent
	github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension => ./pkg/extension/smartagentextension
	github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor => ./pkg/processor/timestampprocessor
	github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver => ./pkg/receiver/smartagentreceiver
	github.com/signalfx/splunk-otel-collector/tests => ./tests
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20250228233359-931557f78bed
)
