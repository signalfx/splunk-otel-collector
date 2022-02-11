// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package databricksreceiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// authClient sends requests with a bearer token to the given URL and returns a
// byte array
type authClient struct {
	httpClient *http.Client
	endpoint   string
	tok        string
}

type errorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func (c authClient) get(path string) ([]byte, error) {
	const method = "authClient.get()"
	req, err := http.NewRequest("GET", c.endpoint+path, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}
	req.Header.Add("Authorization", "Bearer "+c.tok)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}

	// read the response regardless of status code
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: status code: %d: %w", method, resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		er := errorResponse{}
		_ = json.Unmarshal(out, &er)
		return nil, fmt.Errorf(
			"%s: status code: %d: %s: error code: %s message: %s",
			method,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			er.ErrorCode,
			er.Message,
		)
	}

	return out, err
}
