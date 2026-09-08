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