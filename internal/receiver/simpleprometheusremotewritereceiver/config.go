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

	"github.com/jaegertracing/jaeger/pkg/multierror"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confignet"
)

var _ component.Config = (*Config)(nil)

const (
	typeString = "simpleprometheusremotewrite"
)

type Config struct {
	ListenAddr    confignet.NetAddr `mapstructure:",squash"`
	ListenPath    string            `mapstructure:"path"`
	Timeout       time.Duration     `mapstructure:"timeout"`
	BufferSize    int               `mapstructure:"buffer_size"` // Channel buffer size, defaults to blocking each request until processed
	CacheCapacity int               `mapstructure:"cache_size"`
}

func (c *Config) Validate() error {
	var errs []error
	if c.ListenAddr.Endpoint == "" {
		errs = append(errs, errors.New("endpoint must not be empty"))
	}
	if c.ListenAddr.Transport == "" {
		errs = append(errs, errors.New("transport must not be empty"))
	}
	if c.Timeout < time.Millisecond {
		errs = append(errs, errors.New("impractically short timeout"))
	}
	if c.CacheCapacity <= 100 {
		errs = append(errs, errors.New("inadvisably small capacity for cache"))
	}
	if err := componenttest.CheckConfigStruct(c); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return multierror.Wrap(errs)
	}
	return nil
}
