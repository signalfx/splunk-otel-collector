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
	// Host will be filled in by auto-discovery if this monitor has a discovery rule.
	Host string `yaml:"host" json:"host,omitempty"`
	// Port will be filled in by auto-discovery if this monitor has a discovery rule.
	Port uint16 `yaml:"port" json:"port,omitempty"`
	// The service URL for the JMX RMI/JMXMP endpoint. If empty it will be filled in with values from `host` and `port`
	// using a standard JMX RMI template: `service:jmx:rmi:///jndi/rmi://<host>:<port>/jmxrmi`. If overridden, `host`
	// and `port` will have no effect. For JMXMP endpoint the service URL must be specified. The JMXMP endpoint URL
	// format is `service:jmx:jmxmp://<host>:<port>`.
	ServiceURL string `yaml:"serviceURL" json:"serviceURL"`
	// A literal Groovy script that generates datapoints from JMX MBeans. See the top-level `jmx` monitor doc for more
	// information on how to write this script. You can put the Groovy script in a separate file and refer to it here
	// with the [remote config reference](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#configure-the-smart-agent)
	// `{"#from": "/path/to/file.groovy", raw: true}`, or you can put it straight in YAML by using the `|` heredoc
	// syntax.
	GroovyScript string `yaml:"groovyScript" json:"groovyScript" validate:"required"`
	// Username for JMX authentication, if applicable.
	Username string `yaml:"username" json:"username"`
	// Password for JMX authentication, if applicable.
	Password string `yaml:"password" json:"password" neverLog:"true"`
	// The key store path is required if client authentication is enabled on the target JVM.
	KeyStorePath string `yaml:"keyStorePath" json:"keyStorePath"`
	// The key store file password if required.
	KeyStorePassword string `yaml:"keyStorePassword" json:"keyStorePassword" neverLog:"true"`
	// The key store type.
	KeyStoreType string `yaml:"keyStoreType" json:"keyStoreType" default:"jks"`
	// The trusted store path if the TLS profile is required.
	TrustStorePath string `yaml:"trustStorePath" json:"trustStorePath"`
	// The trust store file password if required.
	TrustStorePassword string `yaml:"trustStorePassword" json:"trustStorePassword" neverLog:"true"`
	// Supported JMX remote profiles are TLS in combination with SASL profiles: SASL/PLAIN, SASL/DIGEST-MD5 and
	// SASL/CRAM-MD5. Thus valid `jmxRemoteProfiles` values are: `SASL/PLAIN`, `SASL/DIGEST-MD5`, `SASL/CRAM-MD5`,
	//`TLS SASL/PLAIN`, `TLS SASL/DIGEST-MD5` and `TLS SASL/CRAM-MD5`.
	JmxRemoteProfiles string `yaml:"jmxRemoteProfiles" json:"jmxRemoteProfiles"`
	// The realm is required by profile SASL/DIGEST-MD5.
	Realm string `yaml:"realm" json:"realm"`
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
		JarFilePath:   filepath.Join(conf.BundleDir(), "lib/jmx-monitor.jar"),
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
