package conviva

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/signalfx/signalfx-agent/pkg/core/common/dpmeta"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	metricURLFormat     = "https://api.conviva.com/insights/2.4/metrics.json?metrics=%s&account=%s&filter_ids=%s"
	metricLensURLFormat = metricURLFormat + "&metriclens_dimension_id=%d"
)

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline"`
	// Conviva Pulse username required with each API request.
	Username string `yaml:"pulseUsername" validate:"required"`
	// Conviva Pulse password required with each API request.
	Password       string `yaml:"pulsePassword" validate:"required" neverLog:"true"`
	TimeoutSeconds int    `yaml:"timeoutSeconds" default:"10"`
	// Conviva metrics to fetch. The default is quality_metriclens metric with the "All Traffic" filter applied and all quality_metriclens dimensions.
	MetricConfigs []*metricConfig `yaml:"metricConfigs"`
}

// Monitor for conviva metrics
// This monitor does not implement GetExtraMetrics() in order to get configured extra metrics to allow through because all metrics are included/allowed.
type Monitor struct {
	Output  types.FilteringOutput
	cancel  context.CancelFunc
	ctx     context.Context
	client  httpClient
	timeout time.Duration
	logger  logrus.FieldLogger
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{MetricConfigs: []*metricConfig{{MetricParameter: "quality_metriclens"}}})
}

