package query

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = log.WithField("monitorType", monitorType)

// Config for this monitor
type Config struct {
	config.MonitorConfig  `yaml:",inline" acceptsEndpoints:"true"`
	httpclient.HTTPConfig `yaml:",inline"`

	Host string `yaml:"host" validate:"required"`
	Port string `yaml:"port" validate:"required"`
	// Index that's being queried. If none is provided, given query will be
	// applied across all indexes. To apply the search query to multiple indices,
	// provide a comma separated list of indices
	Index string `yaml:"index" default:"_all"`
	// Takes in an Elasticsearch request body search request. See
	// [here] (https://www.elastic.co/guide/en/elasticsearch/reference/current/search-request-body.html)
	// for details.
	ElasticsearchRequest string `yaml:"elasticsearchRequest" validate:"required"`
}

// Monitor for ES queries
type Monitor struct {
	Output types.FilteringOutput
	cancel context.CancelFunc
	ctx    context.Context
	logger log.FieldLogger
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Configure monitor
func (m *Monitor) Configure(config *Config) error {
	m.logger = logger.WithField("monitorID", config.MonitorID)
	httpClient, err := config.HTTPConfig.Build()
	if err != nil {
		return err
	}

	esClient := NewESQueryClient(config.Host, config.Port, config.HTTPConfig.Scheme(), httpClient)
	m.ctx, m.cancel = context.WithCancel(context.Background())

	var reqBody ElasticsearchQueryBody
	if err = json.Unmarshal([]byte(config.ElasticsearchRequest), &reqBody); err != nil {
		return err
	}

	aggsMeta, err := reqBody.getAggregationsMeta()
	if err != nil {
		return err
	}

	utils.RunOnInterval(m.ctx, func() {
		body, err := esClient.makeHTTPRequestFromConfig(config.Index, config.ElasticsearchRequest)

		if err != nil {
			m.logger.Errorf("Failed to make HTTP request: %s", err)
			return
		}

		var resBody HTTPResponse
		if err := json.Unmarshal(body, &resBody); err != nil {
			m.logger.Errorf("Error processing HTTP response: %s", err)
			return
		}

		dps := collectDatapoints(resBody, aggsMeta, map[string]string{
			"index": config.Index,
		}, log.Fields{"monitorID": config.MonitorID})

		m.Output.SendDatapoints(dps...)
	}, time.Duration(config.IntervalSeconds)*time.Second)
	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
