// Package signalfx contains the SignalFx writer.  The writer is responsible for
// sending datapoints and events to SignalFx ingest.  Ideally all data would
// flow through here, but right now a lot of it is written to ingest by
// collectd.
//
// The writer provides a channel that all monitors can submit datapoints on.
// All monitors should include the "monitorType" key in the `Meta` map of the
// datapoint for use in filtering.
package signalfx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/golib/v3/trace"
	sfxwriter "github.com/signalfx/signalfx-go/writer"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/apm/correlations"
	libtracker "github.com/signalfx/signalfx-agent/pkg/apm/tracetracker"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/dimensions"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/processor"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tap"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tracetracker"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	// There cannot be more than this many events queued to be sent at any
	// given time.  This should be big enough for any reasonable use case.
	eventBufferCapacity = 1000
)

// Writer is what sends events and datapoints to SignalFx ingest.  It
// receives events/datapoints on two buffered channels and writes them to
// SignalFx on a regular interval.
type Writer struct {
	*processor.Processor

	client            *sfxclient.HTTPSink
	correlationClient correlations.CorrelationClient
	dimensionClient   *dimensions.DimensionClient
	datapointWriter   *sfxwriter.DatapointWriter
	spanWriter        *sfxwriter.SpanWriter

	// Monitors should send events to this
	eventChan     chan *event.Event
	dimensionChan chan *types.Dimension

	ctx    context.Context
	cancel context.CancelFunc
	conf   *config.WriterConfig
	logger *utils.ThrottledLogger
	dpTap  *tap.DatapointTap

	// map that holds host-specific ids like AWSUniqueID
	hostIDDims map[string]string

	eventBuffer []*event.Event

	// Keeps track of what service names have been seen in trace spans that are
	// emitted by the agent
	serviceTracker    *libtracker.ActiveServiceTracker
	spanSourceTracker *tracetracker.SpanSourceTracker

	// Datapoints sent in the last minute
	datapointsLastMinute int64
	// Datapoints that tried to be sent but couldn't in the last minute
	datapointsFailedLastMinute int64
	// Events sent in the last minute
	eventsLastMinute int64
	// Spans sent in the last minute
	spansLastMinute int64

	dpChan            chan []*datapoint.Datapoint
	spanChan          chan []*trace.Span
	dpsFailedToSend   int64
	traceSpansDropped int64
	eventsSent        int64
	startTime         time.Time
}

