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
	"runtime/debug"
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

var (
	configPath           = filepath.Join(".", "testdata", "config.yaml")
	anotherConfigPath    = filepath.Join(".", "testdata", "another-config.yaml")
	localGatewayConfig   = filepath.Join("..", "..", "cmd/otelcol/config/collector/gateway_config.yaml")
	localOTLPLinuxConfig = filepath.Join("..", "..", "cmd/otelcol/config/collector/otlp_config_linux.yaml")
	propertiesPath       = filepath.Join(".", "testdata", "properties.yaml")
)

func TestNewSettingsWithUnknownFlagsNotAcceptable(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"--unknown-flag", "100"})
	require.Error(t, err)
	require.Nil(t, settings)
}

func TestNewSettingsWithVersionFlags(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.False(t, settings.versionFlag)

	settings, err = New([]string{"--version"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.True(t, settings.versionFlag)
	require.Equal(t, []string{"--version", "true"}, settings.ColCoreArgs())

	settings, err = New([]string{"-v"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.True(t, settings.versionFlag)
	require.Equal(t, []string{"--version", "true"}, settings.ColCoreArgs())
}

func TestNewSettingsWithHelpFlags(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)

	settings, err = New([]string{"--help"})
	require.Error(t, err)
	require.Equal(t, flag.ErrHelp, err)
	require.Nil(t, settings)

	settings, err = New([]string{"-h"})
	require.Error(t, err)
	require.Equal(t, flag.ErrHelp, err)
	require.Nil(t, settings)
}

func TestNewSettingsConfMapProviders(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)

	confMapProviderFactories := settings.ConfMapProviderFactories()
	require.Len(t, confMapProviderFactories, 6)

	schemas := make([]string, 0, len(confMapProviderFactories))
	for _, provider := range confMapProviderFactories {
		schemas = append(schemas, provider.Create(confmap.ProviderSettings{}).Scheme())
	}
	require.Contains(t, schemas, settings.discovery.PropertyScheme())
	require.Contains(t, schemas, settings.discovery.ConfigDScheme())
	require.Contains(t, schemas, settings.discovery.DiscoveryModeScheme())
	require.Contains(t, schemas, settings.discovery.PropertiesFileScheme())
}

func TestNewSettingsNoConvertConfig(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{
		"--no-convert-config",
		"--config", configPath,
		"--config", anotherConfigPath,
		"--set", "foo",
		"--set", "splunk.discovery.receiver.receiver-type/name.config.field.one=val.one",
		"--set", "bar",
		"--set=baz",
		"--set", "splunk.discovery.receiver.receiver-type/name.config.field.two=val.two",
		"--feature-gates", "foo",
		"--feature-gates", "-bar",
	})
	require.NoError(t, err)

	require.True(t, settings.noConvertConfig)

	require.Equal(t, []string{configPath, anotherConfigPath}, settings.configPaths.value)
	require.Equal(t, []string{"foo", "bar", "baz"}, settings.setProperties)
	require.Equal(t, []string{
		"splunk.discovery.receiver.receiver-type/name.config.field.one=val.one",
		"splunk.discovery.receiver.receiver-type/name.config.field.two=val.two",
	}, settings.discoveryProperties)

	require.Equal(t, []string{
		configPath, anotherConfigPath,
		"splunk.property:splunk.discovery.receiver.receiver-type/name.config.field.one=val.one",
		"splunk.property:splunk.discovery.receiver.receiver-type/name.config.field.two=val.two",
	}, settings.ResolverURIs())
	require.Equal(t, 2, len(settings.ConfMapConverterFactories()))
	require.Equal(t, []string{"--feature-gates", "foo", "--feature-gates", "-bar"}, settings.ColCoreArgs())
}

func TestNewSettingsConvertConfig(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{
		"--config", configPath,
		"--config", anotherConfigPath,
		"--set", "foo",
		"--set=bar",
		"--set", "baz",
		"--feature-gates", "foo",
		"--feature-gates", "-bar",
	})
	require.NoError(t, err)

	require.False(t, settings.versionFlag)
	require.False(t, settings.noConvertConfig)

	require.Equal(t, []string{configPath, anotherConfigPath}, settings.configPaths.value)
	require.Equal(t, []string{"foo", "bar", "baz"}, settings.setProperties)
	require.Equal(t, []string(nil), settings.discoveryProperties)

	require.Equal(t, []string{configPath, anotherConfigPath}, settings.ResolverURIs())
	require.Equal(t, 6, len(settings.ConfMapConverterFactories()))
	require.Equal(t, []string{"--feature-gates", "foo", "--feature-gates", "-bar"}, settings.ColCoreArgs())
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

func TestNewSettingsWithValidate(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"validate"})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, []string{"validate"}, settings.ColCoreArgs())
}

