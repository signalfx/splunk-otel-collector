// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestConfiguredLifecycle(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servicesNS/-/-/data/inputs/http", r.URL.Path)
		assert.Equal(t, "Splunk management-session", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(tokenPageXML(1, 0, 1)))
	}))
	defer api.Close()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.NoError(t, confmap.NewFromStringMap(map[string]any{
		"endpoint": api.URL + "/", "session_key": "management-session",
		"refresh_interval": "1m", "refresh_timeout": "10s", "max_staleness": "5m",
	}).Unmarshal(cfg))
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
	require.NoError(t, cfg.(*Config).Validate())
	for range 2 {
		ext, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
		require.NoError(t, err)
		require.NoError(t, ext.Start(t.Context(), componenttest.NewNopHost()))
		require.NoError(t, ext.Shutdown(t.Context()))
	}
}

// Exercise the same middleware used by upstream receivers, including named IDs
// and propagation of authorization metadata to the request handler.
func TestReceiverAuthentication(t *testing.T) {
	a := newTestAuthenticator(t, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		return []HecTokenConfig{{Token: "valid", DefaultIndex: "main"}, {Token: "disabled", Disabled: true}}, nil
	}), nil)
	id := component.MustNewIDWithName("splunk_auth", "receiver")
	exts := map[component.ID]component.Component{id: a}
	settings := componenttest.NewNopTelemetrySettings()

	httpCfg := confighttp.NewDefaultServerConfig()
	httpCfg.Auth.GetOrInsertDefault().Config.AuthenticatorID = id
	httpServer, err := httpCfg.ToServer(t.Context(), exts, settings, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "main", r.Context().Value(ContextKey).(HecTokenConfig).DefaultIndex)
		w.WriteHeader(http.StatusNoContent)
	}))
	require.NoError(t, err)
	server := httptest.NewServer(httpServer.Handler)
	defer server.Close()

	grpcCfg := configgrpc.NewDefaultServerConfig()
	grpcCfg.Auth.GetOrInsertDefault().AuthenticatorID = id
	grpcServer, err := grpcCfg.ToServer(t.Context(), exts, settings)
	require.NoError(t, err)
	grpc_health_v1.RegisterHealthServer(grpcServer, &authHealthServer{t: t})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		_ = grpcServer.Serve(listener)
	}()
	t.Cleanup(func() { grpcServer.Stop(); <-serveDone })
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := grpc_health_v1.NewHealthClient(conn)
	for _, tc := range []struct {
		name, header string
		valid        bool
	}{
		{name: "valid", header: "Splunk valid", valid: true},
		{name: "missing"},
		{name: "empty", header: ""},
		{name: "empty token", header: "Splunk "},
		{name: "wrong scheme", header: "Bearer valid"},
		{name: "unknown", header: "Splunk unknown"},
		{name: "disabled", header: "Splunk disabled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
			require.NoError(t, err)
			if tc.name != "missing" {
				req.Header.Set("Authorization", tc.header)
			}
			resp, err := server.Client().Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			ctx := t.Context()
			if tc.name != "missing" {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", tc.header)
			}
			_, grpcErr := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			if tc.valid {
				require.Equal(t, http.StatusNoContent, resp.StatusCode)
				require.NoError(t, grpcErr)
			} else {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
				require.Equal(t, codes.Unauthenticated, status.Code(grpcErr))
			}
		})
	}
}

type authHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	t *testing.T
}

func (s *authHealthServer) Check(ctx context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	assert.Equal(s.t, "main", ctx.Value(ContextKey).(HecTokenConfig).DefaultIndex)
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}
