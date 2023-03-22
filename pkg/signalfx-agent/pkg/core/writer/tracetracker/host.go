package tracetracker

import (
	"net"

	lru "github.com/hashicorp/golang-lru"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/golib/v3/trace"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// SpanSourceTracker inserts tags into spans that identify the source of the
// span using data that is available to the agent from the observers
// (specifically the dimensions on the observer output listed in
// DimsToSyncSource).  It also attaches certain properties about the span local
// service name and the global cluster name to those dimensions.
type SpanSourceTracker struct {
	dimChan     chan<- *types.Dimension
	hostTracker *services.EndpointHostTracker
	clusterName string
	dimHistory  *lru.Cache
}

const dimHistoryCacheSize = 1000

// DefaultDimsToAddToSpans are the default dimensions to add as span tags for correlated environments
var DefaultDimsToAddToSpans = map[string]string{
	"container_id":         "container_id",
	"kubernetes_pod_uid":   "kubernetes_pod_uid",
	"kubernetes_pod_name":  "kubernetes_pod_name",
	"kubernetes_namespace": "kubernetes_namespace",
}

func NewSpanSourceTracker(hostTracker *services.EndpointHostTracker, dimChan chan<- *types.Dimension, clusterName string) *SpanSourceTracker {
	dimHistory, _ := lru.New(dimHistoryCacheSize)

	return &SpanSourceTracker{
		clusterName: clusterName,
		dimChan:     dimChan,
		hostTracker: hostTracker,
		dimHistory:  dimHistory,
	}
}

func (st *SpanSourceTracker) AddSourceTagsToSpan(span *trace.Span) {
	sourceIP, ok := span.Meta[constants.DataSourceIPKey].(net.IP)
	if !ok || sourceIP == nil {
		return
	}

	endpoints := st.hostTracker.GetByHost(sourceIP.String())
	found := 0
	for _, endpoint := range endpoints {
		dims := endpoint.Dimensions()
		for _, dim := range DefaultDimsToAddToSpans {
			if val := dims[dim]; val != "" {
				found++

				if span.LocalEndpoint != nil && span.LocalEndpoint.ServiceName != nil {
					st.emitDimensionPropIfNew(dim, val, *span.LocalEndpoint.ServiceName)
				}

				if span.Tags == nil {
					span.Tags = map[string]string{}
				}

				tags := span.Tags
				if _, ok := tags[dim]; ok {
					// Don't overwrite existing span tags
					continue
				}
				tags[dim] = val
			}
		}
		// Short circuit it if we have added all the desired dimensions with
		// this endpoint.
		if found == len(DefaultDimsToAddToSpans) {
			break
		}
	}

	if found == 0 {
		logrus.Debugf("Could not find source of span %v with sourceIP %s", span, sourceIP)
	}
}

func (st *SpanSourceTracker) emitDimensionPropIfNew(dimName, dimValue, serviceName string) {
	key := struct {
		dimName     string
		dimValue    string
		serviceName string
	}{
		dimName:     dimName,
		dimValue:    dimValue,
		serviceName: serviceName,
	}

	_, ok := st.dimHistory.Get(key)

	if !ok {
		st.dimChan <- &types.Dimension{
			Name:  dimName,
			Value: dimValue,
			Properties: map[string]string{
				"service": serviceName,
				"cluster": st.clusterName,
			},
			MergeIntoExisting: true,
		}
		st.dimHistory.Add(key, true)
	}
}

func (st *SpanSourceTracker) InternalMetrics() []*datapoint.Datapoint {
	return append([]*datapoint.Datapoint{
		sfxclient.Cumulative("sfxagent.span_source_tracker_size", nil, int64(st.dimHistory.Len())),
	}, st.hostTracker.InternalMetrics()...)
}
