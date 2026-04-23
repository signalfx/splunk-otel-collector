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

package configsource

import (
	"context"
	"strings"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	_ Hook = (*TelemetryHook)(nil)

	// globalHook instance needed by the extension
	globalHook *TelemetryHook
)

const (
	meterName                    = "github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
	configSourceUsageMetricName  = "otelcol_splunk_configsource_usage"
	configSourceTypeAttributeKey = "config_source_type"
)

// TelemetryHook tracks the usage of custom config sources and exposes them as OpenTelemetry metrics.
type TelemetryHook struct {
	usedSources       map[string]bool
	telemetrySettings *component.TelemetrySettings
	mutex             sync.RWMutex
}

// logger returns the logger from TelemetrySettings once injected, or a no-op logger before that.
func (t *TelemetryHook) logger() *zap.Logger {
	if t.telemetrySettings != nil && t.telemetrySettings.Logger != nil {
		return t.telemetrySettings.Logger
	}
	return zap.NewNop()
}

// NewTelemetryHook creates a new TelemetryHook and registers it as the global instance
// so the configsourcetelemetryextension can inject TelemetrySettings at service startup.
func NewTelemetryHook() *TelemetryHook {
	hook := &TelemetryHook{
		usedSources: make(map[string]bool),
	}
	SetGlobalHook(hook)
	return hook
}

// SetGlobalHook sets the global hook instance.
func SetGlobalHook(hook *TelemetryHook) {
	globalHook = hook
}

// GetGlobalHook returns the global hook instance, or nil if not set.
func GetGlobalHook() *TelemetryHook {
	return globalHook
}

// registerMetrics registers the config source usage metrics using the service's MeterProvider
// injected via SetTelemetrySettings.
func (t *TelemetryHook) registerMetrics() error {
	if t.telemetrySettings == nil || t.telemetrySettings.MeterProvider == nil {
		t.logger().Debug("TelemetrySettings not yet available, skipping metric registration until extension injects it")
		return nil
	}

	meter := t.telemetrySettings.MeterProvider.Meter(meterName)

	// Register an observable gauge that reports which config sources are in use
	var err error
	_, err = meter.Int64ObservableGauge(
		configSourceUsageMetricName,
		metric.WithDescription("Indicates whether a custom config source is in use (1 = in use)"),
		metric.WithUnit("{usage}"),
		metric.WithInt64Callback(t.observeConfigSourceUsage),
	)
	if err != nil {
		t.logger().Error("Failed to register config source usage metric", zap.Error(err))
		return err
	}

	t.logger().Info("Config source telemetry metrics registered successfully")
	return nil
}

// SetTelemetrySettings injects the service component.TelemetrySettings into the hook.
// Called by the configsourcetelemetryextension on Start().
func (t *TelemetryHook) SetTelemetrySettings(settings component.TelemetrySettings) {
	t.mutex.Lock()
	t.telemetrySettings = &settings
	t.mutex.Unlock()

	if err := t.registerMetrics(); err != nil {
		t.logger().Warn("Failed to register config source telemetry metrics", zap.Error(err))
	}
}

// OnNew implements Hook, NoOp.
func (t *TelemetryHook) OnNew() {}

// OnRetrieve is called when a config source retrieves configuration.
// It keeps track of used custom config sources.
func (t *TelemetryHook) OnRetrieve(scheme string, retrieved map[string]any) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if configSourcesMap, ok := retrieved["config_sources"].(map[string]any); ok {
		t.logger().Debug("TelemetryHook: config_sources found in retrieved config", zap.String("scheme", scheme))

		for key := range configSourcesMap {
			// Config source keys follow the component-ID format "type" or "type/name".
			csType := key
			if idx := strings.IndexByte(key, '/'); idx >= 0 {
				csType = key[:idx]
			}

			if isCustomConfigSource(csType) {
				t.logger().Info("TelemetryHook: custom config source detected",
					zap.String("type", csType),
					zap.String("name", key),
					zap.String("scheme", scheme))
				t.usedSources[csType] = true
			}
		}
	}
}

// OnShutdown implements Hook and clears tracked config sources.
func (t *TelemetryHook) OnShutdown() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.usedSources = make(map[string]bool)
}

// observeConfigSourceUsage is the callback function for the observable gauge.
// Only config sources that are actively in use are reported (value = 1).
func (t *TelemetryHook) observeConfigSourceUsage(_ context.Context, observer metric.Int64Observer) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for csType, inUse := range t.usedSources {
		if inUse {
			observer.Observe(1, metric.WithAttributes(
				attribute.String(configSourceTypeAttributeKey, csType),
			))
		}
	}

	return nil
}

// customConfigSources is the fixed set of config source types provided by this distribution.
var customConfigSources = map[string]struct{}{
	"env":       {},
	"include":   {},
	"vault":     {},
	"zookeeper": {},
	"etcd2":     {},
}

// isCustomConfigSource reports whether csType is one of the custom config sources
// provided by this distribution.
func isCustomConfigSource(csType string) bool {
	_, ok := customConfigSources[csType]
	return ok
}

// GetUsedSources returns a copy of the currently used config sources; used for unit test.
func (t *TelemetryHook) GetUsedSources() map[string]bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	result := make(map[string]bool)
	for k, v := range t.usedSources {
		result[k] = v
	}
	return result
}

// GetTelemetrySettings returns the current TelemetrySettings; used for unit test.
func (t *TelemetryHook) GetTelemetrySettings() *component.TelemetrySettings {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.telemetrySettings
}
