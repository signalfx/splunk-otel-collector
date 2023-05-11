package monitors

import (
	"fmt"
	"reflect"

	"github.com/signalfx/defaults"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// ActiveMonitor is a wrapper for an actual monitor instance that keeps some
// metadata about the monitor, such as the set of service endpoints attached to
// the monitor, as well as a copy of its configuration.  It exposes a lot of
// methods to help manage the monitor as well.
type ActiveMonitor struct {
	instance   interface{}
	id         types.MonitorID
	configHash uint64
	agentMeta  *meta.AgentMeta
	output     types.FilteringOutput
	config     config.MonitorCustomConfig
	endpoint   services.Endpoint
	// Is the monitor marked for deletion?
	doomed bool
}

func renderConfig(monConfig config.MonitorCustomConfig, endpoint services.Endpoint) (config.MonitorCustomConfig, error) {
	monConfig = utils.CloneInterface(monConfig).(config.MonitorCustomConfig)
	if err := defaults.Set(monConfig); err != nil {
		return nil, err
	}

	if endpoint != nil {
		err := config.DecodeExtraConfig(endpoint, monConfig, false)
		if err != nil {
			return nil, fmt.Errorf("could not inject endpoint config into monitor config: %w", err)
		}

		for configKey, rule := range monConfig.MonitorConfigCore().ConfigEndpointMappings {
			cem := &services.ConfigEndpointMapping{
				Endpoint:  endpoint,
				ConfigKey: configKey,
				Rule:      rule,
			}
			if err := config.DecodeExtraConfig(cem, monConfig, false); err != nil {
				return nil, fmt.Errorf("could not process config mapping: %s => %s -- %s", configKey, rule, err.Error())
			}
		}
	}

	// Wipe out the other config that has already been decoded since it is not
	// redundant.
	monConfig.MonitorConfigCore().OtherConfig = nil
	return monConfig, nil
}

// Does some reflection magic to pass the right type to the Configure method of
// each monitor
func (am *ActiveMonitor) configureMonitor(monConfig config.MonitorCustomConfig) error {
	monConfig.MonitorConfigCore().MonitorID = am.id
	for k, v := range monConfig.MonitorConfigCore().ExtraDimensions {
		am.output.AddExtraDimension(k, v)
	}

	for k, v := range monConfig.MonitorConfigCore().ExtraDimensionsFromEndpoint {
		val, err := services.EvaluateRule(am.endpoint, v, true, true)
		if err != nil {
			return err
		}
		am.output.AddExtraDimension(k, fmt.Sprintf("%v", val))
	}

	for k, v := range monConfig.MonitorConfigCore().ExtraSpanTags {
		am.output.AddExtraSpanTag(k, v)
	}

	for k, v := range monConfig.MonitorConfigCore().ExtraSpanTagsFromEndpoint {
		val, err := services.EvaluateRule(am.endpoint, v, true, true)
		if err != nil {
			return err
		}
		am.output.AddExtraSpanTag(k, fmt.Sprintf("%v", val))
	}

	for k, v := range monConfig.MonitorConfigCore().DefaultSpanTags {
		am.output.AddDefaultSpanTag(k, v)
	}

	for k, v := range monConfig.MonitorConfigCore().DefaultSpanTagsFromEndpoint {
		val, err := services.EvaluateRule(am.endpoint, v, true, true)
		if err != nil {
			return err
		}
		am.output.AddDefaultSpanTag(k, fmt.Sprintf("%v", val))
	}

	if err := validateConfig(monConfig); err != nil {
		return err
	}

	am.config = monConfig
	am.injectAgentMetaIfNeeded()
	am.injectOutputIfNeeded()

	return config.CallConfigure(am.instance, monConfig)
}

func (am *ActiveMonitor) endpointID() services.ID {
	if am.endpoint == nil {
		return ""
	}
	return am.endpoint.Core().ID
}

func (am *ActiveMonitor) injectOutputIfNeeded() bool {
	outputValue := utils.FindFieldWithEmbeddedStructs(am.instance, "Output",
		reflect.TypeOf((*types.Output)(nil)).Elem())

	if !outputValue.IsValid() {
		// Try and find FilteringOutput type
		outputValue = utils.FindFieldWithEmbeddedStructs(am.instance, "Output",
			reflect.TypeOf((*types.FilteringOutput)(nil)).Elem())
		if !outputValue.IsValid() {
			return false
		}
	}

	outputValue.Set(reflect.ValueOf(am.output))

	return true
}

// Sets the `AgentMeta` field on a monitor if it is present to the agent
// metadata service. Returns whether the field was actually set.
// N.B. that the values in AgentMeta are subject to change at any time.  There
// is no notification mechanism for changes, so a monitor should pull the value
// from the struct each time it needs it and not cache it.
func (am *ActiveMonitor) injectAgentMetaIfNeeded() bool {
	agentMetaValue := utils.FindFieldWithEmbeddedStructs(am.instance, "AgentMeta",
		reflect.TypeOf(&meta.AgentMeta{}))

	if !agentMetaValue.IsValid() {
		return false
	}

	agentMetaValue.Set(reflect.ValueOf(am.agentMeta))

	return true
}

// Shutdown calls Shutdown on the monitor instance if it is provided.
func (am *ActiveMonitor) Shutdown() {
	if sh, ok := am.instance.(Shutdownable); ok {
		sh.Shutdown()
	}
}
