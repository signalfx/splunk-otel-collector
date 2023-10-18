package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpguts"

	"github.com/mitchellh/hashstructure"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/core/propfilters"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	log "github.com/sirupsen/logrus"
)

// WriterConfig holds configuration for the datapoint writer.
type WriterConfig struct {
	// The maximum number of datapoints to include in a batch before sending the
	// batch to the ingest server.  Smaller batch sizes than this will be sent
	// if datapoints originate in smaller chunks.
	DatapointMaxBatchSize int `yaml:"datapointMaxBatchSize" default:"1000"`
	// The maximum number of datapoints that are allowed to be buffered in the
	// agent (i.e. received from a monitor but have not yet received
	// confirmation of successful receipt by the target ingest/gateway server
	// downstream).  Any datapoints that come in beyond this number will
	// overwrite existing datapoints if they have not been sent yet, starting
	// with the oldest.
	MaxDatapointsBuffered int `yaml:"maxDatapointsBuffered" default:"25000"`
	// The analogue of `datapointMaxBatchSize` for trace spans.
	TraceSpanMaxBatchSize int `yaml:"traceSpanMaxBatchSize" default:"1000"`
	// Format to export traces in. Choices are "zipkin" and "sapm"
	TraceExportFormat string `yaml:"traceExportFormat" default:"zipkin"`
	// Deprecated: use `maxRequests` instead.
	DatapointMaxRequests int `yaml:"datapointMaxRequests"`
	// The maximum number of concurrent requests to make to a single ingest server
	// with datapoints/events/trace spans.  This number multiplied by
	// `datapointMaxBatchSize` is more or less the maximum number of datapoints
	// that can be "in-flight" at any given time.  Same thing for the
	// `traceSpanMaxBatchSize` option and trace spans.
	MaxRequests int `yaml:"maxRequests" default:"10"`
	// Timeout specifies a time limit for requests made to the ingest server.
	// The timeout includes connection time, any redirects, and reading the response body.
	// Default is 5 seconds, a Timeout of zero means no timeout.
	Timeout timeutil.Duration `yaml:"timeout" default:"5s"`
	// The agent does not send events immediately upon a monitor generating
	// them, but buffers them and sends them in batches.  The lower this
	// number, the less delay for events to appear in SignalFx.
	EventSendIntervalSeconds int `yaml:"eventSendIntervalSeconds" default:"1"`
	// The analogue of `maxRequests` for dimension property requests.
	PropertiesMaxRequests uint `yaml:"propertiesMaxRequests" default:"20"`
	// How many dimension property updates to hold pending being sent before
	// dropping subsequent property updates.  Property updates will be resent
	// eventually and they are slow to change so dropping them (esp on agent
	// start up) usually isn't a big deal.
	PropertiesMaxBuffered uint `yaml:"propertiesMaxBuffered" default:"10000"`
	// How long to wait for property updates to be sent once they are
	// generated.  Any duplicate updates to the same dimension within this time
	// frame will result in the latest property set being sent.  This helps
	// prevent spurious updates that get immediately overwritten by very flappy
	// property generation.
	PropertiesSendDelaySeconds uint `yaml:"propertiesSendDelaySeconds" default:"30"`
	// Properties that are synced to SignalFx are cached to prevent duplicate
	// requests from being sent, causing unnecessary load on our backend.
	PropertiesHistorySize uint `yaml:"propertiesHistorySize" default:"10000"`
	// If the log level is set to `debug` and this is true, all datapoints
	// generated by the agent will be logged.
	LogDatapoints bool `yaml:"logDatapoints"`
	// The analogue of `logDatapoints` for events.
	LogEvents bool `yaml:"logEvents"`
	// The analogue of `logDatapoints` for trace spans.
	LogTraceSpans bool `yaml:"logTraceSpans"`
	// If `true`, traces and spans which weren't successfully received by the
	// backend, will be logged as json
	LogTraceSpansFailedToShip bool `yaml:"logTraceSpansFailedToShip"`
	// If `true`, dimension updates will be logged at the INFO level.
	LogDimensionUpdates bool `yaml:"logDimensionUpdates"`
	// If true, and the log level is `debug`, filtered out datapoints will be
	// logged.
	LogDroppedDatapoints bool `yaml:"logDroppedDatapoints"`
	// If true, the dimensions specified in the top-level `globalDimensions`
	// configuration will be added to the tag set of all spans that are emitted
	// by the writer.  If this is false, only the "host id" dimensions such as
	// `host`, `AwsUniqueId`, etc. are added to the span tags.
	AddGlobalDimensionsAsSpanTags bool `yaml:"addGlobalDimensionsAsSpanTags"`
	// Whether to send host correlation metrics to correlate traced services
	// with the underlying host
	SendTraceHostCorrelationMetrics *bool `yaml:"sendTraceHostCorrelationMetrics" default:"true"`
	// How long to wait after a trace span's service name is last seen to
	// continue sending the correlation datapoints for that service.  This
	// should be a duration string that is accepted by
	// https://golang.org/pkg/time/#ParseDuration.  This option is irrelevant if
	// `sendTraceHostCorrelationMetrics` is false.
	StaleServiceTimeout timeutil.Duration `yaml:"staleServiceTimeout" default:"5m"`
	// How frequently to purge host correlation caches that are generated from
	// the service and environment names seen in trace spans sent through or by
	// the agent.  This should be a duration string that is accepted by
	// https://golang.org/pkg/time/#ParseDuration.
	TraceHostCorrelationPurgeInterval timeutil.Duration `yaml:"traceHostCorrelationPurgeInterval" default:"1m"`
	// How frequently to send host correlation metrics that are generated from
	// the service name seen in trace spans sent through or by the agent.  This
	// should be a duration string that is accepted by
	// https://golang.org/pkg/time/#ParseDuration.  This option is irrelevant if
	// `sendTraceHostCorrelationMetrics` is false.
	TraceHostCorrelationMetricsInterval timeutil.Duration `yaml:"traceHostCorrelationMetricsInterval" default:"1m"`
	// How many times to retry requests related to trace host correlation
	TraceHostCorrelationMaxRequestRetries uint `yaml:"traceHostCorrelationMaxRequestRetries" default:"2"`
	// How many trace spans are allowed to be in the process of sending.  While
	// this number is exceeded, the oldest spans will be discarded to
	// accommodate new spans generated to avoid memory exhaustion.  If you see
	// log messages about "Aborting pending trace requests..." or "Dropping new
	// trace spans..." it means that the downstream target for traces is not
	// able to accept them fast enough. Usually if the downstream is offline
	// you will get connection refused errors and most likely spans will not
	// build up in the agent (there is no retry mechanism). In the case of slow
	// downstreams, you might be able to increase `maxRequests` to increase the
	// concurrent stream of spans downstream (if the target can make efficient
	// use of additional connections) or, less likely, increase
	// `traceSpanMaxBatchSize` if your batches are maxing out (turn on debug
	// logging to see the batch sizes being sent) and being split up too much.
	// If neither of those options helps, your downstream is likely too slow to
	// handle the volume of trace spans and should be upgraded to more powerful
	// hardware/networking.
	MaxTraceSpansInFlight uint `yaml:"maxTraceSpansInFlight" default:"100000"`
	// Configures the writer specifically writing to Splunk.
	Splunk *SplunkConfig `yaml:"splunk"`
	// If set to `false`, output to SignalFx will be disabled.
	SignalFxEnabled *bool `yaml:"signalFxEnabled" default:"true"`
	// Additional headers to add to any outgoing HTTP requests from the agent.
	ExtraHeaders map[string]string `yaml:"extraHeaders"`
	// The following are propagated from elsewhere
	HostIDDims          map[string]string      `yaml:"-"`
	IngestURL           string                 `yaml:"-"`
	APIURL              string                 `yaml:"-"`
	EventEndpointURL    string                 `yaml:"-"`
	TraceEndpointURL    string                 `yaml:"-"`
	SignalFxAccessToken string                 `yaml:"-"`
	GlobalDimensions    map[string]string      `yaml:"-"`
	GlobalSpanTags      map[string]string      `yaml:"-"`
	MetricsToInclude    []MetricFilter         `yaml:"-"`
	MetricsToExclude    []MetricFilter         `yaml:"-"`
	PropertiesToExclude []PropertyFilterConfig `yaml:"-"`
}

