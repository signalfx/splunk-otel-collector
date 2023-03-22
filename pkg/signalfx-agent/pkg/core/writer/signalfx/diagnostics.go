package signalfx

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Call this in a goroutine to maintain a moving window average DPM, EPM, and
// SPM, updated every 10 seconds.
func (sw *Writer) maintainLastMinuteActivity() {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	var dpSamples [6]int64
	var dpFailedSamples [6]int64
	var eventSamples [6]int64
	var spanSamples [6]int64
	idx := 0
	for {
		select {
		case <-sw.ctx.Done():
			return
		case <-t.C:
			sw.datapointsLastMinute = atomic.LoadInt64(&sw.datapointWriter.TotalSent) - dpSamples[idx]
			dpSamples[idx] += sw.datapointsLastMinute

			sw.datapointsFailedLastMinute = atomic.LoadInt64(&sw.dpsFailedToSend) - dpFailedSamples[idx]
			dpFailedSamples[idx] += sw.datapointsFailedLastMinute

			sw.eventsLastMinute = atomic.LoadInt64(&sw.eventsSent) - eventSamples[idx]
			eventSamples[idx] += sw.eventsLastMinute

			sw.spansLastMinute = atomic.LoadInt64(&sw.spanWriter.TotalSent) - spanSamples[idx]
			spanSamples[idx] += sw.spansLastMinute

			idx = (idx + 1) % 6
		}
	}
}

// DiagnosticText outputs a string that describes the state of the writer to a
// human.
func (sw *Writer) DiagnosticText() string {
	return fmt.Sprintf(
		"Global Dimensions:                %s\n"+
			"GlobalSpanTags:                   %s\n"+
			"Datapoints sent (last minute):    %d\n"+
			"Datapoints failed (last minute):  %d\n"+
			"Datapoints overwritten (total):   %d\n"+
			"Events Sent (last minute):        %d\n"+
			"Trace Spans Sent (last minute):   %d\n"+
			"Trace Spans overwritten (total):  %d",
		utils.FormatStringMapCompact(utils.MergeStringMaps(sw.conf.GlobalDimensions, sw.hostIDDims)),
		sw.conf.GlobalSpanTags,
		sw.datapointsLastMinute,
		sw.datapointsFailedLastMinute,
		atomic.LoadInt64(&sw.datapointWriter.TotalOverwritten),
		sw.eventsLastMinute,
		sw.spansLastMinute,
		atomic.LoadInt64(&sw.spanWriter.TotalOverwritten))
}

// InternalMetrics returns a set of metrics showing how the writer is currently
// doing.
func (sw *Writer) InternalMetrics() []*datapoint.Datapoint {
	return append(append(append(append(append(append([]*datapoint.Datapoint{
		sfxclient.CumulativeP("sfxagent.events_sent", nil, &sw.eventsSent),
		sfxclient.Gauge("sfxagent.datapoint_channel_len", nil, int64(len(sw.dpChan))),
		sfxclient.Gauge("sfxagent.events_buffered", nil, int64(len(sw.eventBuffer))),
		sfxclient.CumulativeP("sfxagent.trace_spans_dropped", nil, &sw.traceSpansDropped),
	}, sw.datapointWriter.InternalMetrics("sfxagent.")...),
		sw.spanWriter.InternalMetrics("sfxagent.")...),
		sw.serviceTracker.InternalMetrics()...),
		sw.dimensionClient.InternalMetrics()...),
		sw.spanSourceTracker.InternalMetrics()...),
		sw.correlationClient.InternalMetrics()...)
}
