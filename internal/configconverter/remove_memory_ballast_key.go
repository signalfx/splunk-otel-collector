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

func removeMemoryBallastStrElementFromSlice(strList []interface{}) []interface{} {
	ret := make([]interface{}, 0)
	for i, v := range strList {
		if v == "memory_ballast" {
			ret = append(ret, strList[:i]...)
			return append(ret, strList[i+1:]...)
		}
	}
	return strList
}

// RemoveMemoryBallastKey removes a memory_ballast on a extension config if it exists.
func RemoveMemoryBallastKey(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return fmt.Errorf("cannot RemoveMemoryBallastKey on nil *confmap.Conf")
	}

	const firstExp = "extensions::memory_ballast.*"
	firstRegExp := regexp.MustCompile(firstExp)
	const secondExp = "service::extensions"
	secondRegExp := regexp.MustCompile(secondExp)

	out := map[string]any{}
	for _, k := range cfgMap.AllKeys() {
		if firstRegExp.MatchString(k) {
			log.Println("[WARNING] `memory_ballast` extension is deprecated. Please remove it from your configuration. " +
				"See https://github.com/signalfx/splunk-otel-collector#from-0961-to-0970 for more details")
			continue
		}
		if secondRegExp.MatchString(k) {
			out[k] = cfgMap.Get(k)
			if extSlice, ok := out[k].([]any); ok {
				ret := removeMemoryBallastStrElementFromSlice(extSlice)
				out[k] = ret
			}
			continue
		}
		out[k] = cfgMap.Get(k)
	}
	*cfgMap = *confmap.NewFromStringMap(out)
	return nil
}
