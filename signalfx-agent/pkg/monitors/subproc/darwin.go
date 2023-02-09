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

//go:build darwin
// +build darwin

package subproc

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// The Darwin specific process attribute that make the Python runner be in the
// same process group as the agent so they get shutdown together.
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func defaultPythonBinaryExecutable() string {
	return filepath.Join(os.Getenv(constants.BundleDirEnvVar), "bin/python")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
