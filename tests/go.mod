module github.com/signalfx/splunk-otel-collector/tests

go 1.21

require (
	github.com/docker/docker v25.0.4+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/go-sql-driver/mysql v1.8.0
	github.com/google/uuid v1.6.0
	github.com/knadh/koanf v1.5.0
	github.com/shirou/gopsutil/v3 v3.24.2
	github.com/stretchr/testify v1.9.0
	github.com/testcontainers/testcontainers-go v0.29.1
	go.opentelemetry.io/collector/component v0.96.0
	go.opentelemetry.io/collector/config/configgrpc v0.96.0
	go.opentelemetry.io/collector/config/confignet v0.96.0
	go.opentelemetry.io/collector/config/configtls v0.96.0
	go.opentelemetry.io/collector/confmap v0.96.0
	go.opentelemetry.io/collector/consumer v0.96.0
	go.opentelemetry.io/collector/exporter v0.96.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.96.0
	go.opentelemetry.io/collector/pdata v1.3.0
	go.opentelemetry.io/collector/receiver v0.96.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.96.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/distribution/reference v0.5.0 // indirect

require (
	dario.cat/mergo v1.0.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/containerd v1.7.14 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/knadh/koanf/v2 v2.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20240226150601-1dcf7310316a // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/collector v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.96.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.3.0 // indirect
	go.opentelemetry.io/collector/config/configretry v0.96.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.96.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.96.0 // indirect
	go.opentelemetry.io/collector/extension v0.96.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.96.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.3.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.47.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240311173647-c811ad7063a7 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

// https://github.com/go-logr/logr/issues/51
replace k8s.io/klog/v2 => k8s.io/klog/v2 v2.80.1

// security updates
replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.18
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/form3tech-oss/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/hashicorp/consul/sdk => github.com/hashicorp/consul/sdk v0.13.1
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	golang.org/x/crypto => golang.org/x/crypto v0.17.0
	golang.org/x/net => golang.org/x/net v0.17.0
)

// required until all deps adopt something like https://github.com/open-telemetry/opentelemetry-collector/commit/6059751e64e0d7b857abff50f1d8ab90424d0306
// updating all cloud.google.com deps in this project doesn't resolve on its own.
replace cloud.google.com/go => cloud.google.com/go v0.110.2
