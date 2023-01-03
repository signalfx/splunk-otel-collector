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

package discovery

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecstaskobserver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/consumer"
	otelcolextension "go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pdata/plog"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const durationEnvVar = "SPLUNK_DISCOVERY_DURATION"

// discoverer provides the mechanism for a "preflight" collector service
// that will stand up the observers and discovery receivers based on the .discovery.yaml
// contents of the config dir, acting as a log consumer to determine which
// of the underlying receivers were successfully discovered by the
// discovery receiver from its emitted log records.
type discoverer struct {
	factories           otelcol.Factories
	expandConverter     confmap.Converter
	configs             map[string]*Config
	extensions          map[component.ID]otelcolextension.Extension
	logger              *zap.Logger
	discoveredReceivers map[component.ID]discovery.StatusType
	discoveredConfig    map[component.ID]map[string]any
	discoveredObservers map[component.ID]discovery.StatusType
	info                component.BuildInfo
	duration            time.Duration
	mu                  sync.Mutex
}

func newDiscoverer(logger *zap.Logger) (*discoverer, error) {
	info := component.BuildInfo{
		Command: "discovery",
		Version: version.Version,
	}
	duration := 10 * time.Second
	if d, ok := os.LookupEnv(durationEnvVar); ok {
		if dur, err := time.ParseDuration(d); err != nil {
			logger.Warn("Invalid SPLUNK_DISCOVERY_DURATION. Using default of 10s", zap.String("duration", d))
		} else {
			duration = dur
		}
	}

	factories, err := components.Get()
	if err != nil {
		return (*discoverer)(nil), err
	}

	m := &discoverer{
		logger:              logger,
		info:                info,
		factories:           factories,
		extensions:          map[component.ID]otelcolextension.Extension{},
		configs:             map[string]*Config{},
		duration:            duration,
		mu:                  sync.Mutex{},
		expandConverter:     expandconverter.New(),
		discoveredReceivers: map[component.ID]discovery.StatusType{},
		discoveredConfig:    map[component.ID]map[string]any{},
		discoveredObservers: map[component.ID]discovery.StatusType{},
	}
	return m, nil
}

