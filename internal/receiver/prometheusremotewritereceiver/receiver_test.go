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

package prometheusremotewritereceiver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/prometheusremotewritereceiver/internal"
)

func TestHappy(t *testing.T) {
	timeout := time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cfg := createDefaultConfig().(*Config)
	freePort, err := internal.GetFreePort()
	require.NoError(t, err)

	cfg.Endpoint = fmt.Sprintf("localhost:%d", freePort)
	cfg.ListenPath = "/metrics"

	nopHost := componenttest.NewNopHost()
	mockSettings := receivertest.NewNopCreateSettings()
	mockConsumer := consumertest.NewNop()
	receiver, err := newPrometheusRemoteWriteReceiver(mockSettings, cfg, mockConsumer)

	assert.NoError(t, err)
	require.NotNil(t, receiver)
	require.NoError(t, receiver.Start(ctx, nopHost))

	// Ensure we can instantiate
	client, err := internal.NewMockPrwClient(
		cfg.Endpoint,
		"metrics",
	)
	require.NoError(t, err)
	require.NotNil(t, client)
}
