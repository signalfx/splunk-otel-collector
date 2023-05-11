package tracetracker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"

	"github.com/signalfx/signalfx-agent/pkg/apm/correlations"
	"github.com/signalfx/signalfx-agent/pkg/apm/log"
)

func setTime(a *ActiveServiceTracker, t time.Time) {
	a.timeNow = func() time.Time { return t }
}

func advanceTime(a *ActiveServiceTracker, minutes int64) {
	newNow := a.timeNow().Add(time.Duration(minutes) * time.Minute)
	a.timeNow = func() time.Time { return newNow }
}

// mergeStringMaps merges n maps with a later map's keys overriding earlier maps
func mergeStringMaps(maps ...map[string]string) map[string]string {
	ret := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			ret[k] = v
		}
	}

	return ret
}

func TestDatapointsAreGenerated(t *testing.T) {
	correlationClient := &correlationTestClient{}

	a := New(log.Nil, 5*time.Minute, correlationClient, nil, true, nil, DefaultDimsToSyncSource)

	a.AddSpans(context.Background(), []*trace.Span{
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("one"),
			},
			Tags: map[string]string{"host": "test"},
		},
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("two"),
			},
			Tags: map[string]string{"host": "test"},
		},
	})

	a.Purge()
	dps := a.CorrelationDatapoints()
	assert.Len(t, dps, 2, "Expected two datapoints")

	var serviceDims []string
	for _, dp := range dps {
		serviceDims = append(serviceDims, dp.Dimensions["sf_hasService"])
	}
	assert.ElementsMatch(t, serviceDims, []string{"one", "two"}, "expected service names 'one' and 'two'")

	assert.Equal(t, dps[0].Value.(datapoint.IntValue).Int(), int64(0), "Expected dp value to be 0")
}

func TestExpiration(t *testing.T) {
	correlationClient := &correlationTestClient{}

	hostIDDims := map[string]string{"host": "test", "AWSUniqueId": "randomAWSUniqueId"}
	a := New(log.Nil, 5*time.Minute, correlationClient, hostIDDims, true, nil, DefaultDimsToSyncSource)
	setTime(a, time.Unix(100, 0))

	a.AddSpans(context.Background(), []*trace.Span{
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("one"),
			},
			Tags: mergeStringMaps(hostIDDims, map[string]string{"environment": "environment1"}),
		},
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("two"),
			},
			Tags: mergeStringMaps(hostIDDims, map[string]string{"environment": "environment2"}),
		},
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("three"),
			},
			Tags: mergeStringMaps(hostIDDims, map[string]string{"environment": "environment3"}),
		},
	})

	assert.Equal(t, int64(3), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(3), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")

	advanceTime(a, 4)

	a.AddSpans(context.Background(), []*trace.Span{
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("two"),
			},
			Tags: mergeStringMaps(hostIDDims, map[string]string{"environment": "environment2"}),
		},
	})

	advanceTime(a, 2)
	a.Purge()
	dps := a.CorrelationDatapoints()
	assert.Len(t, dps, 1, "Expected one datapoint")
	assert.Equal(t, dps[0].Dimensions["sf_hasService"], "two", "expected service two to still be active")

	assert.Equal(t, int64(1), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(1), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")
	assert.Equal(t, int64(2), a.hostServiceCache.PurgedCount, "purgedServiceCount is not properly tracked")
	assert.Equal(t, int64(2), a.hostEnvironmentCache.PurgedCount, "activeEnvironmentCount is not properly tracked")

	advanceTime(a, 4)
	a.Purge()
	assert.Len(t, a.CorrelationDatapoints(), 0, "Expected all datapoints to be expired")
}

type correlationTestClient struct {
	sync.Mutex
	cors             []*correlations.Correlation
	getPayload       map[string]map[string][]string
	getCallback      func()
	getCounter       int64
	deleteCounter    int64
	correlateCounter int64
}

func (c *correlationTestClient) Start() { /*no-op*/ }
func (c *correlationTestClient) Get(dimName string, dimValue string, cb correlations.SuccessfulGetCB) {
	atomic.AddInt64(&c.getCounter, 1)
	go func() {
		cb(c.getPayload[dimValue])
		if c.getCallback != nil {
			c.getCallback()
		}
	}()
}
func (c *correlationTestClient) Correlate(cl *correlations.Correlation, cb correlations.CorrelateCB) {
	c.Lock()
	defer c.Unlock()
	c.cors = append(c.cors, cl)
	cb(cl, nil)
	atomic.AddInt64(&c.correlateCounter, 1)
}
func (c *correlationTestClient) Delete(cl *correlations.Correlation, cb correlations.SuccessfulDeleteCB) {
	c.Lock()
	defer c.Unlock()
	c.cors = append(c.cors, cl)
	cb(cl)
	atomic.AddInt64(&c.deleteCounter, 1)
}
func (c *correlationTestClient) InternalMetrics() []*datapoint.Datapoint {
	return nil
}
func (c *correlationTestClient) getCorrelations() []*correlations.Correlation {
	c.Lock()
	defer c.Unlock()
	return c.cors
}

var _ correlations.CorrelationClient = &correlationTestClient{}

