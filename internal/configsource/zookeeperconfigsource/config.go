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

package zookeeperconfigsource

import (
	"time"

	expcfg "go.opentelemetry.io/collector/config/experimental/config"
)

// Config defines zookeeperconfigsource configuration
type Config struct {
	expcfg.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	// Endpoints is an array of Zookeeper server addresses. Thr ConfigSource will try to connect
	// to these endpoints to access Zookeeper clusters.
	Endpoints []string `mapstructure:"endpoints"`
	// Timeout sets the amount of time for which a session is considered valid after losing
	// connection to a server. Within the session timeout it's possible to reestablish a connection
	// to a different server and keep the same session.
	Timeout time.Duration `mapstructure:"timeout"`
}

func (*Config) Validate() error {
	return nil
}
