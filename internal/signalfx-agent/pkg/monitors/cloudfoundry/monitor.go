package cloudfoundry

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`
	RLPGatewayURL        string `yaml:"rlpGatewayUrl" required:"true"`
	UAAUser              string `yaml:"uaaUser" required:"true"`
	UAAPassword          string `yaml:"uaaPassword" required:"true" neverLog:"true"`
	UAAURL               string `yaml:"uaaUrl" required:"true"`
	ShardID              string `yaml:"shardId" default:"signalfx_nozzle"`
	RLPGatewaySkipVerify bool   `yaml:"rlpGatewaySkipVerify"`
	UAASkipVerify        bool   `yaml:"uaaSkipVerify"`
}

func (c *Config) Validate() error {
	if _, err := url.Parse(c.RLPGatewayURL); err != nil {
		return fmt.Errorf("failed to parse rlpGatewayUrl: %v", err)
	}
	if _, err := url.Parse(c.UAAURL); err != nil {
		return fmt.Errorf("failed to parse uaaUrl: %v", err)
	}
	return nil
}

// Monitor for load
type Monitor struct {
	Output types.Output
	cancel func()
	logger logrus.FieldLogger
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	go func() {
		for {
			uaaToken, err := getUAAToken(conf.UAAURL, conf.UAAUser, conf.UAAPassword, conf.UAASkipVerify)
			if err != nil {
				m.logger.WithError(err).Errorf("Could not get UAA access token for user %s, retrying...", conf.UAAUser)
				time.Sleep(10 * time.Second)
				continue
			}

			client := NewSignalFxGatewayClient(conf.RLPGatewayURL, uaaToken, conf.RLPGatewaySkipVerify, m.logger)
			client.ShardID = conf.ShardID

			m.logger.Info("Running RLP Gateway streamer")
			err = client.Run(ctx, m.Output.SendDatapoints)
			if err != nil && err != context.Canceled {
				m.logger.WithError(err).Error("Gateway streamer shut down unexpectedly, retrying...")
				time.Sleep(10 * time.Second)
				continue
			}
			return
		}
	}()
	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
