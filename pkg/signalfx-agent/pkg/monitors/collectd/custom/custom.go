//go:build linux
// +build linux

// Package custom contains a custom collectd plugin monitor, for which you can
// specify your own config template and parameters.
package custom

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			MonitorCore: *collectd.NewMonitorCore(template.New("custom")),
		}
	}, &Config{})
}

// Config is the configuration for the collectd custom monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	// This should generally not be set manually, but will be filled in by the
	// agent if using service discovery. It can be accessed in the provided
	// config template with `{{.Host}}`.  It will be set to the hostname or IP
	// address of the discovered service. If you aren't using service
	// discovery, you can just hardcode the host/port in the config template
	// and ignore these fields.
	Host string `yaml:"host"`
	// This should generally not be set manually, but will be filled in by the
	// agent if using service discovery. It can be accessed in the provided
	// config template with `{{.Port}}`.  It will be set to the port of the
	// discovered service, if it is a TCP/UDP endpoint.
	Port uint16 `yaml:"port"`
	// This should generally not be set manually, but will be filled in by the
	// agent if using service discovery. It can be accessed in the provided
	// config template with `{{.Name}}`.  It will be set to the name that the
	// observer creates for the endpoint upon discovery.  You can generally
	// ignore this field.
	Name string `yaml:"name"`

	// A config template for collectd.  You can include as many plugin blocks
	// as you want in this value.  It is rendered as a standard Go template, so
	// be mindful of the delimiters `{{` and `}}`.
	Template string `yaml:"template"`
	// A list of templates, but otherwise equivalent to the above `template`
	// option.  This enables you to have a single directory with collectd
	// configuration files and load them all by using a globbed remote config
	// value:
	Templates []string `yaml:"templates"`

	// The number of read threads to use in collectd.  Will default to the
	// number of templates provided, capped at 10, but if you manually specify
	// it there is no limit.
	CollectdReadThreads int `yaml:"collectdReadThreads"`
}

func (c *Config) allTemplates() []string {
	templates := c.Templates
	if c.Template != "" {
		templates = append(templates, c.Template)
	}
	return templates
}

// Validate will check the config that is specific to this monitor
func (c *Config) Validate() error {
	for _, templateText := range c.allTemplates() {
		if _, err := templateFromText(templateText); err != nil {
			return err
		}
	}

	if c.DiscoveryRule != "" && len(c.allTemplates()) > 1 {
		return errors.New("You should not have multiple templates and a discovery " +
			"rule on a custom collectd monitor")
	}

	return nil
}

func templateFromText(templateText string) (*template.Template, error) {
	template, err := collectd.InjectTemplateFuncs(template.New("custom")).Parse(templateText)
	if err != nil {
		return nil, fmt.Errorf("Template text failed to parse: \n%s\n%w", templateText, err)
	}
	return template, nil
}

// Monitor is the core monitor object that gets instantiated by the agent
type Monitor struct {
	collectd.MonitorCore
}

// Configure will render the custom collectd config and queue a collectd
// restart.
func (cm *Monitor) Configure(conf *Config) error {
	templateTextConcatenated := ""
	for _, text := range conf.allTemplates() {
		templateTextConcatenated += "\n" + text + "\n"
	}

	// Allow blank template text so that we have a standard config item that
	// configured the monitor with all of the templates in a possibly
	// non-existent legacy collectd managed_config dir.
	if templateTextConcatenated == "" {
		return nil
	}

	var err error
	cm.Template, err = templateFromText(templateTextConcatenated)
	if err != nil {
		return err
	}
	// always run an isolated collectd instance per monitor instance
	conf.IsolatedCollectd = true
	return cm.SetConfigurationAndRun(conf)
}
