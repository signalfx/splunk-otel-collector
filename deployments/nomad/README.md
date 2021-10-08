# Splunk OpenTelemetry Collector on HashiCorp Nomad

The Splunk OpenTelemetry Collector for HashiCorp Nomad is orchestrator deployment to create job which provides a unified way to receive, process and export metric, and trace data for [Splunk Observability Cloud](https://www.observability.splunk.com/).

_These job files are provided as a reference only and are not designed for production use._

## Deployment

To run the job files you will need access to a Nomad cluster and, optionally, a
Consul cluster as well. You can start a local dev agent for Nomad and Consul by
downloading the [`nomad`](https://www.nomadproject.io/downloads) and
[`consul`](https://www.consul.io/downloads) binary and running the following
commands in two different terminals:

```shell-session
$ nomad agent -dev -network-interface='{{ GetPrivateInterfaces | attr "name" }}'
```

```shell-session
$ consul agent -dev
```
### Usage Guide

To deploy the Splunk OpenTelemetry Collector job on the nomad cluster we need to set environment variable in nomad job configuration. 

```yaml
env {
    SPLUNK_ACCESS_TOKEN = "YOUR_SPLUNK_ACCESS_TOKEN"
    SPLUNK_REALM = "YOUR_SPLUNK_REALM"
    SPLUNK_MEMORY_TOTAL_MIB = 2048
    // We can specify more environment variable to override default values.
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
generators`, to collect metrics and traces and export them using `SignalFx` exporter.

```shell-session
$ nomad run deployments/nomad/otel-demo.nomad
```
