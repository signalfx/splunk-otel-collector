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

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/featuregate"
)

const (
	dropTraceCorrelationPipelineFeatureGateID = "splunk.dropSignalFxTracesCorrelationPipeline.enabled"
)

var dropTraceCorrelationPipelineFeatureGate = featuregate.GlobalRegistry().MustRegister(
	dropTraceCorrelationPipelineFeatureGateID,
	featuregate.StageBeta,
	featuregate.WithRegisterDescription("When enabled (default), the SignalFx Exporter will be removed from all trace pipelines. "+
		"This functionality is no longer needed for trace correlation."),
	featuregate.WithRegisterFromVersion("v0.151.0"),
)

func DropSignalFxTracesExporterIfFeatureGateEnabled(_ context.Context, in *confmap.Conf) error {
	if in == nil || !dropTraceCorrelationPipelineFeatureGate.IsEnabled() {
		return nil
	}

	out := in.ToStringMap()

	service, err := getService(out)
	if err != nil {
		return err
	}

	pipelines, err := getPipelinesFromService(service)
	if err != nil {
		return err
	}
	if len(pipelines) == 0 {
		return nil
	}

	removeSignalFxExportersFromTracePipelines(pipelines)

	*in = *confmap.NewFromStringMap(out)
	return nil
}

func removeSignalFxExportersFromTracePipelines(pipelines map[string]any) {
	filteredPipelines := make(map[string]any, len(pipelines))

	for pipelineName, rawPipeline := range pipelines {
		if !isTracePipeline(pipelineName) {
			filteredPipelines[pipelineName] = rawPipeline
			continue
		}

		pipeline, ok := rawPipeline.(map[string]any)
		if !ok {
			continue
		}

		rawExporters, ok := pipeline["exporters"]
		if !ok || rawExporters == nil {
			filteredPipelines[pipelineName] = pipeline
			continue
		}

		exporters, err := toAnySlice(rawExporters)
		if err != nil {
			continue
		}

		filtered := make([]any, 0, len(exporters))
		for _, exporter := range exporters {
			exporterName, ok := exporter.(string)
			if !ok || isSignalFxExporter(exporterName) {
				continue
			}
			filtered = append(filtered, exporter)
		}

		if len(filtered) == 0 {
			continue
		}
		pipeline["exporters"] = filtered
		filteredPipelines[pipelineName] = pipeline
	}
	for pipelineName := range pipelines {
		delete(pipelines, pipelineName)
	}
	for pipelineName, pipeline := range filteredPipelines {
		pipelines[pipelineName] = pipeline
	}
}
