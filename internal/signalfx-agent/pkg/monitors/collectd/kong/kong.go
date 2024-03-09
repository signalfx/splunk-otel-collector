package kong

import (
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
)

// metricConfigMap is a map of metric names to the metric configuration name
// the Python plugin uses.
var metricConfigMap = map[string]string{
	counterKongConnectionsAccepted: "connections_accepted",
	counterKongConnectionsHandled:  "connections_handled",
	counterKongKongLatency:         "kong_latency",
	counterKongRequestsCount:       "total_requests",
	counterKongRequestsLatency:     "request_latency",
	counterKongRequestsSize:        "request_size",
	counterKongResponsesCount:      "response_count",
	counterKongResponsesSize:       "response_size",
	counterKongUpstreamLatency:     "upstream_latency",
	gaugeKongConnectionsActive:     "connections_active",
	gaugeKongConnectionsReading:    "connections_reading",
	gaugeKongConnectionsWaiting:    "connections_waiting",
	gaugeKongConnectionsWriting:    "connections_writing",
	gaugeKongDatabaseReachable:     "database_reachable",
}

// configMetricMap is the reverse of the above map.
var configMetricMap = map[string]string{}
var log = logrus.WithField("monitorType", monitorType)

func init() {
	if len(metricConfigMap) != len(metricSet) {
		panic("kong metricConfigMap is missing entries")
	}

	for key, val := range metricConfigMap {
		configMetricMap[val] = key
	}

	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			python.PyMonitor{
				MonitorCore: subproc.New(),
			},
		}
	}, &Config{})
}

// Header defines name/value pair for AuthHeader option
type Header struct {
	// Name of header to include with GET
	HeaderName string `yaml:"header" validate:"required"`
	// Value of header
	Value string `yaml:"value" validate:"required"`
}

