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
	"reflect"
	"regexp"

	"go.opentelemetry.io/collector/config"
)

// RenameK8sTagger will replace k8s_tagger processor items with k8sattributes ones.
func RenameK8sTagger(_ context.Context, in *config.Map) error {
	if in == nil {
		return fmt.Errorf("cannot RenameK8sTagger on nil *config.Map")
	}

	tagger := "k8s_tagger(/\\w+:{0,2})?"
	taggerRe := regexp.MustCompile(tagger)
	keyExpr := fmt.Sprintf("processors::%s(.+)?", tagger)
	k8sTaggerKeyRe := regexp.MustCompile(keyExpr)

	const serviceExpr = "service(.+)processors"
	serviceEntryRe := regexp.MustCompile(serviceExpr)

	found := false
	out := config.NewMap()
	for _, k := range in.AllKeys() {
		v := in.Get(k)
		if match := k8sTaggerKeyRe.FindStringSubmatch(k); match != nil {
			k8sAttributesKey := fmt.Sprintf("processors::k8sattributes%s%s", match[1], match[2])
			out.Set(k8sAttributesKey, v)
			if !found {
				log.Println("[WARNING] `k8s_tagger` processor was renamed to `k8sattributes`. Please update your config accordingly.")
			}
			found = true
		} else {
			if serviceEntryRe.MatchString(k) {
				kind := reflect.TypeOf(v).Kind()
				if kind == reflect.Slice {
					if sliceOfInterfaces, ok := v.([]interface{}); ok {
						for i, val := range sliceOfInterfaces {
							if strVal, ok := val.(string); ok {
								if match = taggerRe.FindStringSubmatch(strVal); match != nil {
									k8sAttributeEntry := fmt.Sprintf("k8sattributes%s", match[1])
									sliceOfInterfaces[i] = k8sAttributeEntry
								}
							}
						}
						v = sliceOfInterfaces
					}
				}
			}
			out.Set(k, v)
		}
	}
	*in = *out
	return nil
}
