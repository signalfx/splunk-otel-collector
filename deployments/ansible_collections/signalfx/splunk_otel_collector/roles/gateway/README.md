# Gateway Role

Deploy OpenTelemetry Collector in gateway mode for centralized aggregation, batching, and routing of telemetry data.

## Description

This role deploys the Splunk OpenTelemetry Collector in gateway mode, providing a centralized tier for:
- Aggregating telemetry from multiple collectors
- Batching and buffering data
- Routing data to multiple backends
- Policy enforcement (sampling, filtering, enrichment)
- Load balancing and high availability

## Requirements

- Ansible 2.15 or higher
- Supported OS families: Debian, RedHat, Suse
- Python 3.x on the target hosts

## Role Variables

### Required Variables

- `otel_gateway_exporters`: List of exporters (at least one required)
  ```yaml
  otel_gateway_exporters:
    - name: otlphttp
      endpoint: "https://backend:4318"
    - name: splunk_hec
      endpoint: "https://splunk:8088/services/collector"
      token: "{{ splunk_hec_token }}"
  ```

### Network Configuration

- `otel_gateway_listen_host`: Bind address (default: "0.0.0.0")
- `otel_gateway_otlp_http_port`: OTLP HTTP port (default: 4318)
- `otel_gateway_otlp_grpc_port`: OTLP gRPC port (default: 4317)
- `otel_gateway_health_port`: Health check port (default: 13133)

### TLS Configuration

- `otel_gateway_tls_enabled`: Enable TLS (default: false)
- `otel_gateway_tls_cert`: Path to TLS certificate
- `otel_gateway_tls_key`: Path to TLS private key
- `otel_gateway_tls_ca`: Path to CA certificate (optional)

### Resource Configuration

- `otel_gateway_memory_limit_mib`: Memory limit in MiB (default: 2048)
- `otel_gateway_memory_spike_limit_mib`: Memory spike limit (default: 512)
- `otel_gateway_batch_timeout`: Batch timeout (default: "5s")
- `otel_gateway_batch_send_size`: Batch send size (default: 8192)

### Package Configuration

- `otel_gateway_package_name`: Package name (default: splunk-otel-collector)
- `otel_gateway_version`: Package version (default: latest)
- `otel_gateway_config_path`: Config file path (default: /etc/otel/collector/gateway_config.yaml)

### Service Configuration

- `otel_gateway_service_name`: Systemd service name (default: splunk-otel-collector-gateway)
- `otel_gateway_user`: Service user (default: splunk-otel-collector)
- `otel_gateway_group`: Service group (default: splunk-otel-collector)
- `start_gateway_service`: Enable and start service (default: true)

### Additional Variables

- `otel_gateway_additional_env_vars`: Additional environment variables (dict)

## Dependencies

None

## Example Playbook

### Basic Gateway Deployment

```yaml
- hosts: gateway_servers
  become: true
  roles:
    - role: signalfx.splunk_otel_collector.gateway
      vars:
        otel_gateway_exporters:
          - name: otlphttp
            endpoint: "https://ingest.us0.signalfx.com/v2/datapoint/otlp"
            headers:
              X-SF-Token: "{{ splunk_access_token }}"
```

### Gateway with TLS and Multiple Backends

```yaml
- hosts: gateway_servers
  become: true
  roles:
    - role: signalfx.splunk_otel_collector.gateway
      vars:
        otel_gateway_tls_enabled: true
        otel_gateway_tls_cert: /etc/ssl/certs/gateway.crt
        otel_gateway_tls_key: /etc/ssl/private/gateway.key
        otel_gateway_memory_limit_mib: 4096
        otel_gateway_exporters:
          - name: otlphttp/splunk
            endpoint: "https://ingest.us0.signalfx.com/v2/datapoint/otlp"
            headers:
              X-SF-Token: "{{ splunk_access_token }}"
          - name: otlphttp/backup
            endpoint: "https://backup-collector:4318"
            tls_insecure_skip_verify: false
```

### High-Performance Gateway

```yaml
- hosts: gateway_servers
  become: true
  roles:
    - role: signalfx.splunk_otel_collector.gateway
      vars:
        otel_gateway_memory_limit_mib: 8192
        otel_gateway_memory_spike_limit_mib: 1024
        otel_gateway_batch_timeout: "10s"
        otel_gateway_batch_send_size: 16384
        otel_gateway_exporters:
          - name: otlphttp
            endpoint: "https://ingest.us0.signalfx.com/v2/datapoint/otlp"
            headers:
              X-SF-Token: "{{ splunk_access_token }}"
            timeout: 30s
```

## Architecture

The gateway role deploys the collector with:

1. **Receivers**: OTLP (gRPC + HTTP) with optional TLS
2. **Processors**: 
   - memory_limiter: Prevents OOM conditions
   - batch: Batches telemetry for efficiency
3. **Exporters**: Configured via `otel_gateway_exporters`
4. **Extensions**: health_check for monitoring

All three signal types (traces, metrics, logs) flow through the same pipeline.

## Firewall Configuration

Ensure the following ports are accessible:

- `4317/tcp`: OTLP gRPC
- `4318/tcp`: OTLP HTTP
- `13133/tcp`: Health check endpoint

## Health Checks

Health check endpoint: `http://<gateway_host>:13133/`

## License

Apache-2.0

## Author Information

Cisco/Splunk OpenTelemetry Team
