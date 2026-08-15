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

package persistentqueueconnector

import (
	"errors"
	"time"
)

type Config struct {
	Path            string        `mapstructure:"path"`
	ThroughputLimit int64         `mapstructure:"throughput_limit"`
	MaxBytesPerFile int64         `mapstructure:"max_bytes_per_file"`
	SyncEvery       int64         `mapstructure:"sync_every"`
	SyncTimeout     time.Duration `mapstructure:"sync_timeout"`
	CompactInterval time.Duration `mapstructure:"compact_interval"`
}

func (c *Config) Validate() error {
	if c.Path == "" {
		return errors.New("path must be a valid folder")
	}
	if c.ThroughputLimit < 0 {
		return errors.New("throughput_limit must be zero or positive")
	}
	return nil
}