// Metric is for use with `Metric "metric_name" bool` collectd plugin format
type Metric struct {
	// Name of metric, per collectd-kong
	MetricName string `yaml:"metric" validate:"required"`
	// Whether to report this metric
	ReportBool bool `yaml:"report" validate:"required"`
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	Verbose                *bool `yaml:"verbose"`
	ReportStatusCodes      *bool `yaml:"reportStatusCodes"`
	pyConf                 *python.Config
	ReportStatusCodeGroups *bool   `yaml:"reportStatusCodeGroups"`
	ReportHTTPMethods      *bool   `yaml:"reportHttpMethods"`
	ReportRouteIDs         *bool   `yaml:"reportRouteIds"`
	ReportServiceNames     *bool   `yaml:"reportServiceNames"`
	AuthHeader             *Header `yaml:"authHeader"`
	ReportServiceIDs       *bool   `yaml:"reportServiceIds"`
	ReportAPINames         *bool   `yaml:"reportApiNames"`
	ReportAPIIDs           *bool   `yaml:"reportApiIds"`
	config.MonitorConfig   `yaml:",inline" acceptsEndpoints:"true"`
	CABundle               string `yaml:"caBundle"`
	Host                   string `yaml:"host" validate:"required"`
	ClientCertKey          string `yaml:"clientCertKey"`
	ClientCert             string `yaml:"clientCert"`
	python.CommonConfig    `yaml:",inline"`
	URL                    string   `yaml:"url" default:"http://{{.Host}}:{{.Port}}/signalfx"`
	Name                   string   `yaml:"name"`
	APIIDsBlacklist        []string `yaml:"apiIdsBlacklist"`
	ServiceNames           []string `yaml:"serviceNames"`
	StatusCodesBlacklist   []string `yaml:"statusCodesBlacklist"`
	APIIDs                 []string `yaml:"apiIds"`
	Metrics                []Metric `yaml:"metrics"`
	APINames               []string `yaml:"apiNames"`
	APINamesBlacklist      []string `yaml:"apiNamesBlacklist"`
	ServiceIDs             []string `yaml:"serviceIds"`
	ServiceIDsBlacklist    []string `yaml:"serviceIdsBlacklist"`
	StatusCodes            []string `yaml:"statusCodes"`
	ServiceNamesBlacklist  []string `yaml:"serviceNamesBlacklist"`
	RouteIDs               []string `yaml:"routeIds"`
	RouteIDsBlacklist      []string `yaml:"routeIdsBlacklist"`
	HTTPMethods            []string `yaml:"httpMethods"`
	HTTPMethodsBlacklist   []string `yaml:"httpMethodsBlacklist"`
	Port                   uint16   `yaml:"port" validate:"required"`
	VerifyCerts            bool     `yaml:"verifyCerts"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// GetExtraMetrics returns additional metrics to allow through based on monitor metric configuration
func (c *Config) GetExtraMetrics() []string {
	// if configured in Metrics then add to extra metrics
	var extraMetrics []string

	for _, metricConfig := range c.Metrics {
		if metricConfig.ReportBool {
			if metric, ok := configMetricMap[metricConfig.MetricName]; ok {
				extraMetrics = append(extraMetrics, metric)
			} else {
				log.Warnf("Unknown metric name %s", metric)
			}
		}
	}

	return extraMetrics
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(c *Config) error {
	conf := *c
	conf.Metrics = append([]Metric(nil), c.Metrics...)

	// Track if the user has already configured the metric.
	configuredMetrics := map[string]bool{}

	for _, metricConfig := range c.Metrics {
		configuredMetrics[metricConfig.MetricName] = true
	}

	// For enabled metrics configure the Python metric name unless the user
	// already has.
	for _, metric := range m.Output.EnabledMetrics() {
		metricConfig := metricConfigMap[metric]
		if metricConfig == "" {
			m.Logger().Warnf("Unable to determine metric configuration name for %s", metric)
			continue
		}

		if !configuredMetrics[metricConfig] {
			conf.Metrics = append(conf.Metrics, Metric{metricConfig, true})
		}
	}

	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "kong_plugin",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "kong")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.BundleDir)},
		PluginConfig: map[string]interface{}{
			"URL":                    conf.URL,
			"Interval":               conf.IntervalSeconds,
			"Verbose":                conf.Verbose,
			"Name":                   conf.Name,
			"VerifyCerts":            conf.VerifyCerts,
			"CABundle":               conf.CABundle,
			"ClientCert":             conf.ClientCert,
			"ClientCertKey":          conf.ClientCertKey,
			"ReportAPIIDs":           conf.ReportAPIIDs,
			"ReportAPINames":         conf.ReportAPINames,
			"ReportServiceIDs":       conf.ReportServiceIDs,
			"ReportServiceNames":     conf.ReportServiceNames,
			"ReportRouteIDs":         conf.ReportRouteIDs,
			"ReportHTTPMethods":      conf.ReportHTTPMethods,
			"ReportStatusCodeGroups": conf.ReportStatusCodeGroups,
			"ReportStatusCodes":      conf.ReportStatusCodes,
			"APIIDs": map[string]interface{}{
				"#flatten": true,
				"values":   conf.APIIDs,
			},
			"APIIDsBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.APIIDsBlacklist,
			},
			"APINames": map[string]interface{}{
				"#flatten": true,
				"values":   conf.APINames,
			},
			"APINamesBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.APINamesBlacklist,
			},
			"ServiceIDs": map[string]interface{}{
				"#flatten": true,
				"values":   conf.ServiceIDs,
			},
			"ServiceIDsBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.ServiceIDsBlacklist,
			},
			"ServiceNames": map[string]interface{}{
				"#flatten": true,
				"values":   conf.ServiceNames,
			},
			"ServiceNamesBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.ServiceNamesBlacklist,
			},
			"RouteIDs": map[string]interface{}{
				"#flatten": true,
				"values":   conf.RouteIDs,
			},
			"RouteIDsBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.RouteIDsBlacklist,
			},
			"HTTPMethods": map[string]interface{}{
				"#flatten": true,
				"values":   conf.HTTPMethods,
			},
			"HTTPMethodsBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.HTTPMethodsBlacklist,
			},
			"StatusCodes": map[string]interface{}{
				"#flatten": true,
				"values":   conf.StatusCodes,
			},
			"StatusCodesBlacklist": map[string]interface{}{
				"#flatten": true,
				"values":   conf.StatusCodesBlacklist,
			},
		},
	}

	if conf.AuthHeader != nil {
		conf.pyConf.PluginConfig["AuthHeader"] = []string{conf.AuthHeader.HeaderName, conf.AuthHeader.Value}
	}

	if len(conf.Metrics) > 0 {
		values := make([][]interface{}, 0, len(conf.Metrics))
		for _, m := range conf.Metrics {
			values = append(values, []interface{}{m.MetricName, m.ReportBool})
		}
		conf.pyConf.PluginConfig["Metric"] = map[string]interface{}{
			"#flatten": true,
			"values":   values,
		}
	}

	return m.PyMonitor.Configure(&conf)
}
