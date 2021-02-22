// Copyright 2020 Splunk, Inc.
// Copyright The OpenTelemetry Authors
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
	"strconv"
	"testing"
)

func TestContains(t *testing.T) {
	testArgs := [][]string{
		{"cmd", "--test=foo"},
		{"cmd", "--test", "foo"},
	}
	for _, v := range testArgs {
		result := contains(v, "--test")
		if !(result) {
			t.Errorf("Expected true got false while testing %v", v)
		}
	}
	testArgs = [][]string{
		{"cmd", "--test-fail", "foo"},
		{"cmd", "--test-fail=--test"},
	}
	for _, v := range testArgs {
		result := contains(v, "--test")
		if result {
			t.Errorf("Expected false got true while testing %v", v)
		}
	}
}

func TestUseMemorySizeFromEnvVar(t *testing.T) {
    testArgs := [][]string{
		{"", "0"},
		{"35", "0"},
		{"", "100"},
	}
	for _, v := range testArgs {
		n, _ := strconv.Atoi(v[1])
		os.Setenv(ballastEnvVarName, v[0])
		useMemorySizeFromEnvVar(n)
		result := contains(os.Args, "--mem-ballast-size-mib")
		if !(result) {
			t.Errorf("Expected true got false while testing %v", v)
		}
	}
}

func TestUseConfigFromEnvVar(t *testing.T) {
	result := contains(os.Args, "--config")
	if result {
		t.Error("Expected false got true while looking for --config")
	}
	os.Setenv(configEnvVarName, "../../"+defaultLocalOTLPConfig)
	useConfigFromEnvVar()
	result = contains(os.Args, "--config")
	if !(result) {
		t.Error("Expected true got false while looking for --config")
	}
}
