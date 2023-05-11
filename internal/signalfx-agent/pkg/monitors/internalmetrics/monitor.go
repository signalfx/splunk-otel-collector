package internalmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Config for internal metric monitoring
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	// Defaults to the top-level `internalStatusHost` option
	Host string `yaml:"host"`
	// Defaults to the top-level `internalStatusPort` option
	Port uint16 `yaml:"port" noDefault:"true"`
	// The HTTP request path to use to retrieve the metrics
	Path string `yaml:"path" default:"/metrics"`
}

// Monitor for collecting internal metrics from the simple server that dumps
// them.
type Monitor struct {
	Output    types.Output
	AgentMeta *meta.AgentMeta
	cancel    func()
	logger    log.FieldLogger
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Configure and kick off internal metric collection
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Mark the first interval so that errors are suppressed due to the race
	// between this monitor starting and the internal metrics service getting
	// initialized
	firstTime := true

	utils.RunOnInterval(ctx, func() {
		defer func() { firstTime = false }()

		// Derive the url each time since the AgentMeta data can change but
		// there is no notification system for it.
		host := conf.Host
		if host == "" {
			host = m.AgentMeta.InternalStatusHost
		}

		port := conf.Port
		if port == 0 {
			port = m.AgentMeta.InternalStatusPort
		}

		url := fmt.Sprintf("http://%s:%d%s", host, port, conf.Path)

		logger := m.logger.WithField("url", url)

		resp, err := client.Get(url)
		if err != nil {
			if !firstTime {
				logger.WithError(err).Error("Could not connect to internal metric server")
			}
			return
		}
		defer resp.Body.Close()

		dps := make([]*datapoint.Datapoint, 0)
		err = json.NewDecoder(resp.Body).Decode(&dps)
		if err != nil {
			logger.WithError(err).Error("Could not parse metrics from internal metric server")
			return
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown the internal metric collection
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