func TestCheckRuntimeParams_Default(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "460", os.Getenv(MemLimitMiBEnvVar))
	require.Equal(t, "0.0.0.0", os.Getenv(ListenInterfaceEnvVar))
}

func TestCheckRuntimeParams_MemTotalEnv(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "1000"))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "900", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_ListenInterface(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ListenInterfaceEnvVar, "1.2.3.4"))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "1.2.3.4", os.Getenv(ListenInterfaceEnvVar))
}

func TestCheckRuntimeParams_MemTotalEnvs(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "180", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_LimitEnvs(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemLimitMiBEnvVar, "250"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)

	require.Equal(t, "250", os.Getenv(MemLimitMiBEnvVar))
}

func TestSetDefaultEnvVarsOnlySetsURLsWithRealmSet(t *testing.T) {
	t.Cleanup(clearEnv(t))
	envVars := []string{"SPLUNK_API_URL", "SPLUNK_INGEST_URL", "SPLUNK_TRACE_URL", "SPLUNK_HEC_URL", "SPLUNK_HEC_TOKEN"}
	for _, v := range envVars {
		require.NoError(t, setDefaultEnvVars(nil))
		_, ok := os.LookupEnv(v)
		require.False(t, ok, fmt.Sprintf("Expected %q unset given SPLUNK_REALM is unset", v))
	}
}

func TestSetDefaultEnvVarsOnlySetsHECTokenWithTokenSet(t *testing.T) {
	t.Cleanup(clearEnv(t))
	require.NoError(t, setDefaultEnvVars(nil))
	_, ok := os.LookupEnv("SPLUNK_HEC_TOKEN")
	require.False(t, ok, "Expected SPLUNK_HEC_TOKEN unset given SPLUNK_ACCESS_TOKEN is unset")
}

func TestSetDefaultEnvVarsSetsURLsFromRealm(t *testing.T) {
	t.Cleanup(clearEnv(t))

	realm := "us1"
	os.Setenv("SPLUNK_REALM", realm)
	set := newSettings()
	require.NoError(t, setDefaultEnvVars(set))
	assert.Equal(t, 1, len(set.envVarWarnings))
	assert.Contains(t, set.envVarWarnings["SPLUNK_TRACE_URL"], `"SPLUNK_TRACE_URL" environment variable is deprecated`)

	expectedEnvVars := [][]string{
		{"SPLUNK_API_URL", fmt.Sprintf("https://api.%s.signalfx.com", realm)},
		{"SPLUNK_INGEST_URL", fmt.Sprintf("https://ingest.%s.signalfx.com", realm)},
		{"SPLUNK_TRACE_URL", fmt.Sprintf("https://ingest.%s.signalfx.com/v2/trace", realm)},
		{"SPLUNK_HEC_URL", fmt.Sprintf("https://ingest.%s.signalfx.com/v1/log", realm)},
	}
	for _, v := range expectedEnvVars {
		val, ok := os.LookupEnv(v[0])
		assert.True(t, ok, v[0])
		assert.Equal(t, val, v[1])
	}
}

