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

package pqprocessor

import "errors"

type Config struct {
	Folder    string `mapstructure:"folder"`
	Bandwidth int32  `mapstructure:"bandwidth"`
}

func (c *Config) Validate() error {
	if c.Folder == "" {
		return errors.New("folder must be a valid folder")
	}
	if c.Bandwidth < 0 {
		return errors.New("bandwidth must be zero or positive")
	}
	return nil
}
