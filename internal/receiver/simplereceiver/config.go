// Responsible for defining, receiving and validating the configuration of the receiver.

package githubmetricsreceiver

import (
	"fmt"
	"time"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
)

type Config struct {
    config.ReceiverSettings       `mapstructure:",squash"`
    confighttp.HTTPClientSettings `mapstructure:",squash"`
    Interval                      string `mapstructure:"interval"`
    APIKey                        string `mapstructure:"api_key"`
    RepoName                      string `mapstructure:"repo_name"`
    GitUsername                   string `mapstructure:"git_username"`
}

func (cfg *Config) Validate() error {
    interval, _ := time.ParseDuration(cfg.Interval)
    if (cfg.Endpoint == "") {
        return fmt.Errorf("You must provide a valid endpoint")
    }

    if (cfg.APIKey == "") {
        return fmt.Errorf("A valid API key is required for the snowflake receiver")
    }

    if (interval.Minutes() < 60) {
        return fmt.Errorf("Interval must be set to at least 1 hour")
    }

    return nil
}
