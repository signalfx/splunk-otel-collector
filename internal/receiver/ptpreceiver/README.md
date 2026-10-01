# PTP Receiver

The PTP receiver collects clock synchronization status from a local LinuxPTP `ptp4l` daemon using its management client, `pmc`. It reports the PTP clock and port status; the PTP clock offset is not necessarily the host system clock offset.

| Status | |
| --- | --- |
| Stability | [Development](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development): metrics |
| Supported pipeline types | Metrics |
| Distributions | [Splunk](https://github.com/signalfx/splunk-otel-collector) |

## Configuration

All receiver settings are optional. The collector process must be able to execute `pmc` and access the configured `ptp4l` Unix domain socket.

```yaml
receivers:
  ptp:
    socket_path: /var/run/ptp/ptp4l
    pmc_path: pmc
    domain_number: 0
    collection_interval: 10s
    timeout: 10s
```

- `socket_path` (default: `/var/run/ptp/ptp4l`): Absolute path to the `ptp4l` management socket.
- `pmc_path` (default: `pmc`): Path or executable name for the LinuxPTP management client.
- `domain_number` (default: `0`): PTP domain number, from `0` through `255`.
- `collection_interval` (default: `10s`): Time between metric collections.
- `timeout` (default: `10s`): Maximum time allowed for each collection.

When LinuxPTP reports no remote grandmaster, the receiver reports grandmaster presence as `0` and omits the grandmaster identity, offset, and path-delay measurements. It does not substitute zero for an unavailable offset or delay. The detailed metric definitions are maintained in [documentation.md](./documentation.md).
