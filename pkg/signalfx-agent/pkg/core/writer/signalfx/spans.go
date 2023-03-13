package signalfx

import (
	"context"
	"encoding/json"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/trace"
	log "github.com/sirupsen/logrus"

	apmtracker "github.com/signalfx/signalfx-agent/pkg/apm/tracetracker"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func (sw *Writer) sendSpans(ctx context.Context, spans []*trace.Span) error {
	if sw.serviceTracker != nil {
		sw.serviceTracker.AddSpans(sw.ctx, spans)
	}

	if sw.client != nil {
		// This sends synchonously
		err := sw.client.AddSpans(context.Background(), spans)
		if err != nil {
			var meta log.Fields
			if sw.conf.LogTraceSpansFailedToShip {
				jsonEncodedSpans, _ := json.Marshal(spans)
				meta = log.Fields{
					"error":   err,
					"payload": string(jsonEncodedSpans),
				}
			} else {
				meta = log.Fields{
					"error": err,
				}
			}

			log.WithFields(meta).Error("Error shipping spans to SignalFx")

			// If there is an error sending spans then just forget about them.
			return err
		}
		log.Debugf("Sent %d spans out of the agent", len(spans))
	}
	return nil
}

func (sw *Writer) processSpan(span *trace.Span) bool {
	if !sw.PreprocessSpan(span) {
		return false
	}

	sw.spanSourceTracker.AddSourceTagsToSpan(span)

	if sw.conf.LogTraceSpans {
		jsonEncoded, _ := json.Marshal(span)
		log.Infof("Sending trace span:\n%s", string(jsonEncoded))
	}

	return true
}

func (sw *Writer) startHostCorrelationTracking() *apmtracker.ActiveServiceTracker {
	var sendTraceHostCorrelationMetrics bool
	if sw.conf.SendTraceHostCorrelationMetrics != nil {
		sendTraceHostCorrelationMetrics = *sw.conf.SendTraceHostCorrelationMetrics
	}

	tracker := apmtracker.New(
		utils.NewAPMShim(log.StandardLogger()),
		sw.conf.StaleServiceTimeout.AsDuration(),
		sw.correlationClient,
		sw.hostIDDims,
		sendTraceHostCorrelationMetrics,
		func(dp *datapoint.Datapoint) {
			// Immediately send correlation datapoints when we first see a service
			sw.dpChan <- []*datapoint.Datapoint{dp}
		},
		apmtracker.DefaultDimsToSyncSource,
	)

	// purge the active service tracker periodically
	utils.RunOnInterval(sw.ctx, func() {
		tracker.Purge()
	}, sw.conf.TraceHostCorrelationPurgeInterval.AsDuration())

	// Send the correlation datapoints at a regular interval to keep the
	// service live on the backend.
	utils.RunOnInterval(sw.ctx, func() {
		for _, dp := range tracker.CorrelationDatapoints() {
			sw.dpChan <- []*datapoint.Datapoint{dp}
		}
	}, sw.conf.TraceHostCorrelationMetricsInterval.AsDuration())

	return tracker
}
