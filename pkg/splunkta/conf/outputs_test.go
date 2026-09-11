// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfAndHTTPOut(t *testing.T) {
	payload, err := os.ReadFile("testdata/outputs.conf")
	require.NoError(t, err)

	parsed, err := ParseConf(payload)
	require.NoError(t, err)

	output, err := HTTPOut(parsed)
	require.NoError(t, err)
	require.NotNil(t, output)

	assert.Equal(t, "httpout", output.Configuration.Stanza.Name)
	require.NotNil(t, output.Configuration.Stanza.Params.Get("httpEventCollectorToken"))
	assert.Equal(t, "token-default", output.Configuration.Stanza.Params.Get("httpEventCollectorToken").Value)
	require.NotNil(t, output.Configuration.Stanza.Params.Get("uri"))
	assert.Equal(t, "https://splunk:8088/services/collector/event", output.Configuration.Stanza.Params.Get("uri").Value)
}

func TestHTTPOutMissing(t *testing.T) {
	parsed, err := ParseConf([]byte("[tcpout]\nserver = splunk:9997\n"))
	require.NoError(t, err)
	_, err = HTTPOut(parsed)
	require.ErrorIs(t, err, ErrNoHTTPOut)
}

func TestMergeConf(t *testing.T) {
	base, err := ParseConf([]byte("[httpout]\nhttpEventCollectorToken = base-token\nuri = https://base:8088/services/collector/event\n"))
	require.NoError(t, err)

	override, err := ParseConf([]byte("[httpout]\nhttpEventCollectorToken = override-token\n"))
	require.NoError(t, err)

	merged := MergeConf([]ConfMap{base, override})
	output, err := HTTPOut(merged)
	require.NoError(t, err)

	// override wins for token, base value kept for uri
	assert.Equal(t, "override-token", requireOutputParam(t, output, "httpEventCollectorToken"))
	assert.Equal(t, "https://base:8088/services/collector/event", requireOutputParam(t, output, "uri"))
}

func TestReadOutputGroups(t *testing.T) {
	payload, err := os.ReadFile("testdata/outputs.conf")
	require.NoError(t, err)

	outputs, err := ReadOutputGroups(payload)
	require.NoError(t, err)
	require.Len(t, outputs, 2)

	assert.Equal(t, "httpout", outputs[0].Configuration.Stanza.Name)
	assert.Equal(t, "httpEventCollectorToken", outputs[0].Configuration.Stanza.Params[0].Name)
	assert.Equal(t, "token-default", outputs[0].Configuration.Stanza.Params[0].Value)
	require.NotNil(t, outputs[0].Configuration.Stanza.Params.Get("uri"))
	assert.Equal(t, "https://splunk:8088/services/collector/event", outputs[0].Configuration.Stanza.Params.Get("uri").Value)

	assert.Equal(t, "tcpout", outputs[1].Configuration.Stanza.Name)
	require.NotNil(t, outputs[1].Configuration.Stanza.Params.Get("server"))
	assert.Equal(t, "splunk:9997", outputs[1].Configuration.Stanza.Params.Get("server").Value)
}

func TestReadOutputGroupsNoStanzas(t *testing.T) {
	_, err := ReadOutputGroups([]byte(""))
	require.ErrorIs(t, err, ErrNoOutputStanzas)
}

func requireOutputParam(t *testing.T, output *Output, name string) string {
	t.Helper()
	param := output.Configuration.Stanza.Params.Get(name)
	require.NotNil(t, param)
	return param.Value
}
