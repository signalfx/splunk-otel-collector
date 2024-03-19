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
	"context"
	"fmt"
	"log"
	"regexp"

	"go.opentelemetry.io/collector/confmap"
)

// RemoveMemoryBallastKey is a MapConverter that removes a memory_ballast on a
// extension config if it exists.
type RemoveMemoryBallastKey struct{}

func (RemoveMemoryBallastKey) Convert(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return fmt.Errorf("cannot RemoveMemoryBallastKey on nil *confmap.Conf")
	}

	//const expr = "extensions::memory_ballast(/\\w+)?::size_mib"
	const expr1 = "extensions::memory_ballast"
	memoryBallastKeyRegexp := regexp.MustCompile(expr1)

	out := map[string]any{}
	for _, k := range cfgMap.AllKeys() {
		if memoryBallastKeyRegexp.MatchString(k) {
			log.Println("[WARNING]  `memory_ballast` parameter in extension is " +
				"deprecated/removed. Please update the config according to the guideline: " +
				"https://github.com/signalfx/splunk-otel-collector#from-0340-to-0350.")
		} else {
			out[k] = cfgMap.Get(k)
		}
	}
	*cfgMap = *confmap.NewFromStringMap(out)
	return nil
}
