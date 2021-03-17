# Migrating from the Smart Agent to the Splunk distribution of OpenTelemetry Collector

The [SignalFx Smart Agent's](https://github.com/signalfx/signalfx-agent/blob/master/README.md)
metric monitors allow realtime insights into how your target services and
applications are performing.  These metric gathering utilities have an
equivalent counterpart in the OpenTelemetry Collector, the metric receiver.
The [Smart Agent Receiver](./README.md) is a wrapper utility that allows the
embedding of Smart Agent monitors within your Collector pipelines.

The Smart Agent Receiver works in much the same way your Smart Agent deployment
does and is based on the relocation of your desired monitor configuration into
that of your Collector deployment.

Given an example Smart Agent monitor configuration:

```yaml
signalFxAccessToken: {"#from": "env:SIGNALFX_ACCESS_TOKEN"}
ingestUrl: https://ingest.us1.signalfx.com
apiUrl: https://api.us1.signalfx.com

bundleDir: /opt/my-smart-agent-bundle

observers:
  - type: k8s-api

collectd:
  readThreads: 10
  writeQueueLimitHigh: 1000000
  writeQueueLimitLow: 600000

monitors:
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
```

Here is an equivalent, recommended Collector configuration utilizing the
[Receiver Creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/receiver/receivercreator/README.md):

```yaml
extensions:
  k8s_observer:
    auth_type: serviceAccount
    node: ${K8S_NODE_NAME}
  smartagent:
    bundleDir: /opt/my-smart-agent-bundle
    collectd:
      readThreads: 10
      writeQueueLimitHigh: 1000000
      writeQueueLimitLow: 600000

receivers:
  receiver_creator:
    receivers:
      smartagent/activemq:
        rule: type.port && pod.name == "activemq" && port == 1099
        config:
          type: collectd/activemq
          extraDimensions:
            my_dimension: my_dimension_value
      smartagent/apache:
        rule: type.port && pod.name == "apache" && port == 80
        config:
          type: collectd/apache
      smartagent/postgresql:
        rule: type.port && pod.name == "postgresql" && port == 7199
        config:
          type: postgresql
          extraDimensions:
            my_other_dimension: my_other_dimension_value
    watch_observers:
      - k8s_observer

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

service:
  extensions:
    - k8s_observer
    - smartagent
  pipelines:
    metrics:
      receivers:
        - receivor_creator
      processors:
        - resourcedetection
      exporters:
        - signalfx
```
