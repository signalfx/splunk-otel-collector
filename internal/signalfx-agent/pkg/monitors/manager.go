package monitors

import (
	"fmt"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// MonitorManager coordinates the startup and shutdown of monitors based on the
// configuration provided by the user.  Monitors that have discovery rules can
// be injected with multiple services.  If a monitor does not have a discovery
// rule (a "static" monitor), it will be started immediately (as soon as
// Configure is called).
type MonitorManager struct {
	monitorConfigs map[uint64]config.MonitorCustomConfig
	// Keep track of which services go with which monitor
	activeMonitors []*ActiveMonitor
	badConfigs     map[uint64]*config.MonitorConfig
	lock           sync.Mutex
	// Map of service endpoints that have been discovered
	discoveredEndpoints map[services.ID]services.Endpoint

	collectdConfig     *config.CollectdConfig
	collectdConfigured bool

	DPs              chan<- []*datapoint.Datapoint
	Events           chan<- *event.Event
	DimensionUpdates chan<- *types.Dimension
	TraceSpans       chan<- []*trace.Span

	// TODO: AgentMeta is rather hacky so figure out a better way to share agent
	// metadata with monitors
	agentMeta       *meta.AgentMeta
	intervalSeconds int

	idGenerator func() string
}

// NewMonitorManager creates a new instance of the MonitorManager
func NewMonitorManager(agentMeta *meta.AgentMeta) *MonitorManager {
	return &MonitorManager{
		monitorConfigs:      make(map[uint64]config.MonitorCustomConfig),
		activeMonitors:      make([]*ActiveMonitor, 0),
		badConfigs:          make(map[uint64]*config.MonitorConfig),
		discoveredEndpoints: make(map[services.ID]services.Endpoint),
		idGenerator:         utils.NewIDGenerator(),
		agentMeta:           agentMeta,
	}
}

// Configure receives a list of monitor configurations.  It will start up any
// static monitors and watch discovered services to see if any match dynamic
// monitors.
func (mm *MonitorManager) Configure(confs []config.MonitorConfig, collectdConf *config.CollectdConfig, intervalSeconds int) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	mm.intervalSeconds = intervalSeconds
	for i := range confs {
		confs[i].IntervalSeconds = utils.FirstNonZero(confs[i].IntervalSeconds, intervalSeconds)
	}

	requireSoloTrue := anyMarkedSolo(confs)

	newConfig, deletedHashes := diffNewConfig(confs, mm.allConfigHashes())
	mm.collectdConfig = collectdConf

	if !collectdConf.DisableCollectd {
		// By configuring collectd with the monitor manager, we absolve the monitor
		// instances of having to know about collectd config, which makes it easier
		// to create monitor config from disparate sources such as from observers.
		if err := collectd.ConfigureMainCollectd(collectdConf); err != nil {
			log.WithFields(log.Fields{
				"error":          err,
				"collectdConfig": spew.Sdump(collectdConf),
			}).Error("Could not configure collectd")
		}
		mm.collectdConfigured = true
	}

	for _, hash := range deletedHashes {
		mm.deleteMonitorsByConfigHash(hash)

		delete(mm.monitorConfigs, hash)
		delete(mm.badConfigs, hash)
	}

	for i := range newConfig {
		conf := newConfig[i]
		hash := conf.Hash()

		if requireSoloTrue && !conf.Solo {
			log.Infof("Solo mode is active, skipping monitor of type %s", conf.Type)
			continue
		}

		monConfig, err := mm.handleNewConfig(&conf)
		if err != nil {
			log.WithFields(log.Fields{
				"monitorType": conf.Type,
				"error":       err,
			}).Error("Could not process configuration for monitor")
			conf.ValidationError = err.Error()
			mm.badConfigs[hash] = &conf
			continue
		}

		mm.monitorConfigs[hash] = monConfig
	}
}

func (mm *MonitorManager) allConfigHashes() map[uint64]bool {
	hashes := make(map[uint64]bool)
	for h := range mm.monitorConfigs {
		hashes[h] = true
	}
	for h := range mm.badConfigs {
		hashes[h] = true
	}
	return hashes
}

