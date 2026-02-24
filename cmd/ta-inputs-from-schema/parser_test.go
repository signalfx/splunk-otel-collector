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

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSchemeXML(t *testing.T) {
	xmlContent := `<scheme>
    <title>Test Scheme</title>
    <description>Test Description</description>
    <endpoint>
        <args>
            <arg name="test_arg" defaultValue="default">
                <title>Test Arg</title>
                <description>Test argument description</description>
                <data_type>string</data_type>
                <required_on_edit>true</required_on_edit>
            </arg>
        </args>
    </endpoint>
</scheme>`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "scheme-*.xml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(xmlContent))
	require.NoError(t, err)
	tmpFile.Close()

	// Parse the scheme
	scheme, err := parseSchemeXML(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, scheme)

	assert.Equal(t, "Test Scheme", scheme.Title)
	assert.Equal(t, "Test Description", scheme.Description)
	assert.Len(t, scheme.Endpoint.Args, 1)

	arg := scheme.Endpoint.Args[0]
	assert.Equal(t, "test_arg", arg.Name)
	assert.Equal(t, "default", arg.DefaultValue)
	assert.Equal(t, "Test Arg", arg.Title)
	assert.Equal(t, "Test argument description", arg.Description)
	assert.Equal(t, "string", arg.DataType)
	assert.Equal(t, "true", arg.RequiredOnEdit)
}

func TestParseSchemeXML_InvalidFile(t *testing.T) {
	_, err := parseSchemeXML("nonexistent-file.xml")
	assert.Error(t, err)
}

func TestParseSchemeXML_InvalidXML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid-*.xml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("<invalid xml"))
	require.NoError(t, err)
	tmpFile.Close()

	_, err = parseSchemeXML(tmpFile.Name())
	assert.Error(t, err)
}

func TestArgIsRequired(t *testing.T) {
	tests := []struct {
		name           string
		arg            Arg
		expectedResult bool
	}{
		{
			name: "required with empty default",
			arg: Arg{
				RequiredOnEdit: "true",
				DefaultValue:   "",
			},
			expectedResult: true,
		},
		{
			name: "required with non-empty default",
			arg: Arg{
				RequiredOnEdit: "true",
				DefaultValue:   "default",
			},
			expectedResult: false,
		},
		{
			name: "not required with empty default",
			arg: Arg{
				RequiredOnEdit: "false",
				DefaultValue:   "",
			},
			expectedResult: false,
		},
		{
			name: "not required with non-empty default",
			arg: Arg{
				RequiredOnEdit: "false",
				DefaultValue:   "default",
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.arg.IsRequired())
		})
	}
}
