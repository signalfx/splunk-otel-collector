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

package signalfxgatewayprometheusremotewritereceiver

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/multierr"
)

var _ component.Config = (*Config)(nil)

type Config struct {
	// ListenPath is the path in which the receiver should respond to prometheus remote write requests.
	ListenPath string `mapstructure:"path"`
	// provides generic settings for connecting to HTTP servers as commonly used in opentelemetry
	confighttp.HTTPServerSettings `mapstructure:",squash"`
	// BufferSize is the degree to which metric translations may be buffered without blocking further write requests.
	BufferSize int `mapstructure:"buffer_size"`
}

func (c *Config) Validate() error {
	var errs []error
	if c.Endpoint == "" {
		errs = append(errs, errors.New("endpoint must not be empty"))
	}
	if c.BufferSize < 0 {
		errs = append(errs, errors.New("buffer size must be non-negative"))
	}
	if errs != nil {
		return multierr.Combine(errs...)
	}
	return nil
}
