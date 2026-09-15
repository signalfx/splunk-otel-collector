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

package auth

import (
	"context"
	_ "embed"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensionauth"

	"github.com/signalfx/splunk-otel-collector/internal/auth/authtest"
)

func TestAuth(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	authtest.SetupAuth(t, listener)
	e, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), "http://"+listener.Addr().String(), "foo")
	require.NoError(t, err)
	require.NotNil(t, e)
	se := e.(extensionauth.Server)
	ctx, err := se.Authenticate(context.Background(), map[string][]string{
		"foo":           {"bar"},
		"Authorization": {"Splunk 00000000-0000-0000-0000-0000000000000"},
	})
	require.NoError(t, err)
	require.NotNil(t, ctx)
	cfg := ctx.Value(ContextKey).(HecTokenConfig)
	require.Equal(t, []string{"main", "foo"}, cfg.AllowedIndexes)
	_, err = se.Authenticate(context.Background(), map[string][]string{
		"foo":           {"bar"},
		"Authorization": {"Splunk 00000000-0000-0000-0000-111111111111"},
	})
	require.Error(t, err)
}
