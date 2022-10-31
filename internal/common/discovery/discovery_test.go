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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusTypes(t *testing.T) {
	var asStrings []string
	for _, typ := range StatusTypes {
		asStrings = append(asStrings, string(typ))
	}
	require.Equal(t, asStrings, []string{"successful", "partial", "failed"})
}

func TestIsValidStatusType(t *testing.T) {
	for _, typ := range StatusTypes {
		ok, err := IsValidStatus(typ)
		require.True(t, ok)
		require.Nil(t, err)
	}

	ok, err := IsValidStatus("not a thing")
	require.False(t, ok)
	require.EqualError(t, err, "invalid status \"not a thing\". must be one of [successful partial failed]")
}

func TestNoTypeIsEmpty(t *testing.T) {
	require.Equal(t, "", NoType.String())
}
