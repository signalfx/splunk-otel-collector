# PTP receiver

The PTP receiver collects the clock offset and grandmaster status reported by a local [linuxptp](https://github.com/richardcochran/linuxptp) `ptp4l` daemon. It runs the `pmc` management client over the daemon's Unix domain socket and reads `TIME_STATUS_NP`. The offset is for the PTP clock managed by `ptp4l`; it is not necessarily the host system clock offset. If there is no grandmaster, the receiver reports `ptp.grandmaster.present = 0` and omits `ptp.offset`.

`pmc` must be installed and the collector process must be able to access the socket. This receiver is available in the Splunk distribution on systems where linuxptp is installed.

```yaml
receivers:
  ptp:
    socket_path: /var/run/ptp/ptp4l
    pmc_path: pmc
    domain_number: 0
    collection_interval: 1m
    timeout: 10s
```

| Option | Default | Description |
| --- | --- | --- |
| `socket_path` | `/var/run/ptp/ptp4l` | Absolute path to the `ptp4l` management socket. |
| `pmc_path` | `pmc` | Path or executable name for the linuxptp `pmc` client. |
| `domain_number` | `0` | PTP domain number, from 0 through 255. |
| `collection_interval` | `1m` | Time between queries. |
| `timeout` | `10s` | Maximum time for each query. |

The receiver emits `ptp.offset` in nanoseconds and `ptp.grandmaster.present` as 0 or 1. Both metrics include `ptp.socket_path` as a resource attribute.

Either metric can be disabled under `metrics`, for example:

```yaml
receivers:
  ptp:
    metrics:
      ptp.grandmaster.present:
        enabled: false
```
