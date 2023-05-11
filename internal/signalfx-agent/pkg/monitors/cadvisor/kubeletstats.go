package cadvisor

// Parts of this module are copied from the heapster project, specifically the
// file https://github.com/kubernetes/heapster/blob/master/metrics/sources/kubelet/kubelet_client.go
// We can't just import the heapster project because it depends on the main K8s
// codebase which breaks a lot of stuff if we try and import it transitively
// alongside the k8s client-go library.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	info "github.com/google/cadvisor/info/v1"
	"github.com/sirupsen/logrus"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubelet"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func init() {
	monitors.Register(&kubeletStatsMonitorMetadata, func() interface{} { return &KubeletStatsMonitor{} }, &KubeletStatsConfig{})
}

// KubeletStatsConfig respresents config for the Kubelet stats monitor
type KubeletStatsConfig struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// Kubelet client configuration
	KubeletAPI kubelet.APIConfig `yaml:"kubeletAPI" default:""`
}

// KubeletStatsMonitor will pull container metrics from the /stats/ endpoint of
// the Kubelet API.  This is the same thing that other K8s metric solutions
// like Heapster use and should eventually completely replace out use of the
// cAdvisor endpoint that some K8s deployments expose.  Right now, this assumes
// a certain format of the stats that come off of the endpoints.  TODO: Figure
// out if this is versioned and how to access versioned endpoints.
type KubeletStatsMonitor struct {
	Monitor
	Output types.FilteringOutput
}

// Configure the Kubelet Stats monitor
func (ks *KubeletStatsMonitor) Configure(conf *KubeletStatsConfig) error {
	ks.logger = logrus.WithFields(logrus.Fields{"monitorType": conf.MonitorConfig.Type, "monitorID": conf.MonitorID})
	client, err := kubelet.NewClient(&conf.KubeletAPI, ks.logger)
	if err != nil {
		return err
	}

	return ks.Monitor.Configure(&conf.MonitorConfig, ks.Output.SendDatapoints,
		newKubeletInfoProvider(client, ks.logger), ks.Output.HasEnabledMetricInGroup(groupPodEphemeralStats))
}

type statsRequest struct {
	// The name of the container for which to request stats.
	// Default: /
	ContainerName string `json:"containerName,omitempty"`

	// Max number of stats to return.
	// If start and end time are specified this limit is ignored.
	// Default: 60
	NumStats int `json:"num_stats,omitempty"`

	// Start time for which to query information.
	// If omitted, the beginning of time is assumed.
	Start time.Time `json:"start,omitempty"`

	// End time for which to query information.
	// If omitted, current time is assumed.
	End time.Time `json:"end,omitempty"`

	// Whether to also include information from subcontainers.
	// Default: false.
	Subcontainers bool `json:"subcontainers,omitempty"`
}

type kubeletInfoProvider struct {
	client *kubelet.Client
	logger logrus.FieldLogger
}

func newKubeletInfoProvider(client *kubelet.Client, logger logrus.FieldLogger) *kubeletInfoProvider {
	return &kubeletInfoProvider{
		client: client,
		logger: logger,
	}
}

func (kip *kubeletInfoProvider) SubcontainersInfo(containerName string) ([]info.ContainerInfo, error) {
	containers, err := kip.getAllContainersLatestStats()
	if err != nil {
		return nil, err
	}

	return filterPodContainers(containers), nil
}

func filterPodContainers(containers []info.ContainerInfo) []info.ContainerInfo {
	out := make([]info.ContainerInfo, 0)
	for _, c := range containers {
		// Only get containers that are in pods
		if c.Spec.Labels != nil || len(c.Spec.Labels["io.kubernetes.pod.uid"]) > 0 {
			out = append(out, c)
		}
	}
	return out
}

func (kip *kubeletInfoProvider) GetMachineInfo() (*info.MachineInfo, error) {
	req, err := kip.client.NewRequest("GET", "/spec/", nil)
	if err != nil {
		return nil, err
	}

	machineInfo := info.MachineInfo{}
	err = kip.client.DoRequestAndSetValue(req, &machineInfo)
	if err != nil {
		return nil, err
	}

	return &machineInfo, nil
}

func (kip *kubeletInfoProvider) getAllContainersLatestStats() ([]info.ContainerInfo, error) {
	// Request data from all subcontainers.
	request := statsRequest{
		ContainerName: "/",
		NumStats:      1,
		Subcontainers: true,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	kip.logger.Debugf("Sending body to kubelet stats endpoint: %s", body)
	req, err := kip.client.NewRequest("POST", "/stats/container/", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var containers map[string]info.ContainerInfo
	err = kip.client.DoRequestAndSetValue(req, &containers)
	if err != nil {
		return nil, fmt.Errorf("failed to get all container stats from Kubelet URL %q: %v", req.URL.String(), err)
	}

	result := make([]info.ContainerInfo, 0, len(containers))
	for _, containerInfo := range containers {
		result = append(result, containerInfo)
	}
	return result, nil
}

func (kip *kubeletInfoProvider) GetEphemeralStatsFromPods() ([]stats.PodStats, error) {
	req, err := kip.client.NewRequest("POST", "/stats/summary/", nil)
	if err != nil {
		return nil, err
	}

	var summary stats.Summary
	err = kip.client.DoRequestAndSetValue(req, &summary)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary stats from Kubelet URL %q: %v", req.URL.String(), err)
	}

	return summary.Pods, nil
}
