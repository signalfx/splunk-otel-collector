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
)

func TestConfig_Validate(t *testing.T) {
	cfg := createDefaultConfig()
	rc := cfg.(*Config)
	assert.Error(t, rc.Validate())
	rc.Endpoint = "foo"
	assert.Error(t, rc.Validate())
	rc.InstanceName = "my-instance"
	assert.Error(t, rc.Validate())
	rc.Token = "my-token"
	assert.NoError(t, rc.Validate())
	rc.MaxResults = 26
	assert.Error(t, rc.Validate())
	rc.MaxResults = -1
	err := rc.Validate()
	assert.EqualError(t, err, "max_results must be between 0 and 25, inclusive")
}
