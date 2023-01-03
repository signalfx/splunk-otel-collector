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
	"fmt"
	"os"
	"sync"

	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

var _ confmap.Converter = (*DryRun)(nil)
var _ configprovider.Hook = (*DryRun)(nil)

type DryRun struct {
	*sync.Mutex
	configs []map[string]any
	enabled bool
}

func NewDryRun(enabled bool) *DryRun {
	return &DryRun{
		Mutex:   &sync.Mutex{},
		enabled: enabled,
		configs: []map[string]any{},
	}
}

func (dr *DryRun) OnNew() {}

func (dr *DryRun) OnRetrieve(_ string, retrieved map[string]any) {
	if dr == nil || !dr.enabled {
		return
	}
	dr.Lock()
	defer dr.Unlock()
	dr.configs = append(dr.configs, retrieved)
}

func (dr *DryRun) OnShutdown() {}

// Convert disregards the provided *confmap.Conf so that it will use
// unexpanded values (env vars, config source directives) as
// accrued by OnRetrieve() calls.
func (dr *DryRun) Convert(context.Context, *confmap.Conf) error {
	if dr == nil || !dr.enabled {
		return nil
	}
	cm := confmap.New()
	dr.Lock()
	for _, cfg := range dr.configs {
		if err := cm.Merge(confmap.NewFromStringMap(cfg)); err != nil {
			dr.Unlock()
			return err
		}
	}
	dr.Unlock() // not deferred because we are exiting
	out, err := yaml.Marshal(cm.ToStringMap())
	if err != nil {
		panic(fmt.Errorf("failed marshaling --dry-run config: %w", err))
	}
	fmt.Fprintf(os.Stdout, "%s", out)
	os.Stdout.Sync()
	os.Exit(0)
	return nil
}
