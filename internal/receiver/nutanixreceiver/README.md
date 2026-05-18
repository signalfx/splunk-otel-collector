# Nutanix Receiver

The Nutanix receiver collects Prism metrics through Prism v4 APIs and emits
OTLP metrics directly. It replaces the `~/ai-tools/nutanix-splunk-lab0`
Prometheus exporter scrape with native collector collection.

The receiver uses `github.com/nutanix-cloud-native/prism-go-client/v4` and the
generated Nutanix v4 Go clients. The SDK version currently used by the receiver
is backed by v4.2 API paths.

## Prism v4 API Coverage

Inventory and count data is collected from:

- `/api/clustermgmt/v4.2/config/clusters`
- `/api/clustermgmt/v4.2/config/hosts`
- `/api/clustermgmt/v4.2/config/storage-containers`
- `/api/vmm/v4.2/ahv/config/vms`
- `/api/volumes/v4.2/config/volume-groups`

Runtime statistics are collected from:

- `/api/clustermgmt/v4.2/stats/clusters/{extId}`
- `/api/clustermgmt/v4.2/stats/clusters/{clusterExtId}/hosts/{extId}`
- `/api/clustermgmt/v4.2/stats/storage-containers/{extId}`
- `/api/vmm/v4.2/ahv/stats/vms`
- `/api/volumes/v4.2/stats/volume-groups/{extId}`

Each scrape queries the most recent collection interval with the v4 `LAST`
down-sampling operator.

## Exporter Metric Discovery

The legacy lab exporter creates Prometheus gauges from Prism payload keys:

| Exporter metric pattern | Source data | Receiver metric |
| --- | --- | --- |
| `nutanix_cluster_stats_<stat>` | cluster stats | `nutanix.cluster.stat` with `nutanix.stat.name` |
| `nutanix_cluster_usage_stats_<stat>` | cluster usage stats | `nutanix.cluster.stat` with v4 stat names |
| `nutanix_host_stats_<stat>` | host stats | `nutanix.host.stat` with `nutanix.stat.name` |
| `nutanix_host_usage_stats_<stat>` | host usage stats | `nutanix.host.stat` with v4 stat names |
| `nutanix_storage_container_stats_<stat>` | storage container stats | `nutanix.storage.container.stat` |
| `nutanix_storage_container_usage_stats_<stat>` | storage container usage stats | `nutanix.storage.container.stat` |
| `nutanix_vms_stats_<stat>` | VM stats | `nutanix.vm.stat` |
| `nutanix_vms_usage_stats_<stat>` | VM usage stats | `nutanix.vm.stat` |
| `nutanix_count_vg` | volume group list length | `nutanix.volume_group.count` |
| `nutanix_count_vm`, `nutanix_count_vm_on`, `nutanix_count_vm_off` | VM list totals | `nutanix.vm.count` with optional `nutanix.vm.power_state` |
| `nutanix_count_vcpu` | VM CPU allocation | `nutanix.vm.vcpu.count` |
| `nutanix_count_vram_mib` | VM memory allocation | `nutanix.vm.memory.assigned` |
| `nutanix_count_vdisk`, `nutanix_count_vdisk_ide`, `nutanix_count_vdisk_sata`, `nutanix_count_vdisk_scsi` | VM disk config | `nutanix.vm.disk.count` with optional `nutanix.vm.disk.bus` |
| `nutanix_count_vnic` | VM NIC config | `nutanix.vm.nic.count` |
| `nutanix_cluster_info` | cluster metadata | `nutanix.cluster.info` |

The v4 APIs expose typed stat names such as `controllerNumIops`,
`hypervisorCpuUsagePpm`, and `storageUsageBytes` instead of the exporter v2
snake_case keys. The receiver keeps those source stat names in
`nutanix.stat.name` so one OTel metric can carry many Nutanix stat series
without encoding units or states into the metric name.

The exporter metric names are valid Prometheus names, but they do not follow
OpenTelemetry metric naming style: they use underscore-separated names, encode
state and units in metric names, and use generic labels such as `host`,
`cluster`, `entity`, and `storage_container`. The receiver uses dot-separated
vendor namespaced metric names, UCUM-style units where the unit is known, and
vendor-scoped attributes such as `nutanix.cluster.name`,
`nutanix.host.name`, and `nutanix.storage.container.name`.

The lab collector config labels `deployment_environment` and `telemetry_source`
are Prometheus labels, not OpenTelemetry semantic convention attributes. In an
OTLP pipeline use `deployment.environment.name` or your deployment's established
resource attribute name, and prefer vendor-scoped custom attributes for
Nutanix-specific values.

## Configuration

```yaml
receivers:
  nutanix:
    endpoint: ${env:NUTANIX_PRISM_HOST}
    port: 9440
    username: ${env:NUTANIX_PRISM_USERNAME}
    password: ${env:NUTANIX_PRISM_PASSWORD}
    collection_interval: 30s
    timeout: 20s
    tls:
      insecure_skip_verify: true
    metrics:
      clusters:
        enabled: true
      hosts:
        enabled: true
      storage_containers:
        enabled: true
      vms:
        enabled: true
      volume_groups:
        enabled: true
```

