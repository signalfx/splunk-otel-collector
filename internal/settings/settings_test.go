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

package settings

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/overwritepropertiesconverter"

	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
)

var (
	configPath           = filepath.Join(".", "testdata", "config.yaml")
	anotherConfigPath    = filepath.Join(".", "testdata", "another-config.yaml")
	localGatewayConfig   = filepath.Join("..", "..", "cmd/otelcol/config/collector/gateway_config.yaml")
	localOTLPLinuxConfig = filepath.Join("..", "..", "cmd/otelcol/config/collector/otlp_config_linux.yaml")
)

func TestNewSettingsWithUnknownFlagsAcceptable(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"--unknown-flag", "100"})
	require.NoError(t, err)
	require.NotNil(t, settings)
}

func TestNewSettingsWithVersionFlags(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f := settingsToFlags(t, settings)
	require.False(t, f.versionFlag)

	settings, err = New([]string{"--version"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f = settingsToFlags(t, settings)
	require.True(t, f.versionFlag)

	settings, err = New([]string{"-v"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f = settingsToFlags(t, settings)
	require.True(t, f.versionFlag)
}

func TestNewSettingsWithHelpFlags(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f := settingsToFlags(t, settings)
	require.False(t, f.helpFlag)

	settings, err = New([]string{"--help"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f = settingsToFlags(t, settings)
	require.True(t, f.helpFlag)

	settings, err = New([]string{"-h"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	f = settingsToFlags(t, settings)
	require.True(t, f.helpFlag)
}

func TestNewSettingsNoConvertConfig(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{
		"--no-convert-config",
		"--config", configPath,
		"--config", anotherConfigPath,
		"--mem-ballast-size-mib", "100",
		"--set", "foo",
		"--set", "bar",
		"--set", "baz",
		"--feature-gates", "foo",
		"--feature-gates", "-bar",
	})
	require.NoError(t, err)

	f := settingsToFlags(t, settings)
	require.True(t, f.noConvertConfig)

	require.Equal(t, 100, f.memBallastSizeMiB)
	require.Equal(t, []string{configPath, anotherConfigPath}, f.configPaths.value)
	require.Equal(t, []string{"foo", "bar", "baz"}, f.setProperties.value)

	require.Equal(t, []string{configPath, anotherConfigPath}, settings.ResolverURIs())
	require.Equal(t, []confmap.Converter{
		// nolint: staticcheck
		overwritepropertiesconverter.New(f.setProperties.value), // support until there's an actual replacement
	}, settings.ConfMapConverters())
	require.Equal(t, []string{
		"--config", configPath,
		"--config", anotherConfigPath,
		"--mem-ballast-size-mib", "100",
		"--set", "foo", "--set", "bar", "--set", "baz",
		"--feature-gates", "foo", "--feature-gates", "-bar",
	}, settings.ServiceArgs())
}

func TestNewSettingsConvertConfig(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{
		"--config", configPath,
		"--config", anotherConfigPath,
		"--mem-ballast-size-mib", "100",
		"--set", "foo",
		"--set", "bar",
		"--set", "baz",
		"--feature-gates", "foo",
		"--feature-gates", "-bar",
	})
	require.NoError(t, err)

	f := settingsToFlags(t, settings)
	require.False(t, f.helpFlag)
	require.False(t, f.versionFlag)
	require.False(t, f.noConvertConfig)

	require.Equal(t, 100, f.memBallastSizeMiB)
	require.Equal(t, []string{configPath, anotherConfigPath}, f.configPaths.value)
	require.Equal(t, []string{"foo", "bar", "baz"}, f.setProperties.value)

	require.Equal(t, []string{configPath, anotherConfigPath}, settings.ResolverURIs())
	require.Equal(t, []confmap.Converter{
		// nolint: staticcheck
		overwritepropertiesconverter.New(f.setProperties.value), // support until there's an actual replacement
		configconverter.RemoveBallastKey{},
		configconverter.MoveOTLPInsecureKey{},
		configconverter.MoveHecTLS{},
		configconverter.RenameK8sTagger{},
	}, settings.ConfMapConverters())
	require.Equal(t, []string{
		"--config", configPath,
		"--config", anotherConfigPath,
		"--mem-ballast-size-mib", "100",
		"--set", "foo", "--set", "bar", "--set", "baz",
		"--feature-gates", "foo", "--feature-gates", "-bar",
	}, settings.ServiceArgs())
}

func TestSplunkConfigYamlUtilizedInResolverURIs(t *testing.T) {
	t.Cleanup(clearEnv(t))
	require.NoError(t, os.Setenv(ConfigYamlEnvVar, "some: yaml"))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, []string{"env:SPLUNK_CONFIG_YAML"}, settings.ResolverURIs())
}

func TestSplunkConfigYamlNotUtilizedInResolverURIsWithConfigEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))
	require.NoError(t, os.Setenv(ConfigYamlEnvVar, "some: yaml"))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, []string{localGatewayConfig}, settings.ResolverURIs())
}

func TestCheckRuntimeParams_Default(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "168", os.Getenv(BallastEnvVar))
	require.Equal(t, "460", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_MemTotalEnv(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "1000"))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "330", os.Getenv(BallastEnvVar))
	require.Equal(t, "900", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_MemTotalAndBallastEnvs(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))
	require.NoError(t, os.Setenv(BallastEnvVar, "90"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "90", os.Getenv(BallastEnvVar))
	require.Equal(t, "180", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_LimitAndBallastEnvs(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemLimitMiBEnvVar, "250"))
	require.NoError(t, os.Setenv(BallastEnvVar, "120"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "120", os.Getenv(BallastEnvVar))
	require.Equal(t, "250", os.Getenv(MemLimitMiBEnvVar))
}

func TestSetDefaultEnvVars(t *testing.T) {
	t.Cleanup(clearEnv(t))

	testArgs := []string{"SPLUNK_API_URL", "SPLUNK_INGEST_URL", "SPLUNK_TRACE_URL", "SPLUNK_HEC_URL", "SPLUNK_HEC_TOKEN"}
	for _, v := range testArgs {
		setDefaultEnvVars()
		_, ok := os.LookupEnv(v)
		require.False(t, ok, fmt.Sprintf("Expected %q unset given SPLUNK_ACCESS_TOKEN or SPLUNK_TOKEN is unset", v))
	}

	realm := "us1"
	token := "1234"
	valTest := "test"
	valEmpty := ""

	os.Setenv("SPLUNK_REALM", realm)
	os.Setenv("SPLUNK_ACCESS_TOKEN", token)
	setDefaultEnvVars()
	testArgs2 := [][]string{
		{"SPLUNK_API_URL", fmt.Sprintf("https://api.%s.signalfx.com", realm)},
		{"SPLUNK_INGEST_URL", fmt.Sprintf("https://ingest.%s.signalfx.com", realm)},
		{"SPLUNK_TRACE_URL", fmt.Sprintf("https://ingest.%s.signalfx.com/v2/trace", realm)},
		{"SPLUNK_HEC_URL", fmt.Sprintf("https://ingest.%s.signalfx.com/v1/log", realm)},
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
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))
	require.NoError(t, os.Setenv(MemLimitMiBEnvVar, "150"))
	require.NoError(t, os.Setenv(BallastEnvVar, "50"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "50", os.Getenv(BallastEnvVar))
	require.Equal(t, "150", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_MemTotalEnvAndBallastFlag(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))

	settings, err := New([]string{"--mem-ballast-size-mib=90"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "90", os.Getenv(BallastEnvVar))
	require.Equal(t, "180", os.Getenv(MemLimitMiBEnvVar))
}

func TestUseConfigPathsFromEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))
	os.Setenv(ConfigEnvVar, localGatewayConfig)

	settings, err := New([]string{})
	require.NoError(t, err)
	configPaths := settingsToFlags(t, settings).configPaths.value
	require.Equal(t, []string{localGatewayConfig}, configPaths)
	require.Equal(t, []string{localGatewayConfig}, settings.ResolverURIs())
}

