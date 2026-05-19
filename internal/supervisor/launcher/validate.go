// Copyright Splunk Inc.
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

package launcher

import (
	"bytes"
	"fmt"
	"os/exec"
)

// validateCollectorConfig runs the collector's built-in validation command
// against the local config files before the supervisor starts.
func validateCollectorConfig(collector string, args, configPaths, env []string) error {
	validationArgs := append([]string{}, args...)
	for _, configPath := range configPaths {
		validationArgs = append(validationArgs, "--config", configPath)
	}
	validationArgs = append(validationArgs, "validate")

	cmd := exec.Command(collector, validationArgs...) //nolint:gosec
	cmd.Env = env
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("collector config validation failed: %w\n%s", err, output.String())
	}
	return nil
}