func (wc *WriterConfig) IsSplunkOutputEnabled() bool {
	return wc.Splunk != nil && wc.Splunk.Enabled
}

func (wc *WriterConfig) IsSignalFxOutputEnabled() bool {
	return wc.SignalFxEnabled == nil || *wc.SignalFxEnabled
}

func (wc *WriterConfig) Validate() error {
	if !wc.IsSplunkOutputEnabled() && !wc.IsSignalFxOutputEnabled() {
		return errors.New("both SignalFx and Splunk output are disabled, at least one must be enabled")
	}

	if !httpguts.ValidHeaderFieldValue(wc.SignalFxAccessToken) {
		return errors.New("the SignalFx Access Token does not pass http header validation and is likely malformed")
	}

	if _, err := wc.DatapointFilters(); err != nil {
		return fmt.Errorf("datapoint filters are invalid: %v", err)
	}

	return nil
}

// ParsedIngestURL parses and returns the ingest URL
func (wc *WriterConfig) ParsedIngestURL() *url.URL {
	ingestURL, err := url.Parse(wc.IngestURL)
	if err != nil {
		panic("IngestURL was supposed to be validated already")
	}
	return ingestURL
}

// ParsedAPIURL parses and returns the API server URL
func (wc *WriterConfig) ParsedAPIURL() *url.URL {
	apiURL, err := url.Parse(wc.APIURL)
	if err != nil {
		panic("apiUrl was supposed to be validated already")
	}
	return apiURL
}

