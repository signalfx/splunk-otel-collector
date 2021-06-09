// Copyright Splunk, Inc.
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
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
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
		_, result := getKeyValue(v, "--test")
		if result != v[0] {
			t.Errorf("Expected %v got %v", v[0], v)
		}
	}
}

func TestCheckRuntimeParams(*testing.T) {
	oldArgs := os.Args
	os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig))
	checkConfig()
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
	_, c := getKeyValue(args, "--mem-ballast-size-mib")
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
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	configPath := path.Join("../../", defaultLocalSAPMConfig)
	os.Setenv(configEnvVarName, configPath)
	defer os.Unsetenv(configEnvVarName)
	checkConfig()

	args := os.Args[1:]
	_, c := getKeyValue(args, "--config")
	if c != path.Join("../../", defaultLocalSAPMConfig) {
		t.Error("Config CLI param not set as expected")
	}
}

func TestConfigPrecedence(t *testing.T) {
	validPath1 := path.Join("../../", defaultLocalSAPMConfig)
	validPath2 := path.Join("../../", defaultLocalOTLPConfig)
	validConfig := `receivers:
  hostmetrics:
    collection_interval: 1s
    scrapers:
      cpu:
exporters:
  logging:
    logLevel: debug
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [logging]`

	tests := []struct {
		name                string
		configFlagVal       string // Flag --config value
		splunkConfigVal     string // Environment variable SPLUNK_CONFIG value
		splunkConfigYamlVal string // Environment variable SPLUNK_CONFIG_YAML value
		expectedLogs        []string
		unexpectedLogs      []string
	}{
		{
			name:                "Flag --config precedences env SPLUNK_CONFIG and SPLUNK_CONFIG_YAML",
			configFlagVal:       validPath1,
			splunkConfigVal:     validPath2,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both environment variable SPLUNK_CONFIG and flag '--config' were specified. Using the flag value %s and ignoring the environment variable value %s in this session", validPath1, validPath2),
				fmt.Sprintf("Both environment variable SPLUNK_CONFIG_YAML and flag '--config' were specified. Using the flag value %s and ignoring the environment variable in this session", validPath1),
				fmt.Sprintf("Set config to %v", validPath1),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to %v", validPath2),
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
			},
		},
		{
			name:                "env SPLUNK_CONFIG precedences SPLUNK_CONFIG_YAML",
			configFlagVal:       "",
			splunkConfigVal:     validPath2,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both %s and %s were specified. Using %s environment variable value %s for this session", configEnvVarName, configYamlEnvVarName, configEnvVarName, validPath2),
				fmt.Sprintf("Set config to %v", validPath2),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to %v", validPath1),
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
			},
		},
		{
			name:                "env SPLUNK_CONFIG_YAML used when flag --config and env SPLUNK_CONFIG not specified",
			configFlagVal:       "",
			splunkConfigVal:     "",
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to %v", validPath1),
				fmt.Sprintf("Set config to %v", validPath2),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			func() {
				oldArgs := os.Args
				oldWriter := log.Default().Writer()

				defer func() {
					os.Args = oldArgs
					os.Unsetenv(configEnvVarName)
					os.Unsetenv(configYamlEnvVarName)
					log.Default().SetOutput(oldWriter)
				}()

				actualLogsBuf := new(bytes.Buffer)
				log.Default().SetOutput(actualLogsBuf)
				if test.configFlagVal != "" {
					os.Args = append(os.Args, "--config="+test.configFlagVal)
				}
				os.Setenv(configEnvVarName, test.splunkConfigVal)
				os.Setenv(configYamlEnvVarName, test.splunkConfigYamlVal)

				checkConfig()

				actualLogs := actualLogsBuf.String()

				for _, expectedLog := range test.expectedLogs {
					assert.Contains(t, actualLogs, expectedLog)
				}
				for _, unexpectedLog := range test.unexpectedLogs {
					assert.NotContains(t, actualLogs, unexpectedLog)
				}
			}()
		})
	}
}

func TestSetMemoryBallast(t *testing.T) {
	oldArgs := os.Args
	setMemoryBallast(100)

	HelperTestSetMemoryBallast("33", t)

	os.Args = oldArgs
	os.Setenv(ballastEnvVarName, "50")
	defer os.Unsetenv(ballastEnvVarName)
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
