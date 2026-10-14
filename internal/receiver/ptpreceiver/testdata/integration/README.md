# PTP receiver integration test

Runs the real `ptpreceiver` against a real `ptp4l`/`pmc` (LinuxPTP). This
closes the gap between the mocked-`pmc` unit tests in `scraper_test.go` and
manual testing against a real `ptp4l` deployment.

Containers are built from the same image (`Dockerfile` in this directory):

- a "master" sidecar, running only `ptp4l` with a low `priority1` so it
  wins the Best Master Clock Algorithm (BMCA) and becomes grandmaster. Its
  MAC address is pinned (see `start_master` in `run.sh`), since `ptp4l`
  derives `clockIdentity` from it, and Docker otherwise randomizes an
  unpinned MAC on every `docker start`.
- one or more "slave"/test containers, whose entrypoint
  (`entrypoint.sh`) starts a second `ptp4l` instance with a high
  `priority1` (loses BMCA), waits for its management socket to appear,
  then runs the `go test` command it was given, which points the real
  receiver at that socket.

All `ptp4l` instances run with software timestamping (`-S`), so no
PTP-capable NIC is required. Slave and master run as separate containers on
a shared Docker user-defined bridge network, so each gets a distinct
MAC/clockIdentity -- two `ptp4l` instances sharing one interface (e.g. both
on loopback in a single container) derive identical clockIdentity values
and silently ignore each other's announce messages, so this can't be
collapsed into one container. Slave containers also need
`--cap-add=SYS_TIME` so their clock servo can actually adjust the
(containerized) clock and leave LinuxPTP's UNCALIBRATED state.

Real BMCA negotiation and clock-servo convergence take tens of seconds, so
most tests poll for up to 120s.

## Scenarios

`run.sh` with no arguments runs all of these (this is what CI does):

- `TestPTPReceiverIntegration` (`integration_test.go`): the standard
  successful path -- slave syncs to the master and reaches SLAVE. Asserts
  the full metric contract with real values/attributes (not just metric
  presence), and that the port passed through a pre-sync state
  (LISTENING/UNCALIBRATED) before reaching SLAVE, not just a single
  already-synchronized snapshot. (LISTENING itself isn't asserted
  separately: the master is already broadcasting before the slave
  container even starts, so LISTENING -- at most one ~2s announce interval
  -- is routinely over before this container's uncached `go test` finishes
  downloading modules and compiling.)
- `TestPTPReceiverIntegration_NoRemoteGrandmaster`,
  `TestPTPReceiverIntegration_SocketNotPresent`, and
  `TestPTPReceiverIntegration_PMCTimeout`
  (`integration_negative_test.go`): negative paths with no master sidecar
  or network dependency, so they run first and fail fast. The first
  asserts a real `ptp4l` instance with no competing master self-elects
  MASTER and reports `gmPresent=false`; the second points the receiver at
  a socket path that never exists and asserts it fails every scrape
  cleanly (logging an error, never panicking or blocking) without ever
  producing a non-empty metrics data point. The third points `PMC.Path` at
  a script that hangs forever instead of the real `pmc` binary, and asserts
  the receiver's `exec.CommandContext`-based `Timeout` actually kills the
  hung subprocess and the collector keeps scraping on later intervals
  rather than wedging on the first one. This isn't a real `pmc` hang: an
  empirical probe (a Unix domain socket that accepts `pmc`'s request and
  never responds) showed real `pmc` doesn't block on an unresponsive
  server at all -- it sends the request over `SOCK_DGRAM` and returns
  almost immediately regardless, since there's no connection to wait on. A
  genuinely hung subprocess only happens if the configured binary itself
  hangs, so this test substitutes one via the same `PMC.Path` knob a real
  deployment could point at a wrapper script.
- `TestPTPReceiverIntegration_BoundaryClock`
  (`integration_boundary_test.go`): a single `ptp4l` instance managing two
  ports -- one syncing to the usual master sidecar, one isolated with no
  peers (self-elects MASTER) -- asserting it reports as a boundary clock
  (`ptp.clock.type=BC`) with clock state SLAVE (per `metadata.yaml`'s rule
  for boundary clocks with both a slave and a master port) and two
  distinct port identities. `run.sh` has to `docker create` (not `run`)
  this container with both networks already attached before `docker
  start`: `ptp4l` checks for its second interface once at startup and
  never retries, so a network attached after the container's already
  running arrives too late.
- `TestPTPReceiverIntegration_GrandmasterLossRecovery`
  (`integration_lossrecovery_test.go`): a real grandmaster disappearing and
  coming back mid-collection. Since the test (in a container) and `run.sh`
  (on the host) don't otherwise share a process or mounted volume, they
  hand off control of the master sidecar via marker files under
  `/tmp/ptp-integration-signals`: `run.sh` polls for files with `docker
  exec ... test -f`, the test polls/writes the same path with `os.Stat`/
  `os.WriteFile` from inside its own container filesystem. Asserts the
  receiver reports losing the grandmaster while the master is stopped, and
  re-selects the *same* grandmaster identity once it's restarted.
- `TestPTPReceiverIntegration_FaultyPort` (`integration_faulty_test.go`):
  exercises `clockState()`'s FAULTY branch, which until now had zero
  coverage anywhere (mocked or real). Waits for the usual real sync, then
  signals `run.sh` to `docker network disconnect` this container's network
  entirely, rather than stopping the master as the loss/recovery scenario
  does -- using the same marker-file handshake. `ptp4l` survives losing
  its interface post-startup: the port transitions to FAULTY, but the
  daemon keeps running and keeps answering `pmc` over its
  (network-independent) Unix management socket, which is what makes this a
  real, deterministic way to reach FAULTY. An all-invalid-interface config
  was also tried and ruled out: `ptp4l` can't derive a `clockIdentity`
  without a valid interface at startup and refuses to start at all, rather
  than coming up FAULTY. Asserts both `ptp.port.state` and the overall
  `ptp.clock.state` report FAULTY.

## Run locally

From the repository root:

```sh
internal/receiver/ptpreceiver/testdata/integration/run.sh
```

This builds the image and runs every scenario above, tearing down all
containers/networks on exit. With arguments, it instead runs exactly that
command once against the standard master+slave topology, for manual
debugging -- e.g. to run a single test with `-v` output:

```sh
internal/receiver/ptpreceiver/testdata/integration/run.sh go test -tags=ptp_integration -run TestPTPReceiverIntegration -v ./internal/receiver/ptpreceiver/...
```
