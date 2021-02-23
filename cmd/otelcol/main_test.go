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
	//"strconv"
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

func TestGetKeyValue(t *testing.T) {
	testArgs := [][]string{
		{"", "--bar=foo"},
		{"foo", "--test=foo"},
		{"foo", "--test", "foo"},
	}
	for _, v := range testArgs {
		result := getKeyValue(v, "--test")
		if result != v[0] {
			t.Errorf("Expected %v got %v", v[0], v)
		}
	}
}

func TestCheckRuntimeParams(t *testing.T) {
	oldArgs := os.Args
	os.Setenv(configEnvVarName, "../../"+defaultLocalSAPMConfig)
	setConfig()
	os.Unsetenv(configEnvVarName)
	checkRuntimeParams()

	os.Args = oldArgs
	os.Setenv(memTotalEnvVarName, "1000")
	checkRuntimeParams()

	os.Args = oldArgs
	os.Setenv(ballastEnvVarName, "50")
	setMemoryBallast(100)
	os.Unsetenv(ballastEnvVarName)
	checkRuntimeParams()

	os.Args = oldArgs
	os.Clearenv()
}

func HelperTestSetMemoryBallast(val string, t *testing.T) {
	args := os.Args[1:]
	c := getKeyValue(args, "--mem-ballast-size-mib")
	if c != val {
		t.Errorf("Expected memory ballast CLI param %v got %v", val, c)
	}
	b := os.Getenv(ballastEnvVarName)
	if b != val {
		t.Errorf("Expected memory ballast %v got %v", val, b)
	}
}

func HelperTestSetMemoryLimit(val string, t *testing.T) {
	b := os.Getenv(memLimitMiBEnvVarName)
	if b != val {
		t.Errorf("Expected memory limit %v got %v", val, b)
	}
}

func TestUseConfigFromEnvVar(t *testing.T) {
	os.Setenv(tokenEnvVarName, "12345")
	os.Setenv(realmEnvVarName, "us0")
	os.Setenv(configEnvVarName, "../../"+defaultLocalSAPMConfig)
	setConfig()

	args := os.Args[1:]
	c := getKeyValue(args, "--config")
	if c != "../../"+defaultLocalSAPMConfig {
		t.Error("Config CLI param not set as expected")
	}
}

func TestSetMemoryBallast(t *testing.T) {
	oldArgs := os.Args
	setMemoryBallast(100)

	HelperTestSetMemoryBallast("33", t)

	os.Args = oldArgs
	os.Setenv(ballastEnvVarName, "50")
	setMemoryBallast(100)

	HelperTestSetMemoryBallast("50", t)
	os.Args = oldArgs
}

func TestSetMemoryLimit(t *testing.T) {
	oldArgs := os.Args
	setMemoryLimit(100)

	HelperTestSetMemoryLimit("90", t)

	os.Args = oldArgs
	os.Unsetenv(memLimitMiBEnvVarName)
	setMemoryLimit(100000)

	HelperTestSetMemoryLimit("2048", t)

	os.Args = oldArgs
	os.Setenv(memLimitMiBEnvVarName, "200")
	setMemoryLimit(100)

	HelperTestSetMemoryLimit("200", t)
}
