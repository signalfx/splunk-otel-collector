package splunk

//go:generate sh -c "`go env GOPATH`/bin/genny -ast -pkg splunk -in `go env GOPATH`/pkg/mod/github.com/signalfx/signalfx-go@v1.33.0/writer/template/ring.go gen Instance=logEntry | sed -e s/*logEntry/logEntry/g > ./log_event_ring.gen.go"

//go:generate sh -c "`go env GOPATH`/bin/genny -ast -pkg splunk -in `go env GOPATH`/pkg/mod/github.com/signalfx/signalfx-go@v1.33.0/writer/template/writer.go gen Instance=logEntry | sed -e s/*logEntry/logEntry/g > ./log_event_writer.gen.go"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/processor"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	unknownHost = "unknown"
)

// Output posts data to Splunk HTTP Event Collector.
type Output struct {
	*processor.Processor

	httpClient        *http.Client
	url               string
	token             string
	metricsSource     string
	metricsSourceType string
	metricsIndex      string
	eventsSource      string
	eventsSourceType  string
	eventsIndex       string
	skipTLSVerify     bool
	hostIDDims        map[string]string

	entryWriter *LogEntryWriter

	ctx    context.Context
	cancel context.CancelFunc

	dpChan    chan []*datapoint.Datapoint
	eventChan chan *event.Event
	spanChan  chan []*trace.Span
}

// Build a Splunk Writer.
func New(conf *config.WriterConfig, dpChan chan []*datapoint.Datapoint, eventChan chan *event.Event, spanChan chan []*trace.Span) (*Output, error) {
	out := &Output{
		Processor:         processor.New(conf),
		url:               conf.Splunk.URL,
		token:             conf.Splunk.Token,
		metricsSource:     conf.Splunk.MetricsSource,
		metricsSourceType: conf.Splunk.MetricsSourceType,
		metricsIndex:      conf.Splunk.MetricsIndex,
		eventsSource:      conf.Splunk.EventsSource,
		eventsSourceType:  conf.Splunk.EventsSourceType,
		eventsIndex:       conf.Splunk.EventsIndex,
		skipTLSVerify:     conf.Splunk.SkipTLSVerify,
		hostIDDims:        conf.HostIDDims,
		dpChan:            dpChan,
		eventChan:         eventChan,
		spanChan:          spanChan,
	}

	out.ctx, out.cancel = context.WithCancel(context.Background())

	httpConfig := httpclient.HTTPConfig{
		SkipVerify: conf.Splunk.SkipTLSVerify,
		UseHTTPS:   strings.HasPrefix(conf.Splunk.URL, "https"),
	}

	httpClient, err := httpConfig.Build()
	if err != nil {
		return nil, err
	}
	out.httpClient = httpClient

	out.entryWriter = &LogEntryWriter{
		SendFunc: func(ctx context.Context, entries []logEntry) error {
			err := out.sendToSplunk(ctx, entries)
			if err != nil {
				logrus.WithError(utils.SanitizeHTTPError(err)).Error("Failed to send to Splunk HEC")
			}
			return err
		},
		MaxBatchSize: conf.Splunk.MaxBatchSize,
		MaxRequests:  conf.Splunk.MaxRequests,
		MaxBuffered:  conf.Splunk.MaxBuffered,
		InputChan:    make(chan []logEntry, 1000),
	}

	return out, nil
}

func (o *Output) Start() {
	o.entryWriter.Start(o.ctx)

	logrus.Infof("Sending Splunk HEC entries to %s", o.url)

	go func() {
		for {
			var entries []logEntry

			select {
			case <-o.ctx.Done():
				return
			case dps := <-o.dpChan:
				for i := range dps {
					if !o.PreprocessDatapoint(dps[i]) {
						continue
					}
					entries = append(entries, o.convertDatapoint(dps[i]))
				}
			case event := <-o.eventChan:
				if !o.PreprocessEvent(event) {
					continue
				}
				entries = append(entries, o.convertEvent(event))
			case spans := <-o.spanChan:
				for _, span := range spans {
					if !o.PreprocessSpan(span) {
						continue
					}
					entries = append(entries, o.convertSpan(span))
				}
			}

			if len(entries) > 0 {
				o.entryWriter.InputChan <- entries
			}
		}
	}()
}

