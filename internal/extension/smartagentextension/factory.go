// Copyright OpenTelemetry Authors
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
	"runtime"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/extension/extensionhelper"
)

const (
	typeStr                config.Type = "smartagent"
	defaultIntervalSeconds int         = 10
)

func NewFactory() component.ExtensionFactory {
	return extensionhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		createExtension,
	)
}

var bundleDir = func() string {
	dir := os.Getenv(constants.BundleDirEnvVar)
	if dir == "" {
		if runtime.GOOS == "windows" {
			pfDir := os.Getenv("programfiles")
			if pfDir == "" {
				pfDir = "C:\\Program Files"
			}
			dir = filepath.Join(pfDir, "Splunk", "OpenTelemetry Collector", "agent-bundle")
			if exePath, err := os.Executable(); err == nil {
				if colocatedBundle, err := filepath.Abs(filepath.Join(filepath.Dir(exePath), "agent-bundle")); err == nil {
					if info, err := os.Stat(colocatedBundle); err == nil && info.IsDir() {
						dir = colocatedBundle
					}
				}
			}
		} else {
			dir = "/usr/lib/splunk-otel-collector/agent-bundle"
		}
	}
	return dir
}()

func createDefaultConfig() config.Extension {
	cfg, _ := smartAgentConfigFromSettingsMap(map[string]interface{}{})
	if cfg == nil {
		// We won't truly be using this default in our custom unmarshaler
		// so zero value is adequate
		cfg = &saconfig.Config{}
	}
	cfg.BundleDir = bundleDir
	cfg.Collectd.BundleDir = bundleDir
	cfg.Collectd.ConfigDir = filepath.Join(bundleDir, "run", "collectd")

	return &Config{
		ExtensionSettings: config.NewExtensionSettings(config.NewID(typeStr)),
		Config:            *cfg,
	}
}

func createExtension(
	_ context.Context,
	_ component.ExtensionCreateParams,
	cfg config.Extension,
) (component.Extension, error) {
	return newSmartAgentConfigExtension(cfg.(*Config))
}
