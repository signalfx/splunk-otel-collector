# PTP receiver

The PTP receiver collects clock, grandmaster, path delay, and port status reported by a local [linuxptp](https://github.com/richardcochran/linuxptp) `ptp4l` daemon. It runs the `pmc` management client over the daemon's Unix domain socket and reads `TIME_STATUS_NP`, `CURRENT_DATA_SET`, `PORT_DATA_SET`, and `CLOCK_DESCRIPTION`. The offset is for the PTP clock managed by `ptp4l`; it is not necessarily the host system clock offset.

`pmc` must be installed and the collector process must be able to access the socket. This receiver is available in the Splunk distribution on systems where linuxptp is installed.

```yaml
receivers:
  ptp:
    socket_path: /var/run/ptp/ptp4l
    pmc_path: pmc
    domain_number: 0
    collection_interval: 10s
    timeout: 10s
```

| Option | Default | Description |
| --- | --- | --- |
| `socket_path` | `/var/run/ptp/ptp4l` | Absolute path to the `ptp4l` management socket. |
| `pmc_path` | `pmc` | Path or executable name for the linuxptp `pmc` client. |
| `domain_number` | `0` | PTP domain number, from 0 through 255. |
| `collection_interval` | `10s` | Time between queries. |
| `timeout` | `10s` | Maximum time for each query. |

| Metric | Source | Description |
| --- | --- | --- |
| `ptp.clock.state` | Local port states | Gauge of 1 with `ptp.clock.state` set to `SLAVE`, `MASTER`, `UNCALIBRATED`, `FAULTY`, `UNSYNCHRONIZED`, or `UNKNOWN`. A boundary clock with a slave port reports `SLAVE` even if it also has master ports. This is inferred from port states, not a direct servo lock indication. |
| `ptp.grandmaster.info` | `TIME_STATUS_NP` | Gauge of 1 with `ptp.grandmaster.identity` set to the selected grandmaster. An identity change appears as a new time series. |
| `ptp.grandmaster.present` | `TIME_STATUS_NP` | 1 when a grandmaster is present, otherwise 0. |
| `ptp.offset` | `TIME_STATUS_NP` | Local PTP clock offset from the grandmaster, in nanoseconds. |
| `ptp.path.delay` | `CURRENT_DATA_SET` | Mean path delay to the grandmaster, in nanoseconds. |
| `ptp.port.state` | `PORT_DATA_SET` | Gauge of 1 for each local port, with `ptp.port.identity` and `ptp.port.state` attributes. |

By default, all metrics include `ptp.socket_path` as a resource attribute. When `CLOCK_DESCRIPTION` is available, they also include `ptp.clock.type`: `ordinary`, `boundary`, `p2p_transparent`, `e2e_transparent`, or `unknown`. If a transparent clock does not expose port data through `pmc`, its clock state is `UNKNOWN` and no `ptp.port.state` data point is emitted. If no grandmaster is present, the receiver omits `ptp.grandmaster.info`, `ptp.offset`, and `ptp.path.delay`.

SyncE ESMC/SSM quality levels are outside this receiver's current data source; `ptp4l` does not expose them through these management datasets.

Any metric can be disabled under `metrics`, for example:

```yaml
receivers:
  ptp:
    metrics:
      ptp.grandmaster.present:
        enabled: false
```
