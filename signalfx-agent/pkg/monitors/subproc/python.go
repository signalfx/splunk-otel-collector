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

package subproc

import (
	"os"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// DefaultPythonRuntimeConfig returns the runtime config that uses the bundled Python
// runtime.
func DefaultPythonRuntimeConfig(pkgName string) *RuntimeConfig {
	// The PYTHONHOME envvar is set in agent core when config is processed.
	env := os.Environ()
	env = append(env, config.BundlePythonHomeEnvvar())

	return &RuntimeConfig{
		Binary: defaultPythonBinaryExecutable(),
		Args:   defaultPythonBinaryArgs(pkgName),
		Env:    env,
	}
}
