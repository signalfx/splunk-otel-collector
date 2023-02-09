// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	es "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/client"
)

type ESQueryHTTPClient struct {
	esClient *es.ESClient
}

// NewESQueryClient creates a new ESQueryHTTPClient
func NewESQueryClient(host string, port string, scheme string, client *http.Client) ESQueryHTTPClient {
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
func (es ESQueryHTTPClient) makeHTTPRequestFromConfig(index string, esSearchRequest string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/%s/_search?", es.esClient.Scheme, es.esClient.Host, es.esClient.Port, index)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(esSearchRequest)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := es.esClient.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
