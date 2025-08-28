module github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension

go 1.24.0

require (
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/signalfx-agent v1.0.1-0.20230104182534-9eee411fe305
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v1.39.0
	go.opentelemetry.io/collector/component/componenttest v0.133.0
	go.opentelemetry.io/collector/confmap v1.39.0
	go.opentelemetry.io/collector/extension v1.39.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/creasty/defaults v1.8.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.2.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/golib/v3 v3.3.54 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.39.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.133.0 // indirect
	go.opentelemetry.io/collector/pdata v1.39.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.12.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/log v0.13.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
	google.golang.org/grpc v1.75.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20250228233359-931557f78bed
	github.com/signalfx/signalfx-agent => ../../../internal/signalfx-agent
	github.com/signalfx/splunk-otel-collector/tests => ../../../tests
)
