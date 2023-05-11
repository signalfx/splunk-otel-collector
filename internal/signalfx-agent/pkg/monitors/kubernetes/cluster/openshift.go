package cluster

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster/meta"
)

func init() {
	monitors.Register(&meta.OpenshiftClusterMonitorMetadata,
		func() interface{} { return &Monitor{distribution: OpenShift} }, &Config{})
}
