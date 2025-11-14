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
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	otelcolextension "go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/internal"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/properties"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const logLevelEnvVar = "SPLUNK_DISCOVERY_LOG_LEVEL"

var _ = featuregate.GlobalRegistry().MustRegister(
	"splunk.continuousDiscovery",
	featuregate.StageStable,
	featuregate.WithRegisterDescription("When enabled, service discovery will continuously run and collect metrics from discovered services."),
	featuregate.WithRegisterFromVersion("v0.109.0"),
	featuregate.WithRegisterToVersion("v0.130.0"),
)

// discoverer provides the mechanism for a "preflight" collector service
// that will stand up the observers and discovery receivers based on the .discovery.yaml
// contents of the config dir, acting as a log consumer to determine which
// of the underlying receivers were successfully discovered by the
// discovery receiver from its emitted log records.
type discoverer struct {
	factories otelcol.Factories
	// receiverID -> observerID -> config
	unexpandedReceiverEntries map[component.ID]map[component.ID]map[string]any
	operationalObservers      map[component.ID]component.Component // Only extensions successfully started should be added to this map.
	logger                    *zap.Logger
	configs                   map[string]*Config
	// propertiesConf is a store of all properties from cmdline args and env vars
	// that's merged with receiver/observer configs before creation
	propertiesConf          *confmap.Conf
	info                    component.BuildInfo
	mu                      sync.Mutex
	propertiesFileSpecified bool
}

func newDiscoverer(logger *zap.Logger) (*discoverer, error) {
	info := component.BuildInfo{
		Command: "discovery",
		Version: version.Version,
	}
	factories, err := components.Get()
	if err != nil {
		return (*discoverer)(nil), err
	}
	d := &discoverer{
		logger:                    logger,
		info:                      info,
		factories:                 factories,
		configs:                   map[string]*Config{},
		mu:                        sync.Mutex{},
		unexpandedReceiverEntries: map[component.ID]map[component.ID]map[string]any{},
	}
	d.propertiesConf = d.propertiesConfFromEnv()
	return d, nil
}

func (d *discoverer) resolveConfig(cm map[string]any) (*confmap.Conf, error) {
	out, err := yaml.Marshal(cm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal discovery config for uri: %w", err)
	}
	uris := []string{fmt.Sprintf("yaml:%s", out)}
	resolver, err := confmap.NewResolver(confmap.ResolverSettings{
		URIs:              uris,
		ProviderFactories: []confmap.ProviderFactory{yamlprovider.NewFactory(), envprovider.NewFactory()},
		DefaultScheme:     "env",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a resolver from the given uris. %w", err)
	}

	conf, err := resolver.Resolve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve configuration from the resolver %w", err)
	}
	if err = resolver.Shutdown(context.Background()); err != nil {
		d.logger.Warn("error shutting down resolver", zap.Error(err))
	}
	return conf, nil
}

func (d *discoverer) propertiesConfFromEnv() *confmap.Conf {
	propertiesConf := confmap.New()
	for _, env := range os.Environ() {
		equalsIdx := strings.Index(env, "=")
		if equalsIdx != -1 && len(env) > equalsIdx+1 {
			envVar := env[:equalsIdx]
			if envVar == logLevelEnvVar {
				continue
			}
			if p, ok, e := properties.NewPropertyFromEnvVar(envVar, env[equalsIdx+1:]); ok {
				if e != nil {
					d.logger.Info(fmt.Sprintf("invalid discovery property environment variable %q", env), zap.Error(e))
					continue
				}
				propertiesConf.Merge(confmap.NewFromStringMap(p.ToStringMap()))
			}
		}
	}
	return propertiesConf
}

// discover will create all .discovery.yaml components, start them, wait the configured
// duration, and tear them down before returning the discovery config.
func (d *discoverer) discover(cfg *Config) (map[string]any, error) {
	if !d.propertiesFileSpecified {
		if err := d.mergeDiscoveryPropertiesEntry(cfg); err != nil {
			return nil, fmt.Errorf("failed reconciling properties.discovery: %w", err)
		}
	}
	d.prepareObserverConfigs(cfg)

	if len(cfg.DiscoveryObservers) == 0 {
		fmt.Fprintf(os.Stderr, "No discovery observers have been configured.\n")
		return nil, nil
	}

	cancels := d.startObservers(cfg)
	defer combineCancelFuncs(cancels)()
	defer d.stopObservers()

	discoveryReceiversConfigs, err := d.discoveryReceiversConfigs(cfg)
	if err != nil {
		d.logger.Error("failed preparing discovery receivers", zap.Error(err))
		return nil, err
	}

	return d.continuousDiscoveryConfig(cfg, discoveryReceiversConfigs), nil
}

