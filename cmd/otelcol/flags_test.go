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
	flagSet := flags()
	flagSet.Parse(os.Args[1:])
	assert.False(t, versionFlag)

	// Make sure wrong name doesn't get parsed into given variable
	os.Args = []string{"otelcol", "--ver", "true"}
	flagSet = flags()
	flagSet.Parse(os.Args[1:])
	assert.False(t, versionFlag)

	os.Args = oldArgs
	os.Clearenv()
}

// Test each variable can get a value
func TestFlagParseSuccess(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--version"}
	flagSet := flags()
	flagSet.Parse(os.Args[1:])
	assert.True(t, versionFlag)
	versionFlag = false

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

	flagSet = flags()
	flagSet.Parse(os.Args[1:])

	assert.True(t, versionFlag)
	assert.True(t, helpFlag)
	assert.True(t, noConvertConfigFlag)

	assert.Contains(t, getConfigFlags(), "foo.yml")
	assert.Contains(t, getConfigFlags(), "bar.yml")

	assert.Equal(t, 100, memBallastSizeMibFlag)

	assert.Contains(t, getSetFlags(), "foo")
	assert.Contains(t, getSetFlags(), "bar")
	assert.Contains(t, getSetFlags(), "baz")

	assert.Equal(t, true, gatesList["foo"])
	assert.Equal(t, false, gatesList["bar"])

	os.Args = oldArgs
	os.Clearenv()
}

// Test to make sure different flag names set variable value for same variable
func TestShortenedFlagNames(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"otelcol", "--v", "--h"}
	flagSet := flags()
	flagSet.Parse(os.Args[1:])
	assert.True(t, versionFlag)
	assert.True(t, helpFlag)
	versionFlag = false
	helpFlag = false

	os.Args = oldArgs
	os.Clearenv()
}
