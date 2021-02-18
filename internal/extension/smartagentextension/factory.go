// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentextension

import (
	"context"
	"os"
	"path/filepath"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/extension/extensionhelper"
)

const (
	typeStr configmodels.Type = "smartagent"
)

func NewFactory() component.ExtensionFactory {
	return extensionhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		createExtension)
}

var bundleDir = func() string {
	out := os.Getenv(constants.BundleDirEnvVar)
	if out == "" {
		exePath, err := os.Executable()
		if err != nil {
			panic("Cannot determine agent executable path, cannot continue")
		}
		out, err = filepath.Abs(filepath.Join(filepath.Dir(exePath), ".."))
		if err != nil {
			panic("Cannot determine absolute path of executable parent dir " + exePath)
		}
	}
	return out
}()

func createDefaultConfig() configmodels.Extension {
	return &Config{
		ExtensionSettings: configmodels.ExtensionSettings{
			TypeVal: typeStr,
			NameVal: string(typeStr),
		},
		BundleDir: bundleDir,
		CollectdConfig: CollectdConfig{
			Timeout:             40,
			ReadThreads:         5,
			WriteThreads:        2,
			WriteQueueLimitHigh: 500000,
			WriteQueueLimitLow:  400000,
			LogLevel:            "notice",
			IntervalSeconds:     10,
			WriteServerIPAddr:   "127.9.8.7",
			WriteServerPort:     0,
			ConfigDir:           "/var/run/signalfx-agent/collectd",
		},
	}
}

func createExtension(
	_ context.Context,
	_ component.ExtensionCreateParams,
	cfg configmodels.Extension,
) (component.ServiceExtension, error) {
	return newSmartAgentConfigExtension(cfg)
}
