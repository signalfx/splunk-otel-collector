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
	"fmt"
	"io"
	"net/http"
)

// authClient sends requests with a bearer token to the given URL and returns a
// byte array
type authClient struct {
	baseURL string
	tok     string
}

func (c authClient) get(path string) (out []byte, err error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	req.Header.Add("Authorization", "Bearer "+c.tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authClient.get(): %w", err)
	}
	defer func() { err = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}
