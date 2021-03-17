// Copyright 2021, OpenTelemetry Authors
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
	"go.opentelemetry.io/collector/consumer/pdata"
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
