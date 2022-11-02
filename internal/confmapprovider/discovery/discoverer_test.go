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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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

func TestMergeEntries(t *testing.T) {
	a := confmap.NewFromStringMap(map[string]any{})
	b := confmap.NewFromStringMap(map[string]any{})
	a.Merge(b)

	first := map[string]any{
		"one.key": "one.val",
		"two.key": "two.val",
	}
	second := map[string]any{
		"three.key": "three.val",
		"four.key": map[any]any{
			"four.a.key": "four.a.val",
			"four.b.key": "four.b.val",
		},
	}
	err := mergeMaps(first, second)
	assert.NoError(t, err)
	require.Equal(t, map[string]any{
		"one.key":   "one.val",
		"two.key":   "two.val",
		"three.key": "three.val",
		"four.key": map[string]any{
			"four.a.key": "four.a.val",
			"four.b.key": "four.b.val",
		},
	}, first)

	third := map[string]any{
		"three.key": "three.val^",
		"four.key": map[any]any{
			"four.b.key": "four.b.val^",
			"four.c.key": "four.c.val",
		},
	}
	err = mergeMaps(first, third)
	assert.NoError(t, err)
	require.Equal(t, map[string]any{
		"one.key":   "one.val",
		"two.key":   "two.val",
		"three.key": "three.val^",
		"four.key": map[string]any{
			"four.a.key": "four.a.val",
			"four.b.key": "four.b.val^",
			"four.c.key": "four.c.val",
		},
	}, first)

	fourth := map[string]any{
		"four.key": map[any]any{
			"four.c.key": map[any]any{
				"six.key": "six.val",
			},
			"four.d.key": "four.d.val",
		},
		"five.key": "five.val",
	}
	err = mergeMaps(first, fourth)
	assert.NoError(t, err)
	require.Equal(t, map[string]any{
		"one.key":   "one.val",
		"two.key":   "two.val",
		"three.key": "three.val^",
		"four.key": map[string]any{
			"four.a.key": "four.a.val",
			"four.b.key": "four.b.val^",
			"four.c.key": map[string]any{
				"six.key": "six.val",
			},
			"four.d.key": "four.d.val",
		},
		"five.key": "five.val",
	}, first)
}
