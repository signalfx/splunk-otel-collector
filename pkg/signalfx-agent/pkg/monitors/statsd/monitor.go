package statsd

import (
	"context"
	"errors"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var metricTypeMap = map[string]datapoint.MetricType{
	"g":  datapoint.Gauge,
	"c":  datapoint.Count,
	"ms": datapoint.Gauge,
	"s":  datapoint.Gauge,
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// ConverterInput is to receive configs to setup metric converters
type ConverterInput struct {
	// A pattern to match against StatsD metric names
	Pattern string `yaml:"pattern"`
	// A format to compose a metric name to report to SignalFx
	MetricName string `yaml:"metricName"`
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" singleInstance:"false"`
	// The host/address on which to bind the UDP listener that accepts statsd
	// datagrams
	ListenAddress string `yaml:"listenAddress" default:"localhost"`
	// The port on which to listen for statsd messages (**default:** `8125`)
	ListenPort *uint16 `yaml:"listenPort"`
	// A prefix in metric names that needs to be removed before metric name conversion
	MetricPrefix string `yaml:"metricPrefix"`
	// A list converters to convert StatsD metric names into SignalFx metric names and dimensions
	Converters []ConverterInput `yaml:"converters"`
}

// Validate StatsD monitor config
func (c *Config) Validate() error {
	for _, ci := range c.Converters {
		if ci.Pattern == "" {
			return errors.New("[pattern] is required for a converter")
		}
		if ci.MetricName == "" {
			return errors.New("[metricName] is required for a converter")
		}
	}

	return nil
}

// Monitor that listens to incoming statsd metrics
type Monitor struct {
	Output   types.Output
	cancel   context.CancelFunc
	conf     *Config
	listener *statsDListener
	logger   *utils.ThrottledLogger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(log.WithFields(log.Fields{"monitorType": monitorMetadata.MonitorType, "monitorID": conf.MonitorID}), 30*time.Second)

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// Give default value to ListenPort if not given by user.
	// Cannot use yaml default to take also 0 as a valid value.
	if conf.ListenPort == nil {
		conf.ListenPort = new(uint16)
		*conf.ListenPort = 8125
	}

	m.conf = conf

	var converters []*converter
	for _, ci := range conf.Converters {
		converter := newConverter(&ConverterInput{
			Pattern:    ci.Pattern,
			MetricName: ci.MetricName,
		}, m.logger)
		if converter != nil {
			converters = append(converters, converter)
		}
	}

	m.listener = &statsDListener{
		ipAddr:     conf.ListenAddress,
		port:       *conf.ListenPort,
		tcp:        false, // Will be added to Config when TCP is supported
		prefix:     conf.MetricPrefix,
		converters: converters,
		logger:     m.logger,
	}

	err := m.listener.Listen()
	if err != nil {
		return err
	}

	go m.listener.Read()

	utils.RunOnInterval(ctx, func() {
		metrics := m.listener.FetchMetrics()
		dps := m.convertMetricsToDatapoints(aggregateMetrics(metrics))

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops listening to incoming StatsD metrics
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.listener != nil {
		m.listener.Close()
	}
}

func aggregateMetrics(metrics []*statsDMetric) map[string]*statsDMetric {
	metricsMap := make(map[string]*statsDMetric)

	for _, metric := range metrics {
		if _, exists := metricsMap[metric.rawMetricName]; exists && metricTypeMap[metric.metricType] == datapoint.Count {
			// Add up
			metricsMap[metric.rawMetricName].value += metric.value
		} else {
			// Create a new one or drop older metric by overwriting
			metricsMap[metric.rawMetricName] = metric
		}
	}

	return metricsMap
}

func (m *Monitor) convertMetricsToDatapoints(metrics map[string]*statsDMetric) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	for _, metric := range metrics {
		var dp *datapoint.Datapoint

		// StatsD Metric Types https://github.com/statsd/statsd/blob/master/docs/metric_types.md
		switch metricTypeMap[metric.metricType] {
		case datapoint.Gauge:
			dp = sfxclient.GaugeF(metric.metricName, nil, metric.value)
		case datapoint.Count:
			dp = sfxclient.Counter(metric.metricName, nil, int64(metric.value))
		default:
			m.logger.Errorf("Unsupported StatsD metric type: %s", metric.metricType)
			continue
		}

		dp.Dimensions = metric.dimensions

		dps = append(dps, dp)
	}

	return dps
}
