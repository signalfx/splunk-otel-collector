package types

import (
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/event"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/trace"     //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
)

// Output is the interface that monitors should use to send data to the agent
// core.  It handles adding the proper dimensions and metadata to datapoints so
// that monitors don't have to worry about it themselves.

type Output interface {
	Copy() Output
	SendDatapoints(dps ...*datapoint.Datapoint)
	SendMetrics(metrics ...pmetric.Metric)
	SendEvent(e *event.Event)
	SendSpans(spans ...*trace.Span)
	SendDimensionUpdate(dim *Dimension)
	AddExtraDimension(key, value string)
}

// FilteringOutput is Output enhanced with additional filtering mechanisms.
type FilteringOutput interface {
	Output
	AddDatapointExclusionFilter(filter dpfilters.DatapointFilter)
	EnabledMetrics() []string
	HasEnabledMetricInGroup(group string) bool
	HasAnyExtraMetrics() bool
}