// discover will create all .discovery.yaml components, start them, wait the configured
// duration, and tear them down before returning the discovery config.
func (d *discoverer) discover(cfg *Config) (map[string]any, error) {
	discoveryReceivers, discoveryObservers, err := d.createDiscoveryReceiversAndObservers(cfg)
	if err != nil {
		d.logger.Error("failed preparing discovery components", zap.Error(err))
		return nil, err
	}

	var cancels []context.CancelFunc

	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	for observerID, observer := range discoveryObservers {
		d.logger.Debug(fmt.Sprintf("starting observer %s", observerID.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancels = append(cancels, cancel)
		if e := observer.Start(ctx, d); e != nil {
			d.logger.Warn(
				fmt.Sprintf("%s startup failed. Won't proceed with %s-based discovery", observerID.String(), observerID.Type()),
				zap.Error(e),
			)
		}
	}

	for receiverID, receiver := range discoveryReceivers {
		d.logger.Debug(fmt.Sprintf("starting receiver %s", receiverID.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancels = append(cancels, cancel)
		if err = receiver.Start(ctx, d); err != nil {
			d.logger.Warn(
				fmt.Sprintf("%s startup failed.", receiverID.String()),
				zap.Error(err),
			)
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "Discovering for next %s...\n", d.duration.String())
	select {
	case <-time.After(d.duration):
	case <-context.Background().Done():
	}
	_, _ = fmt.Fprintf(os.Stderr, "Discovery complete.\n")

	for receiverID, receiver := range discoveryReceivers {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancels = append(cancels, cancel)
		if e := receiver.Shutdown(ctx); e != nil {
			d.logger.Warn(fmt.Sprintf("error shutting down receiver %s", receiverID.String()), zap.Error(e))
		}
	}
	for observerID, observer := range discoveryObservers {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancels = append(cancels, cancel)
		if e := observer.Shutdown(ctx); e != nil {
			d.logger.Warn(fmt.Sprintf("error shutting down observer %s", observerID.String()), zap.Error(e))
		}
	}

	discoveryConfig, err := d.discoveryConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed constructing discovery config: %w", err)
	}
	return discoveryConfig, nil
}

func (d *discoverer) createDiscoveryReceiversAndObservers(cfg *Config) (map[component.ID]otelcolreceiver.Logs, map[component.ID]otelcolextension.Extension, error) {
	discoveryObservers := map[component.ID]otelcolextension.Extension{}
	discoveryReceivers := map[component.ID]otelcolreceiver.Logs{}

	discoveryReceiverFactory := discoveryreceiver.NewFactory()
	for _, observerID := range cfg.observersForDiscoveryMode() {
		observer, err := d.createObserver(observerID, cfg)
		if err != nil {
			return nil, nil, err
		}
		d.extensions[observerID] = observer
		discoveryObservers[observerID] = observer

		discoveryReceiverDefaultConfig := discoveryReceiverFactory.CreateDefaultConfig()
		discoveryReceiverConfig, ok := discoveryReceiverDefaultConfig.(*discoveryreceiver.Config)
		if !ok {
			return nil, nil, fmt.Errorf("failed to coerce to receivercreator.Config")
		}

		discoveryReceiverRaw := map[string]any{}
		receivers := map[string]any{}

		for receiverID, receiver := range cfg.ReceiversToDiscover {
			if ok, err = d.updateReceiverForObserver(receiverID, receiver, observerID); err != nil {
				return nil, nil, err
			} else if !ok {
				continue
			}
			receivers[receiverID.String()] = receiver.Entry.ToStringMap()
		}

		discoveryReceiverRaw["receivers"] = receivers
		discoveryReceiverConfMap := confmap.NewFromStringMap(discoveryReceiverRaw)

		if err = d.expandConverter.Convert(context.Background(), discoveryReceiverConfMap); err != nil {
			return nil, nil, fmt.Errorf("error converting environment variables in receiver config: %w", err)
		}

		if err = component.UnmarshalConfig(discoveryReceiverConfMap, discoveryReceiverConfig); err != nil {
			return nil, nil, fmt.Errorf("failed unmarshaling discovery receiver config: %w", err)
		}

		discoveryReceiverConfig.WatchObservers = append(discoveryReceiverConfig.WatchObservers, observerID)
		discoveryReceiverConfig.EmbedReceiverConfig = true

		discoveryReceiverSettings := d.createReceiverCreateSettings()
		discoveryReceiverSettings.ID = observerID
		var lr otelcolreceiver.Logs
		if lr, err = discoveryReceiverFactory.CreateLogsReceiver(context.Background(), discoveryReceiverSettings, discoveryReceiverDefaultConfig, d); err != nil {
			return nil, nil, fmt.Errorf("failed creating discovery receiver: %w", err)
		}
		discoveryReceivers[component.NewIDWithName(discoveryReceiverFactory.Type(), observerID.String())] = lr
	}

	return discoveryReceivers, discoveryObservers, nil
}

func (d *discoverer) createObserver(observerID component.ID, cfg *Config) (otelcolextension.Extension, error) {
	observerFactory, err := factoryForObserverType(observerID.Type())
	if err != nil {
		return nil, err
	}

	observerConfig := observerFactory.CreateDefaultConfig()
	observerCfgMap := confmap.NewFromStringMap(cfg.DiscoveryObservers[observerID].ToStringMap())
	if err = d.expandConverter.Convert(context.Background(), observerCfgMap); err != nil {
		return nil, fmt.Errorf("error converting environment variables in %q config: %w", observerID.String(), err)
	}

	if err = component.UnmarshalConfig(observerCfgMap, observerConfig); err != nil {
		return nil, fmt.Errorf("failed unmarshaling %s config: %w", observerID.String(), err)
	}

	if ce := d.logger.Check(zap.DebugLevel, "unmarshalled observer config"); ce != nil {
		if c, e := yaml.Marshal(observerConfig); e != nil {
			d.logger.Debug("failed marshaling observer config for logging", zap.Error(e))
		} else {
			ce.Write(zap.String("config", string(c)))
		}
	}

	observerSettings := d.createExtensionCreateSettings(observerID.String())
	observer, err := observerFactory.CreateExtension(context.Background(), observerSettings, observerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed creating %s extension: %w", observerID.String(), err)
	}
	return observer, nil
}

func (d *discoverer) updateReceiverForObserver(receiverID component.ID, receiver ReceiverToDiscoverEntry, observerID component.ID) (bool, error) {
	observerRule, hasRule := receiver.Rule[observerID]
	if !hasRule {
		d.logger.Debug(fmt.Sprintf("disregarding %s without a %s rule", receiverID.String(), observerID.String()))
		return false, nil
	}
	receiver.Entry["rule"] = observerRule

	var defaultConfig map[string]any
	defaultConfig, hasDefault := receiver.Config[defaultType]
	if hasDefault {
		receiver.Entry["config"] = defaultConfig
	}
	observerConfigBlock, hasObserverConfigBlock := receiver.Config[observerID]
	if !hasObserverConfigBlock && !hasDefault {
		d.logger.Debug(fmt.Sprintf("disregarding %s without a default and %s config", receiverID.String(), observerID.String()))
		return false, nil
	}
	if hasObserverConfigBlock {
		if hasDefault {
			if err := mergeMaps(defaultConfig, observerConfigBlock); err != nil {
				return false, fmt.Errorf("failed merging %s config for %s: %w", receiverID.String(), observerID.String(), err)
			}
		} else {
			receiver.Entry["config"] = observerConfigBlock
		}
	}
	return true, nil
}

func factoryForObserverType(extType component.Type) (otelcolextension.Factory, error) {
	factories := map[component.Type]otelcolextension.Factory{
		"docker_observer":   dockerobserver.NewFactory(),
		"host_observer":     hostobserver.NewFactory(),
		"k8s_observer":      k8sobserver.NewFactory(),
		"ecs_task_observer": ecstaskobserver.NewFactory(),
	}
	ef, ok := factories[extType]
	if !ok {
		return nil, fmt.Errorf("unsupported discovery observer %q. Please remove its .discovery.yaml from your config directory", extType)
	}
	return ef, nil
}

func (d *discoverer) discoveryConfig(cfg *Config) (map[string]any, error) {
	dCfg := confmap.New()
	receiverAdded := false
	for receiverID, receiverStatus := range d.discoveredReceivers {
		if receiverStatus == discovery.Failed {
			continue
		}
		if receiverCfgMap, ok := d.discoveredConfig[receiverID]; ok {
			receiverCreator := confmap.NewFromStringMap(
				map[string]any{"receivers": map[string]any{"receiver_creator/discovery": receiverCfgMap}},
			)
			if err := dCfg.Merge(receiverCreator); err != nil {
				return nil, fmt.Errorf("failure adding receiver entry to suggested config: %w", err)
			}
			receiverAdded = true
		}
	}
	if receiverAdded {
		dCfg.Merge(confmap.NewFromStringMap(map[string]any{
			"service": map[string]any{
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"receivers": []string{"receiver_creator/discovery"},
					},
				},
			},
		}))
	}

	extensions := confmap.NewFromStringMap(map[string]any{"extensions": map[string]any{}})
	var observers []string
	for observerID, observerStatus := range d.discoveredObservers {
		if observerStatus == discovery.Failed {
			continue
		}
		if observerCfg, ok := cfg.DiscoveryObservers[observerID]; ok {
			obsMap := map[string]any{
				"extensions": map[string]any{
					observerID.String(): observerCfg.ToStringMap(),
				},
			}
			if err := extensions.Merge(confmap.NewFromStringMap(obsMap)); err != nil {
				return nil, fmt.Errorf("failure merging %q with suggested config: %w", observerID.String(), err)
			}
			observers = append(observers, observerID.String())
		}
	}

	if len(observers) > 0 {
		if err := dCfg.Merge(
			confmap.NewFromStringMap(
				map[string]any{"service": map[string]any{"extensions": observers}},
			),
		); err != nil {
			return nil, fmt.Errorf("failed forming suggested discovery observer extensions array: %w", err)
		}
	}

	if err := dCfg.Merge(extensions); err != nil {
		return nil, fmt.Errorf("failed merging discovery observer extensions: %w", err)
	}

	sMap := dCfg.ToStringMap()
	d.logger.Debug("determined discovery config", zap.Any("config", sMap))

	return sMap, nil
}

func (d *discoverer) createExtensionCreateSettings(kind string) otelcolextension.CreateSettings {
	return otelcolextension.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.New(d.logger.Core()).With(zap.String("kind", kind)),
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  metric.NewNoopMeterProvider(),
			MetricsLevel:   configtelemetry.LevelDetailed,
		},
		BuildInfo: d.info,
	}
}

