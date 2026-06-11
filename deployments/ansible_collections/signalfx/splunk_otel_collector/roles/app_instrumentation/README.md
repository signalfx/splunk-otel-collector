# app_instrumentation

Configure OTel Collector to receive and forward application-generated telemetry.

This role configures OTLP receivers (gRPC and HTTP) for applications instrumented with OpenTelemetry SDKs, and forwards the telemetry to a gateway.

## Requirements

- OpenTelemetry Collector must be installed
- The `signalfx.splunk_otel_collector.otel_collector_pipeline` module must be available
- A configured OTel gateway endpoint

## Role Variables

### Required

- `otel_app_gateway`: Gateway OTLP endpoint URL (e.g., `http://gateway:4318`)

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `otel_app_config_path` | `/etc/otel/collector/agent_config.yaml` | Path to OTel collector configuration |
| `otel_app_listen_grpc` | `0.0.0.0:4317` | OTLP gRPC receiver listen address |
| `otel_app_listen_http` | `0.0.0.0:4318` | OTLP HTTP receiver listen address |
| `otel_app_signal_types` | `[traces, metrics, logs]` | Signal types to configure |
| `otel_app_sampling_ratio` | `1.0` | Sampling ratio (1.0 = no sampling, 0.1 = 10%) |
| `otel_app_memory_limit_mib` | `512` | Memory limiter threshold (MiB) |
| `otel_app_memory_spike_limit_mib` | `128` | Memory spike limit (MiB) |
| `otel_app_batch_timeout` | `5s` | Batch processor timeout |
| `otel_app_tls_cert` | `""` | Optional TLS certificate path |
| `otel_app_tls_key` | `""` | Optional TLS key path |
| `otel_app_restart_collector` | `true` | Restart collector after configuration |

## Dependencies

None

## Example Playbook

### Basic usage

```yaml
- hosts: app_servers
  roles:
    - role: signalfx.splunk_otel_collector.app_instrumentation
      vars:
        otel_app_gateway: http://otel-gateway:4318
```

### Traces only with sampling

```yaml
- hosts: app_servers
  roles:
    - role: signalfx.splunk_otel_collector.app_instrumentation
      vars:
        otel_app_gateway: http://otel-gateway:4318
        otel_app_signal_types:
          - traces
        otel_app_sampling_ratio: 0.1  # 10% sampling
```

### With TLS

```yaml
- hosts: app_servers
  roles:
    - role: signalfx.splunk_otel_collector.app_instrumentation
      vars:
        otel_app_gateway: https://otel-gateway:4318
        otel_app_tls_cert: /etc/otel/certs/client.crt
        otel_app_tls_key: /etc/otel/certs/client.key
```

### Custom listen addresses

```yaml
- hosts: app_servers
  roles:
    - role: signalfx.splunk_otel_collector.app_instrumentation
      vars:
        otel_app_gateway: http://otel-gateway:4318
        otel_app_listen_grpc: "127.0.0.1:4317"
        otel_app_listen_http: "127.0.0.1:4318"
```

## Configured Pipelines

This role creates one pipeline per signal type in `otel_app_signal_types`:

### traces/app

Receives OTLP traces from instrumented applications via gRPC (4317) and HTTP (4318), applies optional probabilistic sampling, and forwards to the gateway.

### metrics/app

Receives OTLP metrics from instrumented applications and forwards to the gateway.

### logs/app

Receives OTLP logs from instrumented applications and forwards to the gateway.

## Instrumentation

Applications should be configured to send telemetry to the collector using the OTLP protocol:

### Environment variables

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
```

Or for gRPC:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

### Code example (Python)

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint="http://localhost:4318/v1/traces"))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)
```

## License

Apache-2.0

## Author

Cisco/Splunk OpenTelemetry Team
