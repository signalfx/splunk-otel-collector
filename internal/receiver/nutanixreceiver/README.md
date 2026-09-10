# Nutanix Receiver

The Nutanix receiver collects Prism metrics through either the Prism Element
v2.0 APIs or the Prism Central v4 APIs and emits OTLP metrics directly.

Set `api_version: v2.0` when `endpoint` is a Prism Element cluster. Set
`api_version: v4` when `endpoint` is Prism Central. The Prism Element
implementation uses its v2.0 REST API. The Prism Central implementation uses
the generated Nutanix v4 Go clients for API paths, pagination, authentication,
retries, and version negotiation.

The receiver redirects the generated clients' request logging to
`io.Discard`, so routine HTTP traffic does not appear in the collector console.
Receiver warnings and errors still use the collector logger.

The v4.2 reference page is for the Prism Central API surface. Sending its
`/api/clustermgmt/v4.2/...` paths to a Prism Element endpoint returns 404. PE
inventory endpoints use `/PrismGateway/services/rest/v2.0/...` instead.

## Prism v4 API Coverage

The paths below use v4.2 to match the public API reference. The generated
client for each namespace owns the actual path and minor version used by the
collector and negotiates a supported version with Prism Central when that
endpoint is available; these paths are not hardcoded in the receiver.

Inventory and count data is collected from:

- `/api/clustermgmt/v4.2/config/clusters`
- `/api/clustermgmt/v4.2/config/hosts`
- `/api/clustermgmt/v4.2/config/storage-containers`
- `/api/clustermgmt/v4.2/config/disks`
- `/api/networking/v4.2/config/subnets`
- `/api/vmm/v4.2/ahv/config/vms`
- `/api/volumes/v4.2/config/volume-groups`

Runtime statistics are collected from:

- `/api/clustermgmt/v4.2/stats/clusters/{extId}`
- `/api/clustermgmt/v4.2/stats/clusters/{clusterExtId}/hosts/{extId}`
- `/api/clustermgmt/v4.2/stats/storage-containers/{extId}`
- `/api/clustermgmt/v4.2/stats/disks/{extId}`
- `/api/vmm/v4.2/ahv/stats/vms`
- `/api/volumes/v4.2/stats/volume-groups/{extId}`

Each scrape queries the most recent collection interval with the v4 `LAST`
down-sampling operator.

Optional Prism Central categories add networking entity counts and statistics,
Prism categories/tasks/alerts, protection-policy schedules, protected-VM RPO
counts, recovery-point counts,
microsegmentation inventory, Files inventory and statistics, and Objects
inventory and statistics. These categories are disabled by default because
they require additional Nutanix services and read permissions. Failure of one
optional API is logged and does not discard metrics collected from healthy
services.

## Prism Element v2.0 API Coverage

When `api_version: v2.0` is selected, inventory is collected from the cluster
local endpoints:

- `/PrismGateway/services/rest/v2.0/cluster/` (with a compatibility fallback to `/clusters/`)
- `/PrismGateway/services/rest/v2.0/hosts/`
- `/PrismGateway/services/rest/v2.0/storage_containers/`
- `/PrismGateway/services/rest/v2.0/vms/?include_vm_disk_config=true&include_vm_nic_config=true`
- `/PrismGateway/services/rest/v2.0/volume_groups/`

The receiver flattens numeric values in the `stats` and `usage_stats` objects
returned with those entities. This is the PE-compatible path used for a direct
Prism Element 7.3.1.1 endpoint.

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
| `nutanix_count_vnic` | VM NIC config | `nutanix.vm.nics.count` |
| `nutanix_cluster_info` | cluster metadata | `nutanix.cluster.info` |

