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
	"context"
	"net/http"
	"sync"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/jaegertracing/jaeger/model"
	jaegertranslator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
)

var marshaler = &jsonpb.Marshaler{}

// httpSinkExporter ...
type httpSinkExporter struct {
	endpoint string

	ch      chan *model.Batch
	clients []*client
	mu      sync.Mutex
}

func (e *httpSinkExporter) ConsumeTraces(_ context.Context, td pdata.Traces) error {
	batches, err := jaegertranslator.InternalTracesToJaegerProto(td)
	if err != nil {
		return err
	}
	for _, batch := range batches {
		go func(b *model.Batch) {
			e.ch <- b
		}(batch)
	}
	return nil
}

func (e *httpSinkExporter) addClient(c *client) {
	e.mu.Lock()
	e.clients = append(e.clients, c)
	e.mu.Unlock()
}

func (e *httpSinkExporter) removeClient(c *client) {
	e.mu.Lock()
	index := -1
	for i, v := range e.clients {
		if v == c {
			index = i
			break
		}
	}
	if index != -1 {
		e.clients = append(e.clients[:index], e.clients[index+1:]...)
	}
	e.mu.Unlock()
}

func (e *httpSinkExporter) handler(w http.ResponseWriter, r *http.Request) {
	opts, err := parseOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := newClient(opts)
	e.addClient(c)
	defer e.removeClient(c)

	result, err := c.response()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (e *httpSinkExporter) Start(ctx context.Context, _ component.Host) error {
	go e.startServer(ctx)
	go e.fanOut()
	return nil
}

func (e *httpSinkExporter) fanOut() error {
	e.ch = make(chan *model.Batch)
	for {
		batch := <-e.ch
		e.mu.Lock()
		clients := e.clients
		e.mu.Unlock()
		for _, c := range clients {
			if !c.stopped {
				go func(c *client) {
					c.ch <- batch
				}(c)
			}
		}
	}
}

// Shutdown stops the exporter and is invoked during shutdown.
func (e *httpSinkExporter) Shutdown(context.Context) error {
	// shutdown http server
	return nil

}

func (e *httpSinkExporter) startServer(context.Context) {
	http.HandleFunc("/", e.handler)
	http.ListenAndServe(e.endpoint, nil)
}
