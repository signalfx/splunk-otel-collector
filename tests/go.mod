module github.com/signalfx/splunk-otel-collector/tests

go 1.19

require (
	github.com/docker/docker v20.10.21+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/google/uuid v1.3.0
	github.com/shirou/gopsutil/v3 v3.22.11
	github.com/signalfx/com_signalfx_metrics_protobuf v0.0.2
	github.com/signalfx/signalfx-go v1.25.0
	github.com/stretchr/testify v1.8.1
	github.com/testcontainers/testcontainers-go v0.15.0
	go.opentelemetry.io/collector v0.66.0
	go.opentelemetry.io/collector/component v0.66.0
	go.opentelemetry.io/collector/consumer v0.66.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.66.0
	go.opentelemetry.io/collector/pdata v0.66.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.66.0
	go.opentelemetry.io/otel/metric v0.33.0
	go.opentelemetry.io/otel/trace v1.11.1
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.24.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
	sigs.k8s.io/kind v0.17.0
)

require (
	cloud.google.com/go/compute v1.10.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.4 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/containerd/containerd v1.6.8 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/safetext v0.0.0-20220905092116-b49f7bc46da2 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/knadh/koanf v1.4.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/mount v0.3.3 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mostynb/go-grpc-compression v1.1.17 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/signalfx/golib/v3 v3.3.37 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector/featuregate v0.66.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.4 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.4 // indirect
	go.opentelemetry.io/otel v1.11.1 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/term v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// https://github.com/go-logr/logr/issues/51
replace k8s.io/klog/v2 => k8s.io/klog/v2 v2.80.1

// security updates
replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.13
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/form3tech-oss/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
	golang.org/x/crypto => golang.org/x/crypto v0.3.0
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0
)
