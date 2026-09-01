// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
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
      <param name="start_by_shell">false</param>
    </stanza>
  </configuration>
</input>`

	var config XMLInput
	err := xml.Unmarshal([]byte(input), &config)
	require.NoError(t, err)

	require.Equal(t, "splunk-connect-for-otlp://specialmind", config.Configuration.Stanza.Name)

	xmlCfg := config.Extract()

	require.Equal(t, 4317, xmlCfg.GrpcPort)
	require.Equal(t, "0.0.0.0", xmlCfg.ListenAddress)
	require.Equal(t, 4318, xmlCfg.HTTPPort)
	require.Empty(t, xmlCfg.Source)
	require.Equal(t, "_splunk-connect-for-otlp", xmlCfg.Sourcetype)
	require.Equal(t, "https://127.0.0.1:8089", xmlCfg.ServerURI)
	require.Equal(t, "OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N", xmlCfg.SessionKey)
	require.True(t, xmlCfg.EnableSSL)
	require.Equal(t, "/var/certs/server.cert", xmlCfg.ServerCert)
	require.Equal(t, "/var/certs/server.key", xmlCfg.ServerKey)
}