// New creates a new un-configured writer
func New(conf *config.WriterConfig, dpChan chan []*datapoint.Datapoint, eventChan chan *event.Event,
	dimensionChan chan *types.Dimension, spanChan chan []*trace.Span,
	spanSourceTracker *tracetracker.SpanSourceTracker) (*Writer, error) {

	logger := utils.NewThrottledLogger(log.WithFields(log.Fields{"component": "writer"}), 20*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	dimensionClient, err := dimensions.NewDimensionClient(ctx, conf)
	if err != nil {
		cancel()
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        conf.MaxRequests,
			MaxIdleConnsPerHost: conf.MaxRequests,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	correlationClient, err := correlations.NewCorrelationClient(utils.NewAPMShim(log.StandardLogger()), ctx, client, config.ClientConfigFromWriterConfig(conf))
	if err != nil {
		cancel()
		return nil, err
	}

	sw := &Writer{
		Processor:         processor.New(conf),
		ctx:               ctx,
		cancel:            cancel,
		conf:              conf,
		logger:            logger,
		correlationClient: correlationClient,
		dimensionClient:   dimensionClient,
		hostIDDims:        conf.HostIDDims,
		eventChan:         eventChan,
		dimensionChan:     dimensionChan,
		startTime:         time.Now(),
		spanSourceTracker: spanSourceTracker,
		dpChan:            dpChan,
		spanChan:          spanChan,
	}

	sinkOptions := []sfxclient.HTTPSinkOption{}
	switch strings.ToLower(conf.TraceExportFormat) {
	case config.TraceExportFormatZipkin:
		sinkOptions = append(sinkOptions, sfxclient.WithZipkinTraceExporter())
	case config.TraceExportFormatSAPM:
		sinkOptions = append(sinkOptions, sfxclient.WithSAPMTraceExporter())
	default:
		return nil, fmt.Errorf("trace export format '%s' is not supported", conf.TraceExportFormat)
	}

	sw.client = sfxclient.NewHTTPSink(sinkOptions...)
	sw.client.AuthToken = conf.SignalFxAccessToken
	sw.client.AdditionalHeaders = conf.ExtraHeaders

	sw.client.Client.Timeout = conf.Timeout.AsDuration()

	sw.client.Client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: conf.MaxRequests,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	dpEndpointURL, err := conf.ParsedIngestURL().Parse("v2/datapoint")
	if err != nil {
		logger.WithFields(log.Fields{
			"error":     err,
			"ingestURL": conf.ParsedIngestURL().String(),
		}).Error("Could not construct datapoint ingest URL")
		return nil, err
	}
	sw.client.DatapointEndpoint = dpEndpointURL.String()

	eventEndpointURL := conf.ParsedEventEndpointURL()
	if eventEndpointURL == nil {
		var err error
		eventEndpointURL, err = conf.ParsedIngestURL().Parse("v2/event")
		if err != nil {
			logger.WithFields(log.Fields{
				"error":     err,
				"ingestURL": conf.ParsedIngestURL().String(),
			}).Error("Could not construct event ingest URL")
			return nil, err
		}
	}
	sw.client.EventEndpoint = eventEndpointURL.String()

	traceEndpointURL := conf.ParsedTraceEndpointURL()
	if traceEndpointURL == nil {
		var err error
		traceEndpointURL, err = conf.ParsedIngestURL().Parse(conf.DefaultTraceEndpointPath())
		if err != nil {
			logger.WithFields(log.Fields{
				"error":     err,
				"ingestURL": conf.ParsedIngestURL().String(),
			}).Error("Could not construct trace ingest URL")
			return nil, err
		}
	}
	sw.client.TraceEndpoint = traceEndpointURL.String()

	sw.datapointWriter = &sfxwriter.DatapointWriter{
		PreprocessFunc: sw.processDatapoint,
		SendFunc:       sw.sendDatapoints,
		OverwriteFunc: func() {
			sw.logger.ThrottledWarning(fmt.Sprintf("A datapoint was overwritten in the write buffer, please consider increasing the writer.maxDatapointsBuffered config option to something greater than %d", conf.MaxDatapointsBuffered))
		},
		MaxBatchSize: conf.DatapointMaxBatchSize,
		MaxRequests:  conf.MaxRequests,
		MaxBuffered:  conf.MaxDatapointsBuffered,
		InputChan:    sw.dpChan,
	}

	sw.spanWriter = &sfxwriter.SpanWriter{
		PreprocessFunc: sw.processSpan,
		SendFunc:       sw.sendSpans,
		MaxBatchSize:   conf.TraceSpanMaxBatchSize,
		MaxRequests:    conf.MaxRequests,
		MaxBuffered:    int(conf.MaxTraceSpansInFlight),
		InputChan:      sw.spanChan,
	}

	logger.Infof("Sending datapoints to %s", sw.client.DatapointEndpoint)
	logger.Infof("Sending events to %s", sw.client.EventEndpoint)
	logger.Infof("Sending trace spans to %s", sw.client.TraceEndpoint)

	return sw, nil
}

func (sw *Writer) Start() {
	// The only reason this is on the struct and not a local var is so we can
	// easily get diagnostic metrics from it
	sw.serviceTracker = sw.startHostCorrelationTracking()

	go sw.maintainLastMinuteActivity()

	sw.dimensionClient.Start()
	sw.correlationClient.Start()

	go sw.listenForEventsAndDimensionUpdates()

	sw.datapointWriter.Start(sw.ctx)
	sw.spanWriter.Start(sw.ctx)
}

func (sw *Writer) processDatapoint(dp *datapoint.Datapoint) bool {
	if !sw.PreprocessDatapoint(dp) {
		return false
	}

	utils.TruncateDimensionValuesInPlace(dp.Dimensions)

	if sw.conf.LogDatapoints {
		sw.logger.Debugf("Sending datapoint:\n%s", utils.DatapointToString(dp))
	}

	return true
}

func (sw *Writer) sendDatapoints(ctx context.Context, dps []*datapoint.Datapoint) error {
	// This sends synchronously and retries on transient connection errors
	err := sw.client.AddDatapoints(ctx, dps)
	if err != nil {
		if isTransientError(err) {
			sw.logger.Debugf("retrying datapoint submission after receiving temporary network error: %v\n", err)
			err = sw.client.AddDatapoints(ctx, dps)
		}
		if err != nil {
			sw.logger.WithFields(log.Fields{
				"error": utils.SanitizeHTTPError(err),
			}).Error("Error shipping datapoints to SignalFx")
			// If there is an error sending datapoints then just forget about them.
			return err
		}

	}

	sw.logger.Debugf("Sent %d datapoints out of the agent", len(dps))

	// dpTap.Accept handles the receiver being nil
	sw.dpTap.Accept(dps)

	return nil
}

// isTransientError will walk through errors wrapped by candidate
// and return true if any is temporary, ECONNRESET, or EOF.
func isTransientError(candidate error) bool {
	var isTemporary, isReset, isEOF bool
	for candidate != nil {
		if temp, ok := candidate.(interface{ Temporary() bool }); ok {
			if temp.Temporary() {
				isTemporary = true
				break
			}
		}

		if se, ok := candidate.(syscall.Errno); ok {
			if se == syscall.ECONNRESET {
				isReset = true
				break
			}
		}

		if candidate == io.EOF {
			isEOF = true
			break
		}
		candidate = errors.Unwrap(candidate)
	}

	return isTemporary || isReset || isEOF
}

func (sw *Writer) sendEvents(events []*event.Event) error {
	for i := range events {
		sw.PreprocessEvent(events[i])

		if sw.conf.LogEvents {
			sw.logger.WithFields(log.Fields{
				"event": spew.Sdump(events[i]),
			}).Debug("Sending event")
		}
	}

	if sw.client != nil {
		err := sw.client.AddEvents(context.Background(), events)
		if err != nil {
			return err
		}
	}
	sw.eventsSent += int64(len(events))
	sw.logger.Debugf("Sent %d events to SignalFx", len(events))

	return nil
}

func (sw *Writer) listenForEventsAndDimensionUpdates() {
	eventTicker := time.NewTicker(time.Duration(sw.conf.EventSendIntervalSeconds) * time.Second)
	defer eventTicker.Stop()

	initEventBuffer := func() {
		sw.eventBuffer = make([]*event.Event, 0, eventBufferCapacity)
	}
	initEventBuffer()

	for {
		select {
		case <-sw.ctx.Done():
			return

		case event := <-sw.eventChan:
			if len(sw.eventBuffer) > eventBufferCapacity {
				sw.logger.WithFields(log.Fields{
					"eventType":         event.EventType,
					"eventBufferLength": len(sw.eventBuffer),
				}).Error("Dropping event due to overfull buffer")
				continue
			}
			sw.eventBuffer = append(sw.eventBuffer, event)

		case <-eventTicker.C:
			if len(sw.eventBuffer) > 0 {
				go func(buf []*event.Event) {
					if err := sw.sendEvents(buf); err != nil {

						sw.logger.WithError(utils.SanitizeHTTPError(err)).Error("Error shipping events to SignalFx")
					}
				}(sw.eventBuffer)
				initEventBuffer()
			}
		case dim := <-sw.dimensionChan:
			if err := sw.dimensionClient.AcceptDimension(dim); err != nil {
				sw.logger.WithFields(log.Fields{
					"dimName":  dim.Name,
					"dimValue": dim.Value,
				}).WithError(utils.SanitizeHTTPError(err)).Warn("Dropping dimension update")
			}
		}
	}
}

// SetTap allows you to set one datapoint tap at a time to inspect datapoints
// going out of the agent.
func (sw *Writer) SetTap(dpTap *tap.DatapointTap) {
	sw.dpTap = dpTap
}

// Shutdown the writer and stop sending datapoints
func (sw *Writer) Shutdown() {
	if sw.cancel != nil {
		sw.cancel()
	}
	sw.logger.Debug("Stopped datapoint writer")
}
