package writer

import (
	"context"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/signalfx"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/splunk"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tap"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tracetracker"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// MultiWriter combines the SignalFx and Splunk outputs.
type MultiWriter struct {
	ctx    context.Context
	cancel context.CancelFunc

	signalFxWriter *signalfx.Writer
	splunkWriter   *splunk.Output
}

func New(conf *config.WriterConfig, dpChan chan []*datapoint.Datapoint, eventChan chan *event.Event,
	dimensionChan chan *types.Dimension, spanChan chan []*trace.Span,
	spanSourceTracker *tracetracker.SpanSourceTracker) (*MultiWriter, error) {

	w := new(MultiWriter)
	w.ctx, w.cancel = context.WithCancel(context.Background())

	bothEnabled := conf.IsSignalFxOutputEnabled() && conf.IsSplunkOutputEnabled()

	signalFxDPChan := dpChan
	splunkDPChan := dpChan
	signalFxEventChan := eventChan
	splunkEventChan := eventChan
	signalFxSpanChan := spanChan
	splunkSpanChan := spanChan

	// The channel handling is a bit hacky but we have to broadcast the
	// datapoints to both writers if they are both present given our existing
	// structure of sending datapoints through a channel.  If we ever get more
	// than two writers, definitely refactor all of this.
	if bothEnabled {
		signalFxDPChan = make(chan []*datapoint.Datapoint, cap(dpChan))
		splunkDPChan = make(chan []*datapoint.Datapoint, cap(dpChan))
		signalFxEventChan = make(chan *event.Event, cap(eventChan))
		splunkEventChan = make(chan *event.Event, cap(eventChan))
		splunkSpanChan = make(chan []*trace.Span, cap(spanChan))

		go func() {
			for {
				select {
				case <-w.ctx.Done():
					return
				case dps := <-dpChan:
					signalFxDPChan <- utils.CloneDatapointSlice(dps)
					splunkDPChan <- dps
				case ev := <-eventChan:
					signalFxEventChan <- utils.CloneEvent(ev)
					splunkEventChan <- ev
				case span := <-spanChan:
					signalFxSpanChan <- utils.CloneSpanSlice(span)
					splunkSpanChan <- span
				}
			}
		}()
	}

	if conf.IsSignalFxOutputEnabled() {
		var err error
		w.signalFxWriter, err = signalfx.New(conf, signalFxDPChan, signalFxEventChan, dimensionChan, signalFxSpanChan, spanSourceTracker)
		if err != nil {
			return nil, err
		}
	}

	if conf.IsSplunkOutputEnabled() {
		var err error
		w.splunkWriter, err = splunk.New(conf, splunkDPChan, splunkEventChan, splunkSpanChan)
		if err != nil {
			return nil, err
		}
	}

	return w, nil
}

func (w *MultiWriter) Start() {
	if w.signalFxWriter != nil {
		w.signalFxWriter.Start()
	}

	if w.splunkWriter != nil {
		w.splunkWriter.Start()
	}
}

func (w *MultiWriter) Shutdown() {
	if w.cancel != nil {
		w.cancel()
	}
	if w.signalFxWriter != nil {
		w.signalFxWriter.Shutdown()
	}
	if w.splunkWriter != nil {
		w.splunkWriter.Shutdown()
	}
}

func (w *MultiWriter) InternalMetrics() []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	if w.signalFxWriter != nil {
		dps = append(dps, w.signalFxWriter.InternalMetrics()...)
	}
	if w.splunkWriter != nil {
		dps = append(dps, w.splunkWriter.InternalMetrics()...)
	}

	return dps
}

func (w *MultiWriter) DiagnosticText() string {
	if w.signalFxWriter != nil {
		return w.signalFxWriter.DiagnosticText()
	}
	return "No writer information available"
}

// SetTap allows you to set one datapoint tap at a time to inspect datapoints
// going out of the agent.
func (w *MultiWriter) SetTap(dpTap *tap.DatapointTap) {
	if w.signalFxWriter != nil {
		w.signalFxWriter.SetTap(dpTap)
	}
}
