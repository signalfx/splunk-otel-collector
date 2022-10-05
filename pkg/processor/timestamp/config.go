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
	"fmt"
	"time"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	zeroTs = pcommon.Timestamp(0)
)

type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	Offset                   string                   `mapstructure:"offset"`
}

var _ config.Processor = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	_, err := time.ParseDuration(cfg.Offset)
	if err != nil {
		return fmt.Errorf("invalid offset format %s: %w", cfg.Offset, err)
	}
	return nil
}

func offsetFn(offset time.Duration) func(pcommon.Timestamp) pcommon.Timestamp {
	return func(ts pcommon.Timestamp) pcommon.Timestamp {
		if ts == zeroTs {
			return ts
		}
		t := ts.AsTime()
		t = t.Add(offset)
		return pcommon.NewTimestampFromTime(t)
	}
}
