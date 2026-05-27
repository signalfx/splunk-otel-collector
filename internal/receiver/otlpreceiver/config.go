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

package otlpreceiver

import (
	"fmt"
	"time"

	upstreamotlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
)

// Config defines configuration for the OTLP receiver.
type Config struct {
	upstreamotlpreceiver.Config `mapstructure:",squash"`

	// StartDelay delays starting the upstream OTLP receiver after the collector
	// triggers this receiver's Start.
	StartDelay time.Duration `mapstructure:"start_delay"`
}

// Validate checks the receiver configuration is valid.
func (cfg *Config) Validate() error {
	if cfg.StartDelay < 0 {
		return fmt.Errorf("start_delay must be non-negative")
	}
	if err := cfg.Config.Validate(); err != nil {
		return err
	}
	return nil
}
