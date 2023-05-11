package logstash

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	nodePath = "/_node/"
)

var prefixPathMap = map[string]string{
	"node.os":            fmt.Sprintf("%s%s", nodePath, "os"),
	"node.jvm":           fmt.Sprintf("%s%s", nodePath, "jvm"),
	"node.hot_threads":   fmt.Sprintf("%s%s", nodePath, "hot_threads"),
	"node.stats.jvm":     fmt.Sprintf("%s%s", nodePath, "stats/jvm"),
	"node.stats.process": fmt.Sprintf("%s%s", nodePath, "stats/process"),
	"node.stats.events":  fmt.Sprintf("%s%s", nodePath, "stats/events"),
	"node.stats.reloads": fmt.Sprintf("%s%s", nodePath, "stats/reloads"),
	"node.stats.os":      fmt.Sprintf("%s%s", nodePath, "stats/os"),
}
var pluginPath = fmt.Sprintf("%s%s", nodePath, "plugins")
var pipelinePath = fmt.Sprintf("%s%s", nodePath, "pipelines")
var pipelineStatPath = fmt.Sprintf("%s%s", nodePath, "stats/pipelines")

var dimensionKeyMap = map[string]string{
	"threads": "thread",
	"outputs": "output",
	"inputs":  "input",
	"codecs":  "codec",
	"filters": "filter",
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Monitor that accepts and forwards trace spans
type Monitor struct {
	Output        types.Output
	conf          *Config
	ctx           context.Context
	cancel        context.CancelFunc
	metricTypeMap map[string]datapoint.MetricType
	logger        *utils.ThrottledLogger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID}), 30*time.Second)
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.conf = conf
	m.metricTypeMap = conf.getMetricTypeMap()

	client := &http.Client{
		Timeout: time.Duration(conf.TimeoutSeconds) * time.Second,
	}

	scheme := "http"
	if conf.UseHTTPS {
		scheme = "https"
	}

	dims, err := m.fetchNodeInfo(client, fmt.Sprintf("%s://%s:%d%s", scheme, m.conf.Host, m.conf.Port, nodePath))
	if err != nil {
		m.logger.WithError(err).Error("Couldn't get node info.")
	}

	utils.RunOnInterval(m.ctx, func() {
		var dps []*datapoint.Datapoint
		var fetched []*datapoint.Datapoint

		for prefix, path := range prefixPathMap {
			fetched, err = m.fetchMetrics(client, fmt.Sprintf("%s://%s:%d%s", scheme, m.conf.Host, m.conf.Port, path), prefix, dims)
			if err != nil {
				m.logger.WithError(err).Errorf("Couldn't fetch metrics for path %s", path)
				continue
			}
			dps = append(dps, fetched...)
		}

		if fetched, err = m.fetchPipelineMetrics(client, fmt.Sprintf("%s://%s:%d%s", scheme, m.conf.Host, m.conf.Port, pipelinePath), "node.pipelines", dims); err == nil {
			dps = append(dps, fetched...)
		} else {
			m.logger.WithError(err).Error("Couldn't fetch metrics for pipelines")
		}

		if fetched, err = m.fetchPipelineMetrics(client, fmt.Sprintf("%s://%s:%d%s", scheme, m.conf.Host, m.conf.Port, pipelineStatPath), "node.stats.pipelines", dims); err == nil {
			dps = append(dps, fetched...)
		} else {
			m.logger.WithError(err).Error("Couldn't fetch metrics for pipeline stats")
		}

		if fetched, err = m.fetchPluginMetrics(client, fmt.Sprintf("%s://%s:%d%s", scheme, m.conf.Host, m.conf.Port, pluginPath), "node.plugins", dims); err == nil {
			dps = append(dps, fetched...)
		} else {
			m.logger.WithError(err).Error("Couldn't fetch metrics for plugins")
		}

		now := time.Now()
		for _, dp := range dps {
			dp.Timestamp = now
		}
		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

func (m *Monitor) fetchNodeInfo(client *http.Client, endpoint string) (map[string]string, error) {
	dims := make(map[string]string)

	nodeJSON, err := getJSON(client, endpoint)

	if err != nil {
		return nil, err
	}

	if nodeID, exists := nodeJSON["id"]; exists {
		dims["node_id"], _ = nodeID.(string)
	}
	if nodeName, exists := nodeJSON["name"]; exists {
		dims["node_name"], _ = nodeName.(string)
	}

	return dims, nil
}

func (m *Monitor) fetchMetrics(client *http.Client, endpoint string, prefix string, dimensions map[string]string) ([]*datapoint.Datapoint, error) {
	metricsJSON, err := getJSON(client, endpoint)
	if err != nil {
		return nil, err
	}

	return m.extractDatapoints(prefix, metricsJSON, dimensions), nil
}

func (m *Monitor) fetchPipelineMetrics(client *http.Client, endpoint string, prefix string, dimensions map[string]string) ([]*datapoint.Datapoint, error) {
	var dps []*datapoint.Datapoint

	metricsJSON, err := getJSON(client, endpoint)
	if err != nil {
		return nil, err
	}

	pipelines, exists := metricsJSON["pipelines"]
	if !exists {
		return nil, fmt.Errorf("`pipelines` object doesn't exist in response from %s", endpoint)
	}
	pipelinesObj, isObject := pipelines.(map[string]interface{})
	if !isObject {
		return nil, fmt.Errorf("`pipelines` is not an object in response from %s", endpoint)
	}
	for pipelineName, pipelineObj := range pipelinesObj {
		pipeline, converted := pipelineObj.(map[string]interface{})
		if !converted {
			continue
		}

		dimsClone := utils.CloneStringMap(dimensions)
		dimsClone["pipeline"] = pipelineName

		dps = append(dps, m.extractDatapoints(prefix, pipeline, dimsClone)...)
	}

	return dps, nil
}

func (m *Monitor) fetchPluginMetrics(client *http.Client, endpoint string, prefix string, dimensions map[string]string) ([]*datapoint.Datapoint, error) {
	metricsJSON, err := getJSON(client, endpoint)
	if err != nil {
		return nil, err
	}

	total, exists := metricsJSON["total"]
	if !exists {
		return nil, fmt.Errorf("`total` object doesn't exist in response from %s", endpoint)
	}

	metricName := prefix + ".total"
	metricType, isEnabled := m.metricTypeMap[metricName]
	if !isEnabled {
		return nil, nil
	}
	metricValue, castErr := datapoint.CastMetricValueWithBool(total)
	if castErr != nil {
		return nil, fmt.Errorf("Couldn't cast `total` metric value: %w", castErr)
	}

	return []*datapoint.Datapoint{
		{
			Metric:     metricName,
			MetricType: metricType,
			Value:      metricValue,
			Dimensions: dimensions,
		},
	}, nil
}

func (m *Monitor) extractDatapoints(metricPath string, metricsJSON map[string]interface{}, dims map[string]string) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	for k, v := range metricsJSON {
		childPath := metricPath + "." + k
		if obj, isObject := v.(map[string]interface{}); isObject {
			dps = append(dps, m.extractDatapoints(childPath, obj, dims)...)
		} else if arr, isArray := v.([]interface{}); isArray {
			for _, arrayItem := range arr {
				obj, isObject = arrayItem.(map[string]interface{})
				if !isObject {
					continue
				}

				objectName, exists := obj["name"]
				if !exists {
					continue
				}
				dimsClone := utils.CloneStringMap(dims)
				dimName, exists := dimensionKeyMap[k]
				if !exists {
					dimName = k
				}
				dimsClone[dimName], _ = objectName.(string)

				dps = append(dps, m.extractDatapoints(childPath, obj, dimsClone)...)
			}
		} else if metricType, exists := m.metricTypeMap[childPath]; exists {
			metricValue, err := datapoint.CastMetricValueWithBool(v)
			if err != nil {
				m.logger.WithError(err).Errorf("Couldn't cast value: %s", childPath)
				continue
			}
			dps = append(dps, &datapoint.Datapoint{
				Metric:     childPath,
				MetricType: metricType,
				Value:      metricValue,
				Dimensions: dims,
			})
		}
	}

	return dps
}

// Shutdown the monitor
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

func getJSON(client *http.Client, endpoint string) (map[string]interface{}, error) {
	response, err := client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %w", endpoint, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not connect to %s : %s ", endpoint, http.StatusText(response.StatusCode))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	metricsJSON := make(map[string]interface{})
	err = json.Unmarshal(body, &metricsJSON)
	if err != nil {
		return nil, err
	}

	return metricsJSON, nil
}
