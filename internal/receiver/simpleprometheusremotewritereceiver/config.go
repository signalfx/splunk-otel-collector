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

package simpleprometheusremotewritereceiver

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/multierr"
)

var _ component.Config = (*Config)(nil)

const (
	typeString = "prometheusremotewrite"
)

type Config struct {
	ListenPath                    string `mapstructure:"path"`
	confighttp.HTTPServerSettings `mapstructure:",squash"`
	Timeout                       time.Duration `mapstructure:"timeout"`
	BufferSize                    int           `mapstructure:"buffer_size"`
	CacheCapacity                 int           `mapstructure:"cache_size"`
}

func (c *Config) Validate() error {
	var errs []error
	if c.Endpoint == "" {
		errs = append(errs, errors.New("endpoint must not be empty"))
	}
	if c.Timeout < time.Millisecond {
		errs = append(errs, errors.New("impractically short timeout"))
	}
	if c.BufferSize < 0 {
		errs = append(errs, errors.New("buffer size must be non-negative"))
	}
	if c.CacheCapacity <= 100 {
		errs = append(errs, errors.New("inadvisably small capacity for cache"))
	}
	if errs != nil {
		return multierr.Combine(errs...)
	}
	return nil
}
