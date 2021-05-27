// Copyright Splunk, Inc.
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

package httpsinkexporter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jaegertracing/jaeger/model"
)

type options struct {
	attrs   map[string]string
	names   []string
	count   int
	timeout time.Duration
}

func parseOptions(r *http.Request) (options, error) {
	opts := options{
		count:   1,
		timeout: time.Second * 10,
		names:   []string{},
		attrs:   map[string]string{},
	}

	q := r.URL.Query()

	for _, attr := range q["attr"] {
		parts := strings.SplitN(attr, "=", 2)
		if len(parts) != 2 {
			return opts, fmt.Errorf("attr query string parameter is not formatted correctly")
		}
		opts.attrs[parts[0]] = parts[1]
	}

	opts.names = q["name"]

	if timeout, ok := q["timeout"]; ok {
		timeoutNum, err := strconv.Atoi(timeout[0])
		if err != nil {
			return opts, err
		}
		opts.timeout = time.Second * time.Duration(timeoutNum)
	}

	if count, ok := q["count"]; ok {
		countNum, err := strconv.Atoi(count[0])
		if err != nil {
			return opts, err
		}
		opts.count = countNum
	}

	return opts, nil
}

type client struct {
	ch      chan *model.Batch
	opts    options
	stopped bool
}

func newClient(opts options) *client {
	return &client{
		ch:   make(chan *model.Batch),
		opts: opts,
	}
}

func (c *client) response() ([]byte, error) {
	// TODO: add support to filter by attributes and names
	defer func() {
		c.stopped = true
	}()

	spans := []string{}
	received := 0

	done := make(chan struct{})
	for {
		select {
		case batch := <-c.ch:
			for _, span := range batch.Spans {
				json, err := marshaler.MarshalToString(span)
				if err != nil {
					return nil, err
				}
				spans = append(spans, json)
				received++
				if received == c.opts.count {
					close(done)
				}
			}
		case <-time.After(c.opts.timeout):
			return nil, fmt.Errorf("timed out while waiting for spans")

		case <-done:
			result := "[" + strings.Join(spans, ",") + "]"
			return []byte(result), nil
		}
	}
}
