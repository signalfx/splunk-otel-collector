package services

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// ConfigEndpointMapping creates a mapping between a config key and a value
// that is derived from fields on the endpoint.  It implements the
// CustomConfigurable interface.
type ConfigEndpointMapping struct {
	Endpoint  Endpoint
	ConfigKey string
	Rule      string
}

var _ config.CustomConfigurable = &ConfigEndpointMapping{}

// ExtraConfig evaluates the rule and returns a map that can be merged into the
// final monitor config
func (cem *ConfigEndpointMapping) ExtraConfig() (map[string]interface{}, error) {
	val, err := EvaluateRule(cem.Endpoint, cem.Rule, true, true)
	if err != nil {
		return nil, err
	}

	if s, ok := val.(string); ok {
		val = utils.DecodeValueGenerically(s)
	}

	return map[string]interface{}{
		cem.ConfigKey: val,
	}, nil
}
