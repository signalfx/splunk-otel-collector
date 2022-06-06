// Copyright  Splunk, Inc.
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
)

func TestFlagParseFailure(t *testing.T) {
	oldArgs := os.Args

	// Parsing should stop once an unspecified flag is found
	os.Args = []string{"otelcol", "--invalid-flag", "100", "--version", "true"}
	inputFlags, err := parseFlags(os.Args[1:])
	assert.EqualError(t, err, "flag provided but not defined: -invalid-flag")
	assert.False(t, inputFlags.version)

	// Make sure wrong name doesn't get parsed into given variable
	os.Args = []string{"otelcol", "--ver", "true"}
	inputFlags, err = parseFlags(os.Args[1:])
	assert.EqualError(t, err, "flag provided but not defined: -ver")
	assert.False(t, inputFlags.version)

	os.Args = oldArgs
	os.Clearenv()
}

// Test each variable can get a value
func TestFlagParseSuccess(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--version"}
	inputFlags, err := parseFlags(os.Args[1:])
	assert.NoError(t, err)
	assert.True(t, inputFlags.version)

	os.Args = []string{"otelcol",
		"--version",
		"--help",
		"--no-convert-config",
		"--config", "foo.yml",
		"--config", "bar.yml",
		"--mem-ballast-size-mib", "100",
		"--set", "foo",
		"--set", "bar",
		"--set", "baz",
		"--feature-gates", "foo",
		"--feature-gates", "-bar"}

	inputFlags, err = parseFlags(os.Args[1:])
	assert.NoError(t, err)

	assert.True(t, inputFlags.version)
	assert.True(t, inputFlags.help)
	assert.True(t, inputFlags.noConvertConfig)

	assert.Contains(t, inputFlags.configs.values, "foo.yml")
	assert.Contains(t, inputFlags.configs.values, "bar.yml")

	assert.Equal(t, 100, inputFlags.memBallastSizeMib)

	assert.Contains(t, inputFlags.sets.values, "foo")
	assert.Contains(t, inputFlags.sets.values, "bar")
	assert.Contains(t, inputFlags.sets.values, "baz")

	assert.Equal(t, true, inputFlags.gatesList["foo"])
	assert.Equal(t, false, inputFlags.gatesList["bar"])

	os.Args = oldArgs
	os.Clearenv()
}

// Test to make sure different flag names set variable value for same variable
func TestShortenedFlagNames(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--v", "--h"}
	inputFlags, err := parseFlags(os.Args[1:])
	assert.NoError(t, err)
	assert.True(t, inputFlags.version)
	assert.True(t, inputFlags.help)

	os.Args = oldArgs
	os.Clearenv()
}