// Configure monitor
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	m.timeout = time.Duration(conf.TimeoutSeconds) * time.Second
	m.client = newConvivaClient(&http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}, conf.Username, conf.Password)
	m.ctx, m.cancel = context.WithCancel(context.Background())
	semaphore := make(chan struct{}, maxGoroutinesPerInterval(conf.MetricConfigs))
	interval := time.Duration(conf.IntervalSeconds) * time.Second
	service := newAccountsService(m.ctx, &m.timeout, m.client)
	utils.RunOnInterval(m.ctx, func() {
		for _, metricConf := range conf.MetricConfigs {
			if err := metricConf.init(service); err != nil {
				m.logger.WithError(err).Error("Could not initialize metric configuration")
			}
			if strings.Contains(metricConf.MetricParameter, "metriclens") {
				m.fetchMetricLensMetrics(interval, semaphore, metricConf)
			} else {
				m.fetchMetrics(interval, semaphore, metricConf)
			}
		}
	}, interval)
	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Monitor) fetchMetrics(contextTimeout time.Duration, semaphore chan struct{}, metricConf *metricConfig) {
	semaphore <- struct{}{}
	go func(contextTimeout time.Duration, m *Monitor, metricConf *metricConfig) {
		defer func() { <-semaphore }()
		ctx, cancel := context.WithTimeout(m.ctx, contextTimeout)
		defer cancel()
		numFiltersPerRequest := len(metricConf.filterIDs())
		if metricConf.MaxFiltersPerRequest != 0 {
			numFiltersPerRequest = metricConf.MaxFiltersPerRequest
		}
		var urls []*string
		var low, high int
		numFilters := len(metricConf.filterIDs())
		for i := 1; high < numFilters; i++ {
			if low, high = (i-1)*numFiltersPerRequest, i*numFiltersPerRequest; high > numFilters {
				high = numFilters
			}
			url := fmt.Sprintf(metricURLFormat, metricConf.MetricParameter, metricConf.accountID, strings.Join(metricConf.filterIDs()[low:high], ","))
			urls = append(urls, &url)
		}
		responses := make([]*map[string]metricResponse, len(urls))
		var g errgroup.Group
		for i := range urls {
			idx := i
			g.Go(func() error {
				res := map[string]metricResponse{}
				if _, err := m.client.get(ctx, &res, *urls[idx]); err != nil {
					return fmt.Errorf("GET metric %s failed. %+v", metricConf.MetricParameter, err)
				}
				responses[idx] = &res
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			m.logger.Error(err)
			return
		}
		var dps []*datapoint.Datapoint
		timestamp := time.Now()
		for _, res := range responses {
			for metricParameter, series := range *res {
				metricName := "conviva." + metricParameter
				for filterID, metricValues := range series.FilterIDValuesMap {
					switch series.Type {
					case "time_series":
						if dp := latestTimeSeriesDatapoint(metricName, metricValues, series.Timestamps, metricConf.Account, metricConf.filterName(filterID)); dp != nil {
							dps = append(dps, dp)
						}
					case "label_series":
						if lsdps := labelSeriesDatapoints(metricName, metricValues, series.Xvalues, timestamp, metricConf.Account, metricConf.filterName(filterID)); lsdps != nil {
							dps = append(dps, *lsdps...)
						}
					default:
						if ssdps := simpleSeriesDatapoints(metricName, metricValues, timestamp, metricConf.Account, metricConf.filterName(filterID)); ssdps != nil {
							dps = append(dps, *ssdps...)
						}
					}
				}
			}
		}
		m.Output.SendDatapoints(dps...)
	}(contextTimeout, m, metricConf)
}

func (m *Monitor) fetchMetricLensMetrics(contextTimeout time.Duration, semaphore chan struct{}, metricConf *metricConfig) {
	for _, dim := range metricConf.MetricLensDimensions {
		semaphore <- struct{}{}
		dim = strings.TrimSpace(dim)
		dimID := metricConf.metricLensDimensionMap[dim]
		if dimID == 0.0 {
			m.logger.Errorf("No id for MetricLens dimension %s. Wrong MetricLens dimension name.", dim)
			continue
		}
		go func(contextTimeout time.Duration, m *Monitor, metricConf *metricConfig, metricLensDimension string) {
			defer func() { <-semaphore }()
			ctx, cancel := context.WithTimeout(m.ctx, contextTimeout)
			defer cancel()
			numFiltersPerRequest := len(metricConf.filterIDs())
			if metricConf.MaxFiltersPerRequest != 0 {
				numFiltersPerRequest = metricConf.MaxFiltersPerRequest
			}
			var urls []*string
			var low, high int
			numFilters := len(metricConf.filterIDs())
			for i := 1; high < numFilters; i++ {
				if low, high = (i-1)*numFiltersPerRequest, i*numFiltersPerRequest; high > numFilters {
					high = numFilters
				}
				url := fmt.Sprintf(metricLensURLFormat, metricConf.MetricParameter, metricConf.accountID, strings.Join(metricConf.filterIDs()[low:high], ","), int(dimID))
				urls = append(urls, &url)
			}
			responses := make([]*map[string]metricResponse, len(urls))
			var g errgroup.Group
			for i := range urls {
				idx := i
				g.Go(func() error {
					var res map[string]metricResponse
					if _, err := m.client.get(ctx, &res, *urls[idx]); err != nil {
						return fmt.Errorf("GET metric %s failed. %+v", metricConf.MetricParameter, err)
					}
					responses[idx] = &res
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				m.logger.Error(err)
				return
			}
			var dps []*datapoint.Datapoint
			timestamp := time.Now()
			lensMetrics := metricLensMetrics()
			for _, res := range responses {
				for metricParameter, metricTable := range *res {
					for filterID, tableValue := range metricTable.Tables {
						if tdps := tableDatapoints(lensMetrics[metricParameter], metricLensDimension, tableValue.Rows, metricTable.Xvalues, timestamp, metricConf.Account, metricConf.filterName(filterID)); tdps != nil {
							dps = append(dps, *tdps...)
						}
					}
				}
			}
			m.Output.SendDatapoints(dps...)
		}(contextTimeout, m, metricConf, dim)
	}
}

func maxGoroutinesPerInterval(metricConfigs []*metricConfig) int {
	requests := 0
	for _, metricConfig := range metricConfigs {
		if metricLensDimensionsLength := len(metricConfig.MetricLensDimensions); metricLensDimensionsLength != 0 {
			requests += len(metricConfig.Filters) * metricLensDimensionsLength
		} else {
			requests += len(metricConfig.Filters)
		}
	}
	return int(math.Max(float64(requests), float64(2000)))
}

func latestTimeSeriesDatapoint(metricName string, metricValues []float64, timestamps []int64, accountName string, filterName string) (dp *datapoint.Datapoint) {
	if len(metricValues) > 0 {
		dp = sfxclient.GaugeF(metricName, map[string]string{"account": accountName, "filter": filterName}, metricValues[len(metricValues)-1])
		// Series timestamps are in milliseconds
		dp.Timestamp = time.Unix(timestamps[len(timestamps)-1]/1000, 0)
		dp.Meta[dpmeta.NotHostSpecificMeta] = true
	}
	return
}

func labelSeriesDatapoints(metricName string, metricValues []float64, xvalues []string, timestamp time.Time, accountName string, filterName string) (dps *[]*datapoint.Datapoint) {
	if len(metricValues) > 0 {
		dps = &[]*datapoint.Datapoint{}
		for i, metricValue := range metricValues {
			dp := sfxclient.GaugeF(metricName, map[string]string{"account": accountName, "filter": filterName}, metricValue)
			dp.Dimensions["label"] = xvalues[i]
			dp.Meta[dpmeta.NotHostSpecificMeta] = true
			dp.Timestamp = timestamp
			*dps = append(*dps, dp)
		}
	}
	return
}

func simpleSeriesDatapoints(metricName string, metricValues []float64, timestamp time.Time, accountName string, filterName string) (dps *[]*datapoint.Datapoint) {
	if len(metricValues) > 0 {
		dps = &[]*datapoint.Datapoint{}
		for _, metricValue := range metricValues {
			dp := sfxclient.GaugeF(metricName, map[string]string{"account": accountName, "filter": filterName}, metricValue)
			dp.Meta[dpmeta.NotHostSpecificMeta] = true
			dp.Timestamp = timestamp
			*dps = append(*dps, dp)
		}
	}
	return
}

func tableDatapoints(metricNames []string, dimension string, rows [][]float64, xvalues []string, timestamp time.Time, accountName string, filterName string) (dps *[]*datapoint.Datapoint) {
	//dps := make([]*datapoint.Datapoint, 0)
	if len(rows) > 0 {
		dps = &[]*datapoint.Datapoint{}
		for rowIndex, row := range rows {
			for metricIndex, metricValue := range row {
				dp := sfxclient.GaugeF(metricNames[metricIndex], map[string]string{"account": accountName, "filter": filterName, dimension: xvalues[rowIndex]}, metricValue)
				dp.Timestamp = timestamp
				dp.Meta[dpmeta.NotHostSpecificMeta] = true
				*dps = append(*dps, dp)
			}
		}
	}
	return
}
