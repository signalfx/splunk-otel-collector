// Copyright Splunk, Inc.
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
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

type convFact struct {
	conv confmap.Converter
}

func (cf *convFact) Create(confmap.ConverterSettings) confmap.Converter {
	return cf.conv
}

type conv struct {
	convertFunc func(context.Context, *confmap.Conf) error
}

func (c *conv) Convert(ctx context.Context, cfg *confmap.Conf) error {
	return c.convertFunc(ctx, cfg)
}

// ConverterFactoryFromFunc creates a ConverterFactory from a convert function.
func ConverterFactoryFromFunc(f func(context.Context, *confmap.Conf) error) confmap.ConverterFactory {
	return &convFact{conv: &conv{convertFunc: f}}
}

// ConverterFactoryFromConverter creates a ConverterFactory from a Converter.
func ConverterFactoryFromConverter(conv confmap.Converter) confmap.ConverterFactory {
	return &convFact{conv: conv}
}

func getMetricsPipelineAndReceivers(pipelines map[string]any) (map[string]any, []any, error) {
	metricsPipeline := map[string]any{}
	if mp, ok := pipelines["metrics"]; ok && mp != nil {
		if metricsPipeline, ok = mp.(map[string]any); !ok {
			return nil, nil, fmt.Errorf("metrics pipeline is of unexpected form (%T): %v", mp, mp)
		}
	}
	pipelines["metrics"] = metricsPipeline

	var metricsReceivers []any
	if mr, ok := metricsPipeline["receivers"]; ok && mr != nil {
		var err error
		if metricsReceivers, err = toAnySlice(mr); err != nil {
			return nil, nil, fmt.Errorf("cannot determine metrics pipeline receivers: %w", err)
		}
	}
	return metricsPipeline, metricsReceivers, nil
}

func getPipelinesFromService(service map[string]any) (map[string]any, error) {
	pipelines := map[string]any{}
	if s, hasPipelines := service["pipelines"]; hasPipelines && s != nil {
		var ok bool
		if pipelines, ok = s.(map[string]any); !ok {
			return nil, fmt.Errorf("pipelines is of unexpected form (%T): %v", s, s)
		}
	}
	return pipelines, nil
}

func getService(out map[string]any) (map[string]any, error) {
	service := map[string]any{}
	if s, hasService := out["service"]; hasService && s != nil {
		var ok bool
		if service, ok = s.(map[string]any); !ok {
			return nil, fmt.Errorf("service is of unexpected form (%T): %v", s, s)
		}
	}
	return service, nil
}

func getExtensionsFromService(service map[string]any) ([]any, error) {
	var serviceExtensions []any
	if ses, hasExtensions := service["extensions"]; hasExtensions && ses != nil {
		var err error
		if serviceExtensions, err = toAnySlice(ses); err != nil {
			return nil, fmt.Errorf("cannot determine service extensions: %w", err)
		}
	}
	return serviceExtensions, nil
}

func isSignalFxExporter(exporter string) bool {
	return strings.EqualFold(exporter, "signalfx") || strings.HasPrefix(strings.ToLower(exporter), "signalfx/")
}

func isTracePipeline(pipeline string) bool {
	return strings.EqualFold(pipeline, "traces") || strings.HasPrefix(strings.ToLower(pipeline), "traces/")
}

func toAnySlice(s any) ([]any, error) {
	var out []any
	switch v := s.(type) {
	case []any:
		out = v
	case []string:
		for _, i := range v {
			out = append(out, i)
		}
	default:
		return nil, fmt.Errorf("unexpected form %T", s)
	}
	return out, nil
}
