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

package discovery

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestDiscovererDurationFromEnv(t *testing.T) {
	t.Cleanup(func() func() {
		initial, ok := os.LookupEnv("SPLUNK_DISCOVERY_DURATION")
		os.Unsetenv("SPLUNK_DISCOVERY_DURATION")
		return func() {
			if ok {
				os.Setenv("SPLUNK_DISCOVERY_DURATION", initial)
			}
		}
	}())
	d, err := newDiscoverer(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Second, d.duration)

	os.Setenv("SPLUNK_DISCOVERY_DURATION", "10h")
	d, err = newDiscoverer(zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Hour, d.duration)

	os.Setenv("SPLUNK_DISCOVERY_DURATION", "invalid")

	zc, observedLogs := observer.New(zap.DebugLevel)
	d, err = newDiscoverer(zap.New(zc))
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 10*time.Second, d.duration)

	require.Eventually(t, func() bool {
		for _, m := range observedLogs.All() {
			if strings.Contains(m.Message, "Invalid SPLUNK_DISCOVERY_DURATION. Using default of 10s") {
				return m.ContextMap()["duration"] == "invalid"
			}
		}
		return false
	}, 2*time.Second, time.Millisecond)
}

func TestDetermineCurrentStatus(t *testing.T) {
	for _, test := range []struct {
		current, observed, expected discovery.StatusType
	}{
		{"failed", "failed", "failed"},
		{"failed", "partial", "partial"},
		{"failed", "successful", "successful"},
		{"partial", "failed", "partial"},
		{"partial", "partial", "partial"},
		{"partial", "successful", "successful"},
		{"successful", "failed", "successful"},
		{"successful", "partial", "successful"},
		{"successful", "successful", "successful"},
	} {
		t.Run(fmt.Sprintf("%s:%s->%s", test.current, test.observed, test.expected), func(t *testing.T) {
			require.Equal(t, test.expected, determineCurrentStatus(test.current, test.observed))
		})
	}
}
