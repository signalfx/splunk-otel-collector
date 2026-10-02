# PTP Receiver

The PTP receiver collects clock synchronization status from a local LinuxPTP
`ptp4l` daemon using its management client, `pmc`. It reports the PTP clock and
port status; the PTP clock offset is not necessarily the host system clock
offset.

| Status | |
| --- | --- |
| Stability | [beta]: metrics  |
| Distributions | [Splunk](https://github.com/signalfx/splunk-otel-collector) |

[beta]:
  https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#beta

## Configuration

All receiver settings are optional. The collector process must be able to
execute `pmc` and access the configured `ptp4l` Unix domain socket.

- `socket_path` (default: `/var/run/ptp/ptp4l`): Absolute path to the `ptp4l` management socket.
- `pmc`: Settings for the LinuxPTP management client.
  - `path` (default: `pmc`): Path or executable name for the client.
  - `client_socket_directory` (default: the operating system's temporary
    directory): Writable directory where the receiver creates a private,
    unique client socket directory for each query. This is separate from
    `socket_path`, which identifies the `ptp4l` server socket.
- `domain_number` (default: `0`): PTP domain number, from `0` through `255`.
- `collection_interval` (default: `10s`): Time between metric collections.
- `timeout` (default: `5s`): Maximum time allowed for each collection.

### Example Configuration
```yaml
receivers:
  ptp:
    socket_path: /var/run/ptp/ptp4l
    pmc:
      path: pmc
      client_socket_directory: /tmp
    domain_number: 0
    collection_interval: 10s
    timeout: 5s
```

When LinuxPTP reports no remote grandmaster, the receiver reports grandmaster
presence as `0` and omits the grandmaster identity, offset, and path-delay
measurements. It does not substitute zero for an unavailable offset or delay.
The detailed metric definitions are maintained in
[documentation.md](./documentation.md).

## Metrics

Details about the metrics produced by this receiver can be found in
[metadata.yaml](./metadata.yaml).

### `ptp.clock.type` values

The receiver maps LinuxPTP's `CLOCK_DESCRIPTION` type code to these labels:

- `OC`: Ordinary clock.
- `BC`: Boundary clock.
- `P2P_TC`: Peer-to-peer transparent clock.
- `E2E_TC`: End-to-end transparent clock.
- `UNKNOWN`: Unrecognized clock type code.
