// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package tabuilder

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func requireOutputParam(t *testing.T, output *conf.Output, name string) string {
	t.Helper()
	param := output.Configuration.Stanza.Params.Get(name)
	require.NotNil(t, param)
	return param.Value
}

func TestReadOutputs(t *testing.T) {
	rootDir := filepath.Join("testdata", "outputs")
	tests := []struct {
		expectedErr   error
		name          string
		splunkHome    string
		expectedToken string
		expectedURI   string
	}{
		{
			name:          "system_default_only",
			splunkHome:    filepath.Join(rootDir, "system_default"),
			expectedToken: "token-system-default",
			expectedURI:   "https://system-default:8088/services/collector/event",
		},
		{
			name:          "app_overrides_system_default",
			splunkHome:    filepath.Join(rootDir, "app_overrides"),
			expectedToken: "token-app-default",
			expectedURI:   "https://app-local:8088/services/collector/event",
		},
		{
			name:          "system_local_wins_over_app",
			splunkHome:    filepath.Join(rootDir, "system_local_wins"),
			expectedToken: "token-system-local",
			expectedURI:   "https://system-local:8088/services/collector/event",
		},
		{
			name:          "app_local_wins_over_app_default",
			splunkHome:    filepath.Join(rootDir, "app_local_wins_over_app_default"),
			expectedToken: "token-app-local",
			expectedURI:   "https://app-local:8088/services/collector/event",
		},
		{
			name:        "no_httpout",
			splunkHome:  filepath.Join(rootDir, "no_httpout"),
			expectedErr: conf.ErrNoHTTPOut,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merged, err := ReadOutputs(test.splunkHome)
			if test.expectedErr != nil {
				require.NoError(t, err) // ReadOutputs itself doesn't error on missing stanza
				_, httpErr := HTTPOut(merged)
				require.ErrorIs(t, httpErr, test.expectedErr)
				return
			}
			require.NoError(t, err)
			output, err := HTTPOut(merged)
			require.NoError(t, err)
			require.NotNil(t, output)
			assert.Equal(t, test.expectedToken, requireOutputParam(t, output, "httpEventCollectorToken"))
			assert.Equal(t, test.expectedURI, requireOutputParam(t, output, "uri"))
		})
	}
}

func TestReadOutputGroups(t *testing.T) {
	outputs, err := ReadOutputGroups(filepath.Join("testdata", "outputs", "no_httpout"))
	require.NoError(t, err)
	require.Len(t, outputs, 1)

	assert.Equal(t, "tcpout", outputs[0].Configuration.Stanza.Name)
	assert.Equal(t, "splunk:9997", requireOutputParam(t, &outputs[0], "server"))
}

func TestReadInputsForTA(t *testing.T) {
	rootDir := filepath.Join("testdata", "inputs")
	tests := []struct {
		name               string
		splunkHome         string
		expectedStanza     string
		expectedIndex      string
		expectedSourcetype string
		expectedCount      int
	}{
		{
			name:           "ta_default_only",
			splunkHome:     filepath.Join(rootDir, "ta_default_only"),
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "main",
		},
		{
			name:           "ta_local_overrides_ta_default",
			splunkHome:     filepath.Join(rootDir, "ta_local_overrides"),
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "local_index",
		},
		{
			name:           "ta_default_wins_over_system_default",
			splunkHome:     filepath.Join(rootDir, "system_default_overrides"),
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "ta_index",
		},
		{
			name:           "system_local_wins_over_all",
			splunkHome:     filepath.Join(rootDir, "system_local_wins"),
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "system_local_index",
		},
		{
			name:          "system_stanza_not_in_ta_is_excluded",
			splunkHome:    filepath.Join(rootDir, "system_stanza_excluded"),
			expectedCount: 1,
			// only the TA's stanza is returned, not the system-only one
			expectedStanza: "monitor:///var/log/syslog",
		},
		{
			name:               "system_default_adds_missing_param",
			splunkHome:         filepath.Join(rootDir, "system_adds_param"),
			expectedCount:      1,
			expectedStanza:     "monitor:///var/log/syslog",
			expectedSourcetype: "syslog",
			expectedIndex:      "system_index",
		},
		{
			name:          "no_inputs_returns_nil",
			splunkHome:    filepath.Join(rootDir, "no_inputs"),
			expectedCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taDir := filepath.Join(test.splunkHome, "etc", "apps", "splunk_ta_test")
			inputs, err := ReadInputsForTA(test.splunkHome, taDir)
			require.NoError(t, err)
			require.Len(t, inputs, test.expectedCount)
			if test.expectedCount == 0 {
				return
			}
			assert.Equal(t, test.expectedStanza, inputs[0].Configuration.Stanza.Name)
			if test.expectedIndex != "" {
				param := inputs[0].Configuration.Stanza.Params.Get("index")
				require.NotNil(t, param)
				assert.Equal(t, test.expectedIndex, param.Value)
			}
			if test.expectedSourcetype != "" {
				param := inputs[0].Configuration.Stanza.Params.Get("sourcetype")
				require.NotNil(t, param)
				assert.Equal(t, test.expectedSourcetype, param.Value)
			}
		})
	}
}

func TestReadSystemInputs(t *testing.T) {
	rootDir := filepath.Join("testdata", "inputs")
	tests := []struct {
		name            string
		splunkHome      string
		expectedStanzas []string
	}{
		{
			name:       "no_system_inputs_returns_nil",
			splunkHome: filepath.Join(rootDir, "no_inputs"),
		},
		{
			name:            "returns_system_only_stanza_not_owned_by_ta",
			splunkHome:      filepath.Join(rootDir, "system_stanza_excluded"),
			expectedStanzas: []string{"monitor:///var/log/auth.log"},
		},
		{
			name:       "excludes_stanza_owned_by_ta_returns_only_unowned",
			splunkHome: filepath.Join(rootDir, "system_stanza_owned_by_ta"),
			// monitor:///var/log/syslog is defined in the TA and must be excluded;
			// monitor:///var/log/auth.log is system-only and must be returned.
			expectedStanzas: []string{"monitor:///var/log/auth.log"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputs, err := ReadSystemInputs(test.splunkHome)
			require.NoError(t, err)
			require.Len(t, inputs, len(test.expectedStanzas))
			for i, stanza := range test.expectedStanzas {
				assert.Equal(t, stanza, inputs[i].Configuration.Stanza.Name)
			}
		})
	}
}

func TestReadTransforms(t *testing.T) {
	rootDir := filepath.Join("testdata", "transforms")
	tests := []struct {
		name          string
		splunkHome    string
		expectedName  string
		expectedRegex string
	}{
		{
			name:          "default",
			splunkHome:    filepath.Join(rootDir, "splunk_default"),
			expectedName:  "example_default",
			expectedRegex: "default",
		},
		{
			name:          "local",
			splunkHome:    filepath.Join(rootDir, "splunk_local"),
			expectedName:  "example_local",
			expectedRegex: "local",
		},
		{
			name:          "both",
			splunkHome:    filepath.Join(rootDir, "splunk_both"),
			expectedName:  "example_transform",
			expectedRegex: "local",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transforms, err := ReadTransforms(ConfDirs(test.splunkHome))
			require.NoError(t, err)
			require.Len(t, transforms, 1)
			require.Equal(t, test.expectedName, transforms[0].Name)
			require.Equal(t, test.expectedRegex, transforms[0].Regex)
		})
	}
}
