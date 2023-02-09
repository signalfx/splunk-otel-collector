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

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ESClient struct {
	Scheme     string
	Host       string
	Port       string
	HTTPClient *http.Client
}

// Fetches a JSON response and puts it into an object
func (c *ESClient) FetchJSON(url string, obj interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}

	res, err := c.HTTPClient.Do(req)

	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("received status code that's not 200: %s, url: %s", res.Status, url)
	}

	err = json.NewDecoder(res.Body).Decode(obj)

	if err != nil {
		return fmt.Errorf("could not get url %s: %v", url, err)
	}

	return nil
}
