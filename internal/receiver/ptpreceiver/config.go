// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"

	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver/internal/metadata"
)

// Config controls collection from a local linuxptp ptp4l instance.
type Config struct {
	SocketPath                     string    `mapstructure:"socket_path"`
	PMC                            PMCConfig `mapstructure:"pmc"`
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	DomainNumber                   int `mapstructure:"domain_number"`
}

// PMCConfig controls how the receiver invokes LinuxPTP's pmc client.
type PMCConfig struct {
	Path                  string `mapstructure:"path"`
	ClientSocketDirectory string `mapstructure:"client_socket_directory"`
}

func (c *Config) Validate() error {
	if err := c.ControllerConfig.Validate(); err != nil {
		return err
	}
	if c.SocketPath == "" || !path.IsAbs(c.SocketPath) {
		return errors.New("socket_path must be an absolute path")
	}
	if c.PMC.Path == "" {
		return errors.New("pmc.path must not be empty")
	}
	if c.PMC.ClientSocketDirectory != "" && !filepath.IsAbs(c.PMC.ClientSocketDirectory) {
		return errors.New("pmc.client_socket_directory must be an absolute path")
	}
	if c.DomainNumber < 0 || c.DomainNumber > 255 {
		return fmt.Errorf("domain_number must be between 0 and 255, got %d", c.DomainNumber)
	}
	return nil
}