func TestNoWarningsIfTraceURLSetExplicitly(t *testing.T) {
	t.Cleanup(clearEnv(t))

	os.Setenv("SPLUNK_REALM", "us1")
	os.Setenv("SPLUNK_TRACE_URL", "https://ingest.trace-realm.signalfx.com/v2/trace")
	set := newSettings()
	require.NoError(t, setDefaultEnvVars(set))
	assert.Equal(t, 0, len(set.envVarWarnings))

	val, ok := os.LookupEnv("SPLUNK_INGEST_URL")
	assert.True(t, ok)
	assert.Equal(t, "https://ingest.us1.signalfx.com", val)

	val, ok = os.LookupEnv("SPLUNK_TRACE_URL")
	assert.True(t, ok)
	assert.Equal(t, "https://ingest.trace-realm.signalfx.com/v2/trace", val)
}

func TestSetDefaultEnvVarsSetsHECTokenFromAccessTokenEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))

	token := "1234"
	os.Setenv("SPLUNK_ACCESS_TOKEN", token)
	require.NoError(t, setDefaultEnvVars(nil))

	val, ok := os.LookupEnv("SPLUNK_HEC_TOKEN")
	assert.True(t, ok)
	assert.Equal(t, token, val)
}

func TestSetDefaultEnvVarsSetsTraceURLFromIngestURL(t *testing.T) {
	t.Cleanup(clearEnv(t))

	os.Setenv("SPLUNK_INGEST_URL", "https://ingest.fake-realm.signalfx.com/")
	set := newSettings()
	require.NoError(t, setDefaultEnvVars(set))
	assert.Equal(t, 1, len(set.envVarWarnings))
	assert.Contains(t, set.envVarWarnings["SPLUNK_TRACE_URL"], `"SPLUNK_TRACE_URL" environment variable is deprecated`)

	val, ok := os.LookupEnv("SPLUNK_TRACE_URL")
	assert.True(t, ok)
	assert.Equal(t, "https://ingest.fake-realm.signalfx.com/v2/trace", val)
}

func TestSetDefaultEnvVarsRespectsSetEnvVars(t *testing.T) {
	t.Cleanup(clearEnv(t))
	envVars := []string{"SPLUNK_API_URL", "SPLUNK_INGEST_URL", "SPLUNK_TRACE_URL", "SPLUNK_HEC_URL", "SPLUNK_HEC_TOKEN", "SPLUNK_LISTEN_INTERFACE"}

	someValue := "some.value"
	for _, v := range envVars {
		os.Setenv(v, someValue)
		require.NoError(t, setDefaultEnvVars(newSettings()))
		val, ok := os.LookupEnv(v)
		assert.True(t, ok, v[0])
		assert.Equal(t, someValue, val)
	}

	for _, v := range envVars {
		os.Setenv(v, "")
		require.NoError(t, setDefaultEnvVars(newSettings()))
		val, ok := os.LookupEnv(v)
		assert.True(t, ok, v[0])
		assert.Empty(t, val)
	}
}

func TestSetDefaultEnvVarsSetsInterfaceFromConfigOption(t *testing.T) {
	for _, tc := range []struct{ config, expectedIP string }{
		{"/etc/otel/collector/agent_config.yaml", "127.0.0.1"},
		{"file:/etc/otel/collector/agent_config.yaml", "127.0.0.1"},
		{"/etc/otel/collector/gateway_config.yaml", "0.0.0.0"},
		{"file:/etc/otel/collector/gateway_config.yaml", "0.0.0.0"},
		{"some-other-config.yaml", "0.0.0.0"},
		{"file:some-other-config.yaml", "0.0.0.0"},
	} {
		t.Run(fmt.Sprintf("%v->%v", tc.config, tc.expectedIP), func(t *testing.T) {
			t.Cleanup(clearEnv(t))
			os.Setenv("SPLUNK_REALM", "noop")
			os.Setenv("SPLUNK_ACCESS_TOKEN", "noop")
			s, err := parseArgs([]string{"--config", tc.config})
			require.NoError(t, err)
			require.NoError(t, setDefaultEnvVars(s))

			val, ok := os.LookupEnv("SPLUNK_LISTEN_INTERFACE")
			assert.True(t, ok)
			assert.Equal(t, tc.expectedIP, val)
		})
	}
}

