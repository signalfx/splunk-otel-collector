package processor

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/common/dpmeta"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
)

type Processor struct {
	globalDims                    map[string]string
	globalSpanTags                map[string]string
	addGlobalDimensionsAsSpanTags bool
	hostIDDims                    map[string]string
	datapointFilters              *dpfilters.FilterSet
}

func New(conf *config.WriterConfig) *Processor {
	datapointFilters, _ := conf.DatapointFilters()

	return &Processor{
		hostIDDims:                    conf.HostIDDims,
		globalDims:                    conf.GlobalDimensions,
		globalSpanTags:                conf.GlobalSpanTags,
		addGlobalDimensionsAsSpanTags: conf.AddGlobalDimensionsAsSpanTags,
		datapointFilters:              datapointFilters,
	}
}

func (p *Processor) ShouldSendDatapoint(dp *datapoint.Datapoint) bool {
	return p.datapointFilters == nil || !p.datapointFilters.Matches(dp)
}

func (p *Processor) PreprocessDatapoint(dp *datapoint.Datapoint) bool {
	if !p.ShouldSendDatapoint(dp) {
		return false
	}

	dp.Dimensions = p.addGlobalDims(dp.Dimensions)

	// Some metrics aren't really specific to the host they are running
	// on and shouldn't have any host-specific dims
	if b, ok := dp.Meta[dpmeta.NotHostSpecificMeta].(bool); !ok || !b {
		dp.Dimensions = p.addhostIDFields(dp.Dimensions)
	}

	return true
}

func (p *Processor) PreprocessEvent(event *event.Event) bool {
	event.Dimensions = p.addGlobalDims(event.Dimensions)

	ps := event.Properties
	var notHostSpecific bool
	if ps != nil {
		if b, ok := ps[dpmeta.NotHostSpecificMeta].(bool); ok {
			notHostSpecific = b
			// Clear this so it doesn't leak through to ingest
			delete(ps, dpmeta.NotHostSpecificMeta)
		}
	}
	// Only override host dimension for now and omit other host id dims.
	if !notHostSpecific && p.hostIDDims != nil && p.hostIDDims["host"] != "" {
		event.Dimensions["host"] = p.hostIDDims["host"]
	}

	return true
}

func (p *Processor) PreprocessSpan(span *trace.Span) bool {
	// Some spans aren't really specific to the host they are running
	// on and shouldn't have any host-specific tags.  This is indicated by a
	// special tag key (value is irrelevant).
	if _, ok := span.Tags[dpmeta.NotHostSpecificMeta]; !ok {
		span.Tags = p.addhostIDFields(span.Tags)
	} else {
		// Get rid of the tag so it doesn't pass through to the backend
		delete(span.Tags, dpmeta.NotHostSpecificMeta)
	}

	span.Tags = p.addGlobalSpanTags(span.Tags)

	if p.addGlobalDimensionsAsSpanTags {
		span.Tags = p.addGlobalDims(span.Tags)
	}

	span.Tags["signalfx.smartagent.version"] = constants.Version

	return true
}

// Mutates span tags in place to add global span tags.  Also
// returns tags in case they were nil to begin with, so the return value should
// be assigned back to the span Tags field.
func (p *Processor) addGlobalSpanTags(tags map[string]string) map[string]string {
	if tags == nil {
		tags = make(map[string]string)
	}
	for name, value := range p.globalSpanTags {
		// If the tags are already set, don't override
		if _, ok := tags[name]; !ok {
			tags[name] = value
		}
	}
	return tags
}

// Mutates datapoint dimensions in place to add global dimensions.  Also
// returns dims in case they were nil to begin with, so the return value should
// be assigned back to the dp Dimensions field.
func (p *Processor) addGlobalDims(dims map[string]string) map[string]string {
	if dims == nil {
		dims = make(map[string]string)
	}
	for name, value := range p.globalDims {
		// If the dimensions are already set, don't override
		if _, ok := dims[name]; !ok {
			dims[name] = value
		}
	}
	return dims
}

// Adds the host ids to the given map (e.g. dimensions/span tags), forcibly
// overridding any existing fields of the same name.
func (p *Processor) addhostIDFields(fields map[string]string) map[string]string {
	if fields == nil {
		fields = make(map[string]string)
	}
	for k, v := range p.hostIDDims {
		fields[k] = v
	}
	return fields
}
