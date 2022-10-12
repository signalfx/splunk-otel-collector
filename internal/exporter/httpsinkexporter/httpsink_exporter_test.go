// Copyright Splunk, Inc.
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

package httpsinkexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
)

func Test_httpSinkExporter_Start(t *testing.T) {
	sink := &httpSinkExporter{
		endpoint: "localhost:0",
		ch:       nil,
		clients:  nil,
	}
	err := sink.Start(context.Background(), componenttest.NewNopHost())
	assert.NoError(t, err)
	err = sink.Shutdown(context.Background())
	assert.NoError(t, err)
}