// Returns the any new configs and any removed config hashes
func diffNewConfig(confs []config.MonitorConfig, oldHashes map[uint64]bool) ([]config.MonitorConfig, []uint64) {
	newConfigHashes := make(map[uint64]bool)
	var newConfig []config.MonitorConfig
	for i := range confs {
		hash := confs[i].Hash()
		if !oldHashes[hash] {
			newConfig = append(newConfig, confs[i])
		}

		if newConfigHashes[hash] {
			log.WithFields(log.Fields{
				"monitorType": confs[i].Type,
				"config":      confs[i],
			}).Error("Monitor config is duplicated")
			continue
		}

		newConfigHashes[hash] = true
	}

	var deletedHashes []uint64
	for hash := range oldHashes {
		// If we didn't see it in the latest config slice then we need to
		// delete anything using it.
		if !newConfigHashes[hash] {
			deletedHashes = append(deletedHashes, hash)
		}
	}

	return newConfig, deletedHashes
}

func (mm *MonitorManager) handleNewConfig(conf *config.MonitorConfig) (config.MonitorCustomConfig, error) {
	monConfig, err := getCustomConfigForMonitor(conf)
	if err != nil {
		return nil, err
	}

	if configOnlyAllowsSingleInstance(monConfig) {
		if len(mm.monitorConfigsForType(conf.Type)) > 0 {
			return nil, fmt.Errorf("monitor type %s only allows a single instance at a time", conf.Type)
		}
	}

	// No discovery rule means that the monitor should run from the start
	if conf.DiscoveryRule == "" {
		return monConfig, mm.createAndConfigureNewMonitor(monConfig, nil)
	}

	mm.makeMonitorsForMatchingEndpoints(monConfig)
	// We need to go and see if any discovered endpoints should be
	// monitored by this config, if they aren't already.
	return monConfig, nil
}

func (mm *MonitorManager) makeMonitorsForMatchingEndpoints(conf config.MonitorCustomConfig) {
	for id, endpoint := range mm.discoveredEndpoints {
		// Self configured endpoints are monitored immediately upon being
		// created and never need to be matched against discovery rules.
		if endpoint.Core().IsSelfConfigured() {
			continue
		}

		log.WithFields(log.Fields{
			"monitorType":   conf.MonitorConfigCore().Type,
			"discoveryRule": conf.MonitorConfigCore().DiscoveryRule,
			"endpoint":      endpoint,
		}).Debug("Trying to find config that matches discovered endpoint")

		if mm.isEndpointIDMonitoredByConfig(conf, id) {
			log.Debug("The endpoint is already monitored")
			continue
		}

		if matched, err := mm.monitorEndpointIfRuleMatches(conf, endpoint); matched {
			if err != nil {
				log.WithFields(log.Fields{
					"error":       err,
					"endpointID":  endpoint.Core().ID,
					"monitorType": conf.MonitorConfigCore().Type,
				}).Error("Error monitoring endpoint that matched rule")
			} else {
				log.WithFields(log.Fields{
					"endpointID":  endpoint.Core().ID,
					"monitorType": conf.MonitorConfigCore().Type,
				}).Info("Now monitoring discovered endpoint")
			}
		} else {
			log.Debug("The monitor did not match")
		}
	}
}

func (mm *MonitorManager) isEndpointIDMonitoredByConfig(conf config.MonitorCustomConfig, id services.ID) bool {
	for _, am := range mm.activeMonitors {
		if conf.MonitorConfigCore().Hash() == am.configHash {
			if am.endpointID() == id {
				return true
			}
		}
	}
	return false
}

// Returns true is the service did match a rule in this monitor config
func (mm *MonitorManager) monitorEndpointIfRuleMatches(config config.MonitorCustomConfig, endpoint services.Endpoint) (bool, error) {
	if config.MonitorConfigCore().DiscoveryRule == "" || !services.DoesServiceMatchRule(endpoint, config.MonitorConfigCore().DiscoveryRule, config.MonitorConfigCore().ShouldValidateDiscoveryRule()) {
		return false, nil
	}

	err := mm.createAndConfigureNewMonitor(config, endpoint)
	if err != nil {
		return true, err
	}

	return true, nil
}

