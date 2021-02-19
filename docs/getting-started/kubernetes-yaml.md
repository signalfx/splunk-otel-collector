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

An example configuration for deploying a gateway instance is available
[here](../../examples/kubernetes-yaml/splunk-otel-collector-gateway.yaml).

> Please be advised you must ensure all the TODO items in the YAML are updated
> correctly. You cannot deploy the example YAML as-is.

Once the YAML file is updated you can deploy it with:

```bash
$ kubectl apply -f splunk-otel-collector-gateway.yaml
```

If you want to deploy as an agent, it is recommended to configure a DaemonSet.