func TestConfigPrecedence(t *testing.T) {
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
			configFlagVals:      []string{localGatewayConfig},
			splunkConfigVal:     localOTLPLinuxConfig,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both environment variable SPLUNK_CONFIG and flag '--config' were specified. Using the flag values and ignoring the environment variable value %s in this session", localOTLPLinuxConfig),
				fmt.Sprintf("Set config to [%v]", localGatewayConfig),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to [%v]", localOTLPLinuxConfig),
				fmt.Sprintf("Using environment variable %s for configuration", ConfigYamlEnvVar),
			},
		},
		{
			name:                "env SPLUNK_CONFIG precedences SPLUNK_CONFIG_YAML",
			configFlagVals:      []string{},
			splunkConfigVal:     localOTLPLinuxConfig,
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both %s and %s were specified. Using %s environment variable value %s for this session", ConfigEnvVar, ConfigYamlEnvVar, ConfigEnvVar, localOTLPLinuxConfig),
				fmt.Sprintf("Set config to %v", localOTLPLinuxConfig),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to [%v]", localGatewayConfig),
				fmt.Sprintf("Using environment variable %s for configuration", ConfigYamlEnvVar),
			},
		},
		{
			name:                "env SPLUNK_CONFIG_YAML used when flag --config and env SPLUNK_CONFIG not specified",
			configFlagVals:      []string{},
			splunkConfigVal:     "",
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Using environment variable %s for configuration", ConfigYamlEnvVar),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Set config to %v", localGatewayConfig),
				fmt.Sprintf("Set config to %v", localOTLPLinuxConfig),
			},
		},
		{
			name:                "Flag --config precedences other envvars, works with multiple values",
			configFlagVals:      []string{localGatewayConfig, localOTLPLinuxConfig},
			splunkConfigVal:     "",
			splunkConfigYamlVal: validConfig,
			expectedLogs: []string{
				fmt.Sprintf("Both environment variable %s and flag '--config' were specified. Using the flag values and ignoring the environment variable in this session", ConfigYamlEnvVar),
				fmt.Sprintf("Set config to [%v,%v]", localGatewayConfig, localOTLPLinuxConfig),
			},
			unexpectedLogs: []string{
				fmt.Sprintf("Using environment variable %s for configuration", ConfigYamlEnvVar),
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
					os.Unsetenv(ConfigEnvVar)
					os.Unsetenv(ConfigYamlEnvVar)
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

				os.Setenv(ConfigEnvVar, test.splunkConfigVal)
				os.Setenv(ConfigYamlEnvVar, test.splunkConfigYamlVal)

				set, err := New(os.Args[1:])
				require.NoError(t, err)
				require.NotNil(t, set)

				actualLogs := actualLogsBuf.String()

				for _, expectedLog := range test.expectedLogs {
					require.Contains(t, actualLogs, expectedLog)
				}
				for _, unexpectedLog := range test.unexpectedLogs {
					require.NotContains(t, actualLogs, unexpectedLog)
				}
			}()
		})
	}
}

