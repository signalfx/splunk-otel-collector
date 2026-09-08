module github.com/signalfx/splunk-otel-collector/pkg/receiver/splunkinputsreceiver

go 1.26.4

require (
	github.com/fsnotify/fsnotify v1.10.1
	github.com/splunk/tarunner v0.5.2-0.20260904224642-39b84b2ff193
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.65.0
	go.opentelemetry.io/collector/consumer v1.65.0
	go.opentelemetry.io/collector/pdata v1.65.0
	go.opentelemetry.io/collector/receiver v1.65.0
	go.uber.org/zap v1.28.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v7 v7.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/elastic/lunes v0.2.2 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20260427185012-515ba073c4c1 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.9-0.20260124013517-8f8f42cba0de // indirect
	github.com/google/pprof v0.0.0-20260709232956-b9395ee17fa0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/knadh/koanf/maps v0.1.3 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.1 // indirect
	github.com/knadh/koanf/v2 v2.3.6 // indirect
	github.com/leodido/go-syslog/v4 v4.6.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/pprof v0.159.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk v0.159.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.28 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/client v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.65.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.159.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.65.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.65.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.65.0 // indirect
	go.opentelemetry.io/collector/confmap v1.65.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.159.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.159.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.159.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.159.0 // indirect
	go.opentelemetry.io/collector/exporter v1.65.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.159.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.159.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.159.0 // indirect
	go.opentelemetry.io/collector/extension v1.65.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.65.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.159.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.159.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.65.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.159.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.159.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.159.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.65.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.159.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.159.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.159.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.70.0 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	google.golang.org/grpc v1.83.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
