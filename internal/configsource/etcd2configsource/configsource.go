// Copyright 2020 Splunk, Inc.
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

	"go.etcd.io/etcd/client"
	"go.opentelemetry.io/collector/config/experimental/configsource"
	"go.uber.org/zap"
)

type etcd2ConfigSource struct {
	logger *zap.Logger
	client client.Client
	kapi   client.KeysAPI
}

var _ configsource.ConfigSource = (*etcd2ConfigSource)(nil)

func (s *etcd2ConfigSource) NewSession(context.Context) (configsource.Session, error) {
	return newSession(s.logger, s.kapi), nil
}

func newConfigSource(logger *zap.Logger, cfg *Config) (*etcd2ConfigSource, error) {
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
		logger: logger,
		client: etcdClient,
		kapi:   kapi,
	}, nil
}