func (d *discoverer) createReceiverCreateSettings() otelcolreceiver.CreateSettings {
	return otelcolreceiver.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.New(d.logger.Core()).With(zap.String("kind", "receiver")),
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  metric.NewNoopMeterProvider(),
			MetricsLevel:   configtelemetry.LevelDetailed,
		},
		BuildInfo: d.info,
	}
}

func (c *Config) observersForDiscoveryMode() []component.ID {
	var cids []component.ID
	for k := range c.DiscoveryObservers {
		cids = append(cids, k)
	}
	return cids
}

var _ component.Host = (*discoverer)(nil)

// ReportFatalError is a component.Host method.
func (d *discoverer) ReportFatalError(err error) {
	panic(fmt.Sprintf("--discovery fatal error: %v", err))
}

// GetFactory is a component.Host method used to forward the distribution's components.
func (d *discoverer) GetFactory(kind component.Kind, componentType component.Type) component.Factory {
	switch kind {
	case component.KindExporter:
		return d.factories.Exporters[componentType]
	case component.KindReceiver:
		return d.factories.Receivers[componentType]
	case component.KindExtension:
		return d.factories.Extensions[componentType]
	case component.KindProcessor:
		return d.factories.Processors[componentType]
	}
	return nil
}

// GetExtensions is a component.Host method used to forward discovery observers.
func (d *discoverer) GetExtensions() map[component.ID]otelcolextension.Extension {
	return d.extensions
}

