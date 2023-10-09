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

// go:build windows

package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDefaultEnvVarsSetsInterfaceFromConfigOptionWithProgramData(t *testing.T) {
	pd := os.Getenv("ProgramData")
	for _, tc := range []struct{ config, expectedIP string }{
		{filepath.Join(pd, "Splunk", "OpenTelemetry Collector", "agent_config.yaml"), "127.0.0.1"},
		{fmt.Sprintf("file:%s", filepath.Join(pd, "Splunk", "OpenTelemetry Collector", "agent_config.yaml")), "127.0.0.1"},
		{"\\some-other-config.yaml", "0.0.0.0"},
		{"file:\\some-other-config.yaml", "0.0.0.0"},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v->%v", tc.config, tc.expectedIP), func(t *testing.T) {
			t.Cleanup(clearEnv(t))
			os.Setenv("SPLUNK_REALM", "noop")
			os.Setenv("SPLUNK_ACCESS_TOKEN", "noop")
			s, err := parseArgs([]string{"--config", tc.config})
			require.NoError(t, err)
			require.NoError(t, setDefaultEnvVars(s))

			val, ok := os.LookupEnv("SPLUNK_LISTEN_INTERFACE")
			assert.True(t, ok)
			assert.Equal(t, tc.expectedIP, val)
		})
	}
}