func TestCorrelationEmptyEnvironment(t *testing.T) {
	var wg sync.WaitGroup
	correlationClient := &correlationTestClient{
		getPayload: map[string]map[string][]string{
			"testk8sPodUID": {
				"sf_services":     {"one"},
				"sf_environments": {"environment1"},
			},
			"testContainerID": {
				"sf_services":     {"one"},
				"sf_environments": {"environment1"},
			},
		},
		getCallback: func() {
			wg.Done()
		},
	}

	hostIDDims := map[string]string{"host": "test", "AWSUniqueId": "randomAWSUniqueId"}
	wg.Add(len(hostIDDims))
	containerLevelIDDims := map[string]string{"kubernetes_pod_uid": "testk8sPodUID", "container_id": "testContainerID"}
	a := New(log.Nil, 5*time.Minute, correlationClient, hostIDDims, true, nil, DefaultDimsToSyncSource)
	wg.Wait() // wait for the initial fetch of hostIDDims to complete

	// for each container level ID we're going to perform a GET to check for an environment
	wg.Add(len(containerLevelIDDims))
	a.AddSpans(context.Background(), []*trace.Span{
		{
			LocalEndpoint: &trace.Endpoint{},
			Tags:          mergeStringMaps(hostIDDims, containerLevelIDDims),
		},
		{
			LocalEndpoint: &trace.Endpoint{},
			Tags:          mergeStringMaps(hostIDDims, containerLevelIDDims),
		},
		{
			LocalEndpoint: &trace.Endpoint{},
			Tags:          mergeStringMaps(hostIDDims, containerLevelIDDims),
		},
	})

	wg.Wait() // wait for the gets to complete to check for existing tenant environment values

	// there shouldn't be any active tenant environments.  None of the spans had environments on them,
	// and we don't actively fetch and store environments from the back end.  That's kind of the whole point of this
	// the workaround this is testing.
	assert.Equal(t, int64(0), a.tenantEnvironmentCache.ActiveCount, "tenantEnvironmentCache is not properly tracked")
	// ensure we only have 1 entry per container / pod id
	assert.Equal(t, int64(len(containerLevelIDDims)), a.tenantEmptyEnvironmentCache.ActiveCount, "tenantEmptyEnvironmentCount is not properly tracked")
	// len(hostIDDims) * len(containerLevelIDDims)
	assert.Equal(t, int64(len(containerLevelIDDims)+len(hostIDDims)), atomic.LoadInt64(&correlationClient.getCounter), "")
	// 1 DELETE * len(containerLevelIDDims)
	assert.Equal(t, len(containerLevelIDDims), len(correlationClient.getCorrelations()), "")
}

func TestCorrelationUpdates(t *testing.T) {
	var wg sync.WaitGroup
	correlationClient := &correlationTestClient{
		getPayload: map[string]map[string][]string{
			"test": {
				"sf_services":     {"one"},
				"sf_environments": {"environment1"},
			},
		},
		getCallback: func() {
			wg.Done()
		},
	}

	hostIDDims := map[string]string{"host": "test", "AWSUniqueId": "randomAWSUniqueId"}
	wg.Add(len(hostIDDims))
	containerLevelIDDims := map[string]string{"kubernetes_pod_uid": "testk8sPodUID", "container_id": "testContainerID"}
	a := New(log.Nil, 5*time.Minute, correlationClient, hostIDDims, true, nil, DefaultDimsToSyncSource)
	wg.Wait()
	assert.Equal(t, int64(1), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(1), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")
	advanceTime(a, 6)
	a.Purge()
	assert.Equal(t, int64(0), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(0), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")
	assert.Equal(t, int64(1), a.hostServiceCache.PurgedCount, "hostServiceCache purged count is not properly tracked")
	assert.Equal(t, int64(1), a.hostEnvironmentCache.PurgedCount, "hostEnvironmentCache purged count is not properly tracked")

	setTime(a, time.Unix(100, 0))

	a.AddSpans(context.Background(), []*trace.Span{
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("one"),
			},
			Tags: mergeStringMaps(hostIDDims, mergeStringMaps(containerLevelIDDims, map[string]string{"environment": "environment1"})),
		},
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("two"),
			},
			Tags: mergeStringMaps(hostIDDims, mergeStringMaps(containerLevelIDDims, map[string]string{"environment": "environment2"})),
		},
		{
			LocalEndpoint: &trace.Endpoint{
				ServiceName: pointer.String("three"),
			},
			Tags: mergeStringMaps(hostIDDims, mergeStringMaps(containerLevelIDDims, map[string]string{"environment": "environment3"})),
		},
	})

	assert.Equal(t, int64(3), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(3), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")

	numEnvironments := 3
	numServices := 3
	numHostIDDimCorrelations := len(hostIDDims)*(numEnvironments+numServices) + 4 /* 4 deletes for service & environment fetched at startup */
	numContainerLevelCorrelations := 2 * len(containerLevelIDDims)
	totalExpectedCorrelations := numHostIDDimCorrelations + numContainerLevelCorrelations
	assert.Equal(t, totalExpectedCorrelations, len(correlationClient.getCorrelations()), "# of correlation requests do not match")

	advanceTime(a, 6)
	a.Purge()
	dps := a.CorrelationDatapoints()
	assert.Len(t, dps, 0, "Expected all datapoints to be expired")
	assert.Equal(t, int64(0), a.hostServiceCache.ActiveCount, "activeServiceCount is not properly tracked")
	assert.Equal(t, int64(0), a.hostEnvironmentCache.ActiveCount, "activeEnvironmentCount is not properly tracked")
	assert.Equal(t, int64(4), a.hostServiceCache.PurgedCount, "purgedServiceCount is not properly tracked")
	assert.Equal(t, int64(4), a.hostEnvironmentCache.PurgedCount, "activeEnvironmentCount is not properly tracked")
}
