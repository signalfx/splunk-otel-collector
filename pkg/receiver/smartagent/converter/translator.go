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
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type Translator struct {
	logger *zap.Logger
}

func NewTranslator(logger *zap.Logger) Translator {
	return Translator{logger: logger}
}

func (c Translator) ToMetrics(datapoints []*datapoint.Datapoint) (pmetric.Metrics, error) {
	return sfxDatapointsToPDataMetrics(datapoints, time.Now(), c.logger), nil
}

func (c Translator) ToLogs(event *event.Event) (plog.Logs, error) {
	return sfxEventToPDataLogs(event, c.logger), nil
}

func (c Translator) ToTraces(spans []*trace.Span) (ptrace.Traces, error) {
	return sfxSpansToPDataTraces(spans, c.logger)
}
