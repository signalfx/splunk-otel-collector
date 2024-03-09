package haproxy

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

// Config is the config for this monitor.
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Host                 string            `yaml:"host"`
	Path                 string            `yaml:"path" default:"stats?stats;csv"`
	URL                  string            `yaml:"url"`
	Username             string            `yaml:"username"`
	Password             string            `yaml:"password" neverLog:"true"`
	Proxies              []string          `yaml:"proxies"`
	Timeout              timeutil.Duration `yaml:"timeout" default:"5s"`
	Port                 uint16            `yaml:"port"`
	UseHTTPS             bool              `yaml:"useHTTPS"`
	SSLVerify            bool              `yaml:"sslVerify" default:"true"`
}

func (c *Config) ScrapeURL() string {
	if c.URL != "" {
		return c.URL
	}
	scheme := "http"
	if c.UseHTTPS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/%s", scheme, c.Host, c.Port, c.Path)
}
