// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

type otelCfg struct {
	ConfigSources map[string]interface{} `yaml:"config_sources"`
	Receivers     map[string]interface{}
	Exporters     map[string]interface{}
	Service       map[string]interface{}
}

func saInfoToOtelConfig(cfg saCfgInfo) otelCfg {
	receivers := map[string]interface{}{}
	for _, monitor := range cfg.monitors {
		receiver := saMonitorToOtelReceiver(monitor.(map[interface{}]interface{}))
		for k, v := range receiver {
			receivers[k.(string)] = v
		}
	}
	return otelCfg{
		ConfigSources: map[string]interface{}{
			"include": map[string]interface{}{},
		},
		Receivers: receivers,
		Exporters: sfxExporter(cfg.accessToken, cfg.realm),
		Service: map[string]interface{}{
			"pipelines": map[string]interface{}{
				"metrics": rpe{
					Receivers: receiverList(receivers),
					Exporters: []string{"signalfx"},
				},
			},
		},
	}
}

// rpe == Receivers Processors Exporters. Using this instead of a map for
// deterministic ordering.
type rpe struct {
	Receivers []string
	// Processors field TBD
	Exporters []string
}

func receiverList(receivers map[string]interface{}) []string {
	var keys []string
	for k := range receivers {
		keys = append(keys, k)
	}
	return keys
}

func saMonitorToOtelReceiver(monitor map[interface{}]interface{}) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"smartagent/" + monitor["type"].(string): monitor,
	}
}

func sfxExporter(accessToken, realm string) map[string]interface{} {
	return map[string]interface{}{
		"signalfx": map[interface{}]interface{}{
			"access_token": accessToken,
			"realm":        realm,
		},
	}
}
