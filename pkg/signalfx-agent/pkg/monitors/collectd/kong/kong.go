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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	// Kong host to connect with (used for autodiscovery and URL)
	Host string `yaml:"host" validate:"required"`
	// Port for kong-plugin-signalfx hosting server (used for autodiscovery and URL)
	Port uint16 `yaml:"port" validate:"required"`
	// Registration name when using multiple instances in Smart Agent
	Name string `yaml:"name"`
	// kong-plugin-signalfx metric plugin
	URL string `yaml:"url" default:"http://{{.Host}}:{{.Port}}/signalfx"`
	// Header and its value to use for requests to SFx metric endpoint
	AuthHeader *Header `yaml:"authHeader"`
	// Whether to verify certificates when using ssl/tls
	VerifyCerts bool `yaml:"verifyCerts"`
	// CA Bundle file or directory
	CABundle string `yaml:"caBundle"`
	// Client certificate file (with or without included key)
	ClientCert string `yaml:"clientCert"`
	// Client cert key if not bundled with clientCert
	ClientCertKey string `yaml:"clientCertKey"`
	// Whether to use debug logging for collectd-kong
	Verbose *bool `yaml:"verbose"`
	// List of metric names and report flags. See monitor description for more
	// details.
	Metrics []Metric `yaml:"metrics"`
	// Report metrics for distinct API IDs where applicable
	ReportAPIIDs *bool `yaml:"reportApiIds"`
	// Report metrics for distinct API names where applicable
	ReportAPINames *bool `yaml:"reportApiNames"`
	// Report metrics for distinct Service IDs where applicable
	ReportServiceIDs *bool `yaml:"reportServiceIds"`
	// Report metrics for distinct Service names where applicable
	ReportServiceNames *bool `yaml:"reportServiceNames"`
	// Report metrics for distinct Route IDs where applicable
	ReportRouteIDs *bool `yaml:"reportRouteIds"`
	// Report metrics for distinct HTTP methods where applicable
	ReportHTTPMethods *bool `yaml:"reportHttpMethods"`
	// Report metrics for distinct HTTP status code groups (eg. "5xx") where applicable
	ReportStatusCodeGroups *bool `yaml:"reportStatusCodeGroups"`
	// Report metrics for distinct HTTP status codes where applicable (mutually exclusive with ReportStatusCodeGroups)
	ReportStatusCodes *bool `yaml:"reportStatusCodes"`

	// List of API ID patterns to report distinct metrics for, if reportApiIds is false
	APIIDs []string `yaml:"apiIds"`
	// List of API ID patterns to not report distinct metrics for, if reportApiIds is true or apiIds are specified
	APIIDsBlacklist []string `yaml:"apiIdsBlacklist"`
	// List of API name patterns to report distinct metrics for, if reportApiNames is false
	APINames []string `yaml:"apiNames"`
	// List of API name patterns to not report distinct metrics for, if reportApiNames is true or apiNames are specified
	APINamesBlacklist []string `yaml:"apiNamesBlacklist"`
	// List of Service ID patterns to report distinct metrics for, if reportServiceIds is false
	ServiceIDs []string `yaml:"serviceIds"`
	// List of Service ID patterns to not report distinct metrics for, if reportServiceIds is true or serviceIds are specified
	ServiceIDsBlacklist []string `yaml:"serviceIdsBlacklist"`
	// List of Service name patterns to report distinct metrics for, if reportServiceNames is false
	ServiceNames []string `yaml:"serviceNames"`
	// List of Service name patterns to not report distinct metrics for, if reportServiceNames is true or serviceNames are specified
	ServiceNamesBlacklist []string `yaml:"serviceNamesBlacklist"`
	// List of Route ID patterns to report distinct metrics for, if reportRouteIds is false
	RouteIDs []string `yaml:"routeIds"`
	// List of Route ID patterns to not report distinct metrics for, if reportRouteIds is true or routeIds are specified
	RouteIDsBlacklist []string `yaml:"routeIdsBlacklist"`
	// List of HTTP method patterns to report distinct metrics for, if reportHttpMethods is false
	HTTPMethods []string `yaml:"httpMethods"`
	// List of HTTP method patterns to not report distinct metrics for, if reportHttpMethods is true or httpMethods are specified
	HTTPMethodsBlacklist []string `yaml:"httpMethodsBlacklist"`
	// List of HTTP status code patterns to report distinct metrics for, if reportStatusCodes is false
	StatusCodes []string `yaml:"statusCodes"`
	// List of HTTP status code patterns to not report distinct metrics for, if reportStatusCodes is true or statusCodes are specified
	StatusCodesBlacklist []string `yaml:"statusCodesBlacklist"`
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
		ModulePaths:   []string{collectd.MakePythonPluginPath("kong")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
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
