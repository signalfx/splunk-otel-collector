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

func TestUnmarshaller(t *testing.T) {
	u := databricksUnmarshaller{&testdataDBClient{}}
	list, err := u.jobsList(25, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list.Jobs))
	activeRuns, err := u.activeJobRuns(25, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(activeRuns.Runs))
	completedRuns, err := u.completedJobRuns(288, 25, 0)
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", completedRuns.Runs[0].State.ResultState)
	cl, err := u.clusterList()
	require.NoError(t, err)
	assert.Equal(t, 2, len(cl.Clusters))
}
