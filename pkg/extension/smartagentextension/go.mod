module github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension

go 1.20

require (
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/signalfx-agent v1.0.1-0.20230104182534-9eee411fe305
	github.com/signalfx/splunk-otel-collector/tests v0.83.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/collector/component v0.92.0
	go.opentelemetry.io/collector/confmap v0.92.0
	go.opentelemetry.io/collector/extension v0.92.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/containerd/containerd v1.7.11 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.7+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opencontainers/runc v1.1.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.3 // indirect
	github.com/signalfx/golib/v3 v3.3.53 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/testcontainers/testcontainers-go v0.20.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.92.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.92.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.92.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.92.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.92.0 // indirect
	go.opentelemetry.io/collector/consumer v0.92.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.92.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.1 // indirect
	go.opentelemetry.io/collector/pdata v1.0.1 // indirect
	go.opentelemetry.io/collector/receiver v0.92.0 // indirect
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.92.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-collector/tests => ../../../tests

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.18
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/hashicorp/consul/sdk => github.com/hashicorp/consul/sdk v0.14.0
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20240111190717-3494050f2933
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.36.0
	golang.org/x/crypto => golang.org/x/crypto v0.17.0
	golang.org/x/net => golang.org/x/net v0.17.0
)

replace github.com/signalfx/signalfx-agent => ../../../internal/signalfx-agent
