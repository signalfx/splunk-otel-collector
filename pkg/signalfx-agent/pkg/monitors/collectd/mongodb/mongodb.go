package mongodb

import (
	"errors"

	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"

	"github.com/signalfx/signalfx-agent/pkg/core/config"

	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			python.PyMonitor{
				MonitorCore: subproc.New(),
			},
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	// Host name/IP address of the Mongo instance
	Host string `yaml:"host" validate:"required"`
	// Port of the Mongo instance (default: 27017)
	Port uint16 `yaml:"port" validate:"required"`
	// Name(s) of database(s) that you would like metrics from. Note: the first
	// database in this list must be "admin", as it is used to perform a
	// `serverStatus()` command.
	Databases []string `yaml:"databases" validate:"required"`
	// The MongoDB user to connect as
	Username string `yaml:"username"`
	// The password of the above user
	Password string `yaml:"password" neverLog:"true"`
	// If true, will connect to Mongo using TLS
	UseTLS *bool `yaml:"useTLS"`
	// Path to a CA cert that will be used to verify the certificate that Mongo
	// presents (not needed if not using TLS or if Mongo's cert is signed by a
	// globally trusted issuer already installed in the default location on
	// your OS)
	CACerts string `yaml:"caCerts"`
	// Path to a client certificate (not needed unless your Mongo instance
	// requires x509 client verification)
	TLSClientCert string `yaml:"tlsClientCert"`
	// Path to a client certificate key (not needed unless your Mongo instance
	// requires x509 client verification, or if your client cert above has the
	// key included)
	TLSClientKey string `yaml:"tlsClientKey"`
	// Passphrase for the TLSClientKey above
	TLSClientKeyPassPhrase string `yaml:"tlsClientKeyPassPhrase"`
	// Whether to send collection level metrics or not
	SendCollectionMetrics *bool `yaml:"sendCollectionMetrics"`
	// Whether to send collection level top (timing) metrics or not
	SendCollectionTopMetrics *bool `yaml:"sendCollectionTopMetrics"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Validate will check the config for correctness.
func (c *Config) Validate() error {
	if len(c.Databases) == 0 {
		return errors.New("must specify at least one database for MongoDB")
	}
	return nil
}

// GetExtraMetrics returns a list of metrics that should be let through the
// filtering based on config flags.
func (c *Config) GetExtraMetrics() []string {
	var out []string
	if c.SendCollectionMetrics != nil && *c.SendCollectionMetrics {
		out = append(out, groupMetricsMap[groupCollection]...)
	}
	if c.SendCollectionTopMetrics != nil && *c.SendCollectionTopMetrics {
		out = append(out, groupMetricsMap[groupCollectionTop]...)
	}
	return out
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	sendCollMetrics := conf.SendCollectionMetrics
	sendCollTopMetrics := conf.SendCollectionTopMetrics

	if m.Output.HasEnabledMetricInGroup(groupCollection) {
		sendCollMetrics = pointer.Bool(true)
	}
	if m.Output.HasEnabledMetricInGroup(groupCollectionTop) {
		sendCollTopMetrics = pointer.Bool(true)
	}

	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "mongodb",
		ModulePaths:   []string{collectd.MakePythonPluginPath("mongodb")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"Host":                     conf.Host,
			"Port":                     conf.Port,
			"Database":                 conf.Databases,
			"UseTLS":                   conf.UseTLS,
			"User":                     conf.Username,
			"Password":                 conf.Password,
			"CACerts":                  conf.CACerts,
			"TLSClientCert":            conf.TLSClientCert,
			"TLSClientKey":             conf.TLSClientKey,
			"TLSClientKeyPassphrase":   conf.TLSClientKeyPassPhrase,
			"SendCollectionMetrics":    sendCollMetrics,
			"SendCollectionTopMetrics": sendCollTopMetrics,
		},
	}

	return m.PyMonitor.Configure(conf)
}
