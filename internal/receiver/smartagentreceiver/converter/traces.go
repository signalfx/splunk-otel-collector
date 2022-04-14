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
	"net"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin/zipkinv2"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"github.com/signalfx/golib/v3/trace"
	sfxConstants "github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

const resourceClientIPAttrName = "ip"

var zipkinv2Translator = zipkinv2.ToTranslator{ParseStringTags: false}

func sfxSpansToPDataTraces(spans []*trace.Span, logger *zap.Logger) (ptrace.Traces, error) {
	// we batch sfx spans by the client IP that reported the spans to collector. Each
	// batch gets translated separately to ensure that spans from sources with different
	// IPs don't get bundled together under the same resource.
	batches := map[string][]*trace.Span{}
	for _, span := range spans {
		if span == nil {
			continue
		}
		var sourceIP string
		if val, ok := span.Meta[sfxConstants.DataSourceIPKey]; ok {
			if ip, ok := val.(net.IP); ok {
				sourceIP = ip.String()
			}
		}
		batches[sourceIP] = append(batches[sourceIP], span)
	}

	var lastErr error
	traces := ptrace.NewTraces()
	rss := traces.ResourceSpans()
	for ip, s := range batches {
		// SFx trace is effectively zipkin, so more convenient to convert to it and then rely
		// on existing zipkin receiver translator
		translated, err := zipkinv2Translator.ToTraces(sfxToZipkinSpans(s, logger))
		if err != nil {
			lastErr = err
			continue
		}
		trss := translated.ResourceSpans()
		if ip != "" {
			for i := 0; i < trss.Len(); i++ {
				trss.At(i).Resource().Attributes().UpsertString(resourceClientIPAttrName, ip)
			}
		}
		translated.ResourceSpans().MoveAndAppendTo(rss)
	}
	return traces, lastErr
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
