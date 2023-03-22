package expvar

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	memstatsPauseNsMetricPath       = "memstats.PauseNs.\\.*"
	memstatsPauseEndMetricPath      = "memstats.PauseEnd.\\.*"
	memstatsNumGCMetricPath         = "memstats.NumGC"
	memstatsBySizeSizeMetricPath    = "memstats.BySize.\\.*.Size"
	memstatsBySizeMallocsMetricPath = "memstats.BySize.\\.*.Mallocs"
	memstatsBySizeFreesMetricPath   = "memstats.BySize.\\.*.Frees"
	memstatsBySizeDimensionPath     = "memstats_by_size_index"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Monitor for expvar metrics
type Monitor struct {
	Output types.FilteringOutput
	cancel context.CancelFunc
	ctx    context.Context
	client *http.Client
	logger log.FieldLogger
}

type metricVal struct {
	metric              string
	keys                []string
	value               datapoint.Value
	arrayIndexDimension map[string]string
}

// Configure monitor
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	if m.Output.HasAnyExtraMetrics() {
		conf.EnhancedMetrics = true
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 300 * time.Millisecond,
	}

	url := conf.getURL()
	runInterval := time.Duration(conf.IntervalSeconds) * time.Second
	allMetricConfigs := conf.getAllMetricConfigs()

	utils.RunOnInterval(m.ctx, func() {
		obj, err := m.fetchObj(url)
		if err != nil {
			m.logger.WithError(err).Error("Getting expvar json object failed")
			return
		}
		applicationName, err := getApplicationName(obj)
		if err != nil {
			m.logger.Warn(err)
		}
		valsMap := make(map[string][]*metricVal)
		mostRecentGCPauseIndex := int64(-1)
		// Getting the most recent GC pause index using logic from https://golang.org/pkg/runtime/ in the PauseNs section of the 'type MemStats' section
		for _, mConf := range allMetricConfigs {
			vals := m.getValuesInPath(obj, mConf.JSONPath, mConf.PathSeparator)
			valsMap[mConf.JSONPath] = vals
			if mConf.JSONPath == memstatsNumGCMetricPath {
				if len(vals) > 0 && vals[0].value != nil {
					if numGC, err := strconv.ParseInt(vals[0].value.String(), 10, 0); err == nil {
						mostRecentGCPauseIndex = (numGC + 255) % 256
					}
				}
			}
		}
		// Using the most recent GC pause index to get the most recent gc pause values in arrays PauseNs and PauseEnd and discarding the rest.
		for k, vals := range valsMap {
			if k == memstatsPauseNsMetricPath || k == memstatsPauseEndMetricPath {
				newVals := make([]*metricVal, 0)
				for _, val := range vals {
					indexStr := val.arrayIndexDimension[joinWords(append(snakeCaseSlice(val.keys), "array_index"), "_")]
					index, err := strconv.ParseInt(indexStr, 10, 0)
					if err != nil {
						m.logger.Errorf("Error while getting the most recent GC pause. %+v", err)
						continue
					}
					if index == mostRecentGCPauseIndex {
						if val.metric = memstatsMostRecentGcPauseNs; k == memstatsPauseEndMetricPath {
							val.metric = memstatsMostRecentGcPauseEnd
						}
						newVals = append(newVals, val)
					}
				}
				valsMap[k] = newVals
			}
		}
		dps := make([]*datapoint.Datapoint, 0)
		now := time.Now()
		for _, mConf := range allMetricConfigs {
			for _, metricVal := range valsMap[mConf.JSONPath] {
				dp := m.newDp(metricVal, mConf.Name, mConf.JSONPath, mConf.metricType(), now)
				if applicationName != "" {
					dp.Dimensions["application_name"] = applicationName
				}
				for _, dConf := range mConf.DimensionConfigs {
					if strings.TrimSpace(dConf.Name) != "" && strings.TrimSpace(dConf.Value) != "" {
						dp.Dimensions[dConf.Name] = dConf.Value
					}
				}
				dps = append(dps, dp)
			}
		}
		m.Output.SendDatapoints(dps...)
	}, runInterval)
	return nil
}

