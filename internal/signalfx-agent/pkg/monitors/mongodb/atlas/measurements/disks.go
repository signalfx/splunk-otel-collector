package measurements

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	log "github.com/sirupsen/logrus"
)

// DisksMeasurements are the metric measurements of a particular disk partition in a MongoDB process host.
type DisksMeasurements map[Process]struct {
	PartitionName string
	Measurements  []*mongodbatlas.Measurements
}

// DisksGetter is for fetching metric measurements of disk partitions in the MongoDB processes hosts.
type DisksGetter interface {
	GetMeasurements(ctx context.Context, timeout time.Duration, processes []Process) DisksMeasurements
}

// disksGetter implements DisksGetter
type disksGetter struct {
	projectID         string
	granularity       string
	period            string
	client            *mongodbatlas.Client
	enableCache       bool
	mutex             *sync.Mutex
	measurementsCache *atomic.Value
	disksCache        *atomic.Value
	logger            log.FieldLogger
}

// NewDisksGetter returns a new DisksGetter.
func NewDisksGetter(projectID string, granularity string, period string, client *mongodbatlas.Client, enableCache bool, logger log.FieldLogger) DisksGetter {
	return &disksGetter{
		projectID:         projectID,
		granularity:       granularity,
		period:            period,
		client:            client,
		enableCache:       enableCache,
		mutex:             new(sync.Mutex),
		measurementsCache: new(atomic.Value),
		disksCache:        new(atomic.Value),
		logger:            logger,
	}
}

// GetMeasurements gets metric measurements of disk partitions in the hosts of the given MongoDB processes.
func (getter *disksGetter) GetMeasurements(ctx context.Context, timeout time.Duration, processes []Process) DisksMeasurements {
	var measurements = make(DisksMeasurements)

	partitions := getter.getPartitions(ctx, timeout, processes)

	var wg1 sync.WaitGroup

	wg1.Add(1)

	go func() {
		defer wg1.Done()

		var wg2 sync.WaitGroup
		for process, partitionNames := range partitions {
			for _, partitionName := range partitionNames {
				wg2.Add(1)

				go func(process Process, partitionName string) {
					defer wg2.Done()

					var ctx, cancel = context.WithTimeout(ctx, timeout)
					defer cancel()

					getter.setMeasurements(ctx, measurements, process, partitionName, 1)
				}(process, partitionName)
			}
		}
		wg2.Wait()

		if getter.enableCache {
			getter.measurementsCache.Store(measurements)
		}
	}()

	if getter.measurementsCache.Load() != nil && getter.enableCache {
		return getter.measurementsCache.Load().(DisksMeasurements)
	}

	wg1.Wait()

	return measurements
}

// getPartitions is a helper function for fetching the names of disk partitions is the hosts of given MongoDB processes.
func (getter *disksGetter) getPartitions(ctx context.Context, timeout time.Duration, processes []Process) map[Process][]string {
	var partitions = make(map[Process][]string)

	var wg1 sync.WaitGroup

	wg1.Add(1)

	go func() {
		defer wg1.Done()

		var wg2 sync.WaitGroup
		for _, process := range processes {
			wg2.Add(1)

			go func(process Process) {
				defer wg2.Done()

				var ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()

				partitionNames := getter.getPartitionNames(ctx, process, 1)

				getter.mutex.Lock()
				defer getter.mutex.Unlock()
				partitions[process] = partitionNames
			}(process)
		}
		wg2.Wait()

		if getter.enableCache {
			getter.disksCache.Store(partitions)
		}
	}()

	if getter.disksCache.Load() != nil && getter.enableCache {
		return getter.disksCache.Load().(map[Process][]string)
	}

	wg1.Wait()

	return partitions
}

// getPartitionNames is a helper function of function getPartitions.
func (getter *disksGetter) getPartitionNames(ctx context.Context, process Process, page int) (names []string) {
	list, resp, err := getter.client.ProcessDisks.List(ctx, getter.projectID, process.Host, process.Port, &mongodbatlas.ListOptions{PageNum: page})

	if msg, err := errorMsg(err, resp); err != nil {
		getter.logger.WithError(err).Errorf(msg, "disk partition names", getter.projectID, process.Host, process.Port)
		return names
	}

	if ok, next := nextPage(resp, getter.logger); ok {
		names = append(names, getter.getPartitionNames(ctx, process, next)...)
	}

	for _, r := range list.Results {
		names = append(names, r.PartitionName)
	}

	return names
}

// setMeasurements is a helper function of method GetMeasurements.
func (getter *disksGetter) setMeasurements(ctx context.Context, disksMeasurements DisksMeasurements, process Process, partitionName string, pageNum int) {
	var opts = &mongodbatlas.ProcessMeasurementListOptions{ListOptions: &mongodbatlas.ListOptions{PageNum: pageNum}, Granularity: getter.granularity, Period: getter.period}

	list, resp, err := getter.client.ProcessDiskMeasurements.List(ctx, getter.projectID, process.Host, process.Port, partitionName, opts)

	if msg, err := errorMsg(err, resp); err != nil {
		getter.logger.WithError(err).Errorf(msg, "disk measurements", getter.projectID, process.Host, process.Port)
		return
	}

	if ok, next := nextPage(resp, getter.logger); ok {
		getter.setMeasurements(ctx, disksMeasurements, process, partitionName, next)
	}

	getter.mutex.Lock()
	defer getter.mutex.Unlock()

	disksMeasurements[process] = struct {
		PartitionName string
		Measurements  []*mongodbatlas.Measurements
	}{PartitionName: partitionName, Measurements: list.Measurements}
}