func TestRemoveFlag(t *testing.T) {
	args := []string{"--aaa", "--bbb", "--ccc"}
	removeFlag(&args, "--bbb")
	require.Equal(t, []string{"--aaa", "--ccc"}, args)
	removeFlag(&args, "--ccc")
	require.Equal(t, []string{"--aaa"}, args)
	removeFlag(&args, "--aaa")
	require.Empty(t, args)
}

func TestEnablingConfigD(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{"--config", configPath})
	require.NoError(t, err)
	f := settingsToFlags(t, settings)
	require.False(t, f.configD)
	require.Nil(t, f.configDir.value)

	settings, err = New([]string{"--configd", "--config", configPath})
	require.NoError(t, err)
	f = settingsToFlags(t, settings)
	require.True(t, f.configD)
	require.Nil(t, f.configDir.value)
	require.Equal(t, "/etc/otel/collector/config.d", getConfigDir(f))
}

func TestConfigDirFromArgs(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{"--config-dir", "/from/args", "--config", configPath})
	require.NoError(t, err)
	f := settingsToFlags(t, settings)
	require.False(t, f.configD)
	require.NotNil(t, f.configDir.value)
	require.Equal(t, "/from/args", f.configDir.String())
	require.Equal(t, "/from/args", getConfigDir(f))
}

func TestConfigDirFromEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))
	os.Setenv("SPLUNK_CONFIG_DIR", "/from/env/var")
	settings, err := New([]string{"--config", configPath})
	require.NoError(t, err)
	f := settingsToFlags(t, settings)
	require.Nil(t, f.configDir.value)
	require.Equal(t, "/from/env/var", getConfigDir(f))
}

// to satisfy Settings generation
func setRequiredEnvVars(t *testing.T) func() {
	cleanup := clearEnv(t)
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	return cleanup
}

func clearEnv(t *testing.T) func() {
	toRestore := map[string]string{}
	for _, ev := range os.Environ() {
		i := strings.Index(ev, "=")
		if i < 0 {
			continue
		}
		toRestore[ev[:i]] = os.Getenv(ev[:i])
	}
	os.Clearenv()

	return func() {
		os.Clearenv()
		for k, v := range toRestore {
			require.NoError(t, os.Setenv(k, v))
		}
	}
}

func settingsToFlags(t testing.TB, settings Settings) *flags {
	require.NotNil(t, settings)
	f, ok := settings.(*flags)
	require.True(t, ok)
	require.NotNil(t, f)
	return f
}
