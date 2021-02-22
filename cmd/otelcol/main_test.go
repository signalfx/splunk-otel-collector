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
	}
	for _, v := range testArgs {
		n, _ := strconv.Atoi(v[1])
		os.Setenv(ballastEnvVarName, v[0])
		useMemorySizeFromEnvVar(n)
		result := contains(os.Args, "--mem-ballast-size-mib")
		if result {
			t.Errorf("Expected false got true while testing %v", v)
		}
	}
	testArgs = [][]string{
		{"10", "0"},
		{"", "10"},
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
	os.Setenv(configEnvVarName, "../../"+defaultLocalSAPMNonLinuxConfig)
	useConfigFromEnvVar()
	result = contains(os.Args, "--config")
	if !(result) {
		t.Error("Expected true got false while looking for --config")
	}
	useMemorySettingsMiBFromEnvVar(0)
	_, present := os.LookupEnv(memLimitMiBEnvVarName)
	if present {
		t.Error("Expected false got true while looking for environment variable")
	}
	os.Unsetenv(configEnvVarName)
}

func HelperUseMemorySettingsFromEnvVar(limitEnv string, limit int, spikeEnv string, spike int, t *testing.T) {
	envVarVal, _ := strconv.Atoi(os.Getenv(limitEnv))
	if envVarVal != limit {
		t.Errorf("Expected %d but got %d", limit, envVarVal)
	}
	envVarVal, _ = strconv.Atoi(os.Getenv(spikeEnv))
	if envVarVal != spike {
		t.Errorf("Expected %d but got %d", spike, envVarVal)
	}
	os.Unsetenv(limitEnv)
	os.Unsetenv(spikeEnv)
}

func TestSetMemorySettingsToEnvVar(t *testing.T) {
	setMemorySettingsToEnvVar(10, memLimitMiBEnvVarName, 5, memSpikeMiBEnvVarName)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 10, memSpikeMiBEnvVarName, 5, t)
}
func TestUseMemorySettingsMiBFromEnvVar(t *testing.T) {
	useMemorySettingsMiBFromEnvVar(100)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 90, memSpikeMiBEnvVarName, 25, t)
	useMemorySettingsMiBFromEnvVar(30000)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 27952, memSpikeMiBEnvVarName, 2048, t)

	setMemorySettingsToEnvVar(10, memLimitMiBEnvVarName, 5, memSpikeMiBEnvVarName)
	useMemorySettingsMiBFromEnvVar(0)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 10, memSpikeMiBEnvVarName, 5, t)
	useMemorySettingsMiBFromEnvVar(100)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 90, memSpikeMiBEnvVarName, 25, t)
}
