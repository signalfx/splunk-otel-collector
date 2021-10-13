# Splunk OpenTelemetry Collector for HashiCorp Nomad

The Splunk OpenTelemetry Collector for HashiCorp Nomad is an orchestrator deployment to create a job which provides a unified way to receive, process and export metric, and trace data for [Splunk Observability Cloud](https://www.observability.splunk.com/).

**NOTE**: _Job files are provided as a reference only and are not intended for production use._

## Deployment

To run the job files you need:

- Access to a Nomad cluster
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
// Splunk OpenTelemetry Collector configuration.
EOF
    destination = "local/config/otel-agent-config.yaml"
}

```

### Gateway

The Splunk OpenTelemetry Collector can run as a `gateway` by registering a
[service](https://www.nomadproject.io/docs/schedulers#service) job.

```shell-session
$ nomad run deployments/nomad/otel-gateway.nomad
```

### Agent

The Splunk OpenTelemetry Collector can run as an `agent` by registering a
[system](https://www.nomadproject.io/docs/schedulers#system) job.

```shell-session
$ nomad run deployments/nomad/otel-agent.nomad
```

### Demo

The demo job deploys the Splunk OpenTelemetry Collector as `agent` and `gateway`, `load
generators`, to collect metrics and traces and export them using the `SignalFx` exporter.

```shell-session
$ nomad run deployments/nomad/otel-demo.nomad
```
