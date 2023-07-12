package stats

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/utils"
)

const (
	clusterStatusGreen  = "green"
	clusterStatusYellow = "yellow"
	clusterStatusRed    = "red"
)

// GetClusterStatsDatapoints fetches datapoints for ES cluster level stats
func GetClusterStatsDatapoints(clusterStatsOutput *ClusterStatsOutput, defaultDims map[string]string, enhanced bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			utils.PrepareGaugeHelper("elasticsearch.cluster.initializing-shards", defaultDims, clusterStatsOutput.InitializingShards),
			utils.PrepareGaugeHelper("elasticsearch.cluster.delayed-unassigned-shards", defaultDims, clusterStatsOutput.DelayedUnassignedShards),
			utils.PrepareGaugeHelper("elasticsearch.cluster.pending-tasks", defaultDims, clusterStatsOutput.NumberOfPendingTasks),
			utils.PrepareGaugeHelper("elasticsearch.cluster.in-flight-fetches", defaultDims, clusterStatsOutput.NumberOfInFlightFetch),
			utils.PrepareGaugeHelper("elasticsearch.cluster.task-max-wait-time", defaultDims, clusterStatsOutput.TaskMaxWaitingInQueueMillis),
			utils.PrepareGaugeFHelper("elasticsearch.cluster.active-shards-percent", defaultDims, clusterStatsOutput.ActiveShardsPercentAsNumber),
			utils.PrepareGaugeHelper("elasticsearch.cluster.status", defaultDims, getMetricValueFromClusterStatus(clusterStatsOutput.Status)),
		}...)
	}
	out = append(out, []*datapoint.Datapoint{
		utils.PrepareGaugeHelper("elasticsearch.cluster.active-primary-shards", defaultDims, clusterStatsOutput.ActivePrimaryShards),
		utils.PrepareGaugeHelper("elasticsearch.cluster.active-shards", defaultDims, clusterStatsOutput.ActiveShards),
		utils.PrepareGaugeHelper("elasticsearch.cluster.number-of-data_nodes", defaultDims, clusterStatsOutput.NumberOfDataNodes),
		utils.PrepareGaugeHelper("elasticsearch.cluster.number-of-nodes", defaultDims, clusterStatsOutput.NumberOfNodes),
		utils.PrepareGaugeHelper("elasticsearch.cluster.relocating-shards", defaultDims, clusterStatsOutput.RelocatingShards),
		utils.PrepareGaugeHelper("elasticsearch.cluster.unassigned-shards", defaultDims, clusterStatsOutput.UnassignedShards),
	}...)
	return out
}

// Map cluster status to a numeric value
func getMetricValueFromClusterStatus(s *string) *int64 {
	// For whatever reason if the monitor did not get cluster status return nil
	if s == nil {
		return nil
	}
	out := new(int64)
	status := *s

	switch status {
	case clusterStatusGreen:
		*out = 0
	case clusterStatusYellow:
		*out = 1
	case clusterStatusRed:
		*out = 2
	}

	return out
}
