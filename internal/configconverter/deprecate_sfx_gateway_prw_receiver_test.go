// Copyright Splunk, Inc.
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

package configconverter

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestWarnOnSFXGatewayPRWReceiver_Present(t *testing.T) {
	cfgMap := confmap.NewFromStringMap(map[string]any{
		"receivers": map[string]any{
			"signalfxgatewayprometheusremotewrite": map[string]any{},
		},
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	err := WarnOnSFXGatewayPRWReceiver(context.Background(), cfgMap)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "signalfxgatewayprometheusremotewrite")
	assert.Contains(t, buf.String(), "deprecated")
	assert.Contains(t, buf.String(), "No migration path")
}

func TestWarnOnSFXGatewayPRWReceiver_Absent(t *testing.T) {
	for _, cfgMap := range []*confmap.Conf{
		confmap.NewFromStringMap(map[string]any{
			"receivers": map[string]any{
				"prometheusremotewrite": map[string]any{},
			},
		}),
		confmap.NewFromStringMap(map[string]any{}),
	} {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(os.Stderr) })

		err := WarnOnSFXGatewayPRWReceiver(context.Background(), cfgMap)
		require.NoError(t, err)
		assert.Empty(t, buf.String())
	}
}
