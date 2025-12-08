// Copyright Splunk, Inc.
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

package configconverter

import (
	"context"
	"expvar"
	"sync"

	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
)

var (
	_ confmap.Converter = (*ExpvarConverter)(nil)
	_ configsource.Hook = (*ExpvarConverter)(nil)
)

var (
	expvarConverterInstance *ExpvarConverter
	expvarConverterOnce     sync.Once
)

type ExpvarConverter struct {
	initial        map[string]any
	effective      map[string]any
	initialMutex   sync.RWMutex
	effectiveMutex sync.RWMutex
}

// GetExpvarConverter returns the singleton instance of expvarConverter that publishes
// two entries to the `expvar` JSON map:
//   - "splunk.config.effective": a map with the effective configuration being used by the collector.
//   - "splunk.config.initial": a map with the initial configuration being used by the collector.
func GetExpvarConverter() *ExpvarConverter {
	expvarConverterOnce.Do(func() {
		expvarConverterInstance = &ExpvarConverter{
			initial:        make(map[string]any),
			effective:      make(map[string]any),
			initialMutex:   sync.RWMutex{},
			effectiveMutex: sync.RWMutex{},
		}

		expvar.Publish("splunk.config.effective", expvar.Func(func() any {
			instance := expvarConverterInstance
			instance.effectiveMutex.RLock()
			defer instance.effectiveMutex.RUnlock()
			configYAML, _ := yaml.Marshal(simpleRedact(instance.effective))
			return string(configYAML)
		}))
		expvar.Publish("splunk.config.initial", expvar.Func(func() any {
			instance := expvarConverterInstance
			instance.initialMutex.RLock()
			defer instance.initialMutex.RUnlock()
			configYAML, _ := yaml.Marshal(simpleRedact(instance.initial))
			return string(configYAML)
		}))
	})

	return expvarConverterInstance
}

func (e *ExpvarConverter) OnNew() {}

func (e *ExpvarConverter) OnRetrieve(scheme string, retrieved map[string]any) {
	e.initialMutex.Lock()
	defer e.initialMutex.Unlock()
	e.initial[scheme] = retrieved
}

func (e *ExpvarConverter) OnShutdown() {}

func (e *ExpvarConverter) Convert(_ context.Context, conf *confmap.Conf) error {
	e.effectiveMutex.Lock()
	defer e.effectiveMutex.Unlock()
	e.effective = conf.ToStringMap()
	return nil
}
