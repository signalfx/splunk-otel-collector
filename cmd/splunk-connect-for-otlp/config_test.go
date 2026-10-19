// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     InputConfig
		wantErr bool
	}{
		{
			name:    "SSL disabled no certs",
			cfg:     InputConfig{EnableSSL: false},
			wantErr: false,
		},
		{
			name:    "SSL enabled with both cert and key",
			cfg:     InputConfig{EnableSSL: true, ServerCert: "/certs/server.crt", ServerKey: "/certs/server.key"},
			wantErr: false,
		},
		{
			name:    "SSL enabled missing cert",
			cfg:     InputConfig{EnableSSL: true, ServerKey: "/certs/server.key"},
			wantErr: true,
		},
		{
			name:    "SSL enabled missing key",
			cfg:     InputConfig{EnableSSL: true, ServerCert: "/certs/server.crt"},
			wantErr: true,
		},
		{
			name:    "SSL enabled missing both",
			cfg:     InputConfig{EnableSSL: true},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.AuthConfig = splunkauthextension.NewDefaultConfig()
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseInput(t *testing.T) {
	input := `
<?xml version="1.0" encoding="UTF-8"?>
<input>
  <server_host>773c28971b2a</server_host>
  <server_uri>https://127.0.0.1:8089</server_uri>
  <session_key>OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N</session_key>
  <checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs/splunk-connect-for-otlp</checkpoint_dir>
  <configuration>
    <stanza name="splunk-connect-for-otlp://specialmind" app="search">
      <param name="grpc_port">4317</param>
      <param name="host">$decideOnStartup</param>
      <param name="http_port">4318</param>
      <param name="listen_address">0.0.0.0</param>
      <param name="sourcetype">_splunk-connect-for-otlp</param>
      <param name="enableSSL">1</param>
      <param name="serverCert">/var/certs/server.cert</param>
      <param name="serverKey">/var/certs/server.key</param>
      <param name="realm">us1</param>
      <param name="access_token">test-access-token</param>
      <param name="refresh_interval">2m</param>
      <param name="refresh_timeout">15s</param>
      <param name="max_staleness">8m</param>
      <param name="start_by_shell">false</param>
    </stanza>
  </configuration>
</input>`

	var config XMLInput
	err := xml.Unmarshal([]byte(input), &config)
	require.NoError(t, err)

	require.Equal(t, "splunk-connect-for-otlp://specialmind", config.Configuration.Stanza.Name)

	xmlCfg, err := config.Extract()
	require.NoError(t, err)

	require.Equal(t, 4317, xmlCfg.GrpcPort)
	require.Equal(t, "0.0.0.0", xmlCfg.ListenAddress)
	require.Equal(t, 4318, xmlCfg.HTTPPort)
	require.Empty(t, xmlCfg.Source)
	require.Equal(t, "_splunk-connect-for-otlp", xmlCfg.Sourcetype)
	require.Equal(t, "773c28971b2a", xmlCfg.ServerHost)
	require.Equal(t, "https://127.0.0.1:8089", xmlCfg.ServerURI)
	require.Equal(t, "OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N", xmlCfg.SessionKey)
	require.True(t, xmlCfg.EnableSSL)
	require.Equal(t, "/var/certs/server.cert", xmlCfg.ServerCert)
	require.Equal(t, "/var/certs/server.key", xmlCfg.ServerKey)
	require.Equal(t, "us1", xmlCfg.Realm)
	require.Equal(t, "test-access-token", xmlCfg.AccessToken)
	require.Equal(t, 2*time.Minute, xmlCfg.AuthConfig.RefreshInterval)
	require.Equal(t, 15*time.Second, xmlCfg.AuthConfig.RefreshTimeout)
	require.Equal(t, 8*time.Minute, xmlCfg.AuthConfig.MaxStaleness)
	require.NoError(t, xmlCfg.Validate())
}

func TestAuthInputSettings(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wantErr string
		params  []XMLParam
	}{
		{name: "defaults"},
		{name: "partial override", params: []XMLParam{{Name: "refresh_interval", Value: " 2m "}}},
		{name: "missing units", params: []XMLParam{{Name: "refresh_interval", Value: "180"}}, wantErr: "invalid refresh_interval"},
		{name: "malformed timeout", params: []XMLParam{{Name: "refresh_timeout", Value: "soon"}}, wantErr: "invalid refresh_timeout"},
		{name: "empty staleness", params: []XMLParam{{Name: "max_staleness", Value: ""}}, wantErr: "invalid max_staleness"},
		{name: "duration overflow", params: []XMLParam{{Name: "max_staleness", Value: "999999999999h"}}, wantErr: "invalid max_staleness"},
		{name: "zero interval", params: []XMLParam{{Name: "refresh_interval", Value: "0s"}}, wantErr: "refresh_interval must be positive"},
		{name: "negative timeout", params: []XMLParam{{Name: "refresh_timeout", Value: "-1s"}}, wantErr: "refresh_timeout must be positive"},
		{name: "zero staleness", params: []XMLParam{{Name: "max_staleness", Value: "0s"}}, wantErr: "max_staleness must be positive"},
		{name: "insufficient staleness", params: []XMLParam{{Name: "max_staleness", Value: "3m30s"}}, wantErr: "max_staleness must be greater than refresh_interval + refresh_timeout"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := append([]XMLParam{{Name: "enableSSL", Value: "0"}}, tc.params...)
			input := XMLInput{Configuration: XMLConfig{Stanza: XMLStanza{Params: params}}}
			cfg, err := input.Extract()
			if err == nil {
				err = cfg.Validate()
			}
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			expected := splunkauthextension.NewDefaultConfig()
			if len(tc.params) != 0 {
				expected.RefreshInterval = 2 * time.Minute
			}
			require.Equal(t, expected, cfg.AuthConfig)
		})
	}
}
