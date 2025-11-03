# Scripted Inputs Receiver Deprecation

The Scripted Inputs Receiver has been deprecated and will be removed in a future release.

## Replacement guidance

The Scripted Inputs Receiver was designed to replicate log collection behavior of the Splunk Universal Forwarder when the [Unix and Linux Technical Add-on](https://docs.splunk.com/Documentation/AddOns/released/UnixLinux/About) is installed. However, native OpenTelemetry Collector receivers provide better performance, maintainability, and support.

### Recommended Replacements

For system metrics collection, we recommend using the [Host Metrics Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver), which provides comprehensive system monitoring capabilities.

#### Migration Mapping

Below is a mapping of scripted input scripts to their recommended OTel Collector receiver alternatives:

| Scripted Input | Recommended Receiver | Notes |
|----------------|---------------------|--------|
| `cpu`, `iostat`, `vmstat` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | Use CPU, disk, memory, and process scrapers |
| `df` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | Use filesystem scraper |
| `netstat`, `bandwidth`, `interfaces`, `protocol` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | Use network scraper |
| `nfsiostat` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | Use NFS scraper |
| `ps`, `top`, `lsof` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) | Use process scraper |
| `package`, `service`, `version`, `hardware` | [filelogreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) | For log-based data collection |
| `lastlog`, `who`, `passwd`, `usersWithLoginPrivs` | [filelogreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) | Monitor system log files |
| `rlog` | [filelogreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) | Monitor `/var/log/audit/audit.log` |
| `openPorts`, `openPortsEnhanced` | Custom scripting with [execreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/execreceiver) | Or monitor via network scraper |
| `selinuxChecker`, `sshdChecker`, `vsftpdChecker` | [filelogreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) | Monitor configuration files |
| `time`, `uptime` | [hostmetricsreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) or [ntpreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/ntpreceiver) | For uptime and time sync monitoring |
| `update` | Custom scripting with [execreceiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/execreceiver) | For package update information |

### Example: Migrating from df script to Host Metrics Receiver

**Before (Scripted Inputs):**
```yaml
receivers:
  scripted_inputs/df:
    script_name: df
    collection_interval: 10s
    source: df
    sourcetype: df

service:
  pipelines:
    logs:
      receivers: [scripted_inputs/df]
      processors: [memory_limiter, batch]
      exporters: [splunk_hec]
```

**After (Host Metrics Receiver):**
```yaml
receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      filesystem:
        metrics:
          system.filesystem.usage:
            enabled: true
          system.filesystem.utilization:
            enabled: true

processors:
  resource:
    attributes:
      - key: com.splunk.source
        value: hostmetrics
        action: upsert
      - key: com.splunk.sourcetype
        value: otel
        action: upsert

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [resource, memory_limiter, batch]
      exporters: [splunk_hec]
```

### Additional Resources

- [Host Metrics Receiver Documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver)
- [File Log Receiver Documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [Exec Receiver Documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/execreceiver) (for custom script execution)
- [Splunk OpenTelemetry Collector Configuration Examples](https://github.com/signalfx/splunk-otel-collector/tree/main/examples)

## Timeline

- **Deprecation Notice**: Current release
- **Planned Removal**: Future release (to be announced)

Please plan to migrate to the recommended alternatives at your earliest convenience.

