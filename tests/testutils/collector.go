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

package testutils

import (
	"regexp"
	"testing"

	"go.uber.org/zap"
)

type Collector interface {
	WithConfigPath(path string) Collector
	WithArgs(args ...string) Collector
	WithEnv(env map[string]string) Collector
	WithLogger(logger *zap.Logger) Collector
	WithLogLevel(level string) Collector
	WillFail(fail bool) Collector
	WithMount(path string, mountPoint string) Collector
	Build() (Collector, error)
	Start() error
	Shutdown() error
	InitialConfig(t testing.TB, port uint16) map[string]any
	EffectiveConfig(t testing.TB, port uint16) map[string]any
}

var configFromArgsPattern = regexp.MustCompile("--config($|[^d-]+)")

func configIsSetByArgs(args []string) bool {
	for _, c := range args {
		if configFromArgsPattern.Match([]byte(c)) {
			return true
		}
	}
	return false
}
