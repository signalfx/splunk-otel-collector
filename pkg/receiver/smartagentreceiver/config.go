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
	_ "github.com/signalfx/signalfx-agent/pkg/core" // required to invoke monitor registration via init() calls
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

const defaultIntervalSeconds = 10

var (
	_ confmap.Unmarshaler = (*Config)(nil)

	errDimensionClientValue = fmt.Errorf("dimensionClients must be an array of compatible exporter names")
	nonWindowsMonitors      = map[string]bool{
		"collectd/activemq": true, "collectd/apache": true, "collectd/cassandra": true, "collectd/chrony": true,
		"collectd/cpu": true, "collectd/cpufreq": true, "collectd/custom": true, "collectd/df": true, "collectd/disk": true,
		"collectd/genericjmx": true, "collectd/hadoopjmx": true, "collectd/kafka": true, "collectd/kafka_consumer": true,
		"collectd/kafka_producer": true, "collectd/load": true, "collectd/memcached": true, "collectd/memory": true,
		"collectd/mysql": true, "collectd/netinterface": true, "collectd/nginx": true, "collectd/php-fpm": true,
		"collectd/postgresql": true, "collectd/processes": true, "collectd/protocols": true,
		"collectd/signalfx-metadata": true, "collectd/statsd": true, "collectd/uptime": true, "collectd/vmem": true,
	}
)

type Config struct {
	monitorConfig saconfig.MonitorCustomConfig
	// Generally an observer/receivercreator-set value via Endpoint.Target.
	// Will expand to MonitorCustomConfig Host and Port values if unset.
	Endpoint         string   `mapstructure:"endpoint"`
	DimensionClients []string `mapstructure:"dimensionClients"`
	acceptsEndpoints bool
}

func (cfg *Config) validate() error {
	if cfg == nil || cfg.monitorConfig == nil {
		return fmt.Errorf("you must supply a valid Smart Agent Monitor config")
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
	// AllSettings() is the user provided config and intoCfg is the default Config instance we populate to
	// form the final desired version. To do so, we manually obtain all Config items, leaving only Smart Agent
	// monitor config settings to be unmarshalled to their respective custom monitor config types.
	allSettings := componentParser.ToStringMap()
	monitorType, ok := allSettings["type"].(string)
	if !ok || monitorType == "" {
		return fmt.Errorf("you must specify a \"type\" for a smartagent receiver")
	}

	var endpoint any
	if endpoint, ok = allSettings["endpoint"]; ok {
		cfg.Endpoint = fmt.Sprintf("%s", endpoint)
		delete(allSettings, "endpoint")
	}

	var err error
	cfg.DimensionClients, err = getStringSliceFromAllSettings(allSettings, "dimensionClients", errDimensionClientValue)
	if err != nil {
		return err
	}

	// monitors.ConfigTemplates is a map that all monitors use to register their custom configs in the Smart Agent.
	// The values are always pointers to an actual custom config.
	var customMonitorConfig saconfig.MonitorCustomConfig
	if customMonitorConfig, ok = monitors.ConfigTemplates[monitorType]; !ok {
		if unsupported := nonWindowsMonitors[monitorType]; runtime.GOOS == "windows" && unsupported {
			return fmt.Errorf("smart agent monitor type %q is not supported on windows platforms", monitorType)
		}
		return fmt.Errorf("no known monitor type %q", monitorType)
	}
	monitorConfigType := reflect.TypeOf(customMonitorConfig).Elem()
	monitorConfig := reflect.New(monitorConfigType).Interface()

	asBytes, err := yaml.Marshal(allSettings)
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

func getStringSliceFromAllSettings(allSettings map[string]any, key string, errToReturn error) ([]string, error) {
	var items []string
	if value, ok := allSettings[key]; ok {
		items = []string{}
		if valueAsSlice, isSlice := value.([]any); isSlice {
			for _, c := range valueAsSlice {
				if client, isString := c.(string); isString {
					items = append(items, client)
				} else {
					return nil, errToReturn
				}
			}
		} else {
			return nil, errToReturn
		}
		delete(allSettings, key)
	}
	return items, nil
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
