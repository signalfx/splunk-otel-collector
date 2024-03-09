package jmx

import (
	"fmt"
	"path/filepath"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc/signalfx/java"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			Monitor: java.NewMonitorCore(),
		}
	}, &Config{})
}

// Config for the JMX Monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Password             string `yaml:"password" json:"password" neverLog:"true"`
	ServiceURL           string `yaml:"serviceURL" json:"serviceURL"`
	GroovyScript         string `yaml:"groovyScript" json:"groovyScript" validate:"required"`
	Username             string `yaml:"username" json:"username"`
	Host                 string `yaml:"host" json:"host,omitempty"`
	KeyStorePath         string `yaml:"keyStorePath" json:"keyStorePath"`
	KeyStorePassword     string `yaml:"keyStorePassword" json:"keyStorePassword" neverLog:"true"`
	KeyStoreType         string `yaml:"keyStoreType" json:"keyStoreType" default:"jks"`
	TrustStorePath       string `yaml:"trustStorePath" json:"trustStorePath"`
	TrustStorePassword   string `yaml:"trustStorePassword" json:"trustStorePassword" neverLog:"true"`
	JmxRemoteProfiles    string `yaml:"jmxRemoteProfiles" json:"jmxRemoteProfiles"`
	Realm                string `yaml:"realm" json:"realm"`
	Port                 uint16 `yaml:"port" json:"port,omitempty"`
}

type Monitor struct {
	*java.Monitor
}

func (m *Monitor) Configure(conf *Config) error {
	serviceURL := conf.ServiceURL
	if serviceURL == "" {
		serviceURL = fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%d/jmxrmi", conf.Host, conf.Port)
	}
	return m.Monitor.Configure(&java.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		JarFilePath:   filepath.Join(conf.BundleDir, "lib/jmx-monitor.jar"),
		CustomConfig: map[string]interface{}{
			"serviceURL":         serviceURL,
			"groovyScript":       conf.GroovyScript,
			"username":           conf.Username,
			"password":           conf.Password,
			"keyStorePath":       conf.KeyStorePath,
			"keyStorePassword":   conf.KeyStorePassword,
			"keyStoreType":       conf.KeyStoreType,
			"trustStorePath":     conf.TrustStorePath,
			"trustStorePassword": conf.TrustStorePassword,
			"jmxRemoteProfiles":  conf.JmxRemoteProfiles,
			"realm":              conf.Realm,
		},
	})
}
