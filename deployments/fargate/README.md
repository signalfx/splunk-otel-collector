# AWS Fargate Deployment
Familiarity with AWS Fargate (Fargate) is assumed. Consult the 
[User Guide for AWS Fargate](https://docs.aws.amazon.com/AmazonECS/latest/userguide/what-is-fargate.html)
for further reading.

Unless stated otherwise, the
[Splunk OpenTelemetry Connector](https://github.com/signalfx/splunk-otel-collector)
(Collector) is deployed as a **sidecar** (additional container) to ECS tasks.

Requires Connector release v0.33.0 or newer which corresponds to image tag 0.33.0 and newer.
See image repository [here](https://quay.io/repository/signalfx/splunk-otel-collector?tab=tags).

## Getting Started
Copy the default Collector container definition JSON below. Replace `MY_SPLUNK_ACCESS_TOKEN`
and `MY_SPLUNK_REALM` with valid values. Update the image tag to the newest
version then add the JSON to the `containerDefinitions` section of your task definition
JSON.
```json
{
  "environment": [
    {
      "name": "SPLUNK_ACCESS_TOKEN",
      "value": "MY_SPLUNK_ACCESS_TOKEN"
    },
    {
      "name": "SPLUNK_REALM",
      "value": "MY_SPLUNK_REALM"
    },
    {
      "name": "SPLUNK_CONFIG",
      "value": "/etc/otel/collector/fargate_config.yaml"
    },
    {
      "name": "ECS_METADATA_EXCLUDED_IMAGES",
      "value": "[\"quay.io/signalfx/splunk-otel-collector\"]"
    }
  ],
  "image": "quay.io/signalfx/splunk-otel-collector:0.33.0",
  "essential": true,
  "name": "splunk_otel_collector"
}
```
In the above container definition the Collector is configured to use the default
configuration file `/etc/otel/collector/fargate_config.yaml`. The Collector image Dockerfile
is available [here](../../cmd/otelcol/Dockerfile) and the contents of the default
configuration file can be seen [here](../../cmd/otelcol/config/collector/fargate_config.yaml).
Note that the receiver `smartagent/ecs-metadata` is enabled by default.

In summary, the default Collector container definition does the following:
- Specifies the Collector image.
- Sets the access token using environment variable `SPLUNK_ACCESS_TOKEN`.
- Sets the realm using environment variable `SPLUNK_REALM`.
- Sets the default configuration file path using environment variable `SPLUNK_CONFIG`.
- Excludes `ecs-metadata` metrics from the Collector image using environment variable `ECS_METADATA_EXCLUDED_IMAGES`.

Assign a stringified array of metrics you want excluded to environment variable
`METRICS_TO_EXCLUDE`. You can set the memory limit for the memory limiter processor using
environment variable `SPLUNK_MEMORY_LIMIT_MIB`. The default memory limit is 512 MiB. For
more information about the memory limiter processor, see
[here](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/memorylimiter/README.md)

## Custom Configuration
The example below shows an excerpt of the container definition JSON for the Collector 
configured to use custom configuration file `/path/to/custom/config/file`. 
`/path/to/custom/config/file` is a placeholder value for the actual custom configuration
file path and `0.33.0` is the latest image tag at present. The custom configuration file
should be present in a volume attached to the task.
```json
{
  "environment": [
    {
      "name": "SPLUNK_CONFIG",
      "value": "/path/to/custom/config/file"
    }
  ],
  "image": "quay.io/signalfx/splunk-otel-collector:0.33.0",
  "essential": true,
  "name": "splunk_otel_collector"
}
```
The custom Collector container definition essentially:
- Specifies the Collector image.
- Sets environment variable `SPLUNK_CONFIG` with the custom configuration file path.

Alternatively, you can specify the custom configuration YAML directly using environment
variable `SPLUNK_CONFIG_YAML` as describe [below](#direct-configuration).

### ecs_observer
Use extension
[Amazon Elastic Container Service Observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/ecsobserver#amazon-elastic-container-service-observer)
(`ecs_observer`) in your custom configuration to discover metrics targets
in running tasks, filtered by service names, task definitions and container labels.
`ecs_observer` is currently limited to Prometheus targets and requires the read-only
permissions below. You can add the permissions to the task role by adding them to a 
customer-managed policy that is attached to the task role.
```text
ecs:List*
ecs:Describe*
```

Below is an example of a custom configuration in which the `ecs_observer` is configured to find
Prometheus targets in cluster `lorem-ipsum-cluster`, region `us-west-2`, where the task ARN
pattern is `^arn:aws:ecs:us-west-2:906383545488:task-definition/lorem-ipsum-task:[0-9]+$`.
The results are written to file `/etc/ecs_sd_targets.yaml`. The `prometheus` receiver is
configured to read targets from the results file. The values for `access_token`
and `realm` are read from environment variables `SPLUNK_ACCESS_TOKEN` and `SPLUNK_REALM`
respectively, which must be specified in your container definition.

```yaml
extensions:
  ecs_observer:
    refresh_interval: 10s
    cluster_name: 'lorem-ipsum-cluster'
    cluster_region: 'us-west-2'
    result_file: '/etc/ecs_sd_targets.yaml'
    task_definitions:
      - arn_pattern: "^arn:aws:ecs:us-west-2:906383545488:task-definition/lorem-ipsum-task:[0-9]+$"
        metrics_ports: [9113]
        metrics_path: /metrics
receivers:
  signalfx:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'lorem-ipsum-nginx'
          scrape_interval: 10s
          file_sd_configs:
            - files:
                - '/etc/ecs_sd_targets.yaml'
processors:
  batch:
  resourcedetection:
    detectors: [ecs]
    override: false    
exporters:
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    realm: ${SPLUNK_REALM}
service:
  extensions: [ecs_observer]
  pipelines:
    metrics:
      receivers: [prometheus]
      processors: [batch, resourcedetection]
      exporters: [signalfx]
```
**Note:** The task ARN pattern in the configuration example above will cause `ecs_observer`
to discover targets in running revisions of task `lorem-ipsum-task`. This
means that when multiple revisions of task `lorem-ipsum-task` are running, the
`ecs_observer` will discover targets outside the task in which the Collector sidecar
container is running. In a sidecar deployment the Collector and the monitored containers
are in the same task, so metric targets must be within task. This problem
can be solved by using the complete task ARN as shown below. But, now the
task ARN pattern must be updated to keep pace with task revisions.

```yaml
...
     - arn_pattern: "^arn:aws:ecs:us-west-2:906383545488:task-definition/lorem-ipsum-task:3$"
...
```

### Direct Configuration
In Fargate the filesystem is not readily available. This makes specifying the configuration
YAML directly instead of using a file more convenient. The Collector provides environment
variable `SPLUNK_CONFIG_YAML` for specifying the configuration YAML directly which can be
used instead of `SPLUNK_CONFIG`.

For example, you can store the custom configuration above in a parameter called
`splunk-otel-collector-config` in **AWS Systems Manager Parameter Store**. Then in your
Collector container definition assign the parameter to environment variable 
`SPLUNK_CONFIG_YAML` using `valueFrom`. The example below shows an excerpt of the container
definition JSON for the Collector. `MY_SPLUNK_ACCESS_TOKEN` and `MY_SPLUNK_REALM` are 
placeholder values and image tag `0.33.0` is the latest at present.

```json
{
  "environment": [
    {
      "name": "SPLUNK_ACCESS_TOKEN",
      "value": "MY_SPLUNK_ACCESS_TOKEN"
    },
    {
      "name": "SPLUNK_REALM",
      "value": "MY_SPLUNK_REALM"
    }
  ],
  "secrets": [
    {
      "valueFrom": "splunk-otel-collector-config",
      "name": "SPLUNK_CONFIG_YAML"
    }
  ],
  "image": "quay.io/signalfx/splunk-otel-collector:0.33.0",
  "essential": true,
  "name": "splunk_otel_collector"
}
```

**Note:** You should add policy `AmazonSSMReadOnlyAccess` to the task role in order for
the task to have read access to the Parameter Store.

### Standalone Task
Extension `ecs_observer` is capable of scanning for targets in the entire cluster. This
allows you to collect telemetry data by deploying the Collector in a task that is separate
from tasks containing monitored applications. This is in contrast to the sidecar deployment
where the Collector container, and the monitored application containers are in the same task.
Do not configure the ECS
[resourcedetection](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor#resource-detection-processor) 
processor for the standalone task since it would detect resources in the standalone Collector
task itself as opposed to resources in the tasks containing the monitored applications.
