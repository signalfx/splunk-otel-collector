// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package cyberarkconfigsource

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

// defaultSelectorField is the field returned when the reference selector is empty,
// e.g. ${cyberark:} resolves to the account password.
const defaultSelectorField = "Password"

// errBadSelector is returned when a reference selects a field that was not retrieved.
type errBadSelector struct{ error }

// retriever isolates the CyberArk backend used to fetch a secret's fields. The core
// config source is agnostic to how the bytes are obtained: the "cp" retriever shells out
// to CLIPasswordSDK, while a future "ccp" retriever could call the AIMWebService REST API.
type retriever interface {
	// retrieve fetches the object and returns its fields keyed by logical name
	// (e.g. "Password", "UserName").
	retrieve(ctx context.Context) (map[string]any, error)
}

// cyberarkConfigSource implements the configsource.ConfigSource interface. It is pinned to
// a single CyberArk object: the first Retrieve fetches and caches all fields, later
// Retrieve calls select from the cache.
type cyberarkConfigSource struct {
	logger    *zap.Logger
	retriever retriever

	// fields caches the retrieved object. Nil until the first successful retrieve.
	fields map[string]any

	autoRefresh  bool
	pollInterval time.Duration
}

func newConfigSource(cfg *Config, logger *zap.Logger) (configsource.ConfigSource, error) {
	r, err := newRetriever(cfg)
	if err != nil {
		return nil, err
	}

	return &cyberarkConfigSource{
		logger:       logger,
		retriever:    r,
		autoRefresh:  cfg.AutoRefresh,
		pollInterval: cfg.PollInterval,
	}, nil
}

// newRetriever builds the backend retriever selected by the config's retrieval_mode.
func newRetriever(cfg *Config) (retriever, error) {
	switch cfg.RetrievalMode {
	case retrievalModeCP:
		return newCPRetriever(cfg), nil
	default:
		return nil, &errUnsupportedMode{fmt.Errorf("unsupported retrieval_mode %q", cfg.RetrievalMode)}
	}
}

func (c *cyberarkConfigSource) Retrieve(ctx context.Context, selector string, _ *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	// By default assume that watcher is not supported. The exception will be the first
	// value read from the CyberArk object when auto_refresh is enabled.
	var closeFunc confmap.CloseFunc

	// All fields come from the same object so fetching (and watching) only on the first
	// Retrieve is fine.
	if c.fields == nil {
		fields, err := c.retriever.retrieve(ctx)
		if err != nil {
			return nil, err
		}
		c.fields = fields

		if c.autoRefresh && watcher != nil {
			doneCh := make(chan struct{})
			// The polling watcher runs on its own long-lived background context by
			// design, so it intentionally does not thread ctx through.
			c.buildPollingWatcher(watcher, doneCh) //nolint:contextcheck
			closeFunc = func(_ context.Context) error {
				close(doneCh)
				return nil
			}
		}
	}

	key := selector
	if key == "" {
		key = defaultSelectorField
	}

	value, ok := c.fields[key]
	if !ok {
		return nil, &errBadSelector{fmt.Errorf("no value for field %q", key)}
	}

	return confmap.NewRetrieved(value, confmap.WithRetrievedClose(closeFunc))
}

// buildPollingWatcher polls the retriever on pollInterval and triggers a config reload
// when the retrieved fields change. Mirrors the vault config source polling watcher.
func (c *cyberarkConfigSource) buildPollingWatcher(watcher confmap.WatcherFunc, doneCh chan struct{}) {
	original := c.fields

	go func() {
		ticker := time.NewTicker(c.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// The polling loop outlives the originating Retrieve call, so it must
				// not reuse that (potentially canceled) context; a fresh background
				// context scopes each refresh to its own CLI invocation.
				latest, err := c.retriever.retrieve(context.Background())
				if err != nil {
					// Assume the configuration needs to be re-fetched. The reload will
					// surface a persistent error at resolution time.
					watcher(&confmap.ChangeEvent{Error: fmt.Errorf("failed to refresh CyberArk object: %w", err)})
					return
				}

				if !reflect.DeepEqual(original, latest) {
					watcher(&confmap.ChangeEvent{})
					return
				}
			case <-doneCh:
				return
			}
		}
	}()
}
