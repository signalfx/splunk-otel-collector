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

package httpauth

import (
	"fmt"
	"io"
	"net/http"
)

// Client defines an http REST client. Currently only GET is supported.
type Client interface {
	Get(path string) ([]byte, error)
}

func NewClient(httpClient *http.Client, tok string) Client {
	return bearerClient{httpClient: httpClient, tok: tok}
}

// bearerClient is a Client implementation that adds an Authorization Bearer +
// token to its requests.
type bearerClient struct {
	httpClient *http.Client
	tok        string
}

func (c bearerClient) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bearerCclient failed to create new request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.tok)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bearerClient failed to send http request: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	// read response JSON even if the status code is not http.StatusOK
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil && resp.StatusCode == http.StatusOK {
		return nil, fmt.Errorf("bearerClient failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		text := http.StatusText(resp.StatusCode)
		return out, fmt.Errorf("bearerClient got status code: %d: %s", resp.StatusCode, text)
	}

	return out, nil
}
