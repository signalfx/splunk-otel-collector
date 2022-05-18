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

func testInitFlags() {
	configFlags = new(stringArrayValue)
	setFlags = new(stringArrayValue)
	flagSet := flags()
	flagSet.Parse(os.Args[1:])
}

func testCheckRuntimeParams() {
	testInitFlags()
	checkRuntimeParams()
}

func TestCheckRuntimeParams_Default(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	testCheckRuntimeParams()
	assert.Equal(t, "168", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "460", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
}

func TestCheckRuntimeParams_MemTotalEnv(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	assert.NoError(t, os.Setenv(memTotalEnvVarName, "1000"))
	testCheckRuntimeParams()
	assert.Equal(t, "330", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "900", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
}

func TestCheckRuntimeParams_MemTotalAndBallastEnvs(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	assert.NoError(t, os.Setenv(memTotalEnvVarName, "200"))
	assert.NoError(t, os.Setenv(ballastEnvVarName, "90"))
	testCheckRuntimeParams()
	assert.Equal(t, "90", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "180", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
}

func TestCheckRuntimeParams_LimitAndBallastEnvs(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	assert.NoError(t, os.Setenv(memLimitMiBEnvVarName, "250"))
	assert.NoError(t, os.Setenv(ballastEnvVarName, "120"))
	testCheckRuntimeParams()
	assert.Equal(t, "120", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "250", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
}

func TestSetDefaultEnvVars(t *testing.T) {
	testArgs := []string{"SPLUNK_API_URL", "SPLUNK_INGEST_URL", "SPLUNK_TRACE_URL", "SPLUNK_HEC_URL", "SPLUNK_HEC_TOKEN"}
	for _, v := range testArgs {
		setDefaultEnvVars()
		_, ok := os.LookupEnv(v)
		if ok {
			t.Errorf("Expected %v unset given SPLUNK_ACCESS_TOKEN or SPLUNK_TOKEN is unset", v)
		}
	}

	realm := "us1"
	token := "1234"
	valTest := "test"
	valEmpty := ""

	os.Setenv("SPLUNK_REALM", realm)
	os.Setenv("SPLUNK_ACCESS_TOKEN", token)
	setDefaultEnvVars()
	testArgs2 := [][]string{
		{"SPLUNK_API_URL", "https://api." + realm + ".signalfx.com"},
		{"SPLUNK_INGEST_URL", "https://ingest." + realm + ".signalfx.com"},
		{"SPLUNK_TRACE_URL", "https://ingest." + realm + ".signalfx.com/v2/trace"},
		{"SPLUNK_HEC_URL", "https://ingest." + realm + ".signalfx.com/v1/log"},
		{"SPLUNK_HEC_TOKEN", token},
	}
	for _, v := range testArgs2 {
		val, _ := os.LookupEnv(v[0])
		if val != v[1] {
			t.Errorf("Expected %v got %v for %v", v[1], val, v[0])
		}
	}

	for _, v := range testArgs {
		os.Setenv(v, valTest)
		setDefaultEnvVars()
		val, _ := os.LookupEnv(v)
		if val != valTest {
			t.Errorf("Expected %v got %v for %v", valTest, val, v)
		}
	}

	for _, v := range testArgs {
		os.Setenv(v, valEmpty)
		setDefaultEnvVars()
		val, _ := os.LookupEnv(v)
		if val != valEmpty {
			t.Errorf("Expected %v got %v for %v", valEmpty, val, v)
		}
	}
}

func TestCheckRuntimeParams_MemTotalLimitAndBallastEnvs(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	assert.NoError(t, os.Setenv(memTotalEnvVarName, "200"))
	assert.NoError(t, os.Setenv(memLimitMiBEnvVarName, "150"))
	assert.NoError(t, os.Setenv(ballastEnvVarName, "50"))
	testCheckRuntimeParams()
	assert.Equal(t, "50", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "150", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
}

func TestCheckRuntimeParams_MemTotalEnvAndBallastFlag(t *testing.T) {
	oldArgs := os.Args
	assert.NoError(t, os.Setenv(configEnvVarName, path.Join("../../", defaultLocalSAPMConfig)))
	assert.NoError(t, os.Setenv(memTotalEnvVarName, "200"))
	os.Args = []string{"otelcol", "--mem-ballast-size-mib=90"}
	testCheckRuntimeParams()
	assert.Equal(t, "90", os.Getenv(ballastEnvVarName))
	assert.Equal(t, "180", os.Getenv(memLimitMiBEnvVarName))

	os.Args = oldArgs
	os.Clearenv()
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

	testInitFlags()
	if configPath != path.Join("../../", defaultLocalSAPMConfig) {
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
		configFlagVals      []string // Flag --config value
		splunkConfigVal     string   // Environment variable SPLUNK_CONFIG value
		splunkConfigYamlVal string   // Environment variable SPLUNK_CONFIG_YAML value
		expectedLogs        []string
		unexpectedLogs      []string
	}{
		{
			name:                "Flag --config precedences env SPLUNK_CONFIG and SPLUNK_CONFIG_YAML",
			configFlagVals:      []string{validPath1},
			splunkConfigVal:     validPath2,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both environment variable SPLUNK_CONFIG and flag '--config' were specified. Using the flag values and ignoring the environment variable value %s in this session", validPath2),
				fmt.Sprintf("Set config to [%v]", validPath1),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to [%v]", validPath2),
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
			},
		},
		{
			name:                "env SPLUNK_CONFIG precedences SPLUNK_CONFIG_YAML",
			configFlagVals:      []string{},
			splunkConfigVal:     validPath2,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both %s and %s were specified. Using %s environment variable value %s for this session", configEnvVarName, configYamlEnvVarName, configEnvVarName, validPath2),
				fmt.Sprintf("Set config to %v", validPath2),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to [%v]", validPath1),
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
			},
		},
		{
			name:                "env SPLUNK_CONFIG_YAML used when flag --config and env SPLUNK_CONFIG not specified",
			configFlagVals:      []string{},
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
		{
			name:                "Flag --config precedences other envvars, works with multiple values",
			configFlagVals:      []string{validPath1, validPath2},
			splunkConfigVal:     "",
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both environment variable %s and flag '--config' were specified. Using the flag values and ignoring the environment variable in this session", configYamlEnvVarName),
				fmt.Sprintf("Set config to [%v,%v]", validPath1, validPath2),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Using environment variable %s for configuration", configYamlEnvVarName),
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
				if len(test.configFlagVals) != 0 {
					os.Args = []string{"otelcol"}
					for _, configVal := range test.configFlagVals {
						os.Args = append(os.Args, "--config="+configVal)
					}
				}

				os.Setenv(configEnvVarName, test.splunkConfigVal)
				os.Setenv(configYamlEnvVarName, test.splunkConfigYamlVal)

				testInitFlags()
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

func TestRemoveFlag(t *testing.T) {
	args := []string{"--aaa", "--bbb", "--ccc"}
	removeFlag(&args, "--bbb")
	assert.Equal(t, []string{"--aaa", "--ccc"}, args)
	removeFlag(&args, "--ccc")
	assert.Equal(t, []string{"--aaa"}, args)
	removeFlag(&args, "--aaa")
	assert.Nil(t, args)
}
