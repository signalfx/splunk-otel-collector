package forwarder

import (
	"context"
	"fmt"
	"time"

	goliblog "github.com/signalfx/golib/v3/log"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	SendInternalMetrics  *bool `yaml:"sendInternalMetrics" default:"false"`
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" singleInstance:"true"`
	ListenAddress        string            `yaml:"listenAddress" default:"127.0.0.1:9080"`
	ServerTimeout        timeutil.Duration `yaml:"serverTimeout" default:"5s"`
}

// Monitor that accepts and forwards SignalFx data
type Monitor struct {
	Output      types.Output
	cancel      context.CancelFunc
	logger      *utils.ThrottledLogger
	golibLogger goliblog.Logger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID}), 30*time.Second)
	m.golibLogger = &utils.LogrusGolibShim{FieldLogger: m.logger.FieldLogger}

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	sink := &outputSink{Output: m.Output}
	listenerMetrics, err := m.startListening(ctx, conf.ListenAddress, conf.ServerTimeout.AsDuration(), sink)
	if err != nil {
		return fmt.Errorf("could not start forwarder listener: %w", err)
	}

	if *conf.SendInternalMetrics {
		utils.RunOnInterval(ctx, func() {
			m.Output.SendDatapoints(listenerMetrics.Datapoints()...)
		}, time.Duration(conf.IntervalSeconds)*time.Second)
	}

	return nil
}

// Shutdown stops the forwarder and correlation MTSs
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