func TestSetDefaultFeatureGatesRespectsOverrides(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	for _, args := range [][]string{
		{
			"--feature-gates", "some-gate", "--feature-gates", "telemetry.useOtelForInternalMetrics", "--feature-gates",
			"another-gate",
		},
		{
			"--feature-gates", "some-gate", "--feature-gates", "+telemetry.useOtelForInternalMetrics",
			"--feature-gates", "another-gate",
		},
		{
			"--feature-gates", "some-gate", "--feature-gates", "-telemetry.useOtelForInternalMetrics",
			"--feature-gates", "another-gate",
		},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			settings, err := New(args)
			require.NoError(t, err)
			require.Equal(t, args, settings.ColCoreArgs())
		})
	}
}

func TestSetSoftMemLimitWithoutGoMemLimitEnvVar(t *testing.T) {
	// if GOLIMIT is not set, we expect soft limit to be 90% of the total memory env var or 90% of default total memory  512 Mib.
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))
	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, int64(188743680), debug.SetMemoryLimit(100))

	t.Cleanup(setRequiredEnvVars(t))
	settings, err = New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, int64(482344960), debug.SetMemoryLimit(-1))
}

func TestUseConfigPathsFromEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))
	os.Setenv(ConfigEnvVar, localGatewayConfig)

	settings, err := New([]string{})
	require.NoError(t, err)
	configPaths := settings.configPaths.value
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
  debug:
    verbosity: detailed
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [debug]`

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

func TestEnablingConfigD(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{"--config", configPath})
	require.NoError(t, err)
	require.False(t, settings.configD)
	require.Nil(t, settings.configDir.value)

	settings, err = New([]string{"--configd", "--config", configPath})
	require.NoError(t, err)
	require.True(t, settings.configD)
	require.Nil(t, settings.configDir.value)
	require.Equal(t, "/etc/otel/collector/config.d", getConfigDir(settings))
}

func TestConfigDirFromArgs(t *testing.T) {
	t.Cleanup(clearEnv(t))
	for _, args := range [][]string{
		{"--config-dir", "/from/args", "--config", configPath},
		{"--config-dir=/from/args", "--config=" + configPath},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			settings, err := New(args)
			require.NoError(t, err)
			require.False(t, settings.configD)
			require.NotNil(t, settings.configDir.value)
			require.Equal(t, "/from/args", settings.configDir.String())
			require.Equal(t, "/from/args", getConfigDir(settings))
		})
	}
}

func TestConfigDirFromEnvVar(t *testing.T) {
	t.Cleanup(clearEnv(t))
	os.Setenv("SPLUNK_CONFIG_DIR", "/from/env/var")
	settings, err := New([]string{"--config", configPath})
	require.NoError(t, err)
	require.Nil(t, settings.configDir.value)
	require.Equal(t, "/from/env/var", getConfigDir(settings))
}

func TestConfigArgFileURIForm(t *testing.T) {
	t.Cleanup(clearEnv(t))
	uriPath := fmt.Sprintf("file:%s", configPath)
	settings, err := New([]string{"--config", uriPath})
	require.NoError(t, err)
	require.Equal(t, []string{uriPath}, settings.configPaths.value)
	require.Equal(t, settings.configPaths.value, settings.ResolverURIs())
}

func TestConfigArgEnvURIForm(t *testing.T) {
	t.Cleanup(clearEnv(t))
	settings, err := New([]string{"--config", "env:SOME_ENV_VAR"})
	require.NoError(t, err)
	require.Equal(t, []string{"env:SOME_ENV_VAR"}, settings.configPaths.value)
	require.Equal(t, settings.configPaths.value, settings.ResolverURIs())
}

func TestCheckRuntimeParams_MemTotal(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemTotalEnvVar, "200"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "180", os.Getenv(MemLimitMiBEnvVar))
}

func TestCheckRuntimeParams_Limit(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	require.NoError(t, os.Setenv(ConfigEnvVar, localGatewayConfig))
	require.NoError(t, os.Setenv(MemLimitMiBEnvVar, "337"))

	settings, err := New([]string{})
	require.NoError(t, err)
	require.NotNil(t, settings)
	require.Equal(t, "337", os.Getenv(MemLimitMiBEnvVar))
}

func TestDefaultDiscoveryConfigDir(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"--discovery"})
	require.NoError(t, err)
	require.True(t, settings.discoveryMode)
	require.False(t, settings.configD)
	require.Nil(t, settings.discoveryPropertiesFile.value)

	require.Equal(t, []string{
		localGatewayConfig,
		"splunk.discovery:/etc/otel/collector/config.d",
	}, settings.ResolverURIs())
}

func TestInheritedDiscoveryConfigDir(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"--discovery", "--config-dir", "/some/config.d"})
	require.NoError(t, err)
	require.True(t, settings.discoveryMode)
	require.False(t, settings.configD)

	require.Equal(t, []string{
		localGatewayConfig,
		"splunk.discovery:/some/config.d",
	}, settings.ResolverURIs())
}

func TestInheritedDiscoveryConfigDirWithConfigD(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{
		"--discovery", "--config-dir", "/some/config.d", "--configd", "--discovery-properties", propertiesPath,
	})
	require.NoError(t, err)
	require.True(t, settings.discoveryMode)
	require.True(t, settings.configD)

	require.NotNil(t, settings.discoveryPropertiesFile.value)
	require.Equal(t, propertiesPath, settings.discoveryPropertiesFile.String())

	require.Equal(t, []string{
		localGatewayConfig,
		fmt.Sprintf("splunk.properties:%s", propertiesPath),
		"splunk.configd:/some/config.d",
		"splunk.discovery:/some/config.d",
	}, settings.ResolverURIs())
}

func TestDiscoveryPropertiesMustExist(t *testing.T) {
	t.Cleanup(setRequiredEnvVars(t))
	settings, err := New([]string{"--discovery", "--discovery-properties", "notafile"})
	require.ErrorContains(t, err, "unable to find discovery properties file notafile. Ensure flag '--discovery-properties' is set correctly:")
	require.Nil(t, settings)
}

func TestSetDefaultEnvVarsFileStorageExtension(t *testing.T) {
	t.Cleanup(clearEnv(t))
	_, ok := os.LookupEnv("SPLUNK_FILE_STORAGE_EXTENSION_PATH")
	require.False(t, ok)
	require.NoError(t, setDefaultEnvVars(nil))
	path, ok := os.LookupEnv("SPLUNK_FILE_STORAGE_EXTENSION_PATH")
	require.True(t, ok, "Expected SPLUNK_FILE_STORAGE_EXTENSION_PATH set by default")
	require.Equal(t, path, "/var/lib/otelcol/filelogs")
}

func TestSetNonDefaultEnvVarsFileStorageExtension(t *testing.T) {
	t.Cleanup(clearEnv(t))
	nonDefaultPath := "/var/non/default/path"
	err := os.Setenv("SPLUNK_FILE_STORAGE_EXTENSION_PATH", nonDefaultPath)
	require.NoError(t, err)
	path, ok := os.LookupEnv("SPLUNK_FILE_STORAGE_EXTENSION_PATH")
	require.True(t, ok, "Expected SPLUNK_FILE_STORAGE_EXTENSION_PATH set by default")
	require.Equal(t, path, nonDefaultPath)
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
