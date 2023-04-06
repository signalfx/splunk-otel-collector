package cadvisor

import (
	"fmt"
	"time"

	"github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func init() {
	monitors.Register(&cadvisorMonitorMetadata, func() interface{} { return &Cadvisor{} }, &CHTTPConfig{})
}

// CHTTPConfig is the monitor-specific config for cAdvisor
type CHTTPConfig struct {
	config.MonitorConfig `yaml:",inline"`
	// Where to find cAdvisor
	CAdvisorURL string `yaml:"cadvisorURL" default:"http://localhost:4194"`
}

// Cadvisor is the monitor that goes straight to the exposed cAdvisor port to
// get metrics
type Cadvisor struct {
	Monitor
	Output types.Output
}

// Configure the cAdvisor monitor
func (c *Cadvisor) Configure(conf *CHTTPConfig) error {
	cadvisorClient, err := client.NewClient(conf.CAdvisorURL)
	if err != nil {
		return fmt.Errorf("could not create cAdvisor client: %w", err)
	}

	return c.Monitor.Configure(&conf.MonitorConfig, c.Output.SendDatapoints, newCadvisorInfoProvider(cadvisorClient), false)
}

type cadvisorInfoProvider struct {
	cc         *client.Client
	lastUpdate time.Time
}

func (cip *cadvisorInfoProvider) GetEphemeralStatsFromPods() ([]stats.PodStats, error) {
	// cadvisor does not collect Pod level metrics
	return nil, nil
}

func (cip *cadvisorInfoProvider) SubcontainersInfo(containerName string) ([]info.ContainerInfo, error) {
	curTime := time.Now()
	info, err := cip.cc.AllDockerContainers(&info.ContainerInfoRequest{Start: cip.lastUpdate, End: curTime})
	if len(info) > 0 {
		cip.lastUpdate = curTime
	}
	return info, err
}

func (cip *cadvisorInfoProvider) GetMachineInfo() (*info.MachineInfo, error) {
	return cip.cc.MachineInfo()
}

func newCadvisorInfoProvider(cadvisorClient *client.Client) *cadvisorInfoProvider {
	return &cadvisorInfoProvider{
		cc:         cadvisorClient,
		lastUpdate: time.Now(),
	}
}
