module github.com/signalfx/splunk-otel-collector

go 1.16

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-zookeeper/zk v1.0.2
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/hashicorp/vault v1.7.0
	github.com/hashicorp/vault-plugin-auth-gcp v0.9.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sapmexporter v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarder v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.25.1-0.20210427211609-b747cdc7e2d9 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusexecreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.25.1-0.20210427211609-b747cdc7e2d9 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sapmreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/signalfxreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/signalfx/defaults v1.2.2-0.20180531161417-70562fe60657
	github.com/signalfx/golib/v3 v3.3.33
	github.com/signalfx/signalfx-agent v1.0.1-0.20210503202607-8862b3ba9c0d
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.3.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200425165423-262c93980547
	go.opentelemetry.io/collector v0.25.1-0.20210427221155-df80b9f9f8bc
	go.uber.org/zap v1.16.0
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/internal/stanza v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.25.1-0.20210427211609-b747cdc7e2d9
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.0.0-00010101000000-000000000000 => github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.25.1-0.20210427211609-b747cdc7e2d9
)

// each of these is required for the smartagentreceiver
replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20201211214327-200738592ced
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201105135750-00f16d1ac3a4
	github.com/signalfx/signalfx-agent/pkg/apm => github.com/signalfx/signalfx-agent/pkg/apm v0.0.0-20210503202607-8862b3ba9c0d
	github.com/soheilhy/cmux => github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // required for smartagentreceiver to drop google.golang.org/grpc/examples/helloworld/helloworld test dep
	google.golang.org/grpc => google.golang.org/grpc v1.29.1 // required for smartagentreceiver's go.etcd.io/etcd dep
)
