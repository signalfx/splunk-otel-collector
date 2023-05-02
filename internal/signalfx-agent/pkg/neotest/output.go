package neotest

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// TestOutput can be used in place of the normal monitor outut to provide a
// simpler way of testing monitor output.
type TestOutput struct {
	dpChan    chan []*datapoint.Datapoint
	eventChan chan *event.Event
	spanChan  chan []*trace.Span
	dimChan   chan *types.Dimension
}

// NewTestOutput creates a new initialized TestOutput instance
func NewTestOutput() *TestOutput {
	return &TestOutput{
		dpChan:    make(chan []*datapoint.Datapoint, 1000),
		eventChan: make(chan *event.Event, 1000),
		spanChan:  make(chan []*trace.Span, 1000),
		dimChan:   make(chan *types.Dimension, 1000),
	}
}

// Copy the output object
func (to *TestOutput) Copy() types.Output {
	return to
}

func (to *TestOutput) SendDatapoints(dps ...*datapoint.Datapoint) {
	to.dpChan <- dps
}

// SendEvent accepts an event and sticks it in a buffered queue
func (to *TestOutput) SendEvent(event *event.Event) {
	to.eventChan <- event
}

// SendSpans accepts a trace span and sticks it in a buffered queue
func (to *TestOutput) SendSpans(spans ...*trace.Span) {
	for i := range spans {
		if spans[i].Meta == nil {
			spans[i].Meta = map[interface{}]interface{}{}
		}
	}
	to.spanChan <- spans
}

// SendDimensionUpdate accepts a dim prop update and sticks it in a buffered queue
func (to *TestOutput) SendDimensionUpdate(dims *types.Dimension) {
	to.dimChan <- dims
}

// AddExtraDimension is a noop here
func (to *TestOutput) AddExtraDimension(key, value string) {}

// RemoveExtraDimension is a noop here
func (to *TestOutput) RemoveExtraDimension(key string) {}

// AddExtraSpanTag is a noop here
func (to *TestOutput) AddExtraSpanTag(key, value string) {}

// RemoveExtraSpanTag is a noop here
func (to *TestOutput) RemoveExtraSpanTag(key string) {}

// AddDefaultSpanTag is a noop here
func (to *TestOutput) AddDefaultSpanTag(key, value string) {}

// RemoveDefaultSpanTag is a noop here
func (to *TestOutput) RemoveDefaultSpanTag(key string) {}

// FlushDatapoints returns all of the datapoints injected into the channel so
// far.
func (to *TestOutput) FlushDatapoints() []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	for {
		select {
		case dps := <-to.dpChan:
			out = append(out, dps...)
		default:
			return out
		}
	}
}

// FlushEvents returns all of the datapoints injected into the channel so
// far.
func (to *TestOutput) FlushEvents() []*event.Event {
	var out []*event.Event
	for {
		select {
		case event := <-to.eventChan:
			out = append(out, event)
		default:
			return out
		}
	}
}

// FlushSpans returns all of the spans injected into the channel so far
func (to *TestOutput) FlushSpans() []*trace.Span {
	var out []*trace.Span
	for {
		select {
		case span := <-to.spanChan:
			out = append(out, span...)
		default:
			return out
		}
	}
}

// WaitForDPs will keep pulling datapoints off of the internal queue until it
// either gets the expected count or waitSeconds seconds have elapsed.  It then
// returns those datapoints.  It will never return more than 'count' datapoints.
func (to *TestOutput) WaitForDPs(count, waitSeconds int) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint
	timeout := time.After(time.Duration(waitSeconds) * time.Second)

loop:
	for {
		select {
		case latestDPs := <-to.dpChan:
			dps = append(dps, latestDPs...)
			if len(dps) >= count {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	return dps
}

// WaitForDimensionProps will keep pulling dimension property updates off of
// the internal queue until it either gets the expected count or waitSeconds
// seconds have elapsed.  It then returns those dimension property updates.  It
// will never return more than 'count' objects.
func (to *TestOutput) WaitForDimensions(count, waitSeconds int) []*types.Dimension {
	var dps []*types.Dimension
	timeout := time.After(time.Duration(waitSeconds) * time.Second)

loop:
	for {
		select {
		case dp := <-to.dimChan:
			dps = append(dps, dp)
			if len(dps) >= count {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	return dps
}

// AddDatapointExclusionFilter is a noop here.
func (to *TestOutput) AddDatapointExclusionFilter(f dpfilters.DatapointFilter) {
}
