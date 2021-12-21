// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collectorconfig

import (
	"bytes"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmapprovider"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
	"github.com/signalfx/splunk-otel-collector/internal/configsources"
)

func NewConfigMapProvider(
	info component.BuildInfo,
	hasNoConvertConfigFlag bool,
	configFlagPath string,
	configYamlFromEnv string,
	properties []string,
) configmapprovider.Provider {
	provider := newExpandProvider(info, configFlagPath, configYamlFromEnv, properties, hasNoConvertConfigFlag)
	if !hasNoConvertConfigFlag {
		provider = configconverter.Provider(
			provider,
			configconverter.RemoveBallastKey,
			configconverter.MoveOTLPInsecureKey,
			configconverter.MoveHecTLS,
			configconverter.RenameK8sTagger,
		)
	}
	return provider
}

func newExpandProvider(
	info component.BuildInfo,
	configPath string,
	configYamlFromEnv string,
	properties []string,
	hasNoConvertConfigFlag bool,
) configmapprovider.Provider {
	provider := newBaseProvider(configPath, configYamlFromEnv, properties)
	if !hasNoConvertConfigFlag {
		// we have to convert any $${foo:bar} *before* the expand provider runs,
		// so we do it here rather than in NewConfigMapProvider
		provider = configconverter.Provider(
			provider,
			configconverter.ReplaceDollarDollar,
		)
	}
	return configmapprovider.NewExpand(configprovider.NewConfigSourceParserProvider(
		provider,
		zap.NewNop(), // The service logger is not available yet, setting it to NoP.
		info,
		configsources.Get()...,
	))
}

func newBaseProvider(configPath string, configYamlFromEnv string, properties []string) configmapprovider.Provider {
	if configPath == "" && configYamlFromEnv != "" {
		return configmapprovider.NewInMemory(bytes.NewBufferString(configYamlFromEnv))
	}
	return configmapprovider.NewMerge(
		configmapprovider.NewFile(configPath),
		configmapprovider.NewProperties(properties),
	)
}