func combineCancelFuncs(cancels []context.CancelFunc) context.CancelFunc {
	return func() {
		for _, cancel := range cancels {
			cancel()
		}
	}
}

func (d *discoverer) startObservers(cfg *Config) []context.CancelFunc {
	var cancels []context.CancelFunc
	d.operationalObservers = make(map[component.ID]component.Component, len(cfg.DiscoveryObservers))
	for observerID, observerEntry := range cfg.DiscoveryObservers {
		if observerEntry.Enabled != nil && !*observerEntry.Enabled {
			d.logger.Debug(fmt.Sprintf("skipping observer %q as it is disabled", observerID))
			continue
		}
		d.logger.Debug(fmt.Sprintf("creating observer %q", observerID))
		observerFactory, err := factoryForObserverType(observerID.Type())
		if err != nil {
			d.logger.Warn(err.Error())
			continue
		}

		obsCfg, err := d.resolveConfig(observerEntry.Config)
		if err != nil {
			d.logger.Warn("error resolving observer config", zap.Error(err))
			continue
		}

		observerConfig := observerFactory.CreateDefaultConfig()
		if err = obsCfg.Unmarshal(&observerConfig); err != nil {
			d.logger.Warn(fmt.Sprintf("failed unmarshaling %q config", observerID), zap.Error(err))
			continue
		}

		if ce := d.logger.Check(zap.DebugLevel, "unmarshalled observer config"); ce != nil {
			ce.Write(zap.String("config", fmt.Sprintf("%#v\n", observerConfig)))
		}

		observerSettings := d.createExtensionCreateSettings(observerID)
		observer, err := observerFactory.Create(context.Background(), observerSettings, observerConfig)
		if err != nil {
			d.logger.Warn(fmt.Sprintf("failed creating %q extension", observerID), zap.Error(err))
			continue
		}

		d.logger.Debug(fmt.Sprintf("starting observer %q", observerID))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancels = append(cancels, cancel)
		if e := observer.Start(ctx, nil); e != nil {
			d.logger.Warn(
				fmt.Sprintf("%q startup failed. Won't proceed with %q-based discovery", observerID, observerID.Type()),
				zap.Error(e),
			)
			continue
		}
		d.operationalObservers[observerID] = observer
	}
	return cancels
}

func (d *discoverer) stopObservers() {
	for observerID, observer := range d.operationalObservers {
		if e := observer.Shutdown(context.Background()); e != nil {
			d.logger.Warn(fmt.Sprintf("error shutting down observer %q", observerID), zap.Error(e))
		}
	}
}

func (d *discoverer) continuousDiscoveryConfig(cfg *Config, discoveryReceiversConfigs map[string]any) map[string]any {
	extensions := map[string]any{}
	var observerIDs []string
	for observerID, observerEntry := range cfg.DiscoveryObservers {
		_, ok := d.operationalObservers[observerID]
		if !ok {
			continue
		}
		extensions[observerID.String()] = observerEntry.Config.ToStringMap()
		observerIDs = append(observerIDs, observerID.String())
	}

	receiverIDs := make([]string, 0, len(discoveryReceiversConfigs))
	for receiverID := range discoveryReceiversConfigs {
		receiverIDs = append(receiverIDs, receiverID)
	}

	dCfg := map[string]any{
		"extensions": extensions,
		"receivers":  discoveryReceiversConfigs,
		"service": map[string]any{
			discovery.DiscoReceiversKey:  receiverIDs,
			discovery.DiscoExtensionsKey: observerIDs,
		},
	}
	d.logger.Debug("determined continuous discovery config", zap.Any("config", dCfg))
	return dCfg
}

