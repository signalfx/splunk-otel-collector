# Migrating from the SignalFx Smart Agent

The Splunk Distribution of OpenTelemetry Collector is the next-generation agent
and gateway for Splunk Observability products. As such, it is the replacement
for the [SignalFx Smart Agent](https://github.com/signalfx/signalfx-agent).
This distribution provides helpful components to assist current Smart Agent
users in their transition to OpenTelemetry Collector and ensure no functionality
loss.  The [Smart Agent
Receiver](../internal/receiver/smartagentreceiver/README.md), its associated
[extension](../internal/extension/smartagentextension/README.md), and other
Collector components provide a means of integrating all Smart Agent metric
monitors into your Collector pipelines.

The [SignalFx Smart Agent's](https://github.com/signalfx/signalfx-agent/blob/main/README.md)
metric monitors allow real-time insights into how your target services and
applications are performing.  These metric gathering utilities have an
equivalent counterpart in the OpenTelemetry Collector, the metric receiver.
The [Smart Agent Receiver](../internal/receiver/smartagentreceiver/README.md)
is a wrapper utility that allows the embedding of Smart Agent monitors within
your Collector pipelines.

Based on the relocation of your desired monitor configurations in your Collector
deployment, the Smart Agent Receiver works in much the same way your Smart Agent
deployment does.

Given an example Smart Agent monitor configuration:

```yaml
signalFxAccessToken: {"#from": "env:SIGNALFX_ACCESS_TOKEN"}
ingestUrl: https://ingest.us1.signalfx.com
apiUrl: https://api.us1.signalfx.com

bundleDir: /opt/my-smart-agent-bundle

procPath: /my_custom_proc
etcPath: /my_custom_etc
varPath: /my_custom_var
runPath: /my_custom_run
sysPath: /my_custom_sys

observers:
  - type: k8s-api

collectd:
  readThreads: 10
  writeQueueLimitHigh: 1000000
  writeQueueLimitLow: 600000
  configDir: "/tmp/signalfx-agent/collectd"

monitors:
  - type: signalfx-forwarder
    listenAddress: 0.0.0.0:9080
  - type: collectd/activemq
    discoveryRule: container_image =~ "activemq" && private_port == 1099
    extraDimensions:
      my_dimension: my_dimension_value
  - type: collectd/apache
    discoveryRule: container_image =~ "apache" && private_port == 80
  - type: postgresql
    discoveryRule: container_image =~ "postgresql" && private_port == 7199
    extraDimensions:
      my_other_dimension: my_other_dimension_value
  - type: processlist
```

Below is an equivalent, recommended Collector configuration.  Notice that the
`signalfx-forwarder` monitor's associated `smartagent/signalfx-forwarder` receiver instance
is part of both `metrics` and `traces` pipelines using the `signalfx` and `sapm` exporters,
respectively. Also note the `processlist` monitor's associated `smartagent/processlist` receiver
instance is part of `logs` pipeline using the `resourcedetection` processor and a `signalfx` exporter.
The additional metric monitors utilize the
[Receiver Creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/receivercreator/README.md):

```yaml
extensions:
  k8s_observer:
    auth_type: serviceAccount
    node: ${K8S_NODE_NAME}
  smartagent:
    bundleDir: /opt/my-smart-agent-bundle
    procPath: /my_custom_proc
    etcPath: /my_custom_etc
    varPath: /my_custom_var
    runPath: /my_custom_run
    sysPath: /my_custom_sys
    collectd:
      readThreads: 10
      writeQueueLimitHigh: 1000000
      writeQueueLimitLow: 600000
      configDir: "/tmp/signalfx-agent/collectd"

receivers:
  smartagent/signalfx-forwarder:
    type: signalfx-forwarder
    listenAddress: 0.0.0.0:9080
  receiver_creator:
    receivers:
      smartagent/activemq:
        rule: type == "port" && pod.name matches "activemq" && port == 1099
        config:
          type: collectd/activemq
          extraDimensions:
            my_dimension: my_dimension_value
      smartagent/apache:
        rule: type == "port" && pod.name matches "apache" && port == 80
        config:
          type: collectd/apache
      smartagent/postgresql:
        rule: type == "port" && pod.name matches "postgresql" && port == 7199
        config:
          type: postgresql
          extraDimensions:
            my_other_dimension: my_other_dimension_value
    watch_observers:
      - k8s_observer
  smartagent/processlist:
    type: processlist

processors:
  resourcedetection:
    detectors:
      - system
      - env
  k8s_tagger:
    extract:
      metadata:
        - namespace
        - node
        - podName
        - podUID
    filter:
      node_from_env_var: K8S_NODE_NAME
  resource/add_cluster_name:
    attributes:
      - action: upsert
        key: k8s.cluster.name
        value: my_desired_cluster_name

exporters:
  signalfx:
    access_token: "${SIGNALFX_ACCESS_TOKEN}"
    realm: us1
  sapm:
    access_token: "${SIGNALFX_ACCESS_TOKEN}"
    endpoint: https://ingest.us1.signalfx.com/v2/trace

service:
  extensions:
    - k8s_observer
    - smartagent
  pipelines:
    metrics:
      receivers:
        - receiver_creator
        - smartagent/signalfx-forwarder
      processors:
        - resourcedetection
      exporters:
        - signalfx
    traces:
      receivers:
        - smartagent/signalfx-forwarder
      processors:
        - resourcedetection
      exporters:
        - sapm
    logs:
      receivers:
        - smartagent/processlist
      processors:
        - resourcedetection
      exporters:
        - signalfx
```