## Test Plan

1. Create a read-only Prism user that can read clusters, hosts, storage
   containers, VMs, volume groups, and v4 stats endpoints.
2. Verify API reachability from the collector host:

   ```bash
   curl -k -u "$NUTANIX_PRISM_USERNAME:$NUTANIX_PRISM_PASSWORD" \
     "https://$NUTANIX_PRISM_HOST:9440/api/clustermgmt/v4.2/config/clusters"
   ```

3. Verify one stats endpoint:

   ```bash
   START="<UTC timestamp 5 minutes ago, for example YYYY-MM-DDTHH:MM:SSZ>"
   END="<UTC timestamp now, for example YYYY-MM-DDTHH:MM:SSZ>"
   curl -k -u "$NUTANIX_PRISM_USERNAME:$NUTANIX_PRISM_PASSWORD" \
     -G "https://$NUTANIX_PRISM_HOST:9440/api/vmm/v4.2/ahv/stats/vms" \
     --data-urlencode "\$startTime=$START" \
     --data-urlencode "\$endTime=$END" \
     --data-urlencode "\$samplingInterval=30" \
     --data-urlencode "\$statType=LAST"
   ```

4. Run the collector with the debug exporter config below and confirm the log
   contains `nutanix.cluster.info`, `nutanix.cluster.stat`,
   `nutanix.host.stat`, `nutanix.storage.container.stat`,
   `nutanix.vm.stat`, `nutanix.vm.count`, and `nutanix.vm.vcpu.count`.
5. Compare counts against Prism UI inventory:

   - VM total, on, and off counts
   - vCPU and assigned memory totals
   - VM disk and NIC totals
   - volume group count
   - storage container count

6. Temporarily disable one metric category at a time and confirm the related
   Prism v4 API calls stop and the related metrics disappear.
7. Run with the production exporter after debug validation.

Full debug configuration:

```yaml
extensions:
  health_check:
    endpoint: 0.0.0.0:13133

receivers:
  nutanix:
    endpoint: ${env:NUTANIX_PRISM_HOST}
    port: 9440
    username: ${env:NUTANIX_PRISM_USERNAME}
    password: ${env:NUTANIX_PRISM_PASSWORD}
    collection_interval: 30s
    timeout: 20s
    tls:
      insecure_skip_verify: true
    metrics:
      clusters:
        enabled: true
      hosts:
        enabled: true
      storage_containers:
        enabled: true
      vms:
        enabled: true
      volume_groups:
        enabled: true

processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128
  resource/nutanix:
    attributes:
      - key: service.name
        value: nutanix-prism
        action: upsert
      - key: deployment.environment.name
        value: dev
        action: upsert
      - key: splunk.realm
        value: lab0
        action: upsert
  batch:
    timeout: 10s

exporters:
  debug:
    verbosity: detailed

service:
  extensions: [health_check]
  telemetry:
    logs:
      level: info
    metrics:
      address: 0.0.0.0:8888
  pipelines:
    metrics:
      receivers: [nutanix]
      processors: [memory_limiter, resource/nutanix, batch]
      exporters: [debug]
```

Full Splunk Observability Cloud configuration:

```yaml
extensions:
  health_check:
    endpoint: 0.0.0.0:13133

receivers:
  nutanix:
    endpoint: ${env:NUTANIX_PRISM_HOST}
    port: 9440
    username: ${env:NUTANIX_PRISM_USERNAME}
    password: ${env:NUTANIX_PRISM_PASSWORD}
    collection_interval: 30s
    timeout: 20s
    tls:
      insecure_skip_verify: true
    metrics:
      clusters:
        enabled: true
      hosts:
        enabled: true
      storage_containers:
        enabled: true
      vms:
        enabled: true
      volume_groups:
        enabled: true

processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: ${env:SPLUNK_MEMORY_LIMIT_MIB}
    spike_limit_mib: 128
  resource/nutanix:
    attributes:
      - key: service.name
        value: nutanix-prism
        action: upsert
      - key: deployment.environment.name
        value: dev
        action: upsert
      - key: splunk.realm
        value: lab0
        action: upsert
  batch:
    timeout: 10s

exporters:
  signalfx:
    access_token: ${env:SPLUNK_ACCESS_TOKEN}
    api_url: ${env:SPLUNK_API_URL}
    ingest_url: ${env:SPLUNK_INGEST_URL}
    sync_host_metadata: false

service:
  extensions: [health_check]
  telemetry:
    logs:
      level: info
    metrics:
      address: 0.0.0.0:8888
  pipelines:
    metrics:
      receivers: [nutanix]
      processors: [memory_limiter, resource/nutanix, batch]
      exporters: [signalfx]
```