// discoveryReceiversConfigs merges properties into the discovery receiver configs and returns a map of all discovery receiver configs
// that can be created for the enabled observers.
func (d *discoverer) discoveryReceiversConfigs(cfg *Config) (map[string]any, error) {
	discoveryReceiversConfigs := map[string]any{}
	for observerID := range d.operationalObservers {
		discoveryReceiverRaw := map[string]any{}
		discoveryReceiverRaw["watch_observers"] = []string{observerID.String()}
		discoveryReceiverRaw["embed_receiver_config"] = true
		receiversSection := map[string]any{}
		discoveryReceiverRaw["receivers"] = receiversSection

		receiversPropertiesConf := confmap.New()
		if d.propertiesConf.IsSet("receivers") {
			var err error
			receiversPropertiesConf, err = d.propertiesConf.Sub("receivers")
			if err != nil {
				return nil, fmt.Errorf("failed obtaining receivers properties config: %w", err)
			}
		}
		for receiverID, receiver := range cfg.ReceiversToDiscover {
			if ok, updErr := d.updateReceiverForObserver(receiverID, &receiver, observerID); updErr != nil {
				return nil, updErr
			} else if !ok {
				continue
			}
			enabled := true
			if e := receiver.Enabled; e != nil {
				enabled = *e
			}
			receiverEntry := receiver.Entry.ToStringMap()
			if receiversPropertiesConf.IsSet(receiverID.String()) {
				receiverPropertiesConf, e := receiversPropertiesConf.Sub(receiverID.String())
				if e != nil {
					return nil, fmt.Errorf("failed obtaining receiver properties config: %w", e)
				}
				entryConf := confmap.NewFromStringMap(receiverEntry)

				if receiverPropertiesConf.IsSet("enabled") {
					if b, convErr := strconv.ParseBool(strings.ToLower(fmt.Sprintf("%v", receiverPropertiesConf.Get("enabled")))); convErr == nil {
						// convErr would have been detected in properties
						enabled = b
					}
					pc := receiverPropertiesConf.ToStringMap()
					delete(pc, "enabled")
					receiverPropertiesConf = confmap.NewFromStringMap(pc)
				}

				if err := entryConf.Merge(receiverPropertiesConf); err != nil {
					return nil, fmt.Errorf("failed merging receiver %q properties config: %w", receiverID, err)
				}
				receiverEntry = entryConf.ToStringMap()
			}

			if !enabled {
				continue
			}

			d.addUnexpandedReceiverConfig(receiverID, observerID, receiverEntry)
			receiversSection[receiverID.String()] = receiverEntry
		}

		receiverName := component.MustNewIDWithName(discoveryreceiver.NewFactory().Type().String(), observerID.String()).String()
		discoveryReceiversConfigs[receiverName] = discoveryReceiverRaw
	}

	return discoveryReceiversConfigs, nil
}

func (d *discoverer) prepareObserverConfigs(cfg *Config) {
	for _, observerID := range cfg.observersForDiscoveryMode() {
		if err := d.prepareObserverConfig(observerID, cfg); err != nil {
			d.logger.Info(fmt.Sprintf("failed configuring %q extension. "+
				"no service discovery possible on this platform", observerID), zap.Error(err))
		}
	}
}

// prepareObserverConfigs merges properties into the observer configs.
func (d *discoverer) prepareObserverConfig(observerID component.ID, cfg *Config) error {
	enabled := true
	if e := cfg.DiscoveryObservers[observerID].Enabled; e != nil {
		enabled = *e
	}

	observerDiscoveryConf := confmap.NewFromStringMap(
		cfg.DiscoveryObservers[observerID].Config.ToStringMap(),
	)

	if d.propertiesConf.IsSet("extensions") {
		propertiesConf, e := d.propertiesConf.Sub("extensions")
		if e != nil {
			return fmt.Errorf("failed obtaining extensions properties config: %w", e)
		}
		if propertiesConf.IsSet(observerID.String()) {
			var observerProperties *confmap.Conf
			if observerProperties, e = propertiesConf.Sub(observerID.String()); e != nil {
				return fmt.Errorf("failed obtaining observer properties: %w", e)
			}
			if propertiesConf, e = observerProperties.Sub("config"); e != nil {
				return fmt.Errorf("failed obtaining observer properties config: %w", e)
			}
			if observerProperties.IsSet("enabled") {
				if b, convErr := strconv.ParseBool(strings.ToLower(fmt.Sprintf("%v", observerProperties.Get("enabled")))); convErr == nil {
					// convErr would have been detected in properties
					enabled = b
				}
			}
			if err := observerDiscoveryConf.Merge(propertiesConf); err != nil {
				return fmt.Errorf("failed merging observer properties config: %w", err)
			}

			// update the discovery config item for later retrieval in unexpanded form
			cfg.DiscoveryObservers[observerID] = ObserverEntry{
				Enabled: &enabled,
				Config:  observerDiscoveryConf.ToStringMap(),
			}
		}
	}
	return nil
}

