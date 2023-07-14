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
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

var errTooManyRequests = errors.New("429 Too Many Requests")

// Client defines an http REST client. Currently only GET is supported.
type Client interface {
	Get(path string) ([]byte, error)
}

func NewClient(httpDoer HTTPDoer, tok string, logger *zap.Logger) Client {
	return bearerClient{
		httpDoer: httpDoer,
		tok:      tok,
		ebo:      backoff.NewExponentialBackOff(),
		logger:   logger,
	}
}

// bearerClient is a Client implementation that adds an Authorization Bearer +
// token to its requests. It also retries HTTP 429 responses with exponential backoff.
type bearerClient struct {
	httpDoer HTTPDoer
	ebo      backoff.BackOff
	logger   *zap.Logger
	tok      string
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c bearerClient) Get(url string) (out []byte, err error) {
	for {
		out, err = c.doGet(url)
		if err != errTooManyRequests {
			c.ebo.Reset()
			break
		}
		nextBackoff := c.ebo.NextBackOff()
		c.logger.Debug("bearerClient: rate limit exceeded, will retry", zap.Duration("nextBackoff", nextBackoff))
		time.Sleep(nextBackoff)
	}
	return out, err
}

func (c bearerClient) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bearerCclient failed to create new request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.tok)
	resp, err := c.httpDoer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bearerClient failed to send http request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusTooManyRequests:
		return nil, errTooManyRequests
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