// EndpointAdded should be called when a new service is discovered
func (mm *MonitorManager) EndpointAdded(endpoint services.Endpoint) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	ensureProxyingDisabledForService(endpoint)
	mm.discoveredEndpoints[endpoint.Core().ID] = endpoint

	// If the endpoint has a monitor type specified, then it is expected to
	// have all of its configuration already set in the endpoint and discovery
	// rules will be ignored.
	if endpoint.Core().IsSelfConfigured() {
		if err := mm.monitorSelfConfiguredEndpoint(endpoint); err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"monitorType": endpoint.Core().MonitorType,
				"endpoint":    endpoint,
			}).Error("Could not create monitor for self-configured endpoint")
		}
		return
	}

	mm.findConfigForMonitorAndRun(endpoint)
}

func (mm *MonitorManager) monitorSelfConfiguredEndpoint(endpoint services.Endpoint) error {
	monitorType := endpoint.Core().MonitorType
	conf := &config.MonitorConfig{
		Type: monitorType,
		// This will get overridden by the endpoint configuration if interval
		// was specified
		IntervalSeconds: mm.intervalSeconds,
	}

	monConfig, err := getCustomConfigForMonitor(conf)
	if err != nil {
		return err
	}

	if err = mm.createAndConfigureNewMonitor(monConfig, endpoint); err != nil {
		return err
	}
	return nil
}

func (mm *MonitorManager) findConfigForMonitorAndRun(endpoint services.Endpoint) {
	monitoring := false

	for _, config := range mm.monitorConfigs {
		matched, err := mm.monitorEndpointIfRuleMatches(config, endpoint)
		monitoring = matched || monitoring
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err,
				"config":   config,
				"endpoint": endpoint,
			}).Error("Could not monitor new endpoint")
		}
	}

	if !monitoring {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
		}).Debug("Endpoint added that doesn't match any discovery rules")
	}
}

// endpoint may be nil for static monitors
func (mm *MonitorManager) createAndConfigureNewMonitor(config config.MonitorCustomConfig, endpoint services.Endpoint) error {
	id := types.MonitorID(mm.idGenerator())
	coreConfig := config.MonitorConfigCore()
	monitorType := coreConfig.Type

	// This accounts for the possibility that collectd was marked as disabled
	// when the agent was first started but later an self-configured endpoint
	// that depends on collectd is discovered.
	if coreConfig.IsCollectdBased() && !mm.collectdConfigured {
		if err := collectd.ConfigureMainCollectd(mm.collectdConfig); err != nil {
			log.WithFields(log.Fields{
				"error":          err,
				"collectdConfig": spew.Sdump(mm.collectdConfig),
			}).Error("Could not configure collectd")
		}
		mm.collectdConfigured = true
	}

	log.WithFields(log.Fields{
		"monitorType":   monitorType,
		"discoveryRule": coreConfig.DiscoveryRule,
		"monitorID":     id,
	}).Info("Creating new monitor")

	instance := newMonitor(config.MonitorConfigCore().Type)
	if instance == nil {
		return fmt.Errorf("Could not create new monitor of type %s", monitorType)
	}

	// Make metadata nil if we aren't using built in filtering and then none of
	// the new filtering logic will apply.
	metadata, ok := MonitorMetadatas[monitorType]
	if !ok || metadata == nil {
		// This indicates a programming error in not specifying metadata, not
		// bad user input
		panic(fmt.Sprintf("could not find monitor metadata of type %s", monitorType))
	}

	configHash := config.MonitorConfigCore().Hash()

	renderedConf, err := renderConfig(config, endpoint)
	if err != nil {
		return err
	}

	monFiltering, err := newMonitorFiltering(renderedConf, metadata)
	if err != nil {
		return err
	}

	am := &ActiveMonitor{
		id:         id,
		configHash: configHash,
		instance:   instance,
		endpoint:   endpoint,
		agentMeta:  mm.agentMeta,
	}

	metricNameTransformations, err := renderedConf.MonitorConfigCore().MetricNameExprs()
	if err != nil {
		return err
	}

	output := &monitorOutput{
		monitorType:               renderedConf.MonitorConfigCore().Type,
		monitorID:                 id,
		notHostSpecific:           renderedConf.MonitorConfigCore().DisableHostDimensions,
		disableEndpointDimensions: renderedConf.MonitorConfigCore().DisableEndpointDimensions,
		configHash:                configHash,
		endpoint:                  endpoint,
		dpChan:                    mm.DPs,
		eventChan:                 mm.Events,
		dimensionChan:             mm.DimensionUpdates,
		spanChan:                  mm.TraceSpans,
		extraDims:                 map[string]string{},
		extraSpanTags:             map[string]string{},
		defaultSpanTags:           map[string]string{},
		dimensionTransformations:  renderedConf.MonitorConfigCore().DimensionTransformations,
		metricNameTransformations: metricNameTransformations,
		monitorFiltering:          monFiltering,
	}

	am.output = output

	if err := am.configureMonitor(renderedConf); err != nil {
		return err
	}
	mm.activeMonitors = append(mm.activeMonitors, am)

	return nil
}

