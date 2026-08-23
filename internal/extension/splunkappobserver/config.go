// Copyright Splunk, Inc.
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

package splunkappobserver

import (
	"errors"
	"time"
)

const defaultAppsPath = "$SPLUNK_HOME/etc/apps"

type Config struct {
	// AppsPath is the directory whose direct child directories are reported as
	// Splunk app endpoints.
	AppsPath string `mapstructure:"apps_path"`

	// RefreshInterval determines how frequently the observer checks for app
	// directory additions and removals.
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`

	// prevent unkeyed literal initialization
	_ struct{}
}

func (cfg *Config) Validate() error {
	if cfg.AppsPath == "" {
		return errors.New("apps_path must be specified")
	}
	if cfg.RefreshInterval <= 0 {
		return errors.New("refresh_interval must be greater than zero")
	}
	return nil
}
