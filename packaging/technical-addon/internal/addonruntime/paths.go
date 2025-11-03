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

package addonruntime

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetBinaryPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %v", err)
	}
	return execPath, nil
}

func GetTaPlatformDir() (string, error) {
	execPath, err := GetBinaryPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(execPath)), nil //  ./(windows_x86_64|linux_x86_64)/bin/<binary> -> ../../
}

func GetTaHome() (string, error) {
	splunkTaPlatformHome, err := GetTaPlatformDir()
	if err != nil {
		return "", err
	}
	return filepath.Dir(splunkTaPlatformHome), nil //  <Name of TA>/(windows_x86_64|linux_x86_64) -> ../
}

func GetSplunkHome() (string, error) {
	splunkTaHome, err := GetTaHome()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(splunkTaHome)), nil //  <Splunk_Home>/etc/(apps|deployment_apps)/<Name of TA> -> ../../
}
