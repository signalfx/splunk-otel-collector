# host_monitoring

Turnkey host metrics and logs collection role for OpenTelemetry Collector.

This role configures OTel Collector pipelines for host-level observability, including system metrics and log file monitoring.

## Requirements

- OpenTelemetry Collector must be installed
- The `signalfx.splunk_otel_collector.otel_collector_pipeline` module must be available
- A configured OTel gateway endpoint

## Role Variables

### Required

- `otel_host_monitoring_gateway`: Gateway OTLP endpoint URL (e.g., `http://gateway:4318`)

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `otel_host_monitoring_config_path` | `/etc/otel/collector/agent_config.yaml` | Path to OTel collector configuration |
| `otel_host_monitoring_metrics` | `true` | Enable host metrics collection |
| `otel_host_monitoring_log_paths` | `[/var/log/syslog, /var/log/messages]` | Log file paths to monitor |
| `otel_host_monitoring_syslog_enabled` | `true` | Enable syslog receiver |
| `otel_host_monitoring_syslog_protocol` | `rfc5424` | Syslog protocol (rfc3164 or rfc5424) |
| `otel_host_monitoring_syslog_listen` | `0.0.0.0:514` | Syslog listen address |
| `otel_host_monitoring_memory_limit_mib` | `256` | Memory limiter threshold (MiB) |
| `otel_host_monitoring_memory_spike_limit_mib` | `64` | Memory spike limit (MiB) |
| `otel_host_monitoring_batch_timeout` | `5s` | Batch processor timeout |
| `otel_host_monitoring_tls_cert` | `""` | Optional TLS certificate path |
| `otel_host_monitoring_tls_key` | `""` | Optional TLS key path |
| `otel_host_monitoring_restart_collector` | `true` | Restart collector after configuration |

## Dependencies

None

## Example Playbook

### Basic usage

```yaml
- hosts: servers
  roles:
    - role: signalfx.splunk_otel_collector.host_monitoring
      vars:
        otel_host_monitoring_gateway: http://otel-gateway:4318
```

### Custom log paths

```yaml
- hosts: webservers
  roles:
    - role: signalfx.splunk_otel_collector.host_monitoring
      vars:
        otel_host_monitoring_gateway: http://otel-gateway:4318
        otel_host_monitoring_log_paths:
          - /var/log/nginx/access.log
          - /var/log/nginx/error.log
          - /var/log/syslog
```

### With TLS

```yaml
- hosts: servers
  roles:
    - role: signalfx.splunk_otel_collector.host_monitoring
      vars:
        otel_host_monitoring_gateway: https://otel-gateway:4318
        otel_host_monitoring_tls_cert: /etc/otel/certs/client.crt
        otel_host_monitoring_tls_key: /etc/otel/certs/client.key
```

### Metrics only (no logs)

```yaml
- hosts: servers
  roles:
    - role: signalfx.splunk_otel_collector.host_monitoring
      vars:
        otel_host_monitoring_gateway: http://otel-gateway:4318
        otel_host_monitoring_log_paths: []
        otel_host_monitoring_syslog_enabled: false
```

## Configured Pipelines

### metrics/host

Collects host-level metrics:
- CPU usage
- Memory usage
- Disk I/O
- Network statistics
- Filesystem usage

### logs/host

Collects logs from:
- File paths specified in `otel_host_monitoring_log_paths`
- Syslog (if enabled)

## License

Apache-2.0

## Author

Cisco/Splunk OpenTelemetry Team
