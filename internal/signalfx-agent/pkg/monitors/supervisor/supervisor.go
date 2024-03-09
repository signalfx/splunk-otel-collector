package supervisor

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/mattn/go-xmlrpc"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`
	// The host/ip address of the Supervisor XML-RPC API. This is used to construct
	// the `url` option if not provided.
	Host string `yaml:"host"`
	// The port of the Supervisor XML-RPC API. This is used to construct the `url`
	// option if not provided. (i.e. `localhost`)
	Port uint16 `yaml:"port" default:"9001"`
	// If true, the monitor will connect to Supervisor via HTTPS instead of
	// HTTP.
	UseHTTPS bool `yaml:"useHTTPS" default:"false"`
	// The URL path to use for the scrape URL for Supervisor.
	Path string `yaml:"path" default:"/RPC2"`
	// URL on which to scrape Supervisor XML-RPC API. If this is not provided,
	// it will be derive from the `host`, `port`, `useHTTPS`, and `path`
	// options. (i.e. `http://localhost:9001/RPC2`)
	URL string `yaml:"url"`
}

// Monitor that collect metrics
type Monitor struct {
	Output types.FilteringOutput
	cancel func()
	logger logrus.FieldLogger
}

// Process contains Supervisor properties
type Process struct {
	Name  string `xmlrpc:"name"`
	Group string `xmlrpc:"group"`
	State int    `xmlrpc:"state"`
}

func (c *Config) Scheme() string {
	if c.UseHTTPS {
		return "https"
	}
	return "http"
}

// ScrapeURL from config options
func (c *Config) ScrapeURL() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("%s://%s:%d%s", c.Scheme(), c.Host, c.Port, c.Path)
}

// Configure and kick off internal metric collection
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// Start the metric gathering process here
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	url, err := url.Parse(conf.ScrapeURL())
	if err != nil {
		return fmt.Errorf("cannot parse url %s status. %v", conf.ScrapeURL(), err)
	}
	utils.RunOnInterval(ctx, func() {
		client := xmlrpc.NewClient(url.String())
		res, err := client.Call("supervisor.getAllProcessInfo")
		if err != nil {
			m.logger.WithError(err).Error("unable to call supervisor xmlrpc")
			return
		}

		var process Process
		for _, p := range res.(xmlrpc.Array) {
			for k, v := range p.(xmlrpc.Struct) {
				switch k {
				case "name":
					process.Name = v.(string)
				case "group":
					process.Group = v.(string)
				case "state":
					process.State = v.(int)
				}
			}
			dimensions := map[string]string{
				"name":  process.Name,
				"group": process.Group,
			}
			m.Output.SendDatapoints([]*datapoint.Datapoint{
				datapoint.New(
					supervisorState,
					dimensions,
					datapoint.NewIntValue(int64(process.State)),
					datapoint.Gauge,
					time.Time{},
				),
			}...)
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown the monitor
func (m *Monitor) Shutdown() {
	// Stop any long-running go routines here
	if m.cancel != nil {
		m.cancel()
	}
}
