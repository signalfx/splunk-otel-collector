# PTP receiver manual verification on Ubuntu

This guide sets up LinuxPTP `ptp4l` on an Ubuntu Linux host, verifies its
management socket with `pmc`, and runs the Splunk OpenTelemetry Collector PTP
receiver. The receiver queries a local `ptp4l` daemon; it does not start or
configure `ptp4l` itself.

## 1. Install LinuxPTP and identify the interface

Install LinuxPTP and the optional `ethtool` diagnostic:

```bash
sudo apt update
sudo apt install linuxptp ethtool
```

List the network interfaces and choose the one connected to the PTP network:

```bash
ip -br link
ip route
```

Do not assume the default-route interface is PTP-capable or connected to a PTP
grandmaster. Check timestamping support for the chosen interface:

```bash
sudo ethtool -T ens5
```

Replace `ens5` below with the interface name from your host.

## 2. Configure and start `ptp4l`

Ubuntu's LinuxPTP package provides a `ptp4l@.service` template. Inspect the
unit to confirm which config file it uses and how the instance name is passed:

```bash
sudo systemctl cat 'ptp4l@.service'
```

The packaged unit commonly runs `ptp4l -f /etc/linuxptp/ptp4l.conf -i %I`,
where `%I` is replaced by the interface in the instance name. Inspect the
config file and ensure its `domainNumber` matches the receiver's
`domain_number`:

```bash
sudo grep -nE '^\[|^[[:space:]]*(time_stamping|domainNumber|uds_address|uds_ro_address|uds_file_mode|uds_ro_file_mode)' \
  /etc/linuxptp/ptp4l.conf
```

LinuxPTP defaults to hardware timestamping. If `ptp4l` fails with an
unsupported timestamping mode, and the interface supports software timestamping,
edit `/etc/linuxptp/ptp4l.conf` and set this under its existing `[global]`
section:

```ini
time_stamping software
```

Software timestamping is less precise than hardware timestamping. Use it when
hardware timestamping is unavailable and the goal is to validate collection.

LinuxPTP's usual Unix socket defaults are `/var/run/ptp4l` for the control
socket and `/var/run/ptp4lro` for its read-only socket. Prefer the read-only
socket for the receiver, which only issues `GET` requests. If the config file
overrides `uds_ro_address`, use that configured path instead. Check the actual
socket and permissions after starting the service.

Start the service for the chosen interface and enable it at boot:

```bash
sudo systemctl enable --now 'ptp4l@ens5.service'
```

Verify the unit and recent logs:

```bash
sudo systemctl status 'ptp4l@ens5.service' --no-pager
sudo journalctl -u 'ptp4l@ens5.service' -n 50 --no-pager
```

Look for `Active: active (running)` and port state transitions. If the port
transitions from `LISTENING` to `MASTER` after the announce timeout, no remote
grandmaster has been received; the local clock elected itself master. For a
slave test with offset and path-delay metrics, connect to a PTP network with a
reachable grandmaster and confirm the port reaches `SLAVE`.

## 3. Verify the management socket as the collector user

The `pmc` client creates its own local Unix socket. When running as an
unprivileged user, choose a writable path with `-i`; otherwise `pmc` may fail
to bind under `/var/run`. The `-s` option is the `ptp4l` server socket, while
`-i` is the local PMC client socket path.

Run the check as the same user that will run the collector. Set `-s` to the
configured read-only socket path and `-d` to the PTP domain:

```bash
pmc -u -b 0 -i "/tmp/pmc-$(id -u)-$$" \
  -s /var/run/ptp4lro -d 0 \
  'GET TIME_STATUS_NP' \
  'GET CURRENT_DATA_SET' \
  'GET PORT_DATA_SET' \
  'GET CLOCK_DESCRIPTION'
```

Successful output includes headers such as `RESPONSE MANAGEMENT TIME_STATUS_NP`
and `RESPONSE MANAGEMENT PORT_DATA_SET`. If no responses appear, check that
`ptp4l` is running, the socket path and domain match its config, and the
collector user can access the read-only socket. Do not make `/run` world
writable.

## 4. Build and configure the collector

Use a build that includes the PTP receiver. From the collector repository on
the PTP receiver branch, build the Ubuntu/Linux amd64 binary:

```bash
make GOOS=linux GOARCH=amd64 otelcol
```

The binary is `bin/otelcol_linux_amd64`. Copy it to the Ubuntu host if it was
built elsewhere. Create `/home/splunker/config.yaml` (or another path) with the
following example. Set `socket_path` to the `uds_ro_address` checked above:

```yaml
receivers:
  ptp:
    socket_path: /var/run/ptp4lro
    pmc_path: /usr/sbin/pmc
    pmc_client_socket_directory: /tmp
    domain_number: 0
    collection_interval: 10s
    timeout: 10s

processors:
  batch:

exporters:
  signalfx:
    access_token: ${SPLUNK_ACCESS_TOKEN}
    realm: ${SPLUNK_REALM}
  debug:
    verbosity: detailed

service:
  pipelines:
    metrics/ptp:
      receivers: [ptp]
      processors: [batch]
      exporters: [signalfx, debug]
```

`pmc_path` selects the LinuxPTP management client executable. The receiver
creates a unique private client socket directory under
`pmc_client_socket_directory` for every query and removes it afterward. This
setting is separate from `socket_path`, which points to the `ptp4l` server
socket. If omitted, the receiver uses the operating system's temporary
directory.

Export the Splunk Observability Cloud access token and realm in the environment
of the collector process:

```bash
export SPLUNK_ACCESS_TOKEN='<your-access-token>'
export SPLUNK_REALM='<your-realm>'
```

## 5. Run and verify the collector

Run it as the same unprivileged user used for the successful `pmc` check:

```bash
./bin/otelcol_linux_amd64 --config=/home/splunker/config.yaml
```

Check the detailed debug exporter output and search for these metrics in
Splunk Observability Cloud's Metrics/Data Explorer:

- `ptp.clock.state`
- `ptp.port.state`
- `ptp.grandmaster.present`
- `ptp.grandmaster.info`
- `ptp.offset`
- `ptp.path.delay`

With no remote grandmaster, `ptp.grandmaster.present` is `0`; the receiver
omits grandmaster identity, offset, and path delay rather than reporting
misleading zero measurements. A local `MASTER` state is therefore valid for a
socket/receiver smoke test, even though it does not verify synchronization to a
remote grandmaster.

If the collector reports `pmc response missing supported management datasets`,
run the same `pmc` query from section 3 using the exact `socket_path` and
`domain_number` in the collector config. If that query succeeds but collector
scrapes fail, verify that the running binary was built from the PTP receiver
branch and that `pmc_client_socket_directory` exists and is writable by the
collector user.