func (mm *MonitorManager) monitorsForEndpointID(id services.ID) (out []*ActiveMonitor) {
	for i := range mm.activeMonitors {
		if mm.activeMonitors[i].endpointID() == id {
			out = append(out, mm.activeMonitors[i])
		}
	}
	return // Named return value
}

func (mm *MonitorManager) monitorConfigsForType(monitorType string) []*config.MonitorCustomConfig {
	var out []*config.MonitorCustomConfig
	for i := range mm.monitorConfigs {
		conf := mm.monitorConfigs[i]
		if conf.MonitorConfigCore().Type == monitorType {
			out = append(out, &conf)
		}
	}
	return out
}

// EndpointRemoved should be called by observers when a service endpoint was
// removed.
func (mm *MonitorManager) EndpointRemoved(endpoint services.Endpoint) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	delete(mm.discoveredEndpoints, endpoint.Core().ID)

	monitors := mm.monitorsForEndpointID(endpoint.Core().ID)
	for _, am := range monitors {
		am.doomed = true
	}
	mm.deleteDoomedMonitors()

	log.WithFields(log.Fields{
		"endpoint": endpoint,
	}).Debug("No longer considering endpoint")
}

func (mm *MonitorManager) isEndpointMonitored(endpoint services.Endpoint) bool {
	monitors := mm.monitorsForEndpointID(endpoint.Core().ID)
	return len(monitors) > 0
}

func (mm *MonitorManager) deleteMonitorsByConfigHash(hash uint64) {
	for i := range mm.activeMonitors {
		if mm.activeMonitors[i].configHash == hash {
			log.WithFields(log.Fields{
				"config": mm.activeMonitors[i].config,
			}).Info("Shutting down monitor due to config hash change")
			mm.activeMonitors[i].doomed = true
		}
	}
	mm.deleteDoomedMonitors()
}

func (mm *MonitorManager) deleteDoomedMonitors() {
	newActiveMonitors := []*ActiveMonitor{}

	for i := range mm.activeMonitors {
		am := mm.activeMonitors[i]
		if am.doomed {
			log.WithFields(log.Fields{
				"monitorID":     am.id,
				"monitorType":   am.config.MonitorConfigCore().Type,
				"discoveryRule": am.config.MonitorConfigCore().DiscoveryRule,
			}).Info("Shutting down monitor")

			am.Shutdown()
		} else {
			newActiveMonitors = append(newActiveMonitors, am)
		}
	}

	mm.activeMonitors = newActiveMonitors
}

// Shutdown will shutdown all managed monitors and deinitialize the manager.
func (mm *MonitorManager) Shutdown() {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	for i := range mm.activeMonitors {
		mm.activeMonitors[i].doomed = true
	}
	mm.deleteDoomedMonitors()

	mm.activeMonitors = nil
	mm.discoveredEndpoints = nil
}
