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
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

type SapmToOtlp struct{}

func (SapmToOtlp) Convert(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return fmt.Errorf("cannot SapmToOtlp on nil *confmap.Conf")
	}

	out := map[string]any{}
	sapmExporterFound := false
	sapmPipelineFound := false
	var traceExporters []any
	for _, k := range in.AllKeys() {
		v := in.Get(k)
		if k == "service::pipelines::traces::exporters" {
			if exportersList, ok := v.([]any); ok {
				for _, exporter := range exportersList {
					if exporter == "sapm" {
						sapmPipelineFound = true
					} else {
						traceExporters = append(traceExporters, exporter)
					}
				}
			}
		}
		if strings.HasPrefix(k, "exporters::sapm") {
			sapmExporterFound = true
		} else {
			out[k] = v
		}
	}
	if sapmExporterFound {
		log.Println(
			"[WARNING] `exporters` -> `sapm`" +
				"is deprecated. Please update your configuration to use otlp instead.",
		)
	}
	if sapmPipelineFound {
		log.Println(
			"[WARNING] `service` -> `pipelines` -> `traces` -> `exporters` -> `sapm`" +
				"is deprecated. Please update your configuration to use otlp instead.",
		)
	}
	if sapmExporterFound && sapmPipelineFound {
		out["exporters::otlp/fromsapm::endpoint"] = "${SPLUNK_INGEST_URL}"
		out["exporters::otlp/fromsapm::headers::X-SF-Token"] = "${SPLUNK_ACCESS_TOKEN}"
		traceExporters = append(traceExporters, "otlp/fromsapm")
		out["service::pipelines::traces::exporters"] = traceExporters

	}

	*in = *confmap.NewFromStringMap(out)
	return nil
}
