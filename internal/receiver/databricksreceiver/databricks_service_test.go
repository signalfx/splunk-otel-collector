// Copyright The OpenTelemetry Authors
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

package databricksreceiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabricksClient(t *testing.T) {
	const ignored = 25
	c := newDatabricksService(&testdataDBClient{}, ignored)
	jobs, err := c.jobs()
	require.NoError(t, err)
	assert.Equal(t, 6, len(jobs))
	active, err := c.activeJobRuns()
	require.NoError(t, err)
	assert.Equal(t, 2, len(active))
	completed, err := c.completedJobRuns(288, -1)
	require.NoError(t, err)
	assert.Equal(t, 98, len(completed))
}

func TestDatabricksClient_CompletedRuns(t *testing.T) {
	const ignored = 25
	c := newDatabricksService(&testdataDBClient{}, ignored)

	// 1642777677522 is from completed-job-runs-0-0.json
	runs, err := c.completedJobRuns(288, 1642777677522)
	require.NoError(t, err)
	assert.Equal(t, 30, len(runs))

	// 1642775877669 is from completed-job-runs-1-1.json
	runs, err = c.completedJobRuns(288, 1642775877669)
	require.NoError(t, err)
	assert.Equal(t, 67, len(runs))
}
