package utils

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

func PrepareGaugeHelper(metricName string, dims map[string]string, metricValue *int64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.Gauge(metricName, dims, *metricValue)
}

func PrepareGaugeFHelper(metricName string, dims map[string]string, metricValue *float64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.GaugeF(metricName, dims, *metricValue)
}

func PrepareCumulativeHelper(metricName string, dims map[string]string, metricValue *int64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.Cumulative(metricName, dims, *metricValue)
}
