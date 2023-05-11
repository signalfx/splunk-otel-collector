package monitors

import (
	"regexp"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func helperTestMonitorOuput() (*monitorOutput, error) {
	config := &config.MonitorConfig{}
	var metadata *Metadata

	monFiltering, err := newMonitorFiltering(config, metadata)
	if err != nil {
		return nil, err
	}

	output := &monitorOutput{
		monitorType:      "testMonitor",
		monitorID:        "testMonitor1",
		monitorFiltering: monFiltering,
	}
	return output, nil
}

func TestSendDatapoint(t *testing.T) {
	// Setup our 'fixture' super basic monitorOutput
	testMO, err := helperTestMonitorOuput()
	assert.Nil(t, err)

	// And our Datapoint channel to receive Datapoints
	dpChan := make(chan []*datapoint.Datapoint)
	testMO.dpChan = dpChan

	// And our reference timestamp
	dpTimestamp := time.Now()

	// Create a test Datapoint
	testDp := datapoint.New("test.metric.name", nil, datapoint.NewIntValue(1), datapoint.Gauge, dpTimestamp)

	// Send the datapoint
	go func() { testMO.SendDatapoints(testDp) }()

	// Receive the datapoint
	resultDps := <-dpChan

	// Make sure it's come through as expected
	assert.Equal(t, "test.metric.name", resultDps[0].Metric)
	assert.Equal(t, map[string]string{}, resultDps[0].Dimensions)
	assert.Equal(t, datapoint.NewIntValue(1), resultDps[0].Value)
	assert.Equal(t, datapoint.Gauge, resultDps[0].MetricType)
	assert.Equal(t, dpTimestamp, resultDps[0].Timestamp)

	// Let's add some extra dimensions to our monitorOutput
	testMO.extraDims = map[string]string{"testDim1": "testValue1"}

	// Resend the datapoint
	go func() { testMO.SendDatapoints(testDp) }()

	// Receive the datapoint
	resultDps = <-dpChan

	// Make sure it's come through as expected
	assert.Equal(t, map[string]string{"testDim1": "testValue1"}, resultDps[0].Dimensions)

	// Add some dimensions in the test Datapoint
	go func() {
		testDp.Dimensions = map[string]string{"testDim2": "testValue2"}
		testMO.SendDatapoints(testDp)
	}()

	// Receive the datapoint
	resultDps = <-dpChan

	// Make sure it's come through as expected
	assert.Equal(t, map[string]string{"testDim1": "testValue1", "testDim2": "testValue2"}, resultDps[0].Dimensions)

	t.Run("dimensionTransformations", func(t *testing.T) {
		// Test using the dimension transformation
		testMO.dimensionTransformations = map[string]string{"testDim2": "testDim3"}

		// Send the datapoint with a dimension that matches our transform
		go func() {
			testDp.Dimensions = map[string]string{"testDim2": "testValue2"}
			testMO.SendDatapoints(testDp)
		}()

		// Receive the datapoint
		resultDps = <-dpChan

		// Make sure it's come through as expected
		assert.Equal(t, map[string]string{"testDim1": "testValue1", "testDim3": "testValue2"}, resultDps[0].Dimensions)

		testMO.dimensionTransformations = map[string]string{"highCardDim": ""}

		// Send the datapoint with a matching dimension
		go func() {
			testDp.Dimensions = map[string]string{"highCardDim": "highCardValue"}
			testMO.SendDatapoints(testDp)
		}()

		// Receive the datapoint
		resultDps = <-dpChan

		// Make sure it's come through as expected
		assert.Equal(t, map[string]string{"testDim1": "testValue1"}, resultDps[0].Dimensions)
	})

	t.Run("metricNameTransformations", func(t *testing.T) {
		testMO.metricNameTransformations = []*config.RegexpWithReplace{
			{Regexp: regexp.MustCompile("^cpu.cores$"), Replacement: "other.cores"},
			{Regexp: regexp.MustCompile(`^cpu\.(.*)$`), Replacement: "mycpu.$1"},
			// This is a more specific match after a less specific one that
			// overlaps so it should have no effect.
			{Regexp: regexp.MustCompile("^cpu.utilization$"), Replacement: "other.utilization"},
		}

		dp := utils.CloneDatapoint(testDp)

		cases := map[string]string{
			"cpu.utilization": "mycpu.utilization",
			"cpu.cores":       "other.cores",
			"cpu.user":        "mycpu.user",
			"memory.total":    "memory.total",
		}

		for old, newName := range cases {
			dp.Metric = old

			go testMO.SendDatapoints(dp)
			outDP := (<-dpChan)[0]

			assert.Equal(t, newName, outDP.Metric)
		}
	})
}

func TestSendSpan(t *testing.T) {
	// Setup our 'fixture' super basic monitorOutput
	testMO, err := helperTestMonitorOuput()
	assert.Nil(t, err)

	// And our Span channel to receive Spans
	spanChan := make(chan []*trace.Span)
	testMO.spanChan = spanChan

	// And our reference timestamp
	spanTimestamp := time.Now()

	// Create a test Span
	testSpan := &trace.Span{
		Name:      pointer.String("testSpan"),
		TraceID:   "testTraceID",
		ParentID:  pointer.String("testParentID"),
		Kind:      pointer.String("CLIENT"),
		Timestamp: pointer.Int64(spanTimestamp.UnixNano() / 1000),
		Duration:  pointer.Int64(1 + time.Since(spanTimestamp).Microseconds()),
		LocalEndpoint: &trace.Endpoint{
			ServiceName: pointer.String("testService"),
		},
	}

	// Send the span
	go func() { testMO.SendSpans(testSpan) }()

	// Receive the span
	resultSpans := <-spanChan

	// Make sure it's come through as expected
	assert.NotEmpty(t, resultSpans)
	assert.Equal(t, "testSpan", *resultSpans[0].Name)
	assert.Equal(t, "testTraceID", resultSpans[0].TraceID)
	assert.Equal(t, "testParentID", *resultSpans[0].ParentID)
	assert.Equal(t, "CLIENT", *resultSpans[0].Kind)
	assert.Equal(t, spanTimestamp.UnixNano()/1000, *resultSpans[0].Timestamp)
	assert.True(t, *resultSpans[0].Duration > 0)
	assert.Equal(t, "testService", *resultSpans[0].LocalEndpoint.ServiceName)

	// Let's add some default tags to our monitorOutput
	testMO.defaultSpanTags = map[string]string{"defaultSpanTag": "testValue1"}

	// Resend the span
	go func() { testMO.SendSpans(testSpan) }()

	// Receive the span
	resultSpans = <-spanChan

	// Make sure it's come through as expected
	assert.Equal(t, map[string]string{"defaultSpanTag": "testValue1"}, resultSpans[0].Tags)

	// Add some tags to the test Span
	go func() {
		testSpan.Tags = map[string]string{"testTag2": "testValue2"}
		testMO.SendSpans(testSpan)
	}()

	// Receive the span
	resultSpans = <-spanChan

	// Make sure it's come through as expected
	assert.Equal(t, map[string]string{"defaultSpanTag": "testValue1", "testTag2": "testValue2"}, resultSpans[0].Tags)

	// Add a tag that collides with defaultSpanTags
	go func() {
		testSpan.Tags["defaultSpanTag"] = "testValue3"
		testMO.SendSpans(testSpan)
	}()

	// Receive the span
	resultSpans = <-spanChan

	// Make sure that defaultSpanTags does not overwrite the existing tag
	assert.Equal(t, map[string]string{"defaultSpanTag": "testValue3", "testTag2": "testValue2"}, resultSpans[0].Tags)

	// Add extraSpanTags to our monitorOutput
	testMO.extraSpanTags = map[string]string{"extraSpanTag": "testValue4"}

	// Resend the span
	go func() { testMO.SendSpans(testSpan) }()

	// Receive the span
	resultSpans = <-spanChan

	// Make sure the extra span tag was added
	assert.Equal(t, map[string]string{"extraSpanTag": "testValue4", "defaultSpanTag": "testValue3", "testTag2": "testValue2"}, resultSpans[0].Tags)

	// Add a span tag that collides with the extraSpanTag
	go func() {
		testSpan.Tags["extraSpanTag"] = "testValue5"
		testMO.SendSpans(testSpan)
	}()

	// Receive the span
	resultSpans = <-spanChan

	// Make sure that extraSpanTags overwrites the tag that is already present
	assert.Equal(t, map[string]string{"extraSpanTag": "testValue4", "defaultSpanTag": "testValue3", "testTag2": "testValue2"}, resultSpans[0].Tags)
}
