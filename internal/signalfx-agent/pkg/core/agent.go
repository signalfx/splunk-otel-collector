// Package core contains the central frame of the agent that hooks up the
// various subsystems.
package core

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/hostid"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/core/writer"
	"github.com/signalfx/signalfx-agent/pkg/core/writer/tracetracker"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/observers"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

//go:generate sh -c "test -e `go env GOPATH`/bin/genny > /dev/null || (cd ../.. && go install github.com/mauricelam/genny)"

const (
	// Items should stay in these channels only very briefly as there should be
	// goroutines dedicated to pulling them at all times.  Having the capacity
	// be non-zero is more an optimization to keep monitors from seizing up
	// under extremely heavy load.
	datapointChanCapacity = 3000
	eventChanCapacity     = 100
	dimensionChanCapacity = 100
	traceSpanChanCapacity = 3000
)

// Agent is what hooks up observers, monitors, and the datapoint writer.
type Agent struct {
	observers           *observers.ObserverManager
	monitors            *monitors.MonitorManager
	writer              *writer.MultiWriter
	meta                *meta.AgentMeta
	lastConfig          *config.Config
	dpChan              chan []*datapoint.Datapoint
	eventChan           chan *event.Event
	dimensionChan       chan *types.Dimension
	spanChan            chan []*trace.Span
	endpointHostTracker *services.EndpointHostTracker

	diagnosticServer     *http.Server
	profileServerRunning bool
	startTime            time.Time
}

// NewAgent creates an unconfigured agent instance
func NewAgent() *Agent {
	agent := Agent{
		dpChan:              make(chan []*datapoint.Datapoint, datapointChanCapacity),
		eventChan:           make(chan *event.Event, eventChanCapacity),
		dimensionChan:       make(chan *types.Dimension, dimensionChanCapacity),
		spanChan:            make(chan []*trace.Span, traceSpanChanCapacity),
		endpointHostTracker: services.NewEndpointHostTracker(),
		startTime:           time.Now(),
	}

	agent.observers = &observers.ObserverManager{
		CallbackTargets: &observers.ServiceCallbacks{
			Added:   agent.endpointAdded,
			Removed: agent.endpointRemoved,
		},
	}

	agent.meta = &meta.AgentMeta{}
	agent.monitors = monitors.NewMonitorManager(agent.meta)
	agent.monitors.DPs = agent.dpChan
	agent.monitors.Events = agent.eventChan
	agent.monitors.DimensionUpdates = agent.dimensionChan
	agent.monitors.TraceSpans = agent.spanChan
	return &agent
}

func (a *Agent) configure(conf *config.Config) {
	log.SetFormatter(conf.Logging.LogrusFormatter())

	level := conf.Logging.LogrusLevel()
	if level != nil {
		log.SetLevel(*level)
	}

	log.Infof("Using log level %s", log.GetLevel().String())

	hostDims := map[string]string{}
	if !conf.DisableHostDimensions {
		hostDims = hostid.Dimensions(conf)
		log.Infof("Using host id dimensions %v", hostDims)
		conf.Writer.HostIDDims = hostDims
	}

	if conf.EnableProfiling {
		a.ensureProfileServerRunning(conf.ProfilingHost, conf.ProfilingPort)
	}

	if a.lastConfig == nil || a.lastConfig.Writer.Hash() != conf.Writer.Hash() || a.lastConfig.Cluster != conf.Cluster {
		if a.writer != nil {
			a.writer.Shutdown()
		}

		spanSourceTracker := tracetracker.NewSpanSourceTracker(a.endpointHostTracker, a.dimensionChan, conf.Cluster)

		var err error
		a.writer, err = writer.New(
			&conf.Writer,
			a.dpChan,
			a.eventChan,
			a.dimensionChan,
			a.spanChan,
			spanSourceTracker)
		if err != nil {
			// This is a catastrophic error if we can't write datapoints.
			log.WithError(err).Error("Could not configure SignalFx datapoint writer, unable to start up")
			os.Exit(4)
		}
		a.writer.Start()
	}

	if conf.Cluster != "" {
		startSyncClusterProperty(a.dimensionChan, conf.Cluster, hostDims, conf.SyncClusterOnHostDimension)
	}

	a.meta.InternalStatusHost = conf.InternalStatusHost
	a.meta.InternalStatusPort = conf.InternalStatusPort

	// The order of Configure calls is very important!
	a.monitors.Configure(conf.Monitors, &conf.Collectd, conf.IntervalSeconds)
	a.observers.Configure(conf.Observers)
	a.lastConfig = conf
}

