package vsphere

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type Monitor struct {
	Output types.Output
	cancel func()
	logger logrus.FieldLogger
}

func init() {
	monitors.Register(
		&monitorMetadata,
		func() interface{} { return &Monitor{} },
		&model.Config{},
	)
}

func (m *Monitor) Configure(conf *model.Config) error {
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	m.logger.Warn("[NOTICE] The smartagent/vsphere monitor is deprecated and will be removed on or after December 2026. Use the vcenter receiver instead.")
	r := newRunner(ctx, m.logger, conf, m)
	// 20 seconds is the fixed, real-time metrics interval for vsphere/esxi
	utils.RunOnInterval(ctx, r.run, model.RealtimeMetricsInterval*time.Second)
	return nil
}

func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
