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

package collectd

import (
	"os"
	"path/filepath"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// MakePythonPluginPath takes file path components below the BundleDir for
// Python plugins and returns an os appropriate file path.  The environment
// variable SIGNALFX_BUNDLE_DIR is used as the root of the path
func MakePythonPluginPath(components ...string) string {
	components = append([]string{os.Getenv(constants.BundleDirEnvVar), "collectd-python"}, components...)
	return filepath.Join(components...)
}

// DefaultTypesDBPath returns the default types.db path based on the bundle dir
// envvar.
func DefaultTypesDBPath() string {
	return filepath.Join(os.Getenv(constants.BundleDirEnvVar), "types.db")
}
