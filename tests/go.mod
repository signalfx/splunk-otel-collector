module github.com/signalfx/splunk-otel-collector/tests

go 1.15

require (
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver v0.25.0
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.10.0
	go.opentelemetry.io/collector v0.25.0
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.25.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.25.0
)
