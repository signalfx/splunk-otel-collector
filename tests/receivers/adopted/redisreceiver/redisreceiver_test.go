package tests

import (
	"path"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestRedisReceiverProvidesAllMetrics(t *testing.T) {

	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithExposedPorts("6379:6379").WithName("redis-server").WillWaitForPorts("6379").WillWaitForLogs("Ready to accept connections")

	containers := []testutils.Container{server}

	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", containers)
}


func TestRedisReceiverProvidesAllMetricsWithServer(t *testing.T) {

	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithExposedPorts("6379:6379").WithNetworks("redis_network").WithName("redis-server").WillWaitForLogs("Ready to accept connections")

	client := testutils.NewContainer().WithContext(path.Join(".", "testdata", "client")).WithName("redis-client").WithNetworks("redis_network").WillWaitForLogs("redis client started")

	containers := []testutils.Container{server, client}

	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", containers)
}
