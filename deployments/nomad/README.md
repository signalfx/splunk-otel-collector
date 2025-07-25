# Splunk OpenTelemetry Collector for HashiCorp Nomad

The Splunk OpenTelemetry Collector for HashiCorp Nomad is an orchestrator deployment to create a job which provides a unified way to receive, process and export metric, and trace data for [Splunk Observability Cloud](https://docs.splunk.com/Observability).

**Note**: _Job files are provided as a reference only and are not intended for production use._

## Deployment

To run the job files you need:

- Access to a Nomad cluster (**version 1.6.2 to 1.9.7**)
- (Optional) Access to a Consul cluster - **Attention**: _the `*.nomad` examples provided
  in this repository assume that Consul is being used._

**Note**: _If not using Consul, ensure that `provider: nomad` is specified for each
[service block](https://developer.hashicorp.com/nomad/docs/job-specification/service#provider)
defined in your job file._

To start a local dev agent for Nomad and Consul, download the
[`nomad`](https://www.nomadproject.io/downloads) binary file and
[`consul`](https://www.consul.io/downloads) binary and run the following
commands in two different terminals:

```shell-session
nomad agent -dev -network-interface='{{ GetPrivateInterfaces | attr "name" }}'
```

```shell-session
consul agent -dev
```

### Usage Guide

To deploy the Splunk OpenTelemetry Collector job on the Nomad cluster we need to set environment variable in Nomad job configuration. 

```yaml
env {
    SPLUNK_ACCESS_TOKEN = "<SPLUNK_ACCESS_TOKEN>"
    SPLUNK_REALM = "<SPLUNK_REALM>"
    SPLUNK_MEMORY_TOTAL_MIB = 2048
    // You can specify more environment variables to override default values.
}
```

To use your own Splunk OpenTelemetry Collector configuration file, you can specify content in the [template stanza](https://www.nomadproject.io/docs/job-specification/template).

```yaml
template {
    data        = <<EOF
// Find the below config example for setting up your own Splunk OpenTelemetry Collector configuration file.
extensions:
  health_check:
    endpoint: 0.0.0.0:13133
  zpages: null
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu: null
      disk: null
      filesystem: null
      load: null
      memory: null
      network: null
      paging: null
      processes: null
processors:
  batch: null
  memory_limiter:
    check_interval: 2s
    limit_mib: ${SPLUNK_MEMORY_LIMIT_MIB}
exporters:
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    api_url: https://api.${SPLUNK_REALM}.signalfx.com
    correlation: null
    ingest_url: https://ingest.${SPLUNK_REALM}.signalfx.com
    sync_host_metadata: true
  debug:
    verbosity: detailed
service:
  extensions:
  - health_check
  - zpages
  pipelines:
    metrics:
      exporters:
      - logging
      - signalfx
      processors:
      - memory_limiter
      - batch
      receivers:
      - hostmetrics
      - signalfx
EOF
    destination = "local/config/otel-agent-config.yaml"
}

```

### Gateway

The Splunk OpenTelemetry Collector can run as a `gateway` by registering a
[service](https://www.nomadproject.io/docs/schedulers#service) job.

```shell-session
git clone https://github.com/signalfx/splunk-otel-collector.git
cd splunk-otel-collector/deployments/nomad
nomad run otel-gateway.nomad
```

### Agent

The Splunk OpenTelemetry Collector can run as an `agent` by registering a
[system](https://www.nomadproject.io/docs/schedulers#system) job.

```shell-session
git clone https://github.com/signalfx/splunk-otel-collector.git
cd splunk-otel-collector/deployments/nomad
nomad run otel-agent.nomad
```

### Collecting Metrics from Nomad

While there are no specific Nomad receivers in the Splunk OpenTelemetry Collector,
you can collect [Nomad metrics](https://developer.hashicorp.com/nomad/docs/operations/metrics-reference)
by adding a properly configured [Prometheus receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/prometheusreceiver/README.md#prometheus-receiver)
to the configuration file used by the Splunk OpenTelemetry Collector.

If Nomad is using a Docker driver, you can also configure the
[Docker Stats Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/dockerstatsreceiver#docker-stats-receiver)
to collect container metrics.
