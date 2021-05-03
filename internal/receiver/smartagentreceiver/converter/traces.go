// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"encoding/json"

	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"github.com/signalfx/golib/v3/trace"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/translator/trace/zipkin"
	"go.uber.org/zap"
)

func sfxSpansToPDataTraces(spans []*trace.Span, logger *zap.Logger) (pdata.Traces, error) {
	// SFx trace is effectively zipkin, so more convenient to convert to it and then rely
	// on existing zipkin receiver translator
	zipkinSpans := sfxToZipkinSpans(spans, logger)
	return zipkin.V2SpansToInternalTraces(zipkinSpans, false)
}

func sfxToZipkinSpans(spans []*trace.Span, logger *zap.Logger) []*zipkinmodel.SpanModel {
	var zipkinSpans []*zipkinmodel.SpanModel
	for _, span := range spans {
		if span == nil {
			continue
		}

		spanJSON, err := json.Marshal(span)
		if err != nil { // generally results from reflection panic, which is unlikely here
			logger.Debug("failed to marshall SFx span", zap.Error(err))
			continue
		}
		var zipkinSpan zipkinmodel.SpanModel
		err = json.Unmarshal(spanJSON, &zipkinSpan)
		if err != nil {
			logger.Debug("failed to unmarshall SFx span as Zipkin", zap.Error(err))
			continue
		}
		zipkinSpans = append(zipkinSpans, &zipkinSpan)
	}
	return zipkinSpans
}
