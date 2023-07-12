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
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

var sinkFactoryMu = sync.Mutex{}
var sinks = map[string]*sink{}

type sink struct {
	server    *http.Server
	logger    *zap.Logger
	endpoint  string
	_clients  []*client
	startOnce sync.Once
	mu        sync.Mutex
}

func newSink(logger *zap.Logger, endpoint string) *sink {
	sinkFactoryMu.Lock()
	defer sinkFactoryMu.Unlock()
	s, ok := sinks[endpoint]
	if !ok {
		s = &sink{
			logger:    logger,
			endpoint:  endpoint,
			startOnce: sync.Once{},
		}
		sinks[endpoint] = s
	}
	return s
}

func (s *sink) start(ctx context.Context) {
	s.logger.Info("starting sink", zap.String("endpoint", s.endpoint))
	s.startOnce.Do(func() {
		mux := http.NewServeMux()
		mux.Handle("/", http.HandlerFunc(s.handleDefault))
		mux.Handle("/spans", http.HandlerFunc(s.handle))
		mux.Handle("/metrics", http.HandlerFunc(s.handle))

		s.server = &http.Server{
			Addr:    s.endpoint,
			Handler: mux,
			BaseContext: func(listener net.Listener) context.Context {
				return ctx
			},
			ReadHeaderTimeout: 5 * time.Second,
		}
		go s.server.ListenAndServe()
	})
}

func (s *sink) shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *sink) clients(dType dataType) []*client {
	s.mu.Lock()
	defer s.mu.Unlock()
	var clients []*client
	for _, c := range s._clients {
		if c.opts.dataType == dType {
			clients = append(clients, c)
		}
	}
	return clients
}

func (s *sink) addClient(c *client) {
	s.mu.Lock()
	s._clients = append(s._clients, c)
	s.mu.Unlock()
}

func (s *sink) removeClient(c *client) {
	s.mu.Lock()
	index := -1
	for i, v := range s._clients {
		if v == c {
			index = i
			break
		}
	}
	if index != -1 {
		s._clients = append(s._clients[:index], s._clients[index+1:]...)
	}
	s.mu.Unlock()
}

func (s *sink) handleDefault(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("404 not found - use one of the following endpoints: \n\n \t- /spans\n \t- /metrics"))
}

func (s *sink) handle(w http.ResponseWriter, r *http.Request) {
	opts, err := parseOptions(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := newClient(opts)
	s.addClient(c)
	defer s.removeClient(c)

	result, err := c.response(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
