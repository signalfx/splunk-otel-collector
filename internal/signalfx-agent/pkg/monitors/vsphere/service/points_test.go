package service

import (
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

var testLog = logrus.WithField("monitorType", "vsphere-test")

func TestRetrievePoints(t *testing.T) {
	gateway := newFakeGateway(1)
	inventorySvc := NewInventorySvc(gateway, testLog, nopFilter{}, model.GuestIP)
	metricsSvc := NewMetricsService(gateway, testLog)
	infoSvc := NewVSphereInfoService(inventorySvc, metricsSvc)
	vsphereInfo, _ := infoSvc.RetrieveVSphereInfo()
	var points []*datapoint.Datapoint
	ptsSvc := NewPointsSvc(gateway, testLog, 0, func(dp ...*datapoint.Datapoint) {
		points = append(points, dp...)
	})
	ptsSvc.FetchPoints(vsphereInfo, 1)
	pt := points[0]
	require.Equal(t, "vsphere.cpu_core_utilization_percent", pt.Metric)
	require.Equal(t, datapoint.Count, pt.MetricType)
	require.EqualValues(t, 1.11, pt.Value)
	require.Equal(t, "my-vc", pt.Dimensions["vcenter"])
}
