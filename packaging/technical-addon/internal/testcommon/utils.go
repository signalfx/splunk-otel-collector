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

package testcommon

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func AssertFilesMatch(tt *testing.T, expectedPath, actualPath string) {
	require.FileExists(tt, actualPath)
	require.FileExists(tt, expectedPath)
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		tt.Fatalf("Failed to read expected file: %v", err)
	}

	actual, err := os.ReadFile(actualPath)
	if err != nil {
		tt.Fatalf("Failed to read actual file: %v", err)
	}

	if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
		tt.Errorf("File contents mismatch (-expected +actual)\npaths: (%s, %s):\n%s", expectedPath, actualPath, diff)
	}
}
