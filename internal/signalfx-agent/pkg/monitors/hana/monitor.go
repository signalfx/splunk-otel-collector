package hana

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/SAP/go-hdb/driver"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/sql"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for the SAP Hana monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	ConnectionString     string
	Host                 string
	Username             string
	Password             string

	// ServerName to verify the hostname. Defaults to Host if not specified.
	TLSServerName string `yaml:"tlsServerName"`
	// Controls whether a client verifies the server's certificate chain and host name. Defaults to false.
	TLSInsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// Path to root certificate(s) (optional)
	TLSRootCAFiles []string `yaml:"rootCAFiles"`

	Port                int
	LogQueries          bool
	MaxExpensiveQueries int
}

// Monitor that collects SAP Hana stats
type Monitor struct {
	Output     types.FilteringOutput
	sqlMonitor *sql.Monitor
}

// Configure the monitor and kick off metric collection
func (m *Monitor) Configure(conf *Config) error {
	var err error
	m.sqlMonitor, err = configureSQLMonitor(
		m.Output.Copy(),
		conf.MonitorConfig,
		cfgToConnString(conf),
		conf.LogQueries,
		conf.MaxExpensiveQueries,
	)
	if err != nil {
		return fmt.Errorf("could not configure Hana SQL monitor: %v", err)
	}
	return nil
}

func cfgToConnString(c *Config) string {
	if c.ConnectionString != "" {
		return c.ConnectionString
	}
	host := c.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 443
	}
	tlsServerName := c.TLSServerName
	if tlsServerName == "" {
		tlsServerName = host
	}
	return createDSN(host, c.Username, c.Password, tlsServerName, port, c.TLSInsecureSkipVerify, c.TLSRootCAFiles)
}

func createDSN(host, username, password, tlsServerName string, port int, tlsInsecureSkipVerify bool, rootCAFiles []string) string {
	query := url.Values{}
	query.Add(driver.DSNTLSServerName, tlsServerName)
	query.Add(driver.DSNTLSInsecureSkipVerify, strconv.FormatBool(tlsInsecureSkipVerify))
	for _, f := range rootCAFiles {
		query.Add(driver.DSNTLSRootCAFile, f)
	}
	u := &url.URL{
		Scheme:   "hdb",
		User:     url.UserPassword(username, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		RawQuery: query.Encode(),
	}
	return u.String()
}

func configureSQLMonitor(output types.Output, monCfg config.MonitorConfig, connStr string, logQueries bool, maxExpensiveQueries int) (*sql.Monitor, error) {
	sqlMon := &sql.Monitor{Output: output}
	return sqlMon, sqlMon.Configure(&sql.Config{
		MonitorConfig:    monCfg,
		ConnectionString: connStr,
		DBDriver:         "hdb",
		Queries:          queries(maxExpensiveQueries),
		LogQueries:       logQueries,
	})
}

func (m *Monitor) Shutdown() {
	m.sqlMonitor.Shutdown()
}
