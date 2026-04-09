# collectd/openstack Integration Test

Integration tests for the [`collectd/openstack`](../../../../internal/signalfx-agent/pkg/monitors/collectd/openstack/) monitor.

These tests require a live [OpenStack](https://www.openstack.org/) deployment provided by
[DevStack](https://docs.openstack.org/devstack/latest/).  Because DevStack takes a very long
time to install (~20–30 minutes), you must **build the DevStack Docker image once** before
running the tests.

## Prerequisites

1. Docker with privileged container support (required by DevStack / systemd)
2. At least 20 GB of free disk space
3. A reliable internet connection (the image build downloads several GB of packages)
4. A valid `SPLUNK_OTEL_COLLECTOR_IMAGE` pointing to a built collector image

## Building the DevStack image

```sh
cd tests/receivers/smartagent/collectd-openstack/testdata/devstack
./make-devstack-image.sh
```

This produces a local `devstack:latest` Docker image with OpenStack pre-installed.  The image
only needs to be rebuilt when the DevStack configuration changes.

## Running the tests

```sh
export SPLUNK_OTEL_COLLECTOR_IMAGE=otelcol:latest  # or your collector image tag

cd tests
go test -p 1 -tags=smartagent_integration -v -timeout 60m -count 1 \
    ./receivers/smartagent/collectd-openstack/...
```

The `-timeout 60m` is necessary because DevStack needs several minutes to start all OpenStack
services inside the container at test time.

## Updating golden (expected) metrics files

If the set of metrics emitted by the plugin changes, regenerate the golden files by running
the tests with the `UPDATE_EXPECTED=true` environment variable:

```sh
UPDATE_EXPECTED=true go test -p 1 -tags=smartagent_integration -v -timeout 60m -count 1 \
    ./receivers/smartagent/collectd-openstack/...
```

The updated files will be written to `testdata/expected/`.

## Test structure

```
collectd-openstack/
├── openstack_test.go                         # Go test functions
└── testdata/
    ├── all_metrics_config.yaml               # Collector config (all metrics incl. non-default)
    ├── default_metrics_config.yaml           # Collector config (default metrics only)
    ├── expected/
    │   ├── all.yaml                          # Expected metrics for all_metrics_config
    │   └── default.yaml                      # Expected metrics for default_metrics_config
    └── devstack/
        ├── Dockerfile                        # DevStack image definition
        ├── local.conf                        # DevStack configuration
        ├── start-devstack.sh                 # Service start script (used inside image)
        ├── stop-devstack.sh                  # Service stop script (used inside image)
        ├── devstack.service                  # systemd unit file for DevStack
        └── make-devstack-image.sh            # Helper script to build the devstack:latest image
```
