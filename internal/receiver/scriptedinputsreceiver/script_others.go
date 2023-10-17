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

//go:build !windows

package scriptedinputsreceiver

import (
	_ "embed"
	"fmt"
	"strings"
)

const includePattern = ". \"$(dirname \"$0\")\"/common.sh"

//go:embed scripts/bandwidth.sh
var bandwidthScript string

//go:embed scripts/cpu.sh
var cpuScript string

//go:embed scripts/df.sh
var dfScript string

//go:embed scripts/hardware.sh
var hardwareScript string

//go:embed scripts/interfaces.sh
var interfacesScript string

//go:embed scripts/iostat.sh
var iostatScript string

//go:embed scripts/lastlog.sh
var lastlogScript string

//go:embed scripts/lsof.sh
var lsofScript string

//go:embed scripts/netstat.sh
var netstatScript string

//go:embed scripts/nfsiostat.sh
var nfsiostatScript string

//go:embed scripts/openPorts.sh
var openPortsScript string

//go:embed scripts/openPortsEnhanced.sh
var openPortsEnhancedScript string

//go:embed scripts/package.sh
var packageScript string

//go:embed scripts/passwd.sh
var passwdScript string

//go:embed scripts/protocol.sh
var protocolScript string

//go:embed scripts/ps.sh
var psScript string

//go:embed scripts/rlog.sh
var rlogScript string

//go:embed scripts/selinuxChecker.sh
var selinuxCheckerScript string

//go:embed scripts/service.sh
var serviceScript string

//go:embed scripts/sshdChecker.sh
var sshdCheckerScript string

//go:embed scripts/time.sh
var timeScript string

//go:embed scripts/top.sh
var topScript string

//go:embed scripts/update.sh
var updateScript string

//go:embed scripts/uptime.sh
var uptimeScript string

//go:embed scripts/usersWithLoginPrivs.sh
var usersWithLoginPrivsScript string

//go:embed scripts/version.sh
var versionScript string

//go:embed scripts/vmstat.sh
var vmstatScript string

//go:embed scripts/vsftpdChecker.sh
var vsftpdCheckerScript string

//go:embed scripts/who.sh
var whoScript string

//go:embed scripts/common.sh
var commonScript string

var scripts = func() map[string]string {
	scriptsMap := map[string]string{}
	for _, s := range []struct {
		name    string
		content string
	}{
		{"bandwidth", bandwidthScript},
		{"cpu", cpuScript},
		{"df", dfScript},
		{"hardware", hardwareScript},
		{"interfaces", interfacesScript},
		{"iostat", iostatScript},
		{"lastlog", lastlogScript},
		{"lsof", lsofScript},
		{"netstat", netstatScript},
		{"nfsiostat", nfsiostatScript},
		{"openPorts", openPortsScript},
		{"openPortsEnhanced", openPortsEnhancedScript},
		{"package", packageScript},
		{"passwd", passwdScript},
		{"protocol", protocolScript},
		{"ps", psScript},
		{"rlog", rlogScript},
		{"selinuxChecker", selinuxCheckerScript},
		{"service", serviceScript},
		{"sshdChecker", sshdCheckerScript},
		{"time", timeScript},
		{"top", topScript},
		{"update", updateScript},
		{"uptime", uptimeScript},
		{"usersWithLoginPrivs", usersWithLoginPrivsScript},
		{"version", versionScript},
		{"vmstat", vmstatScript},
		{"vsftpdChecker", vsftpdCheckerScript},
		{"who", whoScript},
	} {
		if _, ok := scriptsMap[s.name]; ok {
			panic(fmt.Errorf("duplicate script_name %q detected", s.name))
		}
		scriptsMap[s.name] = replaceCommon(s.content)
	}

	return scriptsMap
}()

func replaceCommon(script string) string {
	return strings.Replace(script, includePattern, commonScript, 1)
}
