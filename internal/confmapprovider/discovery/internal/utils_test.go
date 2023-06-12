// Copyright Splunk, Inc.
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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestMergeMaps(t *testing.T) {
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
	err := MergeMaps(first, second)
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
	err = MergeMaps(first, third)
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
	err = MergeMaps(first, fourth)
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
