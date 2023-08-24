// Copyright Copyright Splunk, Inc.
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

//go:embed scripts/cpu.sh
var cpuScript string

//go:embed scripts/df.sh
var dfScript string

//go:embed scripts/ps.sh
var psScript string

//go:embed scripts/common.sh
var commonScript string

var scripts = func() map[string]string {
	scripts := map[string]string{}
	for _, s := range []struct {
		name    string
		content string
	}{
		{"cpu", cpuScript},
		{"df", dfScript},
		{"ps", psScript},
	} {
		if _, ok := scripts[s.name]; ok {
			panic(fmt.Errorf("duplicate script_name %q detected", s.name))
		}
		scripts[s.name] = replaceCommon(s.content)
	}

	return scripts
}()

func replaceCommon(script string) string {
	return strings.Replace(script, includePattern, commonScript, 1)
}
