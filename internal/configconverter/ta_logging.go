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
	"log"

	"go.opentelemetry.io/collector/confmap"
)

// RedirectTALogsToStderr keeps collector logs off the modular input's stdout,
// which Splunk reserves for event data. Splunk records stderr in splunkd.log.
func RedirectTALogsToStderr(_ context.Context, cfg *confmap.Conf) error {
	if cfg == nil {
		return nil
	}

	out := cfg.ToStringMap()
	service, ok := out["service"].(map[string]any)
	if !ok {
		return nil
	}
	telemetry, ok := service["telemetry"].(map[string]any)
	if !ok {
		return nil
	}
	logs, ok := telemetry["logs"].(map[string]any)
	if !ok {
		return nil
	}

	changed := false
	for _, key := range []string{"output_paths", "error_output_paths"} {
		paths, ok := logs[key].([]any)
		if !ok {
			continue
		}
		seen := make(map[string]struct{}, len(paths))
		redirected := make([]any, 0, len(paths))
		for _, path := range paths {
			name, ok := path.(string)
			if !ok {
				redirected = append(redirected, path)
				continue
			}
			if name == "stdout" || name == "/dev/stdout" {
				name = "stderr"
				changed = true
			}
			if _, exists := seen[name]; exists && name == "stderr" {
				continue
			}
			seen[name] = struct{}{}
			redirected = append(redirected, name)
		}
		logs[key] = redirected
	}

	if changed {
		log.Print("WARNING: Collector log output to stdout was redirected to stderr because Splunk modular inputs reserve stdout for event data")
		*cfg = *confmap.NewFromStringMap(out)
	}
	return nil
}
