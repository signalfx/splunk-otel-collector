# collectd/openstack Integration Test

Integration tests for the [`collectd/openstack`](../../../../internal/signalfx-agent/pkg/monitors/collectd/openstack/) monitor.

These tests require a live [OpenStack](https://www.openstack.org/) deployment provided by
[DevStack](https://docs.openstack.org/devstack/latest/).  DevStack is deployed directly on the
CI runner using the [gophercloud/devstack-action](https://github.com/gophercloud/devstack-action),
so **these tests are designed to run in GitHub Actions only** and are not typically run on a
developer workstation.

## Running in CI

The tests run automatically via the `integration-test-openstack` job defined in
`.github/workflows/integration-test.yml`.  That job:

1. Deploys DevStack on the runner using `gophercloud/devstack-action`
2. Installs the agent bundle so the `collectd/openstack` plugin is available
3. Runs `make openstack-integration-test` which executes the tests with the
   `openstack_integration` build tag

## Running locally

To run the tests locally you must first deploy DevStack on the host machine by following the
[DevStack quick start guide](https://docs.openstack.org/devstack/latest/).  Use the same
credentials as in the collector config files (`admin`/`secret`, project `demo`).

Once DevStack is running:

```sh
cd tests
go test -p 1 -tags=openstack_integration -v -timeout 10m -count 1 \
    ./receivers/smartagent/collectd-openstack/...
```

## Test structure

```
collectd-openstack/
├── openstack_test.go                         # Go test functions
└── testdata/
    ├── all_metrics_config.yaml               # Collector config (all metrics incl. non-default)
    ├── default_metrics_config.yaml           # Collector config (default metrics only)
    └── expected/
        ├── all.yaml                          # Expected metrics for all_metrics_config
        └── default.yaml                      # Expected metrics for default_metrics_config
```
