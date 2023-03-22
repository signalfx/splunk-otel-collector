package memory

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"

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
	config.MonitorConfig `yaml:",inline" singleInstance:"true" acceptsEndpoints:"false"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	logger log.FieldLogger
}

// EmitDatapoints emits a set of memory datapoints
func (m *Monitor) emitDatapoints() {
	// mem.VirtualMemory is a gopsutil function
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		m.logger.WithError(err).Errorf("Unable to collect memory stats")
		return
	}

	swapInfo, err := mem.SwapMemory()
	if err != nil {
		m.logger.WithError(err).Errorf("Unable to collect swap memory stats")
		return
	}

	dps := m.makeMemoryDatapoints(memInfo, swapInfo, nil)
	m.Output.SendDatapoints(dps...)
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		m.emitDatapoints()
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