func (a *Agent) endpointAdded(service services.Endpoint) {
	a.endpointHostTracker.EndpointAdded(service)
	a.monitors.EndpointAdded(service)
}

func (a *Agent) endpointRemoved(service services.Endpoint) {
	a.monitors.EndpointRemoved(service)
	a.endpointHostTracker.EndpointRemoved(service)
}

func (a *Agent) shutdown() {
	a.observers.Shutdown()
	a.monitors.Shutdown()
	//neopy.Instance().Shutdown()
	a.writer.Shutdown()
	if a.diagnosticServer != nil {
		a.diagnosticServer.Close()
	}
}

// Startup the agent.  Returns a function that can be called to shutdown the
// agent, as well as a channel that will be notified when the agent has
// shutdown.
func Startup(configPath string) (context.CancelFunc, <-chan struct{}) {
	cwc, cancel := context.WithCancel(context.Background())

	configLoads, err := config.LoadConfig(cwc, configPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"configPath": configPath,
		}).Error("Error loading main config")
		os.Exit(1)
	}

	agent := NewAgent()

	shutdownComplete := make(chan struct{})

	go func(ctx context.Context) {
		for {
			select {
			case config := <-configLoads:
				log.Info("New config loaded")

				if config == nil {
					log.WithFields(log.Fields{
						"path": configPath,
					}).Error("Failed to load config, cannot continue!")
					os.Exit(2)
				}

				agent.configure(config)
				log.Info("Done configuring agent")

				if config.InternalStatusHost != "" {
					agent.serveDiagnosticInfo(config.InternalStatusHost, config.InternalStatusPort)
				}

			case <-ctx.Done():
				agent.shutdown()
				close(shutdownComplete)
				return
			}
		}
	}(cwc)

	return cancel, shutdownComplete
}

// Status reads the text from the diagnostic socket and returns it if available.
func Status(configPath string, section string) ([]byte, error) {
	configLoads, err := config.LoadConfig(context.Background(), configPath)
	if err != nil {
		return nil, err
	}

	conf := <-configLoads
	return readStatusInfo(conf.InternalStatusHost, conf.InternalStatusPort, section)
}

// StreamDatapoints reads the text from the diagnostic socket and returns it if available.
func StreamDatapoints(configPath string, metric string, dims string) (io.ReadCloser, error) {
	configLoads, err := config.LoadConfig(context.Background(), configPath)
	if err != nil {
		return nil, err
	}

	conf := <-configLoads
	return streamDatapoints(conf.InternalStatusHost, conf.InternalStatusPort, metric, dims)
}

func startSyncClusterProperty(dimChan chan *types.Dimension, cluster string, hostDims map[string]string, setOnHost bool) {
	hostDims = utils.CloneStringMap(hostDims)
	// Exclude kubernetes_node_uid since it is also managed by the
	// kubernetes-cluster monitor and it is very tricky getting them to
	// merge cleanly without it clobbing the cluster property.
	delete(hostDims, "kubernetes_node_uid")

	for dimName, dimValue := range hostDims {
		if len(hostDims) > 1 && dimName == "host" && !setOnHost {
			// If we also have a platform-specific host-id dimension that isn't
			// the generic 'host' dimension, then skip setting the property on
			// 'host' since it tends to get reused frequently. The property
			// will still show up on all MTSs that come out of this agent.
			continue
		}
		log.Infof("Setting cluster:%s property on %s:%s dimension", cluster, dimName, dimValue)
		dimChan <- &types.Dimension{
			Name:  dimName,
			Value: dimValue,
			Properties: map[string]string{
				"cluster": cluster,
			},
			MergeIntoExisting: true,
		}
	}
}
