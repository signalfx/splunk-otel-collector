package haproxy

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

// Config is the config for this monitor.
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// The host/ip address of the HAProxy instance. This is used to construct the `url` option if not provided.
	Host string `yaml:"host"`
	// The port of the HAProxy instance's stats endpoint (if using HTTP). This is used to construct the `url` option if not provided.
	Port uint16 `yaml:"port"`
	// The path to HAProxy stats. The default is `stats?stats;csv`. This is used to construct the `url` option if not provided.
	Path string `yaml:"path" default:"stats?stats;csv"`
	// Whether to connect on HTTPS or HTTP. If you want to use a UNIX socket, then specify the `url` config option with the format `unix://...` and omit `host`, `port` and `useHTTPS`.
	UseHTTPS bool `yaml:"useHTTPS"`
	// URL on which to scrape HAProxy. Scheme `http://` for http-type and `unix://` socket-type urls. If this is not provided, it will be derive from the `host`, `port`, `path`, and `useHTTPS` options.
	URL string `yaml:"url"`
	// Basic Auth username to use on each request, if any.
	Username string `yaml:"username"`
	// Basic Auth password to use on each request, if any.
	Password string `yaml:"password" neverLog:"true"`
	// Flag that enables SSL certificate verification for the scrape URL.
	SSLVerify bool `yaml:"sslVerify" default:"true"`
	// Timeout for trying to get stats from HAProxy. This should be a duration string that is accepted by https://golang.org/pkg/time/#ParseDuration
	Timeout timeutil.Duration `yaml:"timeout" default:"5s"`
	// A list of the pxname(s) and svname(s) to monitor (e.g. `["http-in", "server1", "backend"]`). If empty then metrics for all proxies will be reported.
	Proxies []string `yaml:"proxies"`
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
