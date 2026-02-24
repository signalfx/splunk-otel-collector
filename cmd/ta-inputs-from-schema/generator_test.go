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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateInputsConf(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:         "test_arg1",
					DefaultValue: "value1",
					Description:  "Test argument 1",
				},
				{
					Name:         "test_arg2",
					DefaultValue: "",
					Description:  "Test argument 2",
				},
			},
		},
	}

	globalSettings := `# Global settings
disabled=false
interval = 0`

	inputName := "Test_TA"

	result := generateInputsConf(scheme, globalSettings, inputName)

	// Verify the header
	assert.Contains(t, result, "[Test_TA://Test_TA]")

	// Verify global settings are included
	assert.Contains(t, result, "disabled=false")
	assert.Contains(t, result, "interval = 0")

	// Verify TA specific settings header
	assert.Contains(t, result, "# TA specific settings")

	// Verify arguments are included
	assert.Contains(t, result, "test_arg1 = value1")
	assert.Contains(t, result, "test_arg2 =")
}

func TestGenerateInputsConf_GlobalSettingsWithoutTrailingNewline(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:         "test_arg",
					DefaultValue: "value",
				},
			},
		},
	}

	globalSettings := "disabled=false"
	inputName := "Test_TA"

	result := generateInputsConf(scheme, globalSettings, inputName)

	// Verify proper spacing between global settings and TA settings
	lines := strings.Split(result, "\n")
	var foundGlobalSettings bool
	var foundTASettings bool
	for i, line := range lines {
		if line == "disabled=false" {
			foundGlobalSettings = true
			// Check that there's a blank line after global settings
			if i+1 < len(lines) {
				assert.Empty(t, lines[i+1], "Expected blank line after global settings")
			}
		}
		if line == "# TA specific settings" {
			foundTASettings = true
		}
	}
	assert.True(t, foundGlobalSettings, "Global settings should be present")
	assert.True(t, foundTASettings, "TA settings header should be present")
}

func TestGenerateInputsConfSpec(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:           "required_arg",
					DefaultValue:   "",
					Description:    "Required argument with no default",
					RequiredOnEdit: "true",
				},
				{
					Name:           "optional_arg",
					DefaultValue:   "default_value",
					Description:    "Optional argument with default",
					RequiredOnEdit: "true",
				},
				{
					Name:           "another_arg",
					DefaultValue:   "value",
					Description:    "Another optional argument",
					RequiredOnEdit: "false",
				},
			},
		},
	}

	inputName := "Test_TA"

	result := generateInputsConfSpec(scheme, inputName)

	// Verify the header
	assert.Contains(t, result, "[Test_TA://<name>]")

	// Verify required argument
	assert.Contains(t, result, "required_arg = <value>")
	assert.Contains(t, result, "* Required argument with no default")
	assert.Contains(t, result, "* Default =")

	// Verify optional argument with default
	assert.Contains(t, result, "optional_arg = <value>")
	assert.Contains(t, result, "* Optional argument with default")
	assert.Contains(t, result, "* Default = default_value")

	// Verify another optional argument
	assert.Contains(t, result, "another_arg = <value>")
	assert.Contains(t, result, "* Another optional argument")
	assert.Contains(t, result, "* Default = value")
}

func TestGenerateInputsConfSpec_DescriptionPreserved(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:           "test_arg",
					DefaultValue:   "value",
					Description:    "Description with capital letter",
					RequiredOnEdit: "false",
				},
			},
		},
	}

	inputName := "Test_TA"

	result := generateInputsConfSpec(scheme, inputName)

	// Verify description is preserved as-is
	assert.Contains(t, result, "* Description with capital letter")
}

func TestGenerateInputsConfSpec_EmptyDescription(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:           "test_arg",
					DefaultValue:   "value",
					Description:    "",
					RequiredOnEdit: "false",
				},
			},
		},
	}

	inputName := "Test_TA"

	result := generateInputsConfSpec(scheme, inputName)

	// Verify it handles empty description
	assert.Contains(t, result, "test_arg = <value>")
	// Should have a line starting with * but empty description
	lines := strings.Split(result, "\n")
	foundEmptyDesc := false
	for _, line := range lines {
		if line == "*" || line == "* " {
			foundEmptyDesc = true
			break
		}
	}
	assert.True(t, foundEmptyDesc, "Should handle empty description")
}

func TestGenerateInputsConfSpec_MultipleArgs(t *testing.T) {
	scheme := &Scheme{
		Endpoint: Endpoint{
			Args: []Arg{
				{
					Name:         "arg1",
					DefaultValue: "val1",
					Description:  "First arg",
				},
				{
					Name:         "arg2",
					DefaultValue: "val2",
					Description:  "Second arg",
				},
			},
		},
	}

	inputName := "Test_TA"

	result := generateInputsConfSpec(scheme, inputName)

	// Verify both arguments are present with blank lines between them
	lines := strings.Split(result, "\n")
	arg1Index := -1
	arg2Index := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "arg1 =") {
			arg1Index = i
		}
		if strings.HasPrefix(line, "arg2 =") {
			arg2Index = i
		}
	}

	assert.NotEqual(t, -1, arg1Index, "arg1 should be present")
	assert.NotEqual(t, -1, arg2Index, "arg2 should be present")
	assert.Greater(t, arg2Index, arg1Index, "arg2 should come after arg1")
}

func TestNormalizeDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line with capital",
			input:    "Description with capital letter",
			expected: "Description with capital letter",
		},
		{
			name:     "multiple lines",
			input:    "Description\n    with multiple lines",
			expected: "Description with multiple lines",
		},
		{
			name:     "extra whitespace",
			input:    "Description  with   extra    whitespace",
			expected: "Description with extra whitespace",
		},
		{
			name:     "tabs and newlines",
			input:    "Description\twith\ttabs\nand\nnewlines",
			expected: "Description with tabs and newlines",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "Description",
			expected: "Description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeDescription(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
