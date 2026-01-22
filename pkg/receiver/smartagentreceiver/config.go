// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"fmt"
	"net"
	"reflect"
	"runtime"
	"strconv"

	"github.com/signalfx/defaults"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/xconfmap"
	"gopkg.in/yaml.v2"

	_ "github.com/signalfx/signalfx-agent/pkg/core" // required to invoke monitor registration via init() calls
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

const defaultIntervalSeconds = 10

var (
	_ confmap.Unmarshaler = (*Config)(nil)
	_ xconfmap.Validator  = (*Config)(nil)

	nonWindowsMonitors = map[string]bool{
		"collectd/apache": true,
		"collectd/cpu":    true, "collectd/cpufreq": true, "collectd/custom": true,
		"collectd/memcached": true, "collectd/memory": true,
		"collectd/nginx": true, "collectd/php-fpm": true,
		"collectd/processes": true, "collectd/protocols": true,
		"collectd/signalfx-metadata": true, "collectd/uptime": true,
	}
)

type Config struct {
	monitorConfig saconfig.MonitorCustomConfig
	MonitorType   string `mapstructure:"type"` // Smart Agent monitor type, e.g. collectd/cpu
	// Generally an observer/receivercreator-set value via Endpoint.Target.
	// Will expand to MonitorCustomConfig Host and Port values if unset.
	Endpoint         string   `mapstructure:"endpoint"`
	DimensionClients []string `mapstructure:"dimensionClients"`
	acceptsEndpoints bool
}

func (cfg *Config) Validate() error {
	if cfg.MonitorType == "" {
		return fmt.Errorf(`you must specify a "type" for a smartagent receiver`)
	}

	monitorConfigCore := cfg.monitorConfig.MonitorConfigCore()
	if monitorConfigCore.IntervalSeconds == 0 {
		monitorConfigCore.IntervalSeconds = defaultIntervalSeconds
	} else if monitorConfigCore.IntervalSeconds < 0 {
		return fmt.Errorf("intervalSeconds must be greater than 0s (%d provided)", monitorConfigCore.IntervalSeconds)
	}

	if err := validation.ValidateStruct(cfg.monitorConfig); err != nil {
		return err
	}
	return validation.ValidateCustomConfig(cfg.monitorConfig)
}

// Unmarshal dynamically creates the desired Smart Agent monitor config
// from the provided receiver config content.
func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	// Load the non-dynamic config normally ignoring unused fields.
	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	// No need to proceed if monitor type isn't specified, this will fail the validation.
	if cfg.MonitorType == "" {
		return nil
	}

	monitorConfmap := componentParser.ToStringMap()
	delete(monitorConfmap, "endpoint")
	delete(monitorConfmap, "dimensionClients")

	// monitors.ConfigTemplates is a map that all monitors use to register their custom configs in the Smart Agent.
	// The values are always pointers to an actual custom config.
	customMonitorConfig, ok := monitors.ConfigTemplates[cfg.MonitorType]
	if !ok {
		if unsupported := nonWindowsMonitors[cfg.MonitorType]; runtime.GOOS == "windows" && unsupported {
			return fmt.Errorf("smart agent monitor type %q is not supported on windows platforms", cfg.MonitorType)
		}
		return fmt.Errorf("no known monitor type %q", cfg.MonitorType)
	}
	monitorConfigType := reflect.TypeOf(customMonitorConfig).Elem()
	monitorConfig := reflect.New(monitorConfigType).Interface()

	asBytes, err := yaml.Marshal(monitorConfmap)
	if err != nil {
		return fmt.Errorf("failed constructing raw Smart Agent Monitor config block: %w", err)
	}

	err = yaml.UnmarshalStrict(asBytes, monitorConfig)
	if err != nil {
		return fmt.Errorf("failed creating Smart Agent Monitor custom config: %w", err)
	}

	err = defaults.Set(monitorConfig)
	if err != nil {
		return fmt.Errorf("failed setting Smart Agent Monitor config defaults: %w", err)
	}

	cfg.acceptsEndpoints, err = monitorAcceptsEndpoints(monitorConfig)
	if err != nil {
		return err
	}

	if cfg.acceptsEndpoints {
		if err = setHostAndPortViaEndpoint(cfg.Endpoint, monitorConfig); err != nil {
			return err
		}
	}

	cfg.monitorConfig = monitorConfig.(saconfig.MonitorCustomConfig)
	return nil
}

// If using the receivercreator, observer-provided endpoints should be used to set
// the Host and Port fields of monitor config structs.  This can only be done by reflection without
// making type assertions over all possible monitor types.
func setHostAndPortViaEndpoint(endpoint string, monitorConfig any) error {
	if endpoint == "" {
		return nil
	}

	var port uint16
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		// best effort
		host = endpoint
	}

	if portStr != "" {
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return fmt.Errorf("cannot determine port via Endpoint: %w", err)
		}
		port = uint16(port64)
	}

	if host != "" {
		// determine if a Host field exists with expected type and ignore if not.
		if field, err := getSettableStructFieldValue(monitorConfig, "Host", reflect.TypeOf(host)); err == nil && field != nil {
			if _, err = setStructFieldIfZeroValue(monitorConfig, "Host", host); err != nil {
				return fmt.Errorf("unable to set monitor Host field using Endpoint-derived value of %s: %w", host, err)
			}
		}
	}

	if port != 0 {
		// Determine if a Port field exists with expected type and ignore if not.
		// Elasticsearch monitors have port fields that are strings so attempt this value if uint16 not found.
		for _, p := range []any{port, portStr} {
			if field, err := getSettableStructFieldValue(monitorConfig, "Port", reflect.TypeOf(p)); err == nil && field != nil {
				if _, err = setStructFieldIfZeroValue(monitorConfig, "Port", p); err != nil {
					return fmt.Errorf("unable to set monitor Port field using Endpoint-derived value of %v: %w", p, err)
				}
				break
			}
		}
	}

	return nil
}

// Monitors can only set the "Host" and "Port" fields if they accept endpoints,
// which is defined as a struct tag for each monitor config.
func monitorAcceptsEndpoints(monitorConfig any) (bool, error) {
	field, ok := reflect.TypeOf(monitorConfig).Elem().FieldByName("MonitorConfig")
	if !ok {
		return false, fmt.Errorf("could not reflect monitor config, top level MonitorConfig does not exist")
	}
	return field.Tag.Get("acceptsEndpoints") == strconv.FormatBool(true), nil
}
