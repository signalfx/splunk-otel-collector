// Package docker contains a monitor for getting metrics about containers running
// in a docker engine.
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	dtypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	dockercommon "github.com/signalfx/signalfx-agent/pkg/core/common/docker"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

const dockerAPIVersion = "v1.24"

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// EnhancedMetricsConfig to decide if it will send out all custom metrics
type EnhancedMetricsConfig struct {
	// Whether it will send all extra block IO metrics as well.
	EnableExtraBlockIOMetrics bool `yaml:"enableExtraBlockIOMetrics" default:"false"`
	// Whether it will send all extra CPU metrics as well.
	EnableExtraCPUMetrics bool `yaml:"enableExtraCPUMetrics" default:"false"`
	// Whether it will send all extra memory metrics as well.
	EnableExtraMemoryMetrics bool `yaml:"enableExtraMemoryMetrics" default:"false"`
	// Whether it will send all extra network metrics as well.
	EnableExtraNetworkMetrics bool `yaml:"enableExtraNetworkMetrics" default:"false"`
}

// Config for this monitor
type Config struct {
	config.MonitorConfig  `yaml:",inline" acceptsEndpoints:"false"`
	EnhancedMetricsConfig `yaml:",inline"`

	// The URL of the docker server
	DockerURL string `yaml:"dockerURL" default:"unix:///var/run/docker.sock"`
	// The maximum amount of time to wait for docker API requests
	TimeoutSeconds int `yaml:"timeoutSeconds" default:"5"`
	// The time to wait before resyncing the list of containers the monitor maintains
	// through the docker event listener example: cacheSyncInterval: "20m"
	CacheSyncInterval timeutil.Duration `yaml:"cacheSyncInterval" default:"60m"`
	// A mapping of container label names to dimension names. The corresponding
	// label values will become the dimension value for the mapped name.  E.g.
	// `io.kubernetes.container.name: container_spec_name` would result in a
	// dimension called `container_spec_name` that has the value of the
	// `io.kubernetes.container.name` container label.
	LabelsToDimensions map[string]string `yaml:"labelsToDimensions"`
	// A mapping of container environment variable names to dimension
	// names.  The corresponding env var values become the dimension values on
	// the emitted metrics.  E.g. `APP_VERSION: version` would result in
	// datapoints having a dimension called `version` whose value is the value
	// of the `APP_VERSION` envvar configured for that particular container, if
	// present.
	EnvToDimensions map[string]string `yaml:"envToDimensions"`
	// A list of filters of images to exclude.  Supports literals, globs, and
	// regex.
	ExcludedImages []string `yaml:"excludedImages"`
}

// Monitor for Docker
type Monitor struct {
	Output  types.FilteringOutput
	cancel  func()
	ctx     context.Context
	client  *docker.Client
	timeout time.Duration
	logger  logrus.FieldLogger
}

