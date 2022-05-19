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
	inputFlags, _ := parseFlags(os.Args[1:])
	assert.False(t, inputFlags.versionFlag)

	// Make sure wrong name doesn't get parsed into given variable
	os.Args = []string{"otelcol", "--ver", "true"}
	inputFlags, _ = parseFlags(os.Args[1:])
	assert.False(t, inputFlags.versionFlag)

	os.Args = oldArgs
	os.Clearenv()
}

// Test each variable can get a value
func TestFlagParseSuccess(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--version"}
	inputFlags, _ := parseFlags(os.Args[1:])
	assert.True(t, inputFlags.versionFlag)

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

	inputFlags, _ = parseFlags(os.Args[1:])

	assert.True(t, inputFlags.versionFlag)
	assert.True(t, inputFlags.helpFlag)
	assert.True(t, inputFlags.noConvertConfigFlag)

	assert.Contains(t, inputFlags.getConfigFlags(), "foo.yml")
	assert.Contains(t, inputFlags.getConfigFlags(), "bar.yml")

	assert.Equal(t, 100, inputFlags.memBallastSizeMibFlag)

	assert.Contains(t, inputFlags.getSetFlags(), "foo")
	assert.Contains(t, inputFlags.getSetFlags(), "bar")
	assert.Contains(t, inputFlags.getSetFlags(), "baz")

	assert.Equal(t, true, inputFlags.gatesList["foo"])
	assert.Equal(t, false, inputFlags.gatesList["bar"])

	os.Args = oldArgs
	os.Clearenv()
}

// Test to make sure different flag names set variable value for same variable
func TestShortenedFlagNames(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--v", "--h"}
	inputFlags, _ := parseFlags(os.Args[1:])
	assert.True(t, inputFlags.versionFlag)
	assert.True(t, inputFlags.helpFlag)

	os.Args = oldArgs
	os.Clearenv()
}
