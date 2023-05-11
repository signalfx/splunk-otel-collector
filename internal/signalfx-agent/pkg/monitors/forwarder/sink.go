package forwarder

import (
	"context"
	"net"
	"net/http"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

type _sourceKey int

var sourceKey _sourceKey

type outputSink struct {
	Output types.Output
}

func (os *outputSink) AddDatapoints(ctx context.Context, dps []*datapoint.Datapoint) error {
	os.Output.SendDatapoints(dps...)
	return nil
}

func (os *outputSink) AddEvents(ctx context.Context, events []*event.Event) error {
	return nil
}

func (os *outputSink) AddSpans(ctx context.Context, spans []*trace.Span) error {
	source, hasSource := ctx.Value(sourceKey).(net.IP)
	if hasSource {
		for i := range spans {
			if spans[i].Meta == nil {
				spans[i].Meta = map[interface{}]interface{}{}
			}
			spans[i].Meta[constants.DataSourceIPKey] = source
		}
	}
	os.Output.SendSpans(spans...)
	return nil
}

func tryToExtractRemoteAddressToContext(ctx context.Context, req *http.Request) context.Context {
	var sourceIP net.IP
	if req.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			sourceIP = net.ParseIP(host)
			if sourceIP != nil {
				return context.WithValue(ctx, sourceKey, sourceIP)
			}
		}
	}
	return ctx
}
