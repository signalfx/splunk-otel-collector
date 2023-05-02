package tracetracker

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/require"
)

const testClusterName = "test-cluster"

func waitForDims(dimCh <-chan *types.Dimension, count, waitSeconds int) []types.Dimension { // nolint: unparam
	var dims []types.Dimension
	timeout := time.After(time.Duration(waitSeconds) * time.Second)

loop:
	for {
		select {
		case dim := <-dimCh:
			dims = append(dims, *dim)
			if len(dims) >= count {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	return dims
}

func TestSourceTracker(t *testing.T) {
	dimChan := make(chan *types.Dimension, 1000)
	hostTracker := services.NewEndpointHostTracker()
	tracker := NewSpanSourceTracker(hostTracker, dimChan, testClusterName)

	t.Run("does basic correlation", func(t *testing.T) {
		const count = 100

		var endpoints []services.Endpoint
		for i := 0; i < count; i++ {
			endpoint := services.NewEndpointCore(fmt.Sprintf("endpoint-%d", i), "test-endpoint", "nothing", map[string]string{
				"container_id":       fmt.Sprintf("container-%d", i),
				"kubernetes_pod_uid": fmt.Sprintf("pod-%d", i),
			})
			endpoint.Core().Host = fmt.Sprintf("10.0.5.%d", i)

			hostTracker.EndpointAdded(endpoint)

			endpoints = append(endpoints, endpoint)
		}

		var spans []*trace.Span
		for i := 0; i < count; i++ {
			span := &trace.Span{
				Name: pointer.String(fmt.Sprintf("span-%d", i)),
				LocalEndpoint: &trace.Endpoint{
					ServiceName: pointer.String(fmt.Sprintf("service-%d", i%5)),
				},
				Meta: map[interface{}]interface{}{
					constants.DataSourceIPKey: net.ParseIP(fmt.Sprintf("10.0.5.%d", i)),
				},
			}
			spans = append(spans, span)

			tracker.AddSourceTagsToSpan(span)

			require.Equal(t, span.Tags, map[string]string{
				"container_id":       fmt.Sprintf("container-%d", i),
				"kubernetes_pod_uid": fmt.Sprintf("pod-%d", i),
			})

			dims := waitForDims(dimChan, 2, 3)
			require.ElementsMatch(t, []types.Dimension{
				{
					Name:  "container_id",
					Value: fmt.Sprintf("container-%d", i),
					Properties: map[string]string{
						"service": fmt.Sprintf("service-%d", i%5),
						"cluster": testClusterName,
					},
					Tags:              nil,
					MergeIntoExisting: true,
				},
				{
					Name:  "kubernetes_pod_uid",
					Value: fmt.Sprintf("pod-%d", i),
					Properties: map[string]string{
						"service": fmt.Sprintf("service-%d", i%5),
						"cluster": testClusterName,
					},
					Tags:              nil,
					MergeIntoExisting: true,
				},
			}, dims)
		}

		// Remove the endpoints from the host tracker and make sure they aren't
		// tracked any more.
		for i := 0; i < count; i++ {
			hostTracker.EndpointRemoved(endpoints[i])

			spans[i].Tags = nil

			tracker.AddSourceTagsToSpan(spans[i])

			require.Equal(t, spans[i].Tags, map[string]string(nil))
		}

		dims := waitForDims(dimChan, 1, 3)
		require.Len(t, dims, 0)
	})
}
