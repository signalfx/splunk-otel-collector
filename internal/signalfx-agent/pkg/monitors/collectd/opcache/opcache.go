//go:build linux
// +build linux

package opcache

//go:generate ../../../../scripts/collectd-template-to-go opcache.tmpl

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			*collectd.NewMonitorCore(CollectdTemplate),
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// The hostname of the webserver (i.e. `127.0.0.1`).
	Host string `yaml:"host"`
	// The port number of the webserver (i.e. `80`).
	Port uint16 `yaml:"port"`
	// If true, the monitor will use HTTPS connection.
	UseHTTPS bool `yaml:"useHTTPS" default:"false"`
	// The URL path to use for the scrape URL for opcache script.
	Path string `yaml:"path" default:"/opcache_stat.php"`
	// The URL, either a final URL or a Go template that will be populated with
	// the `host`, `port` and `path` values.
	URL string `yaml:"url"`
	// This will be sent as the `plugin_instance` dimension and can be any name
	// you like.
	Name string `yaml:"name"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Scheme definition
func (c *Config) Scheme() string {
	if c.UseHTTPS {
		return "https"
	}
	return "http"
}

// Configure configures and runs the plugin in collectd
func (am *Monitor) Configure(conf *Config) error {
	if conf.URL == "" {
		conf.URL = fmt.Sprintf("%s://%s:%d%s", conf.Scheme(), conf.Host, conf.Port, conf.Path)
	}
	return am.SetConfigurationAndRun(conf)
}