The v4 APIs expose typed stat names such as `controllerNumIops`,
`hypervisorCpuUsagePpm`, and `storageUsageBytes` instead of the exporter v2
snake_case keys. The receiver keeps those source stat names in
`nutanix.stat.name` so one OTel metric can carry many Nutanix stat series
without encoding units or states into the metric name. Raw API statistic
families intentionally leave the OpenTelemetry unit unspecified because one
family can contain source fields with different units; the source field name
continues to identify bytes, IOPS, microseconds, ppm, and other measurements.

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

## Build and run locally

From the repository root, build the collector and start it with the local
SignalFx configuration:

```bash
make otelcol

export NUTANIX_PRISM_HOST="<prism-host-or-ip>"
export NUTANIX_PRISM_USERNAME="<read-only-user>"
export NUTANIX_PRISM_PASSWORD="<password>"
export SPLUNK_ACCESS_TOKEN="<access-token>"
export SPLUNK_API_URL="https://api.<realm>.observability.splunkcloud.com"
export SPLUNK_INGEST_URL="https://ingest.<realm>.observability.splunkcloud.com"

./bin/otelcol --config ./config.yaml
```

The collector reports `Everything is ready` after startup. Check the local
health endpoint with `curl http://localhost:13133/`, and stop the collector
with Ctrl-C. For Prism Element, keep `api_version: v2.0`; for Prism Central,
set it to `v4` and enable optional metric categories only when the associated
services are installed and the account has read permission.

## Configuration

```yaml
receivers:
  nutanix:
    endpoint: ${env:NUTANIX_PRISM_HOST}
    # v2.0 for Prism Element; use v4 for Prism Central.
    api_version: v4
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
      disks:
        enabled: false
      hosts:
        enabled: true
      networking:
        enabled: false
      storage_containers:
        enabled: true
      vms:
        enabled: true
      volume_groups:
        enabled: true
      prism_central:
        enabled: false
      data_protection:
        enabled: false
      microsegmentation:
        enabled: false
      files:
        enabled: false
      objects:
        enabled: false
```

## Test Plan

1. Create a read-only Prism user that can read clusters, hosts, storage
   containers, VMs, and volume groups.
2. Verify API reachability from the collector host:

   ```bash
   curl -k -u "$NUTANIX_PRISM_USERNAME:$NUTANIX_PRISM_PASSWORD" \
     "https://$NUTANIX_PRISM_HOST:9440/PrismGateway/services/rest/v2.0/cluster/"
   curl -k -u "$NUTANIX_PRISM_USERNAME:$NUTANIX_PRISM_PASSWORD" \
     "https://$NUTANIX_PRISM_HOST:9440/PrismGateway/services/rest/v2.0/storage_containers/"
   ```

3. If the endpoint is Prism Central and `api_version: v4` is configured,
   verify the v4 API instead:

   ```bash
   curl -k -u "$NUTANIX_PRISM_USERNAME:$NUTANIX_PRISM_PASSWORD" \
     "https://$NUTANIX_PRISM_HOST:9440/api/clustermgmt/v4.2/config/clusters"
   ```

4. Run the collector with the SignalFx exporter configuration below and
   confirm Splunk Observability Cloud receives `nutanix.cluster.info`,
   `nutanix.cluster.stat`, `nutanix.host.stat`,
   `nutanix.storage.container.stat`, `nutanix.vm.stat`,
   `nutanix.vm.count`, and `nutanix.vm.vcpu.count`.
5. Compare counts against Prism UI inventory:

   - VM total, on, and off counts
   - vCPU and assigned memory totals
   - VM disk and NIC totals
   - volume group count
   - storage container count

6. Temporarily disable one metric category at a time and confirm the related
   Prism API calls stop and the related metrics disappear.
7. Enable optional Prism Central categories individually and confirm failures
   from unavailable or unlicensed services are warnings while core metrics
   continue to arrive.

Full local Splunk Observability Cloud configuration:

```yaml
extensions:
  health_check:
    endpoint: 0.0.0.0:13133

receivers:
  nutanix:
    endpoint: ${env:NUTANIX_PRISM_HOST}
    api_version: v2.0
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
