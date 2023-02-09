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

package hostfs

import (
	"os"
)

const (
	// HostProcVar is the environment variable name that is set with the host's /proc path
	HostProcVar = "HOST_PROC"
	// HostSysVar is the environment variable name that is set with the host's /sys path
	HostSysVar = "HOST_SYS"
	// HostRunVar is the environment variable name that is set with the host's /run path
	HostRunVar = "HOST_RUN"
	// HostVarVar is the environment variable name that is set with the host's /var path
	HostVarVar = "HOST_VAR"
	// HostEtcVar is the environment variable name that is set with the host's /etc path
	HostEtcVar = "HOST_ETC"
)

// HostProc returns the configured /proc path
func HostProc() string {
	return os.Getenv(HostProcVar)
}

// HostEtc returns the configured /etc path
func HostEtc() string {
	return os.Getenv(HostEtcVar)
}

// HostRun returns the configured /run path
func HostRun() string {
	return os.Getenv(HostRunVar)
}

// HostVar returns the configured /var path
func HostVar() string {
	return os.Getenv(HostVarVar)
}

// HostSys returns the configured host /sys path
func HostSys() string {
	return os.Getenv(HostSysVar)
}
