// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneInput(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "oneinput.conf"))
	require.NoError(t, err)
	res, err := ReadInput(b, "")
	require.NoError(t, err)
	assert.Equal(
		t,
		[]Input{{
			ServerHost:    "",
			ServerURI:     "",
			SessionKey:    "",
			CheckpointDir: "",
			Configuration: Configuration{
				Stanza: Stanza{
					Name: "otlpinput", App: "tarunner",
					Params: []Param{{
						Name:  "start_by_shell",
						Value: "false",
					}, {Name: "interval", Value: "0"}, {
						Name:  "sourcetype",
						Value: "_otlpinput",
					}, {Name: "index", Value: ""}, {
						Name:  "grpc_port",
						Value: "4317",
					}, {Name: "http_port", Value: "4318"}, {
						Name:  "listen_address",
						Value: "0.0.0.0",
					}},
				},
			},
		}},
		res,
	)
}

func TestTwoInputs(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "twoinputs.conf"))
	require.NoError(t, err)
	res, err := ReadInput(b, "")
	require.NoError(t, err)
	assert.Equal(t, []Input{{
		ServerHost:    "",
		ServerURI:     "",
		SessionKey:    "",
		CheckpointDir: "",
		Configuration: Configuration{
			Stanza: Stanza{
				Name: "otlpinput", App: "tarunner",
				Params: []Param{{
					Name:  "start_by_shell",
					Value: "false",
				}, {
					Name:  "interval",
					Value: "0",
				}, {
					Name:  "sourcetype",
					Value: "_otlpinput",
				}, {
					Name:  "index",
					Value: "",
				}, {
					Name:  "grpc_port",
					Value: "4317",
				}, {
					Name:  "http_port",
					Value: "4318",
				}, {
					Name:  "listen_address",
					Value: "0.0.0.0",
				}},
			},
		},
	}, {
		ServerHost:    "",
		ServerURI:     "",
		SessionKey:    "",
		CheckpointDir: "",
		Configuration: Configuration{
			Stanza: Stanza{
				Name: "otlpinput/2", App: "tarunner",
				Params: []Param{
					{
						Name:  "start_by_shell",
						Value: "true",
					}, {Name: "interval", Value: "0"}, {
						Name:  "sourcetype",
						Value: "foo",
					}, {Name: "index", Value: "nondefault"}, {
						Name:  "grpc_port",
						Value: "1111",
					}, {Name: "http_port", Value: "1112"}, {
						Name:  "listen_address",
						Value: "127.0.0.1",
					},
				},
			},
		},
	}}, res)
}

func TestMergeInputsPartialOverride(t *testing.T) {
	base := []Input{{Configuration: Configuration{Stanza: Stanza{
		Name:   "monitor:///var/log/syslog",
		Params: Params{{Name: "sourcetype", Value: "syslog"}, {Name: "index", Value: "main"}},
	}}}}
	override := []Input{{Configuration: Configuration{Stanza: Stanza{
		Name:   "monitor:///var/log/syslog",
		Params: Params{{Name: "index", Value: "override"}},
	}}}}

	merged := MergeInputs([][]Input{base, override})
	require.Len(t, merged, 1)
	assert.Equal(t, "syslog", merged[0].Configuration.Stanza.Params.Get("sourcetype").Value)
	assert.Equal(t, "override", merged[0].Configuration.Stanza.Params.Get("index").Value)
}

func TestMergeInputsFullOverride(t *testing.T) {
	base := []Input{{Configuration: Configuration{Stanza: Stanza{
		Name:   "monitor:///var/log/syslog",
		Params: Params{{Name: "sourcetype", Value: "syslog"}, {Name: "index", Value: "main"}},
	}}}}
	override := []Input{{Configuration: Configuration{Stanza: Stanza{
		Name:   "monitor:///var/log/syslog",
		Params: Params{{Name: "sourcetype", Value: "sourcetype_override"}, {Name: "index", Value: "index_override"}},
	}}}}

	merged := MergeInputs([][]Input{base, override})
	require.Len(t, merged, 1)
	assert.Equal(t, "sourcetype_override", merged[0].Configuration.Stanza.Params.Get("sourcetype").Value)
	assert.Equal(t, "index_override", merged[0].Configuration.Stanza.Params.Get("index").Value)
}

func TestIsDisabled(t *testing.T) {
	cases := []struct {
		value    string
		disabled bool
	}{
		{"1", true},
		{"true", true},
		{"0", false},
		{"false", false},
		{"", false},
	}
	for _, tc := range cases {
		s := Stanza{Params: Params{{Name: "disabled", Value: tc.value}}}
		assert.Equal(t, tc.disabled, s.IsDisabled(), "disabled=%q", tc.value)
	}
	// No disabled param at all.
	assert.False(t, (&Stanza{}).IsDisabled())
}

func TestToXML(t *testing.T) {
	testStr := `<?xml version="1.0" encoding="UTF-8"?>
<Input>
  <server_host>773c28971b2a</server_host>
  <server_uri>https://127.0.0.1:8089</server_uri>
  <session_key>OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N</session_key>
  <checkpoint_dir>/opt/splunk/var/lib/splunk/modinputs/otlpinput</checkpoint_dir>
  <configuration>
    <stanza name="otlpinput" app="tarunner">
      <param name="start_by_shell">false</param>
      <param name="interval">0</param>
      <param name="sourcetype">_otlpinput</param>
      <param name="index"></param>
      <param name="grpc_port">4317</param>
      <param name="http_port">4318</param>
      <param name="listen_address">0.0.0.0</param>
    </stanza>
  </configuration>
</Input>`
	b, err := os.ReadFile(filepath.Join("testdata", "oneinput.conf"))
	require.NoError(t, err)
	res, err := ReadInput(b, "")
	require.NoError(t, err)
	assert.Len(t, res, 1)
	res[0].ServerHost = "773c28971b2a"
	res[0].ServerURI = "https://127.0.0.1:8089"
	res[0].SessionKey = "OwLHq7jpfgz0WLe5t8KwZuxT4QZRggryMB2io6Phimb2zi5ErifFvx0Eu8WTmfviO^KUKEA8CsGbVltVlCDlYOBM0RE8QoOjOHZhKnHsphk20XoqaK1KXTZj1N"
	res[0].CheckpointDir = "/opt/splunk/var/lib/splunk/modinputs/otlpinput"
	b, err = res[0].ToXML()
	require.NoError(t, err)
	assert.Equal(t, testStr, string(b))
}
