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
	"go.opentelemetry.io/collector/confmap"
	"log"
	"regexp"
)

type NormalizeGcp struct{}

func (NormalizeGcp) Convert(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	const resourceDetector = "processors::resourcedetection(/\\w+)?::detectors"
	resourceDetectorRE := regexp.MustCompile(resourceDetector)
	out := map[string]any{}
	nonNormalizedGcpDetectorFound := false

	for _, k := range in.AllKeys() {
		v := in.Get(k)
		match := resourceDetectorRE.FindStringSubmatch(k)
		if match != nil {
			switch v.(type) {
			case []interface{}:
				vArr := v.([]interface{})
				normalizedV := make([]interface{}, 0, len(vArr))
				for _, item := range vArr {
					switch item.(type) {
					case string:
						if item == "gce" || item == "gke" {
							if !nonNormalizedGcpDetectorFound {
								normalizedV = append(normalizedV, "gcp")
							}
							nonNormalizedGcpDetectorFound = true
						} else if item != nil {
							normalizedV = append(normalizedV, item)
						}
					default:
						if item != nil {
							normalizedV = append(normalizedV, item)
						}
					}
				}
				v = normalizedV
			}
			out[k] = v
		}
	}
	if nonNormalizedGcpDetectorFound {
		log.Println("[WARNING] `processors` -> `resourcedetection` -> `detectors` parameter " +
			"contains a deprecated configuration. Please update the config according to the guideline: " +
			"https://github.com/signalfx/splunk-otel-collector#from-0680-to-0690.")
	}

	in.Merge(confmap.NewFromStringMap(out))
	return nil
}
