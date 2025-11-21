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
	"strings"

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

// SetupDiscovery will find `service::<extensions|receivers>/splunk.discovery` entries
// provided by the discovery confmap.Provider and relocate them to
// `service::extensions` and `service::pipelines::metrics::receivers`,
// by appending them to existing sequences, if any.
func SetupDiscovery(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := in.ToStringMap()
	service, serviceExtensions, err := getServiceExtensions(out)
	if err != nil {
		return err
	}
	out["service"] = service

	discoExtensionsIsSet, discoExtensions, err := getDiscoExtensions(service)
	if err != nil {
		return err
	}

	discoReceiversIsSet, discoReceivers, err := getDiscoReceivers(service)
	if err != nil {
		return err
	}

	// do nothing if discovery provider didn't modify config
	if !discoExtensionsIsSet && !discoReceiversIsSet {
		return nil
	}

	if len(discoExtensions) > 0 {
		service["extensions"] = appendUnique(serviceExtensions, discoExtensions)
	}

	pipelines := map[string]any{}
	if pl, ok := service["pipelines"]; ok && pl != nil {
		pipelines = pl.(map[string]any)
	}
	service["pipelines"] = pipelines

	metricsPipeline, metricsReceivers, err := getMetricsPipelineAndReceivers(pipelines)
	if err != nil {
		return err
	}

	if len(discoReceivers) > 0 {
		metricsPipeline["receivers"] = appendUnique(metricsReceivers, discoReceivers)
	}

	setEntitiesPipelineReceivers(pipelines, discoReceivers)

	setAutoDiscoveryResourceAttribute(service)

	*in = *confmap.NewFromStringMap(out)
	return nil
}

func setAutoDiscoveryResourceAttribute(service map[string]any) {
	telemetry := map[string]any{}
	if tel, ok := service["telemetry"]; ok {
		telemetry = tel.(map[string]any)
	}
	service["telemetry"] = telemetry

	resAttrs := map[string]any{}
	if ra, ok := telemetry["resource"]; ok {
		resAttrs = ra.(map[string]any)
	}
	telemetry["resource"] = resAttrs

	resAttrs["splunk_autodiscovery"] = "true"
}

func getServiceExtensions(out map[string]any) (map[string]any, []any, error) {
	service := map[string]any{}
	var serviceExtensions []any
	if s, hasService := out["service"]; hasService && s != nil {
		var ok bool
		if service, ok = s.(map[string]any); !ok {
			return nil, nil, fmt.Errorf("service is of unexpected form (%T): %v", s, s)
		}
		if ses, hasExtensions := service["extensions"]; hasExtensions && ses != nil {
			var err error
			if serviceExtensions, err = toAnySlice(ses); err != nil {
				return nil, nil, fmt.Errorf("cannot determine service extensions: %w", err)
			}
		}
	}
	return service, serviceExtensions, nil
}

func getDiscoExtensions(service map[string]any) (bool, []any, error) {
	var isSet bool
	var extensions []any
	if des, hasDiscoExtensions := service[discovery.DiscoExtensionsKey]; hasDiscoExtensions {
		isSet = true
		delete(service, discovery.DiscoExtensionsKey)
		var err error
		if extensions, err = toAnySlice(des); err != nil {
			return false, nil, fmt.Errorf("cannot determine discovery extensions: %w", err)
		}
	}
	return isSet, extensions, nil
}

func getDiscoReceivers(service map[string]any) (bool, []any, error) {
	var isSet bool
	var receivers []any
	if des, hasDiscoReceivers := service[discovery.DiscoReceiversKey]; hasDiscoReceivers {
		isSet = true
		delete(service, discovery.DiscoReceiversKey)
		var err error
		if receivers, err = toAnySlice(des); err != nil {
			return false, nil, fmt.Errorf("cannot determine discovery receivers: %w", err)
		}
	}
	return isSet, receivers, nil
}

func getMetricsPipelineAndReceivers(pipelines map[string]any) (map[string]any, []any, error) {
	metricsPipeline := map[string]any{}
	if mp, ok := pipelines["metrics"]; ok && mp != nil {
		metricsPipeline = mp.(map[string]any)
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

func setEntitiesPipelineReceivers(pipelines map[string]any, discoReceivers []any) {
	ep, ok := pipelines["logs/entities"].(map[string]any)
	if !ok {
		// Entities pipeline not set, nothing to do.
		return
	}

	var receivers []any
	if existing, ok := ep["receivers"]; ok && existing != nil {
		receivers = existing.([]any)
	}
	for _, r := range discoReceivers {
		if strings.HasPrefix(r.(string), "discovery") {
			receivers = appendUnique(receivers, []any{r})
		}
	}
	ep["receivers"] = receivers
}

func appendUnique(serviceComponents, discoComponents []any) []any {
	existing := map[any]struct{}{}
	for _, e := range serviceComponents {
		existing[e] = struct{}{}
	}
	for _, e := range discoComponents {
		if _, exists := existing[e]; !exists {
			serviceComponents = append(serviceComponents, e)
		}
	}
	return serviceComponents
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
