# Enable APM-Infra Correlation 

The APM Service Dashboards include charts that indicate the health of
the underlying infrastructure. The default configuration of Splunk OpenTelemetry 
Collector automatically configures this for you, but when specifying a custom 
configuration it is important to keep the following in mind.



## Configuration to enable APM-Infra Correlation from a standalone VM (Agent)
To view your infrastructure data in the APM Service Dashboards, you need to enable the 
following in the opentelemetry collector:

- `hostmetrics` receiver
    - cpu, memory, filesystem and network enabled
   
    <em>This will allow the collection of cpu, mem, disk and network metrics.</em>


- `signalfx` exporter 
    
    <em>The `signalfx` exporter will aggregate the metrics from the `hostmetrics` 
    receiver and send in metrics such as `cpu.utilization`, which are referenced in 
    the relevant APM service charts.</em>

    - `correlation` flag (enabled by default)

    - `signalfx` exporter must be enabled for metrics AND traces pipeline (example below)

    <em>The `correlation` flag of the `signalfx` exporter will allow the 
    collector to make relevant API calls to correlate your spans with the 
    infrastructure metrics, but to do this, the `signalfx` exporter must be placed 
    in the traces pipeline. This flag is enabled by default, so there is no specific 
    configuration needed. You can, however, adjust the `correlation` option further (if 
    needed). See [here](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/signalfxexporter#traces-configuration-correlation-only) for more information on adjusting
    the `correlation` option.</em>

- `resourcedetection` processor 
    - cloud provider/env variable to set `host.name`
    - override enabled
    
    <em>This processor will enable a unique `host.name` value to be set for metrics 
    and traces. The `host.name` is determined by either the ec2 host name, or the 
    system host name.</em>
   

-  `resource/add_environment` processor (optional)
    
    <em>This processor inserts a `deployment.environment` span tag to all spans. The APM
    charts require the environment span tag to be set correctly. You should configure this
    span tag in the instrumentation, but if that is not an option, you can use this processor
    to insert the required `deployment.environment` span tag value.</em>

 
Here are the relevant snippets from each section:
```
...
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu:
      disk:
      filesystem:
      memory:
      network:
...
processors:
  resourcedetection:
    detectors: [system,env,gcp,ec2]
    override: true
  resource/add_environment:
    attributes:
      - action: insert
        value: staging
        key: deployment.environment
...
exporters:
  # Traces
  sapm:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    endpoint: "${SPLUNK_TRACE_URL}"
  # Metrics + Events + APM correlation calls
  signalfx:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    api_url: "${SPLUNK_API_URL}"
    ingest_url: "${SPLUNK_INGEST_URL}"
...
service:
  extensions: [health_check, http_forwarder, zpages]
  pipelines:
    traces:
      receivers: [jaeger, zipkin]
      processors: [memory_limiter, batch, resourcedetection, resource/add_environment]
      exporters: [sapm, signalfx]
    metrics:
      receivers: [hostmetrics]
      processors: [memory_limiter, batch, resourcedetection]
      exporters: [signalfx]
```     

## Configuration to enable APM-Infra Correlation from an Agent -> Gateway
If you need to run the Opentelemetry Collector in both Agent and Gateway mode, refer to
the following sections. 

### Agent
Follow the same steps as mentioned for a standalone VM and also include the following changes.

- `http_forwarder` extension
    - `egress` endpoint

    <em>The `http_forwarder` listens on port 6060 and sends all the REST API calls directly
    to the Splunk Observability Cloud. If your Agent does not have access to talk to the 
    Splunk SaaS backend directly, this should be changed to the URL of the Gateway.</em>

- `signalfx` exporter
    - `api_url` endpoint
    
    <em>The `api_url` endpoint should be set to the URL of the Gateway, and you MUST specify 
    the ingress port of the `http_forwarder` of the Gateway (6060 by default).</em>
    
    - `ingest_url` endpoint 
    
    <em>The `ingest_url` endpoint should be set to the URL of the Gateway, and just like the 
    `api_url`, you must specify the ingress port of the `signalfx` receiver of the Gateway
    (9943 by default).</em>
    
    <em>You can choose to send all your metrics via the `otlp` exporter to the gateway, but you 
    must send the REST API calls, required for trace correlation, via the `signalfx` exporter 
    in the traces pipeline. If you want, you can also use the `signalfx` exporter for metrics.</em>

- All pipelines
    - Metrics, Traces and Logs pipeline should be sent to the appropriate receivers on the 
    Gateway. 

    - `otlp` exporter (optional)
    
    <em>The `otlp` exporter uses the grpc protocol, so the endpoint must be defined as the IP 
    address of the gateway. Using the `otlp` exporter is optional, but recommended for the majority 
    of your traffic from the Agent to the gateway (see NOTE below)</em>

    **NOTE**:
        
    - If you are using the `otlp` exporter for metrics, the `hostmetrics` aggregation must be 
        done at the Gateway (example below)

    - Apart from the requirement of using the `signalfx` exporter [for making REST API 
    calls in the traces pipeline], the rest of the data between the Agent to the Gateway should 
    ideally be sent via the `otlp` exporter. Since `otlp` is the internal format that all data gets 
    converted to upon receival, it is the most efficient way to send data to the Gateway (example below)

Here are the relevant snippets from each section:
```
...
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      cpu:
      disk:
      filesystem:
      memory:
      network:
...
processors:
  resourcedetection:
    detectors: [system,env,gcp,ec2]
    override: true
  resource/add_environment:
    attributes:
      - action: insert
        value: staging
        key: deployment.environment
...
exporters:
  # Traces
  otlp:
    endpoint: "${SPLUNK_GATEWAY_URL}:4317"
    insecure: true
  # Metrics + Events + APM correlation calls
  signalfx:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    api_url: "http://${SPLUNK_GATEWAY_URL}:6060"
    ingest_url: "http://${SPLUNK_GATEWAY_URL}:9943"
...
service:
  extensions: [health_check, http_forwarder, zpages]
  pipelines:
    traces:
      receivers: [jaeger, zipkin]
      processors: [memory_limiter, batch, resourcedetection, resource/add_environment]
      exporters: [otlp, signalfx]
    metrics:
      receivers: [hostmetrics]
      processors: [memory_limiter, batch, resourcedetection]
      exporters: [otlp]
``` 

### Gateway
In Gateway mode, the relevant receivers to match the exporters from the Agent. 
In addition, you need to make the following changes.

- `http_forwarder` extension
    - `egress` endpoint

    <em>The `http_forwarder` listens on port 6060 and sends all the REST API calls directly
    to Splunk Observability Cloud. In Gateway mode, this should be changed to the 
    Splunk Observability Cloud SaaS endpoint.</em>

- `signalfx` exporter
    - `translation_rules` and `exclude_metrics` flag
    
    <em>Both flags should be set to their default value, and thus can be commented out or 
    simply removed. This will ensure that the `hostmetrics` aggregations that are normally 
    performed by the `signalfx` exporter on the Agent, are performed by the `signalfx` 
    exporter on the Gateway instead.</em>

Here are the relevant snippets from each section:
```
...
extensions:
  http_forwarder:
    egress:
      endpoint: "https://api.${SPLUNK_REALM}.signalfx.com"
...
receivers:
  otlp:
    protocols:
      grpc:
      http:
  signalfx:
...
exporters:
  # Traces
  sapm:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    endpoint: "https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace"
  # Metrics + Events
  signalfx:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    realm: "${SPLUNK_REALM}"

...
service:
  extensions: [http_forwarder]
  pipelines:
    traces:
      receivers: [otlp]
      processors:
      - memory_limiter
      - batch
      exporters: [sapm]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [signalfx]
``` 

Alternatively, if you want to use the `signalfx` exporter for metrics on both
Agent and Gateway, you need to disable the aggregation at the Gateway. To do so,
you must set the `translation_rules` and `exclude_metrics` to empty lists (example
below).

The Gateway's configuration can be modified as follows.

```
...
exporters:
  # Traces
  sapm:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    endpoint: "https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace"
  # Metrics + Events
  signalfx:
    access_token: "${SPLUNK_ACCESS_TOKEN}"
    realm: "${SPLUNK_REALM}"
    translation_rules: []
    exclude_metrics: []
...
service:
  extensions: [http_forwarder]
  pipelines:
    traces:
      receivers: [otlp]
      processors:
      - memory_limiter
      - batch
      exporters: [sapm]
    metrics:
      receivers: [signalfx]
      processors: [memory_limiter, batch]
      exporters: [signalfx]
```

