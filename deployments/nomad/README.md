# Splunk OpenTelemetry Collector for HashiCorp Nomad

The Splunk OpenTelemetry Collector for HashiCorp Nomad is an orchestrator deployment to create a job which provides a unified way to receive, process and export metric, and trace data for [Splunk Observability Cloud](https://www.observability.splunk.com/).

**NOTE**: _Job files are provided as a reference only and are not intended for production use._

## Deployment

To run the job files you need:

- Access to a Nomad cluster (**version < 1.3.0**)
- (Optional) Access to a Consul cluster

To start a local dev agent for Nomad and Consul, download the
[`nomad`](https://www.nomadproject.io/downloads) binary file and
[`consul`](https://www.consul.io/downloads) binary and run the following
commands in two different terminals:

```shell-session
$ nomad agent -dev -network-interface='{{ GetPrivateInterfaces | attr "name" }}'
```

```shell-session
$ consul agent -dev
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
  health_check: null
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
    ballast_size_mib: ${SPLUNK_BALLAST_SIZE_MIB}
    check_interval: 2s
    limit_mib: ${SPLUNK_MEMORY_LIMIT_MIB}
exporters:
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    api_url: https://api.${SPLUNK_REALM}.signalfx.com
    correlation: null
    ingest_url: https://ingest.${SPLUNK_REALM}.signalfx.com
    sync_host_metadata: true
  logging:
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
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector/deployments/nomad
$ nomad run otel-gateway.nomad
```

### Agent

The Splunk OpenTelemetry Collector can run as an `agent` by registering a
[system](https://www.nomadproject.io/docs/schedulers#system) job.

```shell-session
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector/deployments/nomad
$ nomad run otel-agent.nomad
```
