// Copyright  Splunk, Inc.
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

package timestamp

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	zeroTs = pcommon.Timestamp(0)
)

var (
	errorInvalidOffsetFormat = errors.New("invalid offset format")
)

type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	Offset                   string                   `mapstructure:"offset"`
}

var _ config.Processor = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Offset[0] != '-' && cfg.Offset[0] != '+' {
		return errorInvalidOffsetFormat
	}
	_, err := time.ParseDuration(cfg.Offset[1:])
	return err
}

func (cfg *Config) offsetFn() func(pcommon.Timestamp) pcommon.Timestamp {
	backwards := cfg.Offset[0] == '-'
	interval, _ := time.ParseDuration(cfg.Offset[1:])
	return func(ts pcommon.Timestamp) pcommon.Timestamp {
		if ts == zeroTs {
			return ts
		}
		t := ts.AsTime()

		if backwards {
			t = t.Add(-1 * interval)
		} else {
			t = t.Add(interval)
		}
		return pcommon.NewTimestampFromTime(t)
	}
}
