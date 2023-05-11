package docker

import (
	"context"
	"sync"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

// ContainerChangeHandler is what gets called when a Docker container is
// initially recognized or changed in some way.  old will be the previous state,
// or nil if no previous state is known.  new is the new state, or nil if the
// container is being destroyed.
type ContainerChangeHandler func(old *dtypes.ContainerJSON, new *dtypes.ContainerJSON)

// ListAndWatchContainers accepts a changeHandler that gets called as containers come and go.
func ListAndWatchContainers(ctx context.Context, client *docker.Client, changeHandler ContainerChangeHandler, imageFilter filter.StringFilter, logger log.FieldLogger, syncTime time.Duration) {
	lock := sync.Mutex{}
	containers := make(map[string]*dtypes.ContainerJSON)

	// Make sure you hold the lock before calling this
	updateContainer := func(id string) bool {
		inspectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		c, err := client.ContainerInspect(inspectCtx, id)
		defer cancel()
		if err != nil {
			logger.WithError(err).Errorf("Could not inspect updated container %s", id)
		} else if imageFilter == nil || !imageFilter.Matches(c.Config.Image) {
			logger.Debugf("Updated Docker container %s", id)
			containers[id] = &c
			return true
		}
		return false
	}

	syncContainerList := func() error {
		f := filters.NewArgs()
		f.Add("status", "running")
		options := dtypes.ContainerListOptions{
			Filters: f,
		}
		containerList, err := client.ContainerList(ctx, options)
		if err != nil {
			return err
		}

		type void struct{}
		containersMap := make(map[string]void, len(containerList))

		wg := sync.WaitGroup{}
		for i := range containerList {
			// The Docker API has a different return type for list vs. inspect, and
			// no way to get the return type of list for individual containers,
			// which makes this harder than it should be.
			// Add new entries and skip containers that are already cached
			if _, ok := containers[containerList[i].ID]; !ok {
				wg.Add(1)
				go func(id string) {
					lock.Lock()
					updateContainer(id)
					changeHandler(nil, containers[id])
					lock.Unlock()
					wg.Done()
				}(containerList[i].ID)
			}
			// Map will be used to find the delta
			containersMap[containerList[i].ID] = void{}
		}
		wg.Wait()
		// Find stale entries and delete them
		for containerID := range containers {
			if _, ok := containersMap[containerID]; !ok {
				logger.Debugf("Docker container %s no longer exists, removing from cache", containerID)
				changeHandler(containers[containerID], nil)
				delete(containers, containerID)
			}
		}
		return nil
	}

	go func() {
		refreshTicker := time.NewTicker(syncTime)
		defer refreshTicker.Stop()
		// This pattern is taken from
		// https://github.com/docker/cli/blob/master/cli/command/container/stats.go
		f := filters.NewArgs()
		f.Add("type", "container")
		f.Add("event", "destroy")
		f.Add("event", "die")
		f.Add("event", "pause")
		f.Add("event", "stop")
		f.Add("event", "start")
		f.Add("event", "unpause")
		f.Add("event", "update")
		lastTime := time.Now()

	START_STREAM:
		for {
			since := lastTime.Format(time.RFC3339Nano)
			options := dtypes.EventsOptions{
				Filters: f,
				Since:   since,
			}

			logger.Infof("Watching for Docker events since %s", since)
			eventCh, errCh := client.Events(ctx, options)

			err := syncContainerList()
			if err != nil {
				logger.WithError(err).Error("Error while syncing container cache")
			}

			for {
				select {
				case <-refreshTicker.C:
					logger.Debugf("sync container cache")
					err := syncContainerList()
					if err != nil {
						logger.WithError(err).Error("Error while periodically syncing container cache")
					}

				case event := <-eventCh:
					lock.Lock()

					switch event.Action {
					// This assumes that all deleted containers get a "destroy"
					// event associated with them, otherwise memory usage could
					// be unbounded.
					case "destroy":
						logger.Debugf("Docker container was destroyed: %s", event.ID)
						if _, ok := containers[event.ID]; ok {
							changeHandler(containers[event.ID], nil)
							delete(containers, event.ID)
						}
					default:
						oldContainer := containers[event.ID]
						if updateContainer(event.ID) {
							changeHandler(oldContainer, containers[event.ID])
						}
					}

					lock.Unlock()

					lastTime = time.Unix(0, event.TimeNano)

				case err := <-errCh:
					logger.WithError(err).Error("Error watching docker container events")
					time.Sleep(3 * time.Second)
					continue START_STREAM

				case <-ctx.Done():
					// Event stream is tied to the same context and will quit
					// also.
					return
				}
			}
		}
	}()
}
