// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"log"
	"regexp"

	"go.opentelemetry.io/collector/config/configparser"
)

// RemoveBallastKey is a CfgMapFunc that removes a ballast_size_mib on a
// memory_limiter processor config if it exists. This config key will go away at
// some point (or already has) at which point its presence in a config will
// prevent the Collector from starting.
func RemoveBallastKey(cfgMap *configparser.ConfigMap) *configparser.ConfigMap {
	const expr = "processors::memory_limiter(/\\w+)?::ballast_size_mib"
	ballastKeyRegexp, _ := regexp.Compile(expr)

	out := configparser.NewParser()
	for _, k := range cfgMap.AllKeys() {
		if ballastKeyRegexp.MatchString(k) {
			log.Println("[WARNING] `ballast_size_mib` parameter in `memory_limiter` processor is " +
				"deprecated. Please update the config according to the guideline: " +
				"https://github.com/signalfx/splunk-otel-collector#from-0340-to-0350.")
		} else {
			out.Set(k, cfgMap.Get(k))
		}
	}
	return out
}
