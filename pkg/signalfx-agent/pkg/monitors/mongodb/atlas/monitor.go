package atlas

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Sectorbob/mlab-ns2/gae/ns/digest"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/mongodb/atlas/measurements"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline"`
	// ProjectID is the Atlas project ID.
	ProjectID string `yaml:"projectID" validate:"required" `
	// PublicKey is the Atlas public API key
	PublicKey string `yaml:"publicKey" validate:"required" `
	// PrivateKey is the Atlas private API key
	PrivateKey string `yaml:"privateKey" validate:"required" neverLog:"true"`
	// Timeout for HTTP requests to get MongoDB process measurements from Atlas.
	// This should be a duration string that is accepted by https://golang.org/pkg/time/#ParseDuration
	Timeout timeutil.Duration `yaml:"timeout" default:"5s"`
	// EnableCache enables locally cached Atlas metric measurements to be used when true. The metric measurements that
	// were supposed to be fetched are in fact always fetched asynchronously and cached.
	EnableCache bool `yaml:"enableCache" default:"true"`
	// Granularity is the duration in ISO 8601 notation that specifies the interval between measurement data points
	// from Atlas over the configured period. The default is shortest duration supported by Atlas of 1 minute.
	Granularity string `yaml:"granularity" default:"PT1M"`
	// Period the duration in ISO 8601 notation that specifies how far back in the past to retrieve measurements from Atlas.
	Period string `yaml:"period" default:"PT20M"`
}

// Monitor for MongoDB Atlas metrics
type Monitor struct {
	Output        types.FilteringOutput
	cancel        context.CancelFunc
	processGetter measurements.ProcessesGetter
	diskGetter    measurements.DisksGetter
	logger        log.FieldLogger
}

// Configure monitor
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	var client *mongodbatlas.Client
	var processMeasurements measurements.ProcessesMeasurements
	var diskMeasurements measurements.DisksMeasurements

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	timeout := conf.Timeout.AsDuration()

	if client, err = newDigestClient(conf.PublicKey, conf.PrivateKey); err != nil {
		return fmt.Errorf("error making HTTP digest client: %+v", err)
	}

	m.processGetter = measurements.NewProcessesGetter(conf.ProjectID, conf.Granularity, conf.Period, client, conf.EnableCache, m.logger)
	m.diskGetter = measurements.NewDisksGetter(conf.ProjectID, conf.Granularity, conf.Period, client, conf.EnableCache, m.logger)

	utils.RunOnInterval(ctx, func() {
		processes := m.processGetter.GetProcesses(ctx, timeout)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			processMeasurements = m.processGetter.GetMeasurements(ctx, timeout, processes)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			diskMeasurements = m.diskGetter.GetMeasurements(ctx, timeout, processes)
		}()

		wg.Wait()

		var dps = make([]*datapoint.Datapoint, 0)

		// Creating metric datapoints from the 1 minute resolution process measurement datapoints
		for k, v := range processMeasurements {
			dps = append(dps, newDps(k, v, "")...)
		}

		// Creating metric datapoints from the 1 minute resolution disk measurement datapoints
		for k, v := range diskMeasurements {
			dps = append(dps, newDps(k, v.Measurements, v.PartitionName)...)
		}

		m.Output.SendDatapoints(dps...)

	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown the monitor
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

func newDigestClient(publicKey, privateKey string) (*mongodbatlas.Client, error) {
	//Setup a transport to handle digest
	transport := digest.NewTransport(publicKey, privateKey)

	client, err := transport.Client()
	if err != nil {
		return nil, err
	}

	return mongodbatlas.NewClient(client), nil
}

func newDps(process measurements.Process, measurementsArr []*mongodbatlas.Measurements, partitionName string) []*datapoint.Datapoint {
	var dps = make([]*datapoint.Datapoint, 0)

	var dimensions = newDimensions(&process, partitionName)

	for _, measures := range measurementsArr {
		metricValue := newFloatValue(measures.DataPoints)

		if metricValue == nil || metricsMap[measures.Name] == "" {
			continue
		}

		dp := &datapoint.Datapoint{
			Metric:     metricsMap[measures.Name],
			MetricType: datapoint.Gauge,
			Value:      metricValue,
			Dimensions: dimensions,
		}

		dps = append(dps, dp)
	}

	return dps
}

func newFloatValue(dataPoints []*mongodbatlas.DataPoints) datapoint.FloatValue {
	if len(dataPoints) == 0 {
		return nil
	}

	var timestamp = dataPoints[0].Timestamp

	var value = dataPoints[0].Value

	// Getting the latest non nil value
	for i := 1; i < len(dataPoints); i++ {
		if dataPoints[i].Timestamp > timestamp && dataPoints[i].Value != nil {
			value = dataPoints[i].Value
			timestamp = dataPoints[i].Timestamp
		}
	}

	if value == nil {
		return nil
	}

	return datapoint.NewFloatValue(float64(*value))
}

func newDimensions(process *measurements.Process, partitionName string) map[string]string {
	var dimensions = map[string]string{"process_id": process.ID, "project_id": process.ProjectID, "host": process.Host, "port": strconv.Itoa(process.Port), "type_name": process.TypeName}

	if process.ReplicaSetName != "" {
		dimensions["replica_set_name"] = process.ReplicaSetName
	}

	if process.ShardName != "" {
		dimensions["shard_name"] = process.ShardName
	}

	if partitionName != "" {
		dimensions["partition_name"] = partitionName
	}

	return dimensions
}
