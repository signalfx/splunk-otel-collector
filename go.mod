module github.com/signalfx/splunk-otel-collector

go 1.15

require (
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.31.0
	github.com/google/addlicense v0.0.0-20200906110928-a0294312aa76
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/ory/go-acc v0.2.6
	github.com/pavius/impi v0.0.3
	github.com/securego/gosec/v2 v2.4.0
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.12.1-0.20201016230751-46aada6e3c3a
	golang.org/x/sys v0.0.0-20200929083018-4d22bbb62b3c
	honnef.co/go/tools v0.0.1-2020.1.6
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.12.1-0.20201017021937-1a20922f151f
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.12.1-0.20201017021937-1a20922f151f
)
