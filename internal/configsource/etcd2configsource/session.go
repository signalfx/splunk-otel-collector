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

package etcd2configsource

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.etcd.io/etcd/client"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const maxBackoffTime = time.Second * 60

// etcd2Session implements the configsource.Session interface.
type etcd2Session struct {
	logger     *zap.Logger
	kapi       client.KeysAPI
	closeFuncs []func()
}

var _ configsource.Session = (*etcd2Session)(nil)

func newSession(logger *zap.Logger, kapi client.KeysAPI) *etcd2Session {
	return &etcd2Session{
		logger:     logger,
		kapi:       kapi,
		closeFuncs: []func(){},
	}
}

func (s *etcd2Session) Retrieve(ctx context.Context, selector string, _ interface{}) (configsource.Retrieved, error) {
	resp, err := s.kapi.Get(ctx, selector, nil)
	if err != nil {
		return nil, err
	}

	watchCtx, cancel := context.WithCancel(context.Background())
	s.closeFuncs = append(s.closeFuncs, cancel)

	return configprovider.NewWatchableRetrieved(resp.Node.Value, s.newWatcher(watchCtx, selector, resp.Node.ModifiedIndex)), nil
}

func (s *etcd2Session) RetrieveEnd(context.Context) error {
	return nil
}

func (s *etcd2Session) Close(context.Context) error {
	for _, cancel := range s.closeFuncs {
		cancel()
	}

	return nil
}

func (s *etcd2Session) newWatcher(ctx context.Context, selector string, index uint64) func() error {
	return func() error {
		watcher := s.kapi.Watcher(selector, &client.WatcherOptions{AfterIndex: index})

		ebo := backoff.NewExponentialBackOff()
		ebo.MaxElapsedTime = maxBackoffTime
		for {
			_, err := watcher.Next(ctx)
			if err == nil {
				return configsource.ErrValueUpdated
			}

			if err == context.Canceled {
				return configsource.ErrSessionClosed
			}

			s.logger.Info("error watching", zap.String("selector", selector), zap.Error(err))
			// if error is recoverable, try again with backoff
			if _, ok := err.(*client.ClusterError); ok {
				select {
				case <-time.After(ebo.NextBackOff()):
					continue
				case <-ctx.Done():
					return configsource.ErrSessionClosed
				}
			}

			return err
		}
	}
}
