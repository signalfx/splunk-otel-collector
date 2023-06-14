// Package cluster contains a Kubernetes cluster monitor.
//
// This plugin collects high level metrics about a K8s cluster and sends them
// to SignalFx.  The basic technique is to pull data from the K8s API and keep
// up-to-date copies of datapoints for each metric that we collect and then
// ship them off at the end of each reporting interval.  The K8s streaming
// watch API is used to efficiently maintain the state between read intervals
// (see `clusterstate.go`).
//
// This plugin requires read-only access to the K8s API.
package cluster

import (
	"fmt"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster/meta"

	"github.com/sirupsen/logrus"

	"k8s.io/client-go/rest"

	"github.com/signalfx/signalfx-agent/pkg/core/common/dpmeta"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster/metrics"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/leadership"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// KubernetesDistribution indicates the particular flavor of Kubernetes.
type KubernetesDistribution int

const (
	// Generic is normal Kubernetes with nothing extra added.
	Generic KubernetesDistribution = iota
	// OpenShift is RedHat's Kubernetes distribution.
	OpenShift
)

var distributionToMonitorType = map[KubernetesDistribution]string{
	Generic:   meta.KubernetesClusterMonitorMetadata.MonitorType,
	OpenShift: meta.OpenshiftClusterMonitorMetadata.MonitorType,
}

// Config for the K8s monitor
type Config struct {
	config.MonitorConfig `yaml:",inline"`
	// If `true`, leader election is skipped and metrics are always reported.
	AlwaysClusterReporter bool `yaml:"alwaysClusterReporter"`
	// If specified, only resources within the given namespace will be
	// monitored.  If omitted (blank) all supported resources across all
	// namespaces will be monitored.
	Namespace string `yaml:"namespace"`
	// Config for the K8s API client
	KubernetesAPI *kubernetes.APIConfig `yaml:"kubernetesAPI" default:"{}"`
	// A list of node status condition types to report as metrics.  The metrics
	// will be reported as datapoints of the form `kubernetes.node_<type_snake_cased>`
	// with a value of `0` corresponding to "False", `1` to "True", and `-1`
	// to "Unknown".
	NodeConditionTypesToReport []string `yaml:"nodeConditionTypesToReport" default:"[\"Ready\"]"`
	// If set to true, the `kubernetes_node` dimension, in addition to the `kubernetes_node_uid` dimension, will get
	// properties about each respective node synced to it. Do not enable this, if node names in the cluster are
	// reused (can lead to colliding or stale properties).
	UpdatesForNodeDimension bool `yaml:"updatesForNodeDimension" default:"false"`
}

// Validate the k8s-specific config
func (c *Config) Validate() error {
	return c.KubernetesAPI.Validate()
}

func (c *Config) GetExtraMetrics() []string {
	var out []string
	for _, cond := range c.NodeConditionTypesToReport {
		out = append(out, "kubernetes.node_"+cond)
	}
	return out
}

// Monitor for K8s Cluster Metrics.  Also handles syncing certain properties
// about pods.
type Monitor struct {
	config       *Config
	distribution KubernetesDistribution
	Output       types.Output
	// Since most datapoints will stay the same or only slightly different
	// across reporting intervals, reuse them
	datapointCache *metrics.DatapointCache
	dimHandler     *metrics.DimensionHandler
	restConfig     *rest.Config
	stop           chan struct{}
	logger         logrus.FieldLogger
}

func init() {
	monitors.Register(&meta.KubernetesClusterMonitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Configure is called by the plugin framework when configuration changes
func (m *Monitor) Configure(config *Config) error {
	var err error
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": distributionToMonitorType[m.distribution], "monitorID": config.MonitorID})

	m.config = config

	if m.restConfig, err = kubernetes.CreateRestConfig(config.KubernetesAPI); err != nil {
		return fmt.Errorf("could not create Kubernetes REST config: %s", err)
	}

	m.datapointCache = metrics.NewDatapointCache(m.config.NodeConditionTypesToReport, m.logger)
	m.dimHandler = metrics.NewDimensionHandler(m.Output.SendDimensionUpdate, m.config.UpdatesForNodeDimension, m.logger)
	m.stop = make(chan struct{})

	return m.Start()
}

// Start starts syncing resources and sending datapoints to ingest
func (m *Monitor) Start() error {
	ticker := time.NewTicker(time.Second * time.Duration(m.config.IntervalSeconds))

	shouldReport := m.config.AlwaysClusterReporter

	clusterState, err := newState(m.distribution, m.restConfig, m.datapointCache, m.dimHandler, m.config.Namespace, m.logger)
	if err != nil {
		return err
	}

	var leaderCh <-chan bool
	var unregister func()

	if m.config.AlwaysClusterReporter {
		clusterState.Start()
	} else {
		var err error
		leaderCh, unregister, err = leadership.RequestLeaderNotification(clusterState.clientset.CoreV1(), clusterState.clientset.CoordinationV1(), m.logger)
		if err != nil {
			return err
		}
	}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-m.stop:
				if unregister != nil {
					unregister()
				}
				clusterState.Stop()
				return
			case isLeader := <-leaderCh:
				if isLeader {
					shouldReport = true
					clusterState.Start()
				} else {
					shouldReport = false
					clusterState.Stop()
				}
			case <-ticker.C:
				if shouldReport {
					m.sendLatestDatapoints()
				}
			}
		}
	}()

	return nil
}

// Synchonously send all of the cached datapoints to ingest
func (m *Monitor) sendLatestDatapoints() {
	dps := m.datapointCache.AllDatapoints()

	now := time.Now()
	for i := range dps {
		dps[i].Timestamp = now
		dps[i].Meta[dpmeta.NotHostSpecificMeta] = true
	}
	m.Output.SendDatapoints(dps...)
}

// Shutdown halts everything that is syncing
func (m *Monitor) Shutdown() {
	if m.stop != nil {
		close(m.stop)
	}
}
