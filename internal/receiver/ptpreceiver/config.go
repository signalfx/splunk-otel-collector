// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"errors"
	"fmt"
	"path"

	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

// Config controls collection from a local linuxptp ptp4l instance.
type Config struct {
	SocketPath                     string `mapstructure:"socket_path"`
	PMCPath                        string `mapstructure:"pmc_path"`
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	DomainNumber                   int `mapstructure:"domain_number"`
}

func (c *Config) Validate() error {
	if err := c.ControllerConfig.Validate(); err != nil {
		return err
	}
	if c.SocketPath == "" || !path.IsAbs(c.SocketPath) {
		return errors.New("socket_path must be an absolute path")
	}
	if c.PMCPath == "" {
		return errors.New("pmc_path must not be empty")
	}
	if c.DomainNumber < 0 || c.DomainNumber > 255 {
		return fmt.Errorf("domain_number must be between 0 and 255, got %d", c.DomainNumber)
	}
	return nil
}
