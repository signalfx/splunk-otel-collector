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

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestRenameK8sTaggerTestRenameK8sTagger(t *testing.T) {
	actual, err := configtest.LoadConfigMap("testdata/k8s-tagger.yaml")
	require.NoError(t, err)
	require.NotNil(t, actual)

	expected, err := configtest.LoadConfigMap("testdata/k8sattributes.yaml")
	require.NoError(t, err)

	err = RenameK8sTagger{}.Convert(context.Background(), actual)
	require.NoError(t, err)

	require.Equal(t, expected.ToStringMap(), actual.ToStringMap())
}
