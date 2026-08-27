package promqlreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Config struct {
	ControllerConfig scraperhelper.ControllerConfig `mapstructure:",squash"`
	ClientConfig     confighttp.ClientConfig        `mapstructure:",squash"`
	Queries          []string                       `mapstructure:"queries"`
}

func (c *Config) Validate() error {
	if len(c.Queries) == 0 {
		return errors.New("queries cannot be empty")
	}
	if c.ClientConfig.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	return nil
}
