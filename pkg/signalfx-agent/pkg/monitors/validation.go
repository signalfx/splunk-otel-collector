package monitors

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
)

// Used to validate configuration that is common to all monitors up front
func validateConfig(monConfig config.MonitorCustomConfig) error {
	conf := monConfig.MonitorConfigCore()

	if _, ok := MonitorFactories[conf.Type]; !ok {
		return errors.New("monitor type not recognized")
	}

	if conf.IntervalSeconds <= 0 {
		return fmt.Errorf("invalid intervalSeconds provided: %d", conf.IntervalSeconds)
	}

	takesEndpoints := configAcceptsEndpoints(monConfig)
	if !takesEndpoints && conf.DiscoveryRule != "" {
		return fmt.Errorf("monitor %s does not support discovery but has a discovery rule", conf.Type)
	}

	// Validate discovery rules
	if conf.DiscoveryRule != "" {
		err := services.ValidateDiscoveryRule(conf.DiscoveryRule)
		if err != nil {
			return errors.New("discovery rule is invalid: " + err.Error())
		}
	}

	if len(conf.ConfigEndpointMappings) > 0 && len(conf.DiscoveryRule) == 0 {
		return errors.New("configEndpointMappings is not useful without a discovery rule")
	}

	if err := validation.ValidateStruct(monConfig); err != nil {
		return err
	}

	return validation.ValidateCustomConfig(monConfig)
}

func configAcceptsEndpoints(monConfig config.MonitorCustomConfig) bool {
	confVal := reflect.Indirect(reflect.ValueOf(monConfig))
	coreConfField, ok := confVal.Type().FieldByName("MonitorConfig")
	if !ok {
		return false
	}
	return coreConfField.Tag.Get("acceptsEndpoints") == "true"
}
