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

func removeMemoryBallast(strList []interface{}) []interface{} {
	ret := make([]interface{}, 0)
	for i, v := range strList {
		if v == "memory_ballast" {
			ret = append(ret, strList[:i]...)
			return append(ret, strList[i+1:]...)
		}
	}
	return ret
}

func (RemoveMemoryBallastKey) Convert(_ context.Context, cfgMap *confmap.Conf) error {
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
			log.Println("[WARNING]  `memory_ballast` parameter in extensions is deprecated. Please remove it from your configuration.")
		} else {
			out[k] = cfgMap.Get(k)
			if secondRegExp.MatchString(k) {
				temp := out[k]
				switch temp := temp.(type) {
				default:
					continue
				case []interface{}:
					ret := removeMemoryBallast(temp)
					if len(ret) > 0 {
						out[k] = ret
					}
				}
			}
		}
	}
	*cfgMap = *confmap.NewFromStringMap(out)
	return nil
}