// GetExporters is a component.Host method.
func (d *discoverer) GetExporters() map[component.DataType]map[component.ID]component.Component {
	return nil
}

var _ consumer.Logs = (*discoverer)(nil)

// Capabilities is a consumer.Logs method.
func (d *discoverer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

// ConsumeLogs will walk through all discovery receiver-emitted logs and store all receiver and observer statuses,
// including reported receiver configs from their discovery.receiver.config attribute. It is a consumer.Logs method.
func (d *discoverer) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	if ld.LogRecordCount() == 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	rlogs := ld.ResourceLogs()
	for i := 0; i < rlogs.Len(); i++ {
		var (
			receiverType, receiverName string
			receiverConfig, obsID      string
			observerID                 component.ID
			err                        error
		)
		rlog := rlogs.At(i)
		rAttrs := rlog.Resource().Attributes()
		if rName, ok := rAttrs.Get(discovery.ReceiverNameAttr); ok {
			receiverName = rName.Str()
		}
		rType, ok := rAttrs.Get(discovery.ReceiverTypeAttr)
		if !ok {
			// nothing we can do without this one
			continue
		}
		receiverType = rType.Str()
		if rConfig, ok := rAttrs.Get(discovery.ReceiverConfigAttr); ok {
			receiverConfig = rConfig.Str()
		}
		if rObsID, ok := rAttrs.Get(discovery.ObserverIDAttr); ok {
			obsID = rObsID.Str()
		}

		if obsID != "" {
			observerID = component.ID{}
			if err = observerID.UnmarshalText([]byte(obsID)); err != nil {
				d.logger.Debug(
					fmt.Sprintf("invalid %s", discovery.ObserverIDAttr),
					zap.String("observer id", obsID), zap.Error(err),
				)
				continue
			}
		}

		receiverID := component.NewIDWithName(component.Type(receiverType), receiverName)
		if receiverConfig != "" {
			rCfg := map[string]any{}
			var dBytes []byte
			if dBytes, err = base64.StdEncoding.DecodeString(receiverConfig); err != nil {
				return err
			}
			if err = yaml.Unmarshal(dBytes, &rCfg); err != nil {
				return err
			}
			d.discoveredConfig[receiverID] = rCfg
		}

		currentReceiverStatus := d.discoveredReceivers[receiverID]
		currentObserverStatus := d.discoveredObservers[observerID]

		slogs := rlog.ScopeLogs()
		for j := 0; j < slogs.Len(); j++ {
			slog := slogs.At(0)
			lrs := slog.LogRecords()
			for k := 0; k < lrs.Len(); k++ {
				lr := lrs.At(k)
				if currentReceiverStatus != discovery.Successful || currentObserverStatus != discovery.Successful {
					if rStatusAttr, ok := lr.Attributes().Get(discovery.StatusAttr); ok {
						rStatus := discovery.StatusType(rStatusAttr.Str())
						if ok, err = discovery.IsValidStatus(rStatus); !ok {
							d.logger.Debug("invalid status from log record", zap.Error(err), zap.Any("lr", lr.Body().AsRaw()))
							continue
						}
						d.discoveredReceivers[receiverID] = determineCurrentStatus(currentReceiverStatus, rStatus)
						d.discoveredObservers[observerID] = determineCurrentStatus(currentObserverStatus, rStatus)
					}
				}
			}
		}
	}

	return nil
}

func determineCurrentStatus(current, observed discovery.StatusType) discovery.StatusType {
	switch {
	case observed == discovery.Successful:
		current = discovery.Successful
	case current == discovery.Failed && observed == discovery.Partial:
		current = discovery.Partial
	default:
		current = observed
	}
	return current
}

func mergeMaps(dst, src map[string]any) error {
	dstMap := confmap.NewFromStringMap(dst)
	srcMap := confmap.NewFromStringMap(src)
	if err := dstMap.Merge(srcMap); err != nil {
		return err
	}
	for k, v := range dstMap.ToStringMap() {
		dst[k] = v
	}
	return nil
}
