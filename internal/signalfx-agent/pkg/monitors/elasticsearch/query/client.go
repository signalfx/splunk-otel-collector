package query

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	es "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/client"
)

type ESQueryHTTPClient struct {
	esClient *es.ESClient
}

// NewESQueryClient creates a new ESQueryHTTPClient
func NewESQueryClient(host, port, scheme string, client *http.Client) ESQueryHTTPClient {
	return ESQueryHTTPClient{
		esClient: &es.ESClient{
			Scheme:     scheme,
			Host:       host,
			Port:       port,
			HTTPClient: client,
		},
	}
}

// Returns a response for a given elasticsearch query
func (es ESQueryHTTPClient) makeHTTPRequestFromConfig(index, esSearchRequest string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/%s/_search?", es.esClient.Scheme, es.esClient.Host, es.esClient.Port, index)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(esSearchRequest))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := es.esClient.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
