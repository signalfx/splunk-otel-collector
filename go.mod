module github.com/signalfx/splunk-otel-collector

go 1.15

require (
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.31.0
	github.com/google/addlicense v0.0.0-20200622132530-df58acafd6d5
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/stackdriverexporter v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.11.0
	github.com/ory/go-acc v0.2.6
	github.com/pavius/impi v0.0.3
	github.com/securego/gosec/v2 v2.4.0
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.11.0
	golang.org/x/sys v0.0.0-20200929083018-4d22bbb62b3c
	honnef.co/go/tools v0.0.1-2020.1.5
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.11.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.11.0
)
