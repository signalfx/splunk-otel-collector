package stats

import (
	"fmt"
	"net/http"

	es "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/client"
)

const (
	nodeStatsEndpoint          = "_nodes/_local/stats/transport,http,process,jvm,indices,thread_pool"
	clusterHealthStatsEndpoint = "_cluster/health"
	nodeInfoEndpoint           = "_nodes/_local"
	masterNodeEndpoint         = "_cluster/state/master_node"
	allIndexStatsEndpoint      = "_all/_stats"
)

// ESStatsHTTPClient holds methods hitting various ES stats endpoints
type ESStatsHTTPClient struct {
	esClient *es.ESClient
}

// NewESClient creates a new esClient
func NewESClient(host string, port string, scheme string, client *http.Client) ESStatsHTTPClient {
	return ESStatsHTTPClient{
		esClient: &es.ESClient{
			Scheme:     scheme,
			Host:       host,
			Port:       port,
			HTTPClient: client,
		},
	}
}

// Method to collect index stats
func (c *ESStatsHTTPClient) GetIndexStats() (*IndexStatsOutput, error) {
	url := fmt.Sprintf("%s://%s:%s/%s", c.esClient.Scheme, c.esClient.Host, c.esClient.Port, allIndexStatsEndpoint)

	var indexStatsOutput IndexStatsOutput

	err := c.esClient.FetchJSON(url, &indexStatsOutput)

	return &indexStatsOutput, err
}

// Method to identify the master node
func (c *ESStatsHTTPClient) GetMasterNodeInfo() (*MasterInfoOutput, error) {
	url := fmt.Sprintf("%s://%s:%s/%s", c.esClient.Scheme, c.esClient.Host, c.esClient.Port, masterNodeEndpoint)

	var masterInfoOutput MasterInfoOutput

	err := c.esClient.FetchJSON(url, &masterInfoOutput)

	return &masterInfoOutput, err
}

// Method to fetch node info
func (c *ESStatsHTTPClient) GetNodeInfo() (*NodeInfoOutput, error) {
	url := fmt.Sprintf("%s://%s:%s/%s", c.esClient.Scheme, c.esClient.Host, c.esClient.Port, nodeInfoEndpoint)

	var nodeInfoOutput NodeInfoOutput

	err := c.esClient.FetchJSON(url, &nodeInfoOutput)

	return &nodeInfoOutput, err
}

// Method to fetch cluster stats
func (c *ESStatsHTTPClient) GetClusterStats() (*ClusterStatsOutput, error) {
	url := fmt.Sprintf("%s://%s:%s/%s", c.esClient.Scheme, c.esClient.Host, c.esClient.Port, clusterHealthStatsEndpoint)

	var clusterStatsOutput ClusterStatsOutput

	err := c.esClient.FetchJSON(url, &clusterStatsOutput)

	return &clusterStatsOutput, err
}

// Method to fetch node stats
func (c *ESStatsHTTPClient) GetNodeAndThreadPoolStats() (*NodeStatsOutput, error) {
	url := fmt.Sprintf("%s://%s:%s/%s", c.esClient.Scheme, c.esClient.Host, c.esClient.Port, nodeStatsEndpoint)

	var nodeStatsOutput NodeStatsOutput

	err := c.esClient.FetchJSON(url, &nodeStatsOutput)

	return &nodeStatsOutput, err
}