type dockerContainer struct {
	*dtypes.ContainerJSON
	EnvMap map[string]string
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	enhancedMetricsConfig := EnableExtraGroups(conf.EnhancedMetricsConfig, m.Output.EnabledMetrics())

	defaultHeaders := map[string]string{"User-Agent": "signalfx-agent"}

	var err error
	m.client, err = docker.NewClient(conf.DockerURL, dockerAPIVersion, nil, defaultHeaders)
	if err != nil {
		return fmt.Errorf("could not create docker client: %w", err)
	}

	m.timeout = time.Duration(conf.TimeoutSeconds) * time.Second

	m.ctx, m.cancel = context.WithCancel(context.Background())

	imageFilter, err := filter.NewBasicStringFilter(conf.ExcludedImages)
	if err != nil {
		return err
	}

	lock := sync.Mutex{}
	containers := map[string]dockerContainer{}
	isRegistered := false

	changeHandler := func(old *dtypes.ContainerJSON, new *dtypes.ContainerJSON) {
		if old == nil && new == nil {
			return
		}

		var id string
		if new != nil {
			id = new.ID
		} else {
			id = old.ID
		}

		lock.Lock()
		defer lock.Unlock()

		if new == nil || (!new.State.Running || new.State.Paused) {
			m.logger.Debugf("Container %s is no longer running", id)
			delete(containers, id)
			return
		}
		m.logger.Infof("Monitoring docker container %s", id)
		containers[id] = dockerContainer{
			ContainerJSON: new,
			EnvMap:        parseContainerEnvSlice(new.Config.Env),
		}
	}

	utils.RunOnInterval(m.ctx, func() {
		// Repeat the watch setup in the face of errors in case the docker
		// engine is non-responsive when the monitor starts.
		if !isRegistered {
			dockercommon.ListAndWatchContainers(m.ctx, m.client, changeHandler, imageFilter, m.logger, conf.CacheSyncInterval.AsDuration())
			isRegistered = true
		}

		// Individual container objects don't need to be protected by the lock,
		// only the map that holds them.
		lock.Lock()
		for id := range containers {
			go m.fetchStats(containers[id], conf.LabelsToDimensions, conf.EnvToDimensions, enhancedMetricsConfig)
		}
		lock.Unlock()

	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Instead of streaming stats like the collectd plugin does, fetch the stats in
// parallel in individual goroutines.  This is much easier on CPU usage since
// we aren't doing something every second across all containers, but only
// something once every metric interval.
func (m *Monitor) fetchStats(container dockerContainer, labelMap map[string]string, envMap map[string]string, enhancedMetricsConfig EnhancedMetricsConfig) {
	ctx, cancel := context.WithTimeout(m.ctx, m.timeout)
	stats, err := m.client.ContainerStats(ctx, container.ID, false)
	if err != nil {
		cancel()
		if isContainerNotFound(err) {
			m.logger.Debugf("container %s is not found in cache", container.ID)
			return
		}
		m.logger.WithError(err).Errorf("Could not fetch docker stats for container id %s", container.ID)
		return
	}

	var parsed dtypes.StatsJSON
	err = json.NewDecoder(stats.Body).Decode(&parsed)
	stats.Body.Close()
	if err != nil {
		cancel()
		// EOF means that there aren't any stats, perhaps because the container
		// is gone.  Just return nothing and no error.
		if err == io.EOF {
			return
		}
		m.logger.WithError(err).Errorf("Could not parse docker stats for container id %s", container.ID)
		return
	}

	dps, err := ConvertStatsToMetrics(container.ContainerJSON, &parsed, enhancedMetricsConfig)
	cancel()
	if err != nil {
		m.logger.WithError(err).Errorf("Could not convert docker stats for container id %s", container.ID)
		return
	}

	for i := range dps {
		for k, dimName := range envMap {
			if v := container.EnvMap[k]; v != "" {
				dps[i].Dimensions[dimName] = v
			}
		}
		for k, dimName := range labelMap {
			if v := container.Config.Labels[k]; v != "" {
				dps[i].Dimensions[dimName] = v
			}
		}
	}
	m.Output.SendDatapoints(dps...)
}

func parseContainerEnvSlice(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, v := range env {
		parts := strings.Split(v, "=")
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			continue
		}
		out[parts[0]] = parts[1]
	}
	return out
}

// EnableExtraGroups enables extra metrics that were individually turned on
// by ExtraMetrics/ExtraGroups configuration
func EnableExtraGroups(initConf EnhancedMetricsConfig, enabledMetrics []string) EnhancedMetricsConfig {
	groupEnableMap := map[string]bool{
		groupBlkio:   initConf.EnableExtraBlockIOMetrics,
		groupCPU:     initConf.EnableExtraCPUMetrics,
		groupMemory:  initConf.EnableExtraMemoryMetrics,
		groupNetwork: initConf.EnableExtraNetworkMetrics,
	}

	for _, metric := range enabledMetrics {
		if metricInfo, ok := metricSet[metric]; ok {
			groupEnableMap[metricInfo.Group] = true
		}
	}

	return EnhancedMetricsConfig{
		EnableExtraBlockIOMetrics: groupEnableMap[groupBlkio],
		EnableExtraCPUMetrics:     groupEnableMap[groupCPU],
		EnableExtraMemoryMetrics:  groupEnableMap[groupMemory],
		EnableExtraNetworkMetrics: groupEnableMap[groupNetwork],
	}
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}

}

// GetExtraMetrics returns additional metrics that should be allowed through.
func (c *Config) GetExtraMetrics() []string {
	var extraMetrics []string

	if c.EnableExtraBlockIOMetrics {
		extraMetrics = append(extraMetrics, groupMetricsMap[groupBlkio]...)
	}

	if c.EnableExtraCPUMetrics {
		extraMetrics = append(extraMetrics, groupMetricsMap[groupCPU]...)
	}

	if c.EnableExtraMemoryMetrics {
		extraMetrics = append(extraMetrics, groupMetricsMap[groupMemory]...)
	}

	if c.EnableExtraNetworkMetrics {
		extraMetrics = append(extraMetrics, groupMetricsMap[groupNetwork]...)
	}

	return extraMetrics
}

func isContainerNotFound(err error) (notfound bool) {
	// ref: https://github.com/moby/moby/blob/master/container/view.go#L116
	if err != nil && strings.Contains(err.Error(), "no such container") {
		notfound = true
	}

	return
}
