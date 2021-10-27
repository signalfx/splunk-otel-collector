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
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
)

// Config defines etcd2configsource configuration
type Config struct {
	expcfg.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	// Authentication defines the authentication method to be used.
	Authentication *Authentication `mapstructure:"auth"`

	// Endpoints is a list of etcd2 server endpoints the etcd2
	// config source should try to connect to.
	Endpoints []string `mapstructure:"endpoints"`
}

// Authentication holds the authentication configuration for Etcd2 config source objects.
type Authentication struct {
	// Username can be optionally used to authenticate with etcd2 cluster.
	Username string `mapstructure:"username"`

	// Password can be optionally used to authenticate with etcd2 cluster.
	Password string `mapstructure:"password"`
}

func (*Config) Validate() error {
	return nil
}
