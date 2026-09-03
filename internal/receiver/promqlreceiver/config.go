// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promqlreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Config struct {
	Queries          []string                       `mapstructure:"queries"`
	ClientConfig     confighttp.ClientConfig        `mapstructure:",squash"`
	ControllerConfig scraperhelper.ControllerConfig `mapstructure:",squash"`
}

func (c *Config) Validate() error {
	if len(c.Queries) == 0 {
		return errors.New("queries cannot be empty")
	}
	if c.ClientConfig.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	return nil
}
