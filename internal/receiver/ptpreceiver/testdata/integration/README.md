# PTP receiver integration test

Runs the real `ptpreceiver` against a real `ptp4l`/`pmc` (LinuxPTP). This
closes the gap between the mocked-`pmc` unit tests in `scraper_test.go` and
manual testing against a real `ptp4l` deployment.

Two containers are used, both built from the same image (`Dockerfile` in
this directory):

- a "master" sidecar, running only `ptp4l` with a low `priority1` so it
  wins the Best Master Clock Algorithm (BMCA) and becomes grandmaster.
- a "slave"/test container, whose entrypoint starts a second `ptp4l`
  instance with a high `priority1` (loses BMCA), waits for its management
  socket to appear, then runs `go test -tags=ptp_integration
  ./internal/receiver/ptpreceiver/...`, which points the real receiver at
  that socket.

Both run with software timestamping (`-S`), so no PTP-capable NIC is
required, and on a shared Docker user-defined bridge network, so each gets
a distinct MAC/clockIdentity -- two `ptp4l` instances sharing one
interface (e.g. both on loopback in a single container) derive identical
clockIdentity values and silently ignore each other's announce messages,
so this can't be collapsed into one container. The slave container also
needs `--cap-add=SYS_TIME` so its clock servo can actually adjust the
(containerized) clock and leave LinuxPTP's UNCALIBRATED state.

Real BMCA negotiation and clock-servo convergence take tens of seconds, so
the test polls for up to 120s.

## Run locally

From the repository root:

```sh
internal/receiver/ptpreceiver/testdata/integration/run.sh
```

This builds the image, creates the Docker network and master sidecar, runs
the slave/test container, and tears both down on exit. Extra arguments are
passed through to the slave container's entrypoint, e.g. to run a specific
test or get verbose `go test` output:

```sh
internal/receiver/ptpreceiver/testdata/integration/run.sh go test -tags=ptp_integration -run TestPTPReceiverIntegration -v ./internal/receiver/ptpreceiver/...
```