func (m *Monitor) fetchObj(url url.URL) (map[string]interface{}, error) {
	resp, err := m.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsonObj := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		return nil, err
	}
	return jsonObj, nil
}

// getValuesInPath gets values in path "path" of "obj"
func (m *Monitor) getValuesInPath(obj map[string]interface{}, path string, pathSeparator string) []*metricVal {
	// Splitting configured path into components. The path and thus its components contain regular expressions
	pathComps, err := utils.SplitString(path, []rune(pathSeparator)[0], escape)
	if err != nil {
		m.logger.Error(err)
		return nil
	}
	return m.getValuesInPathHelper(obj, pathComps)
}

// getValuesInPathHelper gets values in "obj" by traversing "obj" depth-wise using path components "pathComps" as keys.
func (m *Monitor) getValuesInPathHelper(obj interface{}, pathComps []string) (values []*metricVal) {
	switch value := obj.(type) {
	case map[string]interface{}:
		if len(pathComps) == 0 {
			m.logger.Debugf("no key(s) for embedded object %+v", value)
			return
		}
		pathComp, err := regexp.Compile(pathComps[0])
		if err != nil {
			m.logger.Error(err)
			return
		}
		for key := range value {
			if !pathComp.MatchString(key) {
				continue
			}
			var nextPathComps []string
			if len(pathComps) > 1 {
				nextPathComps = pathComps[1:]
			}
			newValues := m.getValuesInPathHelper(value[key], nextPathComps)
			for _, newValue := range newValues {
				newValue.keys = append([]string{key}, newValue.keys...)
				dims := make(map[string]string)
				for k, v := range newValue.arrayIndexDimension {
					dims[joinWords(append(snakeCaseSlice([]string{key}), k), "_")] = v
				}
				newValue.arrayIndexDimension = dims
			}
			values = append(values, newValues...)
		}
		return values
	case []interface{}:
		pathComp, err := regexp.Compile(pathComps[0])
		if err != nil {
			m.logger.Error(err)
			return
		}
		for i := range value {
			index := strconv.Itoa(i)
			if !pathComp.MatchString(index) {
				continue
			}
			var nextPathComps []string
			if len(pathComps) > 1 {
				nextPathComps = pathComps[1:]
			}
			newValues := m.getValuesInPathHelper(value[i], nextPathComps)
			for _, newValue := range newValues {
				newValue.arrayIndexDimension = make(map[string]string)
				newValue.arrayIndexDimension["array_index"] = index
			}
			values = append(values, newValues...)
		}
		return values
	default:
		metricValue, err := datapoint.CastMetricValueWithBool(value)
		if err != nil {
			m.logger.Error(err)
		}
		return []*metricVal{{value: metricValue}}
	}
}

func (m *Monitor) newDp(val *metricVal, name string, path string, metricType datapoint.MetricType, now time.Time) *datapoint.Datapoint {
	dp := datapoint.Datapoint{
		Metric:     strings.TrimSpace(val.metric),
		MetricType: metricType,
		Value:      val.value,
		Timestamp:  now,
	}
	if dp.Metric == "" {
		if strings.TrimSpace(name) == "" {
			dp.Metric = joinWords(snakeCaseSlice(val.keys), ".")
		} else {
			dp.Metric = name
		}
	}
	dp.Dimensions = make(map[string]string)
	for k, v := range val.arrayIndexDimension {
		dp.Dimensions[k] = v
	}
	// Renaming auto created dimension 'memstats.BySize' that stores array index to 'class'
	if path == memstatsBySizeSizeMetricPath || path == memstatsBySizeMallocsMetricPath || path == memstatsBySizeFreesMetricPath {
		dp.Dimensions["class"] = dp.Dimensions[memstatsBySizeDimensionPath]
		delete(dp.Dimensions, memstatsBySizeDimensionPath)
	}
	return &dp
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