func toString(obj interface{}) string {
	if stringer, ok := obj.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%v", obj)
}

func computeTime(timestamp time.Time) int64 {
	if timestamp.IsZero() {
		return time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	}
	return timestamp.UnixNano() / time.Millisecond.Nanoseconds()
}

// LogDatapoint logs a datapoint as a Splunk metric event
func (o *Output) convertDatapoint(d *datapoint.Datapoint) *logMetric {
	fields := make(map[string]string)

	for key, v := range d.Meta {
		if v != nil {
			fields[toString(key)] = toString(v)
		}
	}
	for key, v := range d.Dimensions {
		fields[toString(key)] = toString(v)
	}

	fields["metric_type"] = d.MetricType.String()
	fields["metric_name:"+d.Metric] = d.Value.String()

	host := d.Dimensions["host"]
	if host == "" {
		host = unknownHost
	}
	return &logMetric{
		Time:       computeTime(d.Timestamp),
		Host:       host,
		Source:     o.metricsSource,
		SourceType: o.metricsSourceType,
		Index:      o.metricsIndex,
		Event:      "metric",
		Fields:     fields,
	}
}

// convertEvent converts an event as a Splunk event
func (o *Output) convertEvent(e *event.Event) *logEvent {
	props := make(map[string]string)
	for key, v := range e.Properties {
		if v != nil {
			props[key] = toString(v)
		}
	}

	meta := make(map[string]string)
	for key, v := range e.Meta {
		if v != nil {
			meta[toString(key)] = toString(v)
		}
	}
	host := e.Dimensions["host"]
	if host == "" {
		host = unknownHost
	}
	return &logEvent{
		Time:       computeTime(e.Timestamp),
		Host:       host,
		Source:     o.eventsSource,
		SourceType: o.eventsSourceType,
		Index:      o.eventsIndex,
		Event:      eventdata{Properties: props, Dimensions: e.Dimensions, Meta: meta, EventType: e.EventType, Category: e.Category},
	}
}

// convertSpan converts a span as a Splunk event
func (o *Output) convertSpan(s *trace.Span) *logSpan {

	meta := make(map[string]string)
	for key, v := range s.Meta {
		if v != nil {
			meta[toString(key)] = toString(v)
		}
	}

	b, err := json.Marshal(s.LocalEndpoint)
	var host string
	if err != nil {
		host = unknownHost
	} else {
		host = string(b)
	}

	ts := *s.Timestamp
	if ts == 0 {
		ts = time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	}

	return &logSpan{
		Time:       ts,
		Host:       host,
		Source:     o.eventsSource,
		SourceType: o.eventsSourceType,
		Index:      o.eventsIndex,
		Event:      spandata{TraceID: s.TraceID, ID: s.ID, Name: s.Name, LocalEndpoint: s.LocalEndpoint, RemoteEndpoint: s.RemoteEndpoint, Debug: s.Debug, Duration: s.Duration, Kind: s.Kind, ParentID: s.ParentID, Shared: s.Shared, Tags: s.Tags, Meta: meta, Annotations: s.Annotations},
	}
}

func (o *Output) sendToSplunk(ctx context.Context, entries []logEntry) error {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	for i := range entries {
		err := encoder.Encode(entries[i])
		if err != nil {
			return err
		}

		_, err = buf.WriteString("\r\n\r\n")
		if err != nil {
			return fmt.Errorf("failed to write line separator: %v", err)
		}
	}

	return o.doRequest(ctx, buf)
}

func (o *Output) doRequest(ctx context.Context, b io.Reader) error {
	url := o.url
	req, err := http.NewRequestWithContext(ctx, "POST", url, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk "+o.token)

	res, err := o.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		_, _ = io.Copy(ioutil.Discard, res.Body)
		return nil
	default:
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(res.Body)
		responseBody := buf.String()
		err = fmt.Errorf("non-200 response received (%d): %s", res.StatusCode, responseBody)
	}
	return err
}

func (o *Output) Shutdown() {
	if o.cancel != nil {
		o.cancel()
	}
}

// InternalMetrics returns a set of metrics showing how the writer is currently
// doing.
func (o *Output) InternalMetrics() []*datapoint.Datapoint {
	return o.entryWriter.InternalMetrics("splunk_writer.")
}
