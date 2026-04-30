// Copyright Splunk, Inc.
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

package configsourcetelemetryextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
)

const (
	TypeStr = "config_source_telemetry"
)

var _ extension.Extension = (*configSourceTelemetryExtension)(nil)

// configSourceTelemetryExtension is an extension that injects TelemetrySettings
// into the config source hook to enable proper metric registration.
type configSourceTelemetryExtension struct {
	telemetrySettings component.TelemetrySettings
}

func (e *configSourceTelemetryExtension) Start(_ context.Context, _ component.Host) error {
	hook := configsource.GetGlobalHook()

	if hook != nil {
		e.telemetrySettings.Logger.Info("Injecting TelemetrySettings into config source hook")
		hook.SetTelemetrySettings(e.telemetrySettings)
	} else {
		e.telemetrySettings.Logger.Warn("Config source telemetry hook not found, metrics may not be available at /metrics endpoint")
	}
	return nil
}

func (e *configSourceTelemetryExtension) Shutdown(_ context.Context) error {
	return nil
}

func newConfigSourceTelemetryExtension(settings extension.Settings) *configSourceTelemetryExtension {
	return &configSourceTelemetryExtension{
		telemetrySettings: settings.TelemetrySettings,
	}
}
