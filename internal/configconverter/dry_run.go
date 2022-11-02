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

	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

var _ confmap.Converter = (*DryRun)(nil)

type DryRun struct {
	enabled bool
}

func NewDryRun(enabled bool) DryRun {
	return DryRun{enabled: enabled}
}

// Convert is intended to be called as the final service confmap.Converter so
// that it has access to the final config before exiting, if enabled.
func (dr DryRun) Convert(_ context.Context, conf *confmap.Conf) error {
	if dr.enabled {
		out, err := yaml.Marshal(conf.ToStringMap())
		if err != nil {
			panic(fmt.Errorf("failed marshaling --dry-run config: %w", err))
		}
		fmt.Fprintf(os.Stdout, "%s", out)
		os.Stdout.Sync()
		os.Exit(0)
	}
	return nil
}