// ParsedEventEndpointURL parses and returns the event endpoint server URL
func (wc *WriterConfig) ParsedEventEndpointURL() *url.URL {
	if wc.EventEndpointURL != "" {
		eventEndpointURL, err := url.Parse(wc.EventEndpointURL)
		if err != nil {
			panic("eventEndpointUrl was supposed to be validated already")
		}
		return eventEndpointURL
	}
	return nil
}

// ParsedTraceEndpointURL parses and returns the trace endpoint server URL
func (wc *WriterConfig) ParsedTraceEndpointURL() *url.URL {
	if wc.TraceEndpointURL != "" {
		traceEndpointURL, err := url.Parse(wc.TraceEndpointURL)
		if err != nil {
			panic("traceEndpointUrl was supposed to be validated already")
		}
		return traceEndpointURL
	}
	return nil
}

// DatapointFilters creates the filter set for datapoints
func (wc *WriterConfig) DatapointFilters() (*dpfilters.FilterSet, error) {
	return makeOldFilterSet(wc.MetricsToExclude, wc.MetricsToInclude)
}

// PropertyFilters creates the filter set for dimension properties
func (wc *WriterConfig) PropertyFilters() (*propfilters.FilterSet, error) {
	return makePropertyFilterSet(wc.PropertiesToExclude)
}

// Hash calculates a unique hash value for this config struct
func (wc *WriterConfig) Hash() uint64 {
	hash, err := hashstructure.Hash(wc, nil)
	if err != nil {
		log.WithError(err).Error("Could not get hash of WriterConfig struct")
		return 0
	}
	return hash
}

// DefaultTraceEndpointPath returns the default path based on the export format.
func (wc *WriterConfig) DefaultTraceEndpointPath() string {
	if strings.ToLower(wc.TraceExportFormat) == TraceExportFormatSAPM {
		return "/v2/trace"
	}
	return "/v1/trace"
}

// SplunkConfig configures the writer specifically writing to Splunk.
type SplunkConfig struct {
	// Enable logging to a Splunk Enterprise instance
	Enabled bool `yaml:"enabled"`
	// Full URL (including path) of Splunk HTTP Event Collector (HEC) endpoint
	URL string `yaml:"url"`
	// Splunk HTTP Event Collector token
	Token string `yaml:"token"`
	// Splunk source field value, description of the source of the event
	MetricsSource string `yaml:"source"`
	// Splunk source type, optional name of a sourcetype field value
	MetricsSourceType string `yaml:"sourceType"`
	// Splunk index, optional name of the Splunk index to store the event in
	MetricsIndex string `yaml:"index"`
	// Splunk index, specifically for traces (must be event type)
	EventsIndex string `yaml:"eventsIndex"`
	// Splunk source field value, description of the source of the trace
	EventsSource string `yaml:"eventsSource"`
	// Splunk trace source type, optional name of a sourcetype field value
	EventsSourceType string `yaml:"eventsSourceType"`
	// Skip verifying the certificate of the HTTP Event Collector
	SkipTLSVerify bool `yaml:"skipTLSVerify"`

	// The maximum number of Splunk log entries of all types (e.g. metric,
	// event) to be buffered before old events are dropped.  Defaults to the
	// writer.maxDatapointsBuffered config if not specified.
	MaxBuffered int `yaml:"maxBuffered"`
	// The maximum number of simultaneous requests to the Splunk HEC endpoint.
	// Defaults to the writer.maxBuffered config if not specified.
	MaxRequests int `yaml:"maxRequests"`
	// The maximum number of Splunk log entries to submit in one request to the
	// HEC
	MaxBatchSize int `yaml:"maxBatchSize"`
}
