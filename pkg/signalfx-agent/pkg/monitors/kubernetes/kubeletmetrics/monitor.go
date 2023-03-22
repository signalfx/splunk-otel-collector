package kubeletmetrics

import (
	"context"
	"fmt"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// Kubelet client configuration
	KubeletAPI kubelet.APIConfig `yaml:"kubeletAPI" default:""`
	// If true, this montior will scrape additional metadata from the `/pods`
	// endpoint on the kubelet to enhance the container metrics with the
	// `container_id` dimension that is not otherwise available from the
	// `/stats/summary` endpoint.
	UsePodsEndpoint bool `yaml:"usePodsEndpoint"`
}

// Monitor for K8s volume metrics as reported by kubelet
type Monitor struct {
	Output        types.Output
	cancel        func()
	kubeletClient *kubelet.Client
	logger        log.FieldLogger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	var err error
	m.kubeletClient, err = kubelet.NewClient(&conf.KubeletAPI, m.logger)
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	utils.RunOnInterval(ctx, func() {
		m.logger.Debug("Collecting kubelet /stats/summary metrics")
		dps, err := m.getSummaryMetrics(conf.UsePodsEndpoint)
		if err != nil {
			m.logger.WithError(err).Error("Could not get summary metrics")
			return
		}
		m.logger.Debugf("Sending kubelet metrics: %v", dps)

		now := time.Now()
		for i := range dps {
			dps[i].Timestamp = now
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

func (m *Monitor) getSummaryMetrics(usePodsEndpoint bool) ([]*datapoint.Datapoint, error) {
	var podsByUID map[k8stypes.UID]*v1.Pod
	if usePodsEndpoint {
		var err error
		podsByUID, err = m.getPodsByUID()
		if err != nil {
			return nil, err
		}
	}

	summary, err := m.getSummary()
	if err != nil {
		return nil, err
	}

	var dps []*datapoint.Datapoint

	for _, p := range summary.Pods {
		podResource := podsByUID[k8stypes.UID(p.PodRef.UID)]
		if usePodsEndpoint && podResource == nil {
			continue
		}

		dims := map[string]string{
			"kubernetes_pod_uid":   p.PodRef.UID,
			"kubernetes_pod_name":  p.PodRef.Name,
			"kubernetes_namespace": p.PodRef.Namespace,
		}

		if p.EphemeralStorage != nil {
			if p.EphemeralStorage.CapacityBytes != nil {
				dps = append(dps, sfxclient.Gauge(podEphemeralStorageCapacityBytes, dims, int64(*p.EphemeralStorage.CapacityBytes)))
			}
			if p.EphemeralStorage.UsedBytes != nil {
				dps = append(dps, sfxclient.Gauge(podEphemeralStorageUsedBytes, dims, int64(*p.EphemeralStorage.UsedBytes)))
			}
		}

		if p.Network != nil && p.Network.Interfaces != nil {
			for _, i := range p.Network.Interfaces {
				intfDims := utils.CloneStringMap(dims)

				intfDims["interface"] = i.Name

				if i.RxBytes != nil {
					dps = append(dps, sfxclient.Cumulative(podNetworkReceiveBytesTotal, intfDims, int64(*i.RxBytes)))
				}
				if i.TxBytes != nil {
					dps = append(dps, sfxclient.Cumulative(podNetworkTransmitBytesTotal, intfDims, int64(*i.TxBytes)))
				}
				if i.RxErrors != nil {
					dps = append(dps, sfxclient.Cumulative(podNetworkReceiveErrorsTotal, intfDims, int64(*i.RxErrors)))
				}
				if i.TxErrors != nil {
					dps = append(dps, sfxclient.Cumulative(podNetworkTransmitErrorsTotal, intfDims, int64(*i.TxErrors)))
				}
			}
		}

		var statusBySpecName map[string]*v1.ContainerStatus
		if usePodsEndpoint {
			statusBySpecName = containerStatusBySpecName(podResource.Status.ContainerStatuses)
		}
		for i := range p.Containers {
			status := statusBySpecName[p.Containers[i].Name]
			if usePodsEndpoint && status == nil {
				continue
			}

			dps = append(dps, convertContainerMetrics(&p.Containers[i], status, utils.CloneStringMap(dims))...)
		}
	}
	return dps, nil
}

func (m *Monitor) getSummary() (*stats.Summary, error) {
	req, err := m.kubeletClient.NewRequest("POST", "/stats/summary/", nil)
	if err != nil {
		return nil, err
	}

	var summary stats.Summary
	err = m.kubeletClient.DoRequestAndSetValue(req, &summary)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary stats from Kubelet URL %q: %v", req.URL.String(), err)
	}

	return &summary, nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
