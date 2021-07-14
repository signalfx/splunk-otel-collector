// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type Converter struct {
	logger *zap.Logger
}

func NewConverter(logger *zap.Logger) Converter {
	return Converter{logger: logger}
}

func (c Converter) DatapointsToPDataMetrics(datapoints []*datapoint.Datapoint, timeReceived time.Time) (pdata.Metrics, int) {
	return sfxDatapointsToPDataMetrics(datapoints, timeReceived, c.logger)
}

func (c Converter) EventToPDataLogs(event *event.Event) pdata.Logs {
	return sfxEventToPDataLogs(event, c.logger)
}

func (c Converter) SpansToPDataTraces(spans []*trace.Span) pdata.Traces {
	traces, err := sfxSpansToPDataTraces(spans, c.logger)
	if err != nil {
		c.logger.Error("error converting SFx spans to pdata.Traces", zap.Error(err))
	}
	return traces
}
