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

package volumequotaprocessor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	for _, test := range []struct {
		cfg       Config
		errWanted error
		name      string
	}{
		{
			name:      "no limits",
			cfg:       Config{},
			errWanted: errors.New("no limits set"),
		},
		{
			name:      "negative span limit",
			cfg:       Config{GlobalLimits: GlobalLimits{Spans: -1}},
			errWanted: errors.New("spans global limit must be zero or positive"),
		},
		{
			name:      "negative trace limit",
			cfg:       Config{GlobalLimits: GlobalLimits{Traces: -1}},
			errWanted: errors.New("traces global limit must be zero or positive"),
		},
		{
			name:      "negative service span limit",
			cfg:       Config{Limits: Limits{Spans: map[string]int64{"foo": -1}}},
			errWanted: errors.New(`span limit for service "foo" must be positive`),
		},
		{
			name:      "negative service trace limit",
			cfg:       Config{Limits: Limits{Traces: map[string]int64{"foo": -1}}},
			errWanted: errors.New(`trace limit for service "foo" must be positive`),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.cfg.Validate()
			if test.errWanted != nil {
				require.Equal(t, test.errWanted, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
