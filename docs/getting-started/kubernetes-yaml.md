> The official Splunk documentation for this page is [Install on Kubernetes](https://docs.splunk.com/Observability/gdi/opentelemetry/install-k8s.html). For instructions on how to contribute to the docs, see [CONTRIBUTE.md](../CONTRIBUTE.md).
# Kubernetes YAML

The easiest and recommended way to get started on Kubernetes is to leverage the
[Splunk Connector for
Kubernetes](https://github.com/signalfx/splunk-otel-collector-chart), which
provides a configurable Helm chart. Alternatively, YAML files can be manually
applied and maintained to get started.

## Getting Started

Given the Collector can be deployed as [an agent or
gateway](https://github.com/signalfx/splunk-otel-collector#getting-started),
the YAML required depends on the configuration.

An example configuration for deploying either as an agent or a gateway instance
is available
[here](../../examples/kubernetes-yaml/splunk-otel-collector-gateway.yaml).

> Please be advised you MUST configure at least the SPLUNK_ACCESS_TOKEN and
> SPLUNK_REALM for the configuration to work. Several additional TODO sections
> can be updated as needed.

Once the YAML file is updated you can deploy it with with `kubectl`.

Agent:

```bash
$ kubectl apply -f splunk-otel-collector-agent.yaml
```

> Note agent is configured to send direct to Splunk SaaS endpoints. It can be
> reconfigured to send to a gateway.

Gateway instance:

```bash
$ kubectl apply -f splunk-otel-collector-gateway.yaml
```
