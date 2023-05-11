package monitors

import (
	"fmt"
	"reflect"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// MonitorFactory is a niladic function that creates an unconfigured instance
// of a monitor.
type MonitorFactory func() interface{}

// MonitorFactories holds all of the registered monitor factories
var MonitorFactories = map[string]MonitorFactory{}

// ConfigTemplates are blank (zero-value) instances of the configuration struct
// for a particular monitor type.
var ConfigTemplates = map[string]config.MonitorCustomConfig{}

// MonitorMetadatas contains a mapping of monitor type to its metadata.
var MonitorMetadatas = map[string]*Metadata{}

// InjectableMonitor should be implemented by a dynamic monitor that needs to
// know when services are added and removed.
type InjectableMonitor interface {
	AddService(services.Endpoint)
	RemoveService(services.Endpoint)
}

// MetricInfo contains metadata about a metric.
type MetricInfo struct {
	Type       datapoint.MetricType
	Group      string
	Dimensions map[string]string
}

// Metadata describes information about a monitor.
type Metadata struct {
	MonitorType     string
	SendAll         bool
	SendUnknown     bool
	DefaultMetrics  map[string]bool
	Metrics         map[string]MetricInfo
	Groups          map[string]bool
	GroupMetricsMap map[string][]string
}

// HasMetric returns whether the metric exists at all (custom or included).
func (metadata *Metadata) HasMetric(metric string) bool {
	_, ok := metadata.Metrics[metric]
	return ok
}

// HasDefaultMetric returns whether the metric is an included metric.
func (metadata *Metadata) HasDefaultMetric(metric string) bool {
	return metadata.DefaultMetrics[metric]
}

// HasGroup returns whether the group exists or not.
func (metadata *Metadata) HasGroup(group string) bool {
	return metadata.Groups[group]
}

// NonDefaultMetrics returns list of metrics that are non-included.
// Note that it is not that efficient so cache calls if necessary or change
// implementation.
func (metadata *Metadata) NonDefaultMetrics() []string {
	var metrics []string
	for metric := range metadata.Metrics {
		if !metadata.HasDefaultMetric(metric) {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

// Register a new monitor type with the agent.  This is intended to be called
// from the init function of the module of a specific monitor
// implementation. configTemplate should be a zero-valued struct that is of the
// same type as the parameter to the Configure method for this monitor type.
func Register(metadata *Metadata, factory MonitorFactory, configTemplate config.MonitorCustomConfig) {
	if _, ok := MonitorFactories[metadata.MonitorType]; ok {
		panic("Monitor type '" + metadata.MonitorType + "' already registered")
	}
	MonitorFactories[metadata.MonitorType] = factory
	ConfigTemplates[metadata.MonitorType] = configTemplate
	MonitorMetadatas[metadata.MonitorType] = metadata
}

// DeregisterAll unregisters all monitor types.  Primarily intended for testing
// purposes.
func DeregisterAll() {
	for k := range MonitorFactories {
		delete(MonitorFactories, k)
	}

	for k := range ConfigTemplates {
		delete(ConfigTemplates, k)
	}
}

func newUninitializedMonitor(_type string) interface{} {
	if factory, ok := MonitorFactories[_type]; ok {
		return factory()
	}

	log.WithFields(log.Fields{
		"monitorType": _type,
	}).Error("Monitor type not supported")
	return nil
}

// Creates a new, unconfigured instance of a monitor of _type.  Returns nil if
// the monitor type is not registered.
func newMonitor(_type string) interface{} {
	mon := newUninitializedMonitor(_type)
	if initMon, ok := mon.(Initializable); ok {
		if err := initMon.Init(); err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"monitorType": _type,
			}).Error("Could not initialize monitor")
			return nil
		}
	}
	return mon
}

// Initializable represents a monitor that has a distinct InitMonitor method.
// This should be called once after the monitor is created and before any of
// its other methods are called.  It is useful for things that are not
// appropriate to do in the monitor factory function.
type Initializable interface {
	Init() error
}

// Shutdownable should be implemented by all monitors that need to clean up
// resources before being destroyed.
type Shutdownable interface {
	Shutdown()
}

// Takes a generic MonitorConfig and pulls out monitor-specific config to
// populate a clone of the config template that was registered for the monitor
// type specified in conf.  This will also validate the config and return nil
// if validation fails.
func getCustomConfigForMonitor(conf *config.MonitorConfig) (config.MonitorCustomConfig, error) {
	confTemplate, ok := ConfigTemplates[conf.Type]
	if !ok {
		return nil, fmt.Errorf("unknown monitor type %s", conf.Type)
	}
	monConfig := utils.CloneInterface(confTemplate).(config.MonitorCustomConfig)

	if err := config.FillInConfigTemplate("MonitorConfig", monConfig, conf); err != nil {
		return nil, err
	}

	return monConfig, nil
}

func anyMarkedSolo(confs []config.MonitorConfig) bool {
	for i := range confs {
		if confs[i].Solo {
			return true
		}
	}
	return false
}

func configOnlyAllowsSingleInstance(monConfig config.MonitorCustomConfig) bool {
	confVal := reflect.Indirect(reflect.ValueOf(monConfig))
	coreConfField, ok := confVal.Type().FieldByName("MonitorConfig")
	if !ok {
		return false
	}
	return coreConfField.Tag.Get("singleInstance") == "true"
}
