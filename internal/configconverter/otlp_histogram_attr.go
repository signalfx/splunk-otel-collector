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
	"reflect"
	"regexp"

	"go.opentelemetry.io/collector/confmap"
)

type AddOTLPHistogramAttr struct{}

// Convert updates the service::telemetry::resource to add the attribute send_otlp_histograms=true.
// This additional resource attr is only added if we see any signalfx exporter in use with the config
// send_otlp_histograms set to true.
func (AddOTLPHistogramAttr) Convert(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return nil
	}

	sfxExporters := map[string]struct{}{}
	metricsPlRe := regexp.MustCompile(`^metrics($|/.*$)`)
	signalfxExpRe := regexp.MustCompile(`^signalfx($|/.*$)`)

	exp, err := cfgMap.Sub("exporters")
	if err != nil {
		return nil // Ignore invalid config. Rely on the config validation to catch this.
	}

	// get signalfx exporter names which have send_otlp_histograms enabled
	for expName := range exp.ToStringMap() {
		if signalfxExpRe.MatchString(expName) {
			otlpHistoEnabled := cfgMap.Get(fmt.Sprintf("exporters::%s::send_otlp_histograms", expName))
			if otlpHistoEnabled != nil && reflect.ValueOf(otlpHistoEnabled).Kind() == reflect.Bool {
				if otlpHistoEnabled.(bool) {
					sfxExporters[expName] = struct{}{}
				}
			}
		}
	}

	// no sfx exporters with send_otlp_histograms enabled
	if len(sfxExporters) == 0 {
		return nil
	}

	pl, err := cfgMap.Sub("service::pipelines")
	if err != nil {
		return nil
	}

	// check if metrics pipelines use any of the signalfx exporters which have send_otlp_histograms enabled
	for pipelineName := range pl.ToStringMap() {
		if metricsPlRe.MatchString(pipelineName) {
			mExps := cfgMap.Get(fmt.Sprintf("service::pipelines::%s::exporters", pipelineName))
			if metricsExps, ok := mExps.([]any); ok {
				for _, expName := range metricsExps {
					if expNameStr, ok := expName.(string); ok {
						if _, ok := sfxExporters[expNameStr]; ok {
							resAttrKey := "service::telemetry::resource::splunk_otlp_histograms"
							return cfgMap.Merge(confmap.NewFromStringMap(map[string]any{resAttrKey: "true"}))
						}
					}
				}
			}
		}
	}
	return nil
}
