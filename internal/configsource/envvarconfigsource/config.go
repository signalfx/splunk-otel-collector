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

package envvarconfigsource

import (
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
)

// Config holds the configuration for the creation of environment variable config source objects.
type Config struct {
	// Defaults specify a map to fallback if a given environment variable is not defined.
	Defaults map[string]any `mapstructure:"defaults"`

	expcfg.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
}

func (*Config) Validate() error {
	return nil
}
