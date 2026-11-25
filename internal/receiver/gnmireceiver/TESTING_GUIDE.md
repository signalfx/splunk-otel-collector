## Manual validation against a real device

Unit tests cover the receiver against an in-process mock gNMI server. The procedure
below additionally validates it against real hardware, which is the only way to
exercise vendor behavior such as `notification.Prefix` splitting, real `proto`
encoding, and keyed path handling.

The [Cisco DevNet Sandbox](https://developer.cisco.com/site/sandbox/) "IOS XR
Programmability" pod is used here. An example configuration is provided in
[`testdata/devnet-xrv9k.yaml`](testdata/devnet-xrv9k.yaml).

### 1. Enable gNMI on the device

Reserve the sandbox, connect the AnyConnect VPN, then SSH to the IOS-XR device.

```bash
$ ssh developer@10.10.20.50                                       # DevBox
$ ssh -o HostKeyAlgorithms=+ssh-dss developer@10.10.20.35         # XRv, from DevBox
```

Enable the gRPC server (IOS-XR requires an explicit `commit`):

```
configure
 grpc
  port 57400
  no-tls
 !
commit
end
```

### 2. Reach the port from the collector host

The sandbox VPN does not permit 57400 directly; `nc -vz 10.10.20.35 57400` times out
from a laptop but succeeds from the DevBox. Forward the port:

```bash
ssh -L 57400:10.10.20.35:57400 developer@10.10.20.50   # leave running
nc -vz localhost 57400                                  # should succeed
```

**Set `NO_PROXY`.** 
gRPC honours `HTTP_PROXY`/`HTTPS_PROXY`, so on a machine with a
corporate proxy the receiver will try to dial the device through it and fail DNS
resolution:
```bash
export NO_PROXY=localhost,127.0.0.1,10.10.20.0/24
```

Alternatively, run the collector on the DevBox and use `10.10.20.35:57400` directly.

### 3. Run the receiver

```bash
make otelcol
./bin/otelcol --config internal/receiver/gnmireceiver/testdata/devnet-xrv9k.yaml
```

The example configuration uses the `debug` exporter with `verbosity: detailed`, so
every converted metric is printed.

Shut-down interfaces report `0` for every counter, which is expected. Generate traffic
(for example `ping <gateway> count 100`) and inspect the management interface if you
want non-zero values.

## Manual validation of gRPC tunnel dial-out (Cisco XRd sandbox)

With dial-out the **device dials the collector**, and the **collector then dials gNMI back
through that tunnel**. The TCP connection and the gNMI RPCs run in opposite directions, so
the SSH local-forward procedure from the dial-in section does not apply: the collector must
run somewhere the device can reach.

This procedure was validated against the [Cisco DevNet "XRd Sandbox"][xrd-sandbox]
(`ios-xrd-sandbox`) using its pre-built segment-routing topology: 8 XRd Control Plane
containers on an Ubuntu VM, reachable over the sandbox AnyConnect VPN. It is expected to work
with the IOS XR version in that topology (25.3.1 at the time of writing). The example
collector configuration is
[`testdata/xrd-sandbox-tunnel.yaml`](testdata/xrd-sandbox-tunnel.yaml).

[xrd-sandbox]: https://developer.cisco.com/site/sandbox/

Unless noted otherwise, run the commands below on the sandbox Ubuntu VM (address and
credentials are on the reservation page), except `make` and `scp`, which run on your
workstation.

### 1. Put the collector where the router can reach it

Neither of the obvious locations works:

- **Your workstation.** The sandbox VPN has no route from the internal network back to the
  VPN client pool, so the router cannot connect to it.
- **The Ubuntu VM's own address.** XRd containers sit on a Docker macvlan network
  (`segment-routing_mgmt`), and Linux macvlan isolates a host from its own macvlan children.
  Pinging the VM from the router console fails with a route lookup error, not a timeout.

Instead, run the collector in a container attached to the same macvlan network, which makes
it an L2 peer of the routers. Pick a free address in the subnet:

```bash
docker network inspect segment-routing_mgmt   # confirm the subnet and the IPs in use
docker run -d --name otelcol-test --network segment-routing_mgmt --ip 10.10.20.200 \
  -v /home/developer:/data ubuntu:22.04 sleep infinity
```

The rest of this guide assumes `10.10.20.200`; substitute your address everywhere if you
choose a different one. Verify from the router console (`docker attach xrd-1`, sandbox credentials;
detach with `Ctrl-P Ctrl-Q`):

```
RP/0/RP0/CPU0:xrd-1#ping 10.10.20.200
```

### 2. Configure the router

The tunnel forwards into the router's regular gNMI/gNOI agent, so enable that and the tunnel
destination. IOS-XR requires an explicit `commit`:

```
configure
 grpc
  port 57400
  no-tls
  tunnel
   destination 10.10.20.200 port 57500
    target xr-1
   !
  !
 !
commit
end
```

`10.10.20.200:57500` is the collector's tunnel server, not the router's own port `57400`.
Check with `show grpc tunnel sessions`. Until the collector is running it reports
`Status: Not connected`.

**Certificate.** The `grpc tunnel` submode has no `tls`/`no-tls` keyword, but the outer tunnel
transport still requires TLS. Until a certificate exists, `show grpc tunnel sessions` reports:

```
Error: not-found: certificate /misc/config/grpc/10.10.20.200:57500.pem
```

On the Ubuntu VM, generate a self-signed certificate for the collector and copy it to that
**literal** path inside the router container (not the sibling `dialout/` directory, despite
the name):

```bash
openssl req -x509 -newkey rsa:2048 -keyout tunnel-server.key -out tunnel-server.crt \
  -days 365 -nodes -subj "/CN=10.10.20.200" -addext "subjectAltName=IP:10.10.20.200"
docker cp tunnel-server.crt "xrd-1:/misc/config/grpc/10.10.20.200:57500.pem"
```

Keep `tunnel-server.crt` and `tunnel-server.key` in `/home/developer`, which is mounted at
`/data` in the collector container.

### 3. Run the receiver

On your workstation, build a Linux binary and copy it and the configuration to the VM:

```bash
GOOS=linux GOARCH=amd64 make otelcol
scp ./bin/otelcol_linux_amd64 \
  internal/receiver/gnmireceiver/testdata/xrd-sandbox-tunnel.yaml developer@<ubuntu-vm>:~/
```

On the VM, start the collector in the container and follow its log:

```bash
docker exec -d otelcol-test sh -c \
  "nohup /data/otelcol_linux_amd64 --config /data/xrd-sandbox-tunnel.yaml > /data/otelcol.log 2>&1 &"
tail -f ~/otelcol.log
```

The local agent is `no-tls`, so the inner gNMI session must disable TLS entirely:
`client_tls.insecure: true`. `insecure_skip_verify` is not enough, because it still attempts
a TLS handshake against a plaintext server.

Within one device redial interval (about 30s) expect:

```
info    gnmireceiver/tunnel.go    gNMI tunnel target added  {..., "target": "xr-1", "target_type": "GNMI_GNOI", "subscriptions": 1}
```

followed by `debug` exporter output with non-zero interface counters and the resource
attribute `server.address: xr-1`, which is the tunnel target ID rather than a host and port.
On the router, `show grpc tunnel sessions` should report `Status: Connected`.

IOS-XR always registers the `GNMI_GNOI` target type; it has no CLI knob for it.

### 4. Verify teardown

On the router, remove the tunnel destination and check that both sides clear:

```
configure
 no grpc tunnel destination 10.10.20.200 port 57500
commit
end
show grpc tunnel sessions
```

The router should show no sessions and the collector log should show:

```
info    gnmireceiver/tunnel.go    gNMI tunnel target deleted  {..., "target": "xr-1", "target_type": "GNMI_GNOI"}
```

### 5. Clean up

```bash
docker rm -f otelcol-test
```
