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

	// The base URL to the RLP Gateway server. This is quite often of the form
	// https://log-stream.<CLOUD CONTROLLER SYSTEM DOMAIN> if using PCF 2.4+.
	RLPGatewayURL string `yaml:"rlpGatewayUrl" required:"true"`
	// Whether to skip SSL/TLS verification when using HTTPS to connect to the
	// RLP Gateway
	RLPGatewaySkipVerify bool `yaml:"rlpGatewaySkipVerify"`
	// The UAA username for a user that has the appropriate authority to fetch
	// logs from the firehose (usually the `logs.admin` authority)
	UAAUser string `yaml:"uaaUser" required:"true"`
	// The password for the above UAA user
	UAAPassword string `yaml:"uaaPassword" required:"true" neverLog:"true"`
	// The URL to the UAA server. This monitor will obtain an access token
	// from this server that it will use to authenticate with the RLP Gateway.
	UAAURL string `yaml:"uaaUrl" required:"true"`
	// Whether to skip SSL/TLS verification when using HTTPS to connect to the
	// UAA server
	UAASkipVerify bool `yaml:"uaaSkipVerify"`
	// The nozzle's shard id.  All nozzle instances with the same id will
	// receive an exclusive subset of the data from the firehose. The default
	// should suffice in the vast majority of use cases.
	ShardID string `yaml:"shardId" default:"signalfx_nozzle"`
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
