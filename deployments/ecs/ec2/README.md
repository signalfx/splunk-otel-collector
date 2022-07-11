# Amazon ECS EC2 Deployment
Familiarity with Amazon ECS using launch type EC2 is assumed. Consult the 
[Getting started with the Amazon ECS console using Amazon EC2](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/getting-started-ecs-ec2.html)
for further reading.

The [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector)
(Collector) can be run as a Daemon service in an ECS cluster with EC2 launch type.

Requires Collector release v0.34.1 or newer which corresponds to image tag 0.34.1 and newer.
See image repository [here](https://quay.io/repository/signalfx/splunk-otel-collector?tab=tags).

## Getting Started
### Create Task Definition
Take the task definition JSON for the Collector [here](./splunk-otel-collector.json), replace
`MY_SPLUNK_ACCESS_TOKEN` and `MY_SPLUNK_REALM` with valid values.
We recommend pinning the [image version](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/ecs/ec2/splunk-otel-collector.json#L56) to a specific version instead of latest to avoid upgrade issues. Use the JSON to create a task definition of **EC2 launch type** following
the instructions [here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/create-task-definition.html).
The supplied task definition is a minimal definition. More options for configuration may be found in the
[AWS task definition parameters](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) documentation.

The Collector is configured to use the default configuration file `/etc/otel/collector/ecs_ec2_config.yaml`.
The Collector image Dockerfile is available [here](../../../cmd/otelcol/Dockerfile) and the contents of the default
configuration file can be seen [here](../../../cmd/otelcol/config/collector/ecs_ec2_config.yaml).

The suggested configured network mode for the task is **host**. This means that **task metadata
endpoint version 2** used by receiver `smartagent/ecs-metadata` is not enabled by default. See
[here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html)
to determine if **task metadata endpoint version 4** is enabled by default for your task. If so
the default configuration for ECS with EC2 launch type already uses it to form the metadata and stats endpoints
for the receiver to query to generate metrics and no task definition change is required. If you're using an alternate
task stats or metadata endpoint, configure them via the `ECS_TASK_METADATA_ENDPOINT` and `ECS_TASK_STATS_ENDPOINT`
environment variables list in your container definition as necessary:

```json
"environment": [
...
  {
    "name": "ECS_TASK_METADATA_ENDPOINT",
    "value": "<MY_TASK_METADATA_ENDPOINT>/task"
  },
  {
    "name": "ECS_TASK_STATS_ENDPOINT",
    "value": "<MY_TASK_METADATA_ENDPOINT>/task/stats"
  },
...
]
```

**Note**: You do not need the `smartagent/ecs-metadata` metrics receiver in the default
configuration file if all you want is tracing or logs. You can take the default configuration,
remove the receiver, then use the configuration in a custom configuration following the direction
in the [custom configuration](#custom-configuration) section.

The default configuration includes a filter processor that allows you to specify metrics not to report via the
`METRICS_TO_EXCLUDE` environment variable. Assign a stringified array of metrics you want excluded:

```json
"environment": [
...
  {
    "name": "METRICS_TO_EXCLUDE",
    "value": "[\"excluded.metric.name.regex.pattern.one\", \"excluded.metric.name.regex.pattern.two\"]"
  },
...
]
```

You can set the memory limit for the memory limiter processor using environment variable `SPLUNK_MEMORY_LIMIT_MIB`.
For more information about the memory limiter processor, see [its documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/memorylimiterprocessor/README.md).

### Launch the Collector
The Collector is designed to be run as a Daemon service in an EC2 ECS cluster.

To create a Collector service from the Amazon ECS console:

Go to your cluster in the console
1. Click on the "Services" tab.
2. Click "Create" at the top of the tab.
3. Select:
   - Launch Type -> EC2
   - Task Definition (Family) -> splunk-otel-collector
   - Task Definition (Revision) -> 1 (or whatever the latest is in your case)
   - Service Name -> splunk-otel-collector
   - Service type -> DAEMON
4. Leave everything else at default and click "Next step"
5. Leave everything on this next page at their defaults and click "Next step". 
6. Leave everything on this next page at their defaults and click "Next step". 
7. Click "Create Service" and the collector should be deployed onto each node in the ECS cluster. You should see infrastructure and docker metrics flowing soon.

## Custom Configuration
To use a custom configuration file, replace the value of environment variable
`SPLUNK_CONFIG` with the file path of the custom configuration file in Collector
task definition.

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

### Direct Configuration
The Collector provides environment variable `SPLUNK_CONFIG_YAML` for specifying the
configuration YAML directly which can be used instead of `SPLUNK_CONFIG`.

For example, you can store the custom configuration above in a parameter called
`splunk-otel-collector-config` in **AWS Systems Manager Parameter Store**. Then
assign the parameter to environment variable `SPLUNK_CONFIG_YAML` using `valueFrom`.

Using the aws-cli to create the parameter from a config file:

```bash
aws ssm put-parameter --tier Intelligent-Tiering --type String --name <My Collector Config Parameter Name> --value file://<my-collector-config.yaml>
```

Adding the required `secrets` mapping in the Collector container definition (being sure to remove any existing `SPLUNK_CONFIG` environment variable):

```json
"secrets": [
  {
    "name": "SPLUNK_CONFIG_YAML",
    "valueFrom": "<My Collector Config Parameter Name>"
  }
]
```

**Note:** Using the [`Intelligent-Tiering` tier mode](https://docs.aws.amazon.com/cli/latest/reference/ssm/put-parameter.html) will automatically
determine whether to use [standard or advanced SSM parameter tiers](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-advanced-parameters.html) based on the provided config size. The maximum size for standard is 4KB and 8KB for advanced. Usage charges will apply for configuration requiring the advanced parameter tier.

**Note:** You should add the `AmazonSSMReadOnlyAccess` policy to the task role in order for
the task to have read access to the Parameter Store.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:Describe*",
                "ssm:Get*",
                "ssm:List*"
            ],
            "Resource": "*"
        }
    ]
}
```
