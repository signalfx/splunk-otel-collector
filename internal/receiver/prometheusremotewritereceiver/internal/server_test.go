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

package internal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestSmoke(t *testing.T) {
	mc := make(chan pmetric.Metrics)
	defer close(mc)
	timeout := 5 * time.Second
	addr := "localhost:0"
	reporter := NewMockReporter(0)
	cfg := &ServerConfig{
		Path:               "/metrics",
		Reporter:           reporter,
		Mc:                 mc,
		HTTPServerSettings: confighttp.HTTPServerSettings{},
	}
	cfg.Endpoint = addr
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	receiver, err := NewPrometheusRemoteWriteServer(ctx, cfg)
	assert.NoError(t, err)
	require.NotNil(t, receiver)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		require.NoError(t, receiver.ListenAndServe())
		wg.Done()
	}()

	require.NoError(t, receiver.Shutdown(ctx))
	require.Eventually(t, func() bool { wg.Wait(); return true }, time.Second*10, time.Second)
}
