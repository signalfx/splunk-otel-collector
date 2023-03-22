package service

import (
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/core/common/dpmeta"
	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

type PointsSvc struct {
	log         logrus.FieldLogger
	gateway     IGateway
	perfFetcher perfFetcher
	ptConsumer  func(...*datapoint.Datapoint)
}

func NewPointsSvc(
	gateway IGateway,
	log logrus.FieldLogger,
	batchSize int,
	ptConsumer func(...*datapoint.Datapoint),
) *PointsSvc {
	fetcher := newPerfFetcher(gateway, batchSize, log)
	return &PointsSvc{
		gateway:     gateway,
		log:         log,
		ptConsumer:  ptConsumer,
		perfFetcher: fetcher,
	}
}

// Retrieves datapoints for all of the inventory objects in the passed-in
// VsphereInfo for the number of 20-second intervals indicated by the passed-in
// numSamplesReqd. Also returns the most recent sample time for the returned points.
func (svc *PointsSvc) FetchPoints(vsInfo *model.VsphereInfo, numSamplesReqd int32) time.Time {
	it := svc.perfFetcher.invIterator(vsInfo.Inv.Objects, numSamplesReqd)
	var metrics []types.BasePerfEntityMetricBase
	var err error
	var latestSampleTime time.Time
	hasNext := true
	for hasNext {
		metrics, hasNext, err = it.nextInvPage()
		if err != nil {
			svc.log.WithError(err).Error("queryPerf failed")
			return time.Time{}
		}

		for _, baseMetric := range metrics {
			perfEntityMetric, ok := baseMetric.(*types.PerfEntityMetric)
			if !ok {
				svc.log.WithField(
					"baseMetric", baseMetric,
				).Error("Type coersion to PerfEntityMetric failed")
				continue
			}

			t := perfEntityMetric.SampleInfo[len(perfEntityMetric.SampleInfo)-1].Timestamp
			if t.After(latestSampleTime) {
				latestSampleTime = t
			}

			for _, metric := range perfEntityMetric.Value {
				intSeries, ok := metric.(*types.PerfMetricIntSeries)
				if !ok {
					svc.log.WithField(
						"metric", metric,
					).Error("Type coersion to PerfMetricIntSeries failed")
					continue
				}

				metricInfo := vsInfo.PerfCounterIndex[intSeries.Id.CounterId]
				metricName := metricInfo.MetricName
				sfxMetricType := statsTypeToMetricType(metricInfo.PerfCounterInfo.StatsType)

				cachedDims, ok := vsInfo.Inv.DimensionMap[perfEntityMetric.Entity.Value]
				var dims map[string]string
				if !ok {
					dims = map[string]string{}
				} else {
					dims = copyMap(cachedDims)
				}

				if intSeries.Id.Instance != "" {
					// the vsphere UI calls this dimension 'Object'
					dims["object"] = intSeries.Id.Instance
				}

				dims["vcenter"] = svc.gateway.vcenterName()

				if len(intSeries.Value) > 0 && intSeries.Value[0] > 0 {
					svc.log.Debugf(
						"metric = %s, type = (%s->%s), dims = %v, values = %v",
						metricName,
						metricInfo.PerfCounterInfo.StatsType,
						sfxMetricType,
						dims,
						intSeries.Value,
					)
				}
				for i, value := range intSeries.Value {
					var dpVal datapoint.Value
					if strings.HasSuffix(metricName, "_percent") {
						dpVal = datapoint.NewFloatValue(float64(value) / 100)
					} else {
						dpVal = datapoint.NewIntValue(value)
					}
					dp := datapoint.New(
						metricName,
						dims,
						dpVal,
						sfxMetricType,
						perfEntityMetric.SampleInfo[i].Timestamp,
					)
					// If the host dimension is set, the make sure the agent doesn't overwrite
					// the `host` dimension with the hostname the agent is running on
					if _, ok := dims["host"]; ok {
						dp.Meta[dpmeta.NotHostSpecificMeta] = true
					}
					svc.ptConsumer(dp)
				}
			}
		}
	}
	return latestSampleTime
}

func statsTypeToMetricType(statsType types.PerfStatsType) datapoint.MetricType {
	switch statsType {
	case types.PerfStatsTypeDelta:
		return datapoint.Count
	default:
		return datapoint.Gauge
	}
}
