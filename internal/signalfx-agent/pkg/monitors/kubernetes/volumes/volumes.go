package volumes

import (
	"context"
	"fmt"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	log "github.com/sirupsen/logrus"
	k8s "k8s.io/client-go/kubernetes"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline"`
	// Kubelet kubeletClient configuration
	KubeletAPI kubelet.APIConfig `yaml:"kubeletAPI" default:""`
	// Configuration of the Kubernetes API kubeletClient
	KubernetesAPI *kubernetes.APIConfig `yaml:"kubernetesAPI" default:"{}"`
}

// Monitor for K8s volume metrics as reported by kubelet
type Monitor struct {
	Output        types.Output
	cancel        func()
	kubeletClient *kubelet.Client
	k8sClient     *k8s.Clientset
	dimCache      map[string]map[string]string
	logger        log.FieldLogger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	var err error
	m.kubeletClient, err = kubelet.NewClient(&conf.KubeletAPI, m.logger)
	if err != nil {
		return err
	}

	m.k8sClient, err = kubernetes.MakeClient(conf.KubernetesAPI)
	if err != nil {
		return err
	}

	m.dimCache = make(map[string]map[string]string)

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	utils.RunOnInterval(ctx, func() {
		dps, err := m.getVolumeMetrics()
		if err != nil {
			m.logger.WithError(err).Error("Could not get volume metrics")
			return
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

func (m *Monitor) getVolumeMetrics() ([]*datapoint.Datapoint, error) {
	summary, err := m.getSummary()
	if err != nil {
		return nil, err
	}

	var dps []*datapoint.Datapoint
	for _, p := range summary.Pods {
		for _, v := range p.VolumeStats {
			dims := map[string]string{
				"volume":               v.Name,
				"kubernetes_pod_uid":   p.PodRef.UID,
				"kubernetes_pod_name":  p.PodRef.Name,
				"kubernetes_namespace": p.PodRef.Namespace,
			}

			volumeDims, err := m.volumeIDDimsForPod(p.PodRef.Name, p.PodRef.Namespace, p.PodRef.UID, v.Name)
			if err != nil {
				m.logger.WithFields(log.Fields{
					"error":     err,
					"volName":   v.Name,
					"podName":   p.PodRef.Name,
					"namespace": p.PodRef.Namespace,
				}).Error("Could not get volume-specific dimensions")
			} else {
				for k, v := range volumeDims {
					dims[k] = v
				}
			}

			if v.AvailableBytes != nil {
				// uint64 -> int64 conversion should be safe since disk sizes
				// aren't going to get that big for a long time.
				dps = append(dps, sfxclient.Gauge(kubernetesVolumeAvailableBytes, dims, int64(*v.AvailableBytes)))
			}
			if v.CapacityBytes != nil {
				dps = append(dps, sfxclient.Gauge(kubernetesVolumeCapacityBytes, dims, int64(*v.CapacityBytes)))
			}
			if v.Inodes != nil {
				dps = append(dps, sfxclient.Gauge(kubernetesVolumeInodes, dims, int64(*v.Inodes)))
			}
			if v.InodesFree != nil {
				dps = append(dps, sfxclient.Gauge(kubernetesVolumeInodesFree, dims, int64(*v.InodesFree)))
			}
			if v.InodesUsed != nil {
				dps = append(dps, sfxclient.Gauge(kubernetesVolumeInodesUsed, dims, int64(*v.InodesUsed)))
			}
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
