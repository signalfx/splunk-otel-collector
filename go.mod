module github.com/signalfx/splunk-otel-collector

go 1.15

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/client9/misspell v0.3.4
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/golangci/golangci-lint v1.38.0
	github.com/google/addlicense v0.0.0-20200906110928-a0294312aa76
	github.com/hashicorp/vault/api v1.0.5-0.20201001211907-38d91b749c77
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.24.1-0.20210416173851-62c6f89406ce // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.24.1-0.20210416173851-62c6f89406ce
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/ory/go-acc v0.2.6
	github.com/pavius/impi v0.0.3
	github.com/securego/gosec/v2 v2.6.1
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/golib/v3 v3.3.16
	github.com/signalfx/signalfx-agent v1.0.1-0.20210325135021-a18ea9d77b40
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.3.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.24.1-0.20210416180505-2b33043a5024
	go.uber.org/zap v1.16.0
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4
	gopkg.in/yaml.v2 v2.4.0
	honnef.co/go/tools v0.1.2
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.24.1-0.20210416173851-62c6f89406ce
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.24.1-0.20210416173851-62c6f89406ce
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/dancannon/gorethink => gopkg.in/gorethink/gorethink.v4 v4.0.0
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20201211214327-200738592ced
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201105135750-00f16d1ac3a4
	github.com/signalfx/signalfx-agent/pkg/apm => github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20210325135021-a18ea9d77b40
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep
	google.golang.org/grpc => google.golang.org/grpc v1.29.1 // required for smartagentreceiver's go.etcd.io/etcd dep
)
