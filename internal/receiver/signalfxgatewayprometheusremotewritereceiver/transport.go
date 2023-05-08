// Copyright 2020, OpenTelemetry Authors
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

package signalfxgatewayprometheusremotewritereceiver

import (
	"context"
)

// reporter is used to report (via zPages, logs, metrics, etc) the events
// happening when the Server is receiving and processing data.
type reporter interface {

	// OnError is used to report a translation error from original
	// format to the internal format of the Collector. The context
	// passed to it should be the ones returned by StartMetricsOp.
	OnError(ctx context.Context, reason string, err error)

	// OnMetricsProcessed is called when the received data is passed to next
	// consumer on the pipeline. The context passed to it should be the
	// one returned by StartMetricsOp. The error should be error returned by
	// the next consumer - the otelReporter is expected to handle nil error too.
	OnMetricsProcessed(ctx context.Context, numReceivedMessages int, err error)

	// OnDebugf allows less structured reporting for debugging scenarios.
	OnDebugf(
		template string,
		args ...interface{})

	// StartMetricsOp should always be called first, and the context from such passed onto the calls for
	// OnError (if an issue occurs) xor OnMetricsProcessed (if successful)
	StartMetricsOp(ctx context.Context) context.Context
}
