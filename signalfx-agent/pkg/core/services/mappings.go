// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
