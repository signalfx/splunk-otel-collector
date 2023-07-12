package measurements

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	log "github.com/sirupsen/logrus"
)

// ProcessesMeasurements are the metric measurements of a particular MongoDB Process.
type ProcessesMeasurements map[Process][]*mongodbatlas.Measurements

// ProcessesGetter is for getting metric measurements of MongoDB processes.
type ProcessesGetter interface {
	GetProcesses(ctx context.Context, timeout time.Duration) []Process
	GetMeasurements(ctx context.Context, timeout time.Duration, processes []Process) ProcessesMeasurements
}

// processesGetter implements ProcessesGetter
type processesGetter struct {
	projectID         string
	granularity       string
	period            string
	client            *mongodbatlas.Client
	enableCache       bool
	mutex             *sync.Mutex
	measurementsCache *atomic.Value
	processesCache    *atomic.Value
	logger            log.FieldLogger
}

// NewProcessesGetter returns a new ProcessesGetter.
func NewProcessesGetter(projectID string, granularity string, period string, client *mongodbatlas.Client, enableCache bool, logger log.FieldLogger) ProcessesGetter {
	return &processesGetter{
		projectID:         projectID,
		granularity:       granularity,
		period:            period,
		client:            client,
		enableCache:       enableCache,
		mutex:             new(sync.Mutex),
		measurementsCache: new(atomic.Value),
		processesCache:    new(atomic.Value),
		logger:            logger,
	}
}

// GetProcesses gets all MongoDB processes in the configured project ID.
func (getter *processesGetter) GetProcesses(ctx context.Context, timeout time.Duration) (processes []Process) {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		processes = getter.getProcessesHelper(ctx, 1)

		if getter.enableCache {
			getter.processesCache.Store(processes)
		}
	}()

	if getter.processesCache.Load() != nil && getter.enableCache {
		return getter.processesCache.Load().([]Process)
	}

	wg.Wait()

	return processes
}

// getProcessesHelper is a helper function for method get.
func (getter *processesGetter) getProcessesHelper(ctx context.Context, pageNum int) (processes []Process) {
	list, resp, err := getter.client.Processes.List(ctx, getter.projectID, &mongodbatlas.ListOptions{PageNum: pageNum})

	if err != nil {
		getter.logger.WithError(err).Errorf("the request for getting processes failed (Atlas project: %s)", getter.projectID)
		return
	}

	if resp == nil {
		getter.logger.Errorf("the response for getting processes returned empty (Atlas project: %s)", getter.projectID)
		return
	}

	if err := mongodbatlas.CheckResponse(resp.Response); err != nil {
		getter.logger.WithError(err).Errorf("the response for getting processes returned an error (Atlas project: %s)", getter.projectID)
		return
	}

	if ok, next := nextPage(resp, getter.logger); ok {
		processes = append(processes, getter.getProcessesHelper(ctx, next)...)
	}

	for _, p := range list {
		processes = append(processes, Process{ID: p.ID, ProjectID: p.GroupID, Host: p.Hostname, Port: p.Port, ShardName: p.ShardName, ReplicaSetName: p.ReplicaSetName, TypeName: p.TypeName})
	}

	return processes
}

// GetMeasurements gets metric measurements of the given MongoDB processes.
func (getter *processesGetter) GetMeasurements(ctx context.Context, timeout time.Duration, processes []Process) ProcessesMeasurements {
	var processesMeasurements = make(ProcessesMeasurements)

	var wg1 sync.WaitGroup
	wg1.Add(1)

	// Get measurements and update cache asynchronously.
	go func() {
		defer wg1.Done()

		var wg2 sync.WaitGroup
		for _, process := range processes {
			wg2.Add(1)

			go func(process Process) {
				defer wg2.Done()

				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				getter.setMeasurements(ctx, process, processesMeasurements, 1)
			}(process)
		}
		wg2.Wait()

		if getter.enableCache {
			getter.measurementsCache.Store(processesMeasurements)
		}
	}()

	// Get current cached measurements
	if getter.measurementsCache.Load() != nil && getter.enableCache {
		return getter.measurementsCache.Load().(ProcessesMeasurements)
	}

	wg1.Wait()

	return processesMeasurements
}

// setMeasurements is a helper function of method GetMeasurements.
func (getter *processesGetter) setMeasurements(ctx context.Context, process Process, processesMeasurements ProcessesMeasurements, pageNum int) {
	var opts = &mongodbatlas.ProcessMeasurementListOptions{ListOptions: &mongodbatlas.ListOptions{PageNum: pageNum}, Granularity: getter.granularity, Period: getter.period}

	list, resp, err := getter.client.ProcessMeasurements.List(ctx, getter.projectID, process.Host, process.Port, opts)

	if msg, err := errorMsg(err, resp); err != nil {
		getter.logger.WithError(err).Errorf(msg, "process measurements", getter.projectID, process.Host, process.Port)
		return
	}

	if ok, next := nextPage(resp, getter.logger); ok {
		getter.setMeasurements(ctx, process, processesMeasurements, next)
	}

	getter.mutex.Lock()
	defer getter.mutex.Unlock()

	processesMeasurements[process] = list.Measurements
}