func (d *discoverer) updateReceiverForObserver(receiverID component.ID, receiver *ReceiverToDiscoverEntry, observerID component.ID) (bool, error) {
	observerRule, hasRule := receiver.Rule[observerID]
	if !hasRule {
		d.logger.Debug(fmt.Sprintf("disregarding %q without a %q rule", receiverID, observerID))
		return false, nil
	}

	receiver.Entry = make(Entry)
	receiver.Entry["rule"] = observerRule

	var defaultConfig map[string]any
	defaultConfig, hasDefault := receiver.Config[defaultType]
	if hasDefault {
		receiver.Entry["config"] = defaultConfig
	}
	observerConfigBlock, hasObserverConfigBlock := receiver.Config[observerID]
	if !hasObserverConfigBlock && !hasDefault {
		d.logger.Debug(fmt.Sprintf("disregarding %q without a default and %q config", receiverID, observerID))
		return false, nil
	}
	if hasObserverConfigBlock {
		if hasDefault {
			if err := internal.MergeMaps(defaultConfig, observerConfigBlock); err != nil {
				return false, fmt.Errorf("failed merging %q config for %q: %w", receiverID, observerID, err)
			}
		} else {
			receiver.Entry["config"] = observerConfigBlock
		}
	}
	return true, nil
}

func factoryForObserverType(extType component.Type) (otelcolextension.Factory, error) {
	factories := map[component.Type]otelcolextension.Factory{
		component.MustNewType("docker_observer"): dockerobserver.NewFactory(),
		component.MustNewType("host_observer"):   hostobserver.NewFactory(),
		component.MustNewType("k8s_observer"):    k8sobserver.NewFactory(),
	}

	ef, ok := factories[extType]
	if !ok {
		return nil, fmt.Errorf("unsupported discovery observer %q. Please remove its .discovery.yaml from your config directory", extType)
	}
	return ef, nil
}

func (d *discoverer) createExtensionCreateSettings(observerID component.ID) otelcolextension.Settings {
	return otelcolextension.Settings{
		ID: observerID,
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.New(d.logger.Core()).With(zap.String("kind", observerID.String())),
			TracerProvider: tnoop.NewTracerProvider(),
			MeterProvider:  mnoop.NewMeterProvider(),
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

func (d *discoverer) addUnexpandedReceiverConfig(receiverID, observerID component.ID, cfg map[string]any) {
	d.logger.Debug(fmt.Sprintf("adding unexpanded config[%q][%q]: %v\n", receiverID, observerID, cfg))
	observerMap, ok := d.unexpandedReceiverEntries[receiverID]
	if !ok {
		observerMap = map[component.ID]map[string]any{}
		d.unexpandedReceiverEntries[receiverID] = observerMap
	}
	observerMap[observerID] = cfg
}

// mergeDiscoveryPropertiesEntry validates and merges properties.discovery.yaml content with existing sources.
// Priority is discovery.properties.yaml < env var properties < --set properties. --set and env var properties
// are already resolved at this point.
func (d *discoverer) mergeDiscoveryPropertiesEntry(cfg *Config) error {
	conf, warning, fatal := properties.LoadConf(cfg.DiscoveryProperties.ToStringMap())
	if fatal != nil {
		return fatal
	}
	if warning != nil {
		d.logger.Warn("invalid discovery properties will be disregarded", zap.Error(warning))
	}
	if err := conf.Merge(d.propertiesConf); err != nil {
		return err
	}
	d.propertiesConf = conf
	return nil
}
