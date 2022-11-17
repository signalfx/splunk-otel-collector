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
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.etcd.io/etcd/client/v2"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const maxBackoffTime = time.Second * 60

// etcd2ConfigSource implements the configprovider.Session interface.
type etcd2ConfigSource struct {
	logger *zap.Logger
	kapi   client.KeysAPI
}

func newConfigSource(params configprovider.CreateParams, cfg *Config) (configprovider.ConfigSource, error) {
	var username, password string
	if cfg.Authentication != nil {
		username = cfg.Authentication.Username
		password = cfg.Authentication.Password
	}
	etcdClient, err := client.New(client.Config{
		Endpoints: cfg.Endpoints,
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, err
	}

	kapi := client.NewKeysAPI(etcdClient)

	return &etcd2ConfigSource{
		logger: params.Logger,
		kapi:   kapi,
	}, nil
}

func (s *etcd2ConfigSource) Retrieve(ctx context.Context, selector string, _ *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error) {
	resp, err := s.kapi.Get(ctx, selector, nil)
	if err != nil {
		return nil, err
	}
	if watcher == nil {
		return confmap.NewRetrieved(resp.Node.Value)
	}
	return confmap.NewRetrieved(resp.Node.Value, confmap.WithRetrievedClose(s.newWatcher(selector, resp.Node.ModifiedIndex, watcher)))
}

func (s *etcd2ConfigSource) Shutdown(context.Context) error {
	return nil
}

func (s *etcd2ConfigSource) newWatcher(selector string, index uint64, watcherFunc confmap.WatcherFunc) confmap.CloseFunc {
	watchCtx, cancel := context.WithCancel(context.Background())
	watcher := s.kapi.Watcher(selector, &client.WatcherOptions{AfterIndex: index})
	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = maxBackoffTime

	go func() {
		for {
			_, err := watcher.Next(watchCtx)
			if err == nil {
				// Value updated
				watcherFunc(&confmap.ChangeEvent{Error: nil})
				return
			}

			if errors.Is(err, context.Canceled) {
				return
			}

			s.logger.Info("error watching", zap.String("selector", selector), zap.Error(err))
			// if error is recoverable, try again with backoff
			cErr := &client.ClusterError{}
			if errors.As(err, &cErr) {
				select {
				case <-time.After(ebo.NextBackOff()):
					continue
				case <-watchCtx.Done():
					return
				}
			}
			watcherFunc(&confmap.ChangeEvent{Error: err})
			return
		}
	}()

	return func(ctx context.Context) error {
		cancel()
		return nil
	}
}
