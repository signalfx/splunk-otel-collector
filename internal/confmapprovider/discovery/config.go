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

package discovery

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"sort"

	"github.com/knadh/koanf/maps"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

const (
	typeService            = "service"
	typeReceiver           = "receiver"
	typeExporter           = "exporter"
	typeExtension          = "extension"
	typeProcessor          = "processor"
	typeDiscoveryObserver  = "discovery.extension"
	typeReceiverToDiscover = "discovery.receiver"
)

var (
	defaultType = component.NewID("default")

	discoveryDirRegex = fmt.Sprintf("[^%s]*", compilablePathSeparator)
	serviceEntryRegex = regexp.MustCompile(fmt.Sprintf("%s%s*service\\.(yaml|yml)$", discoveryDirRegex, compilablePathSeparator))

	_, exporterEntryRegex                   = dirAndEntryRegex("exporters")
	extensionsDirRegex, extensionEntryRegex = dirAndEntryRegex("extensions")
	discoveryObserverEntryRegex             = regexp.MustCompile(fmt.Sprintf("%s%s[^%s]*\\.discovery\\.(yaml|yml)$", extensionsDirRegex, compilablePathSeparator, compilablePathSeparator))

	_, processorEntryRegex                = dirAndEntryRegex("processors")
	receiversDirRegex, receiverEntryRegex = dirAndEntryRegex("receivers")
	receiverToDiscoverEntryRegex          = regexp.MustCompile(fmt.Sprintf("%s%s[^%s]*\\.discovery\\.(yaml|yml)$", receiversDirRegex, compilablePathSeparator, compilablePathSeparator))
)

// Config is a model for stitching together the final Collector configuration with additional discovery component
// fields for use w/ discovery mode (not yet implemented). It allows individual yaml files to be added to a config.d
// directory and be sourced in the final config such that small changes don't require a central configuration file,
// and possible eliminates the need for one overall (still in design).
type Config struct {
	logger *zap.Logger
	// Service is for pipelines and final settings
	Service ServiceEntry
	// Exporters is a map of exporters to use in final config.
	// They must be in `config.d/exporters` directory.
	Exporters map[component.ID]ExporterEntry
	// Extensions is a map of extensions to use in final config.
	// They must be in `config.d/extensions` directory.
	Extensions map[component.ID]ExtensionEntry
	// DiscoveryObservers is a map of observer extensions to use in discovery,
	// overriding the default settings. They must be in `config.d/extensions` directory
	// and end with ".discovery.yaml".
	DiscoveryObservers map[component.ID]ExtensionEntry
	// Processors is a map of extensions to use in final config.
	// They must be in `config.d/processors` directory.
	Processors map[component.ID]ProcessorEntry
	// Receivers is a map of receiver entries to use in final config
	// They must be in `config.d/receivers` directory.
	Receivers map[component.ID]ReceiverEntry
	// ReceiversToDiscover is a map of receiver entries to use in discovery mode's
	// underlying discovery receiver. They must be in `config.d/receivers` directory and
	// end with ".discovery.yaml".
	ReceiversToDiscover map[component.ID]ReceiverToDiscoverEntry
}

func NewConfig(logger *zap.Logger) *Config {
	return &Config{
		logger:              logger,
		Service:             ServiceEntry{Entry{}},
		Exporters:           map[component.ID]ExporterEntry{},
		Extensions:          map[component.ID]ExtensionEntry{},
		DiscoveryObservers:  map[component.ID]ExtensionEntry{},
		Processors:          map[component.ID]ProcessorEntry{},
		Receivers:           map[component.ID]ReceiverEntry{},
		ReceiversToDiscover: map[component.ID]ReceiverToDiscoverEntry{},
	}
}

func dirAndEntryRegex(dirName string) (*regexp.Regexp, *regexp.Regexp) {
	dirRegex := regexp.MustCompile(fmt.Sprintf("%s%s*%s", discoveryDirRegex, compilablePathSeparator, dirName))
	entryRegex := regexp.MustCompile(fmt.Sprintf("%s%s[^%s]*\\.(yaml|yml)$", dirRegex, compilablePathSeparator, compilablePathSeparator))
	return dirRegex, entryRegex
}

type keyType interface {
	string | component.ID
}

type entryType interface {
	ErrorF(path string, err error) error
	Self() Entry
	ToStringMap() map[string]any
}

type Entry map[string]any

func (e Entry) Self() Entry {
	return e
}

func (e Entry) ToStringMap() map[string]any {
	cp := map[string]any{}
	for k, v := range e {
		cp[k] = v
	}
	maps.IntfaceKeysToStrings(cp)
	return cp
}

var _ entryType = (*ServiceEntry)(nil)

type ServiceEntry struct {
	Entry `yaml:",inline"`
}

func (ServiceEntry) ErrorF(path string, err error) error {
	return errorF(typeService, path, err)
}

var _ entryType = (*ExtensionEntry)(nil)

type ExtensionEntry struct {
	Entry `yaml:",inline"`
}

func (ExtensionEntry) ErrorF(path string, err error) error {
	return errorF(typeExtension, path, err)
}

var _ entryType = (*ExporterEntry)(nil)

type ExporterEntry struct {
	Entry `yaml:",inline"`
}

func (ExporterEntry) ErrorF(path string, err error) error {
	return errorF(typeExporter, path, err)
}

var _ entryType = (*ObserverEntry)(nil)

type ObserverEntry struct {
	Entry `yaml:",inline"`
}

func (ObserverEntry) ErrorF(path string, err error) error {
	return errorF(typeDiscoveryObserver, path, err)
}

var _ entryType = (*ProcessorEntry)(nil)

type ProcessorEntry struct {
	Entry `yaml:",inline"`
}

func (ProcessorEntry) ErrorF(path string, err error) error {
	return errorF(typeProcessor, path, err)
}

var _ entryType = (*ReceiverEntry)(nil)

type ReceiverEntry struct {
	Entry `yaml:",inline"`
}

func (ReceiverEntry) ErrorF(path string, err error) error {
	return errorF(typeReceiver, path, err)
}

type ReceiverToDiscoverEntry struct {
	// Receiver creator rules by observer extension ID
	Rule map[component.ID]string
	// Platform/observer specific config by observer extension ID.
	// These are merged w/ "default" component.ID in a "config" map
	Config map[component.ID]map[string]any
	// The remaining items used to merge applicable rule and config
	Entry `yaml:",inline"`
}

var _ entryType = (*ReceiverToDiscoverEntry)(nil)

func (r ReceiverToDiscoverEntry) ToStringMap() map[string]any {
	return r.Entry.ToStringMap()
}

func (ReceiverToDiscoverEntry) ErrorF(path string, err error) error {
	return errorF(typeReceiverToDiscover, path, err)
}

// Load will walk the file tree from the configDPath root, loading the component
// files as they are discovered, determined by their parent directory and filename.
func (c *Config) Load(configDPath string) error {
	if c == nil {
		return fmt.Errorf("config must not be nil to be loaded (use NewConfig())")
	}
	return c.LoadFS(os.DirFS(configDPath))
}

// LoadFS will walk the provided filesystem, loading the component files as they are discovered,
// determined by their parent directory and filename.
func (c *Config) LoadFS(dirfs fs.FS) error {
	if c == nil {
		return fmt.Errorf("config must not be nil to be loaded (use NewConfig())")
	}
	err := fs.WalkDir(dirfs, ".", func(path string, d fs.DirEntry, err error) error {
		c.logger.Debug("loading component", zap.String("path", path), zap.String("DirEntry", fmt.Sprintf("%#v", d)), zap.Error(err))
		if err != nil {
			return err
		}

		switch {
		case isServiceEntryPath(path):
			// c.Service is not a map[string]ServiceEntry, so we form a tmp
			// and unmarshal to the underlying ServiceEntry
			tmpSEMap := map[string]ServiceEntry{typeService: c.Service}
			return loadEntry(typeService, dirfs, path, tmpSEMap)
		case isExporterEntryPath(path):
			return loadEntry(typeExporter, dirfs, path, c.Exporters)
		case isExtensionEntryPath(path):
			if isDiscoveryObserverEntryPath(path) {
				return loadEntry(typeDiscoveryObserver, dirfs, path, c.DiscoveryObservers)
			}
			return loadEntry(typeExtension, dirfs, path, c.Extensions)
		case isProcessorEntryPath(path):
			return loadEntry(typeProcessor, dirfs, path, c.Processors)
		case isReceiverEntryPath(path):
			if isReceiverToDiscoverEntryPath(path) {
				return loadEntry(typeReceiverToDiscover, dirfs, path, c.ReceiversToDiscover)
			}
			return loadEntry(typeReceiver, dirfs, path, c.Receivers)
		default:
			c.logger.Debug("Disregarding path", zap.String("path", path))
		}
		return nil
	})
	if err != nil {
		// clean up to prevent using partial config
		c.DiscoveryObservers = nil
		c.ReceiversToDiscover = nil
		c.Receivers = nil
		c.Service = ServiceEntry{nil}
		c.Exporters = nil
		c.Processors = nil
		c.Extensions = nil
	}
	return err
}

// toServiceConfig renders the loaded Config content
// suitable for use as a Collector configuration
func (c *Config) toServiceConfig() map[string]any {
	sc := confmap.New()
	service := c.Service.ToStringMap()
	sc.Merge(confmap.NewFromStringMap(map[string]any{typeService: service}))

	receivers := map[string]any{}
	for k, v := range c.Receivers {
		receivers[k.String()] = v.ToStringMap()
	}
	sc.Merge(confmap.NewFromStringMap(map[string]any{"receivers": receivers}))

	processors := map[string]any{}
	for k, v := range c.Processors {
		processors[k.String()] = v.ToStringMap()
	}
	sc.Merge(confmap.NewFromStringMap(map[string]any{"processors": processors}))

	exporters := map[string]any{}
	for k, v := range c.Exporters {
		exporters[k.String()] = v.ToStringMap()
	}
	sc.Merge(confmap.NewFromStringMap(map[string]any{"exporters": exporters}))

	extensions := map[string]any{}
	for k, v := range c.Extensions {
		extensions[k.String()] = v.ToStringMap()
	}
	sc.Merge(confmap.NewFromStringMap(map[string]any{"extensions": extensions}))

	return sc.ToStringMap()
}

func errorF(entryType, path string, err error) error {
	return fmt.Errorf("failed loading %s from %s: %w", entryType, path, err)
}

func isServiceEntryPath(path string) bool {
	return serviceEntryRegex.MatchString(path)
}

func isExporterEntryPath(path string) bool {
	return exporterEntryRegex.MatchString(path)
}

func isExtensionEntryPath(path string) bool {
	return extensionEntryRegex.MatchString(path)
}

func isDiscoveryObserverEntryPath(path string) bool {
	return discoveryObserverEntryRegex.MatchString(path)
}

func isProcessorEntryPath(path string) bool {
	return processorEntryRegex.MatchString(path)
}

func isReceiverEntryPath(path string) bool {
	return receiverEntryRegex.MatchString(path)
}

func isReceiverToDiscoverEntryPath(path string) bool {
	return receiverToDiscoverEntryRegex.MatchString(path)
}

func loadEntry[K keyType, V entryType](componentType string, fs fs.FS, path string, target map[K]V) error {
	tmpDest := map[K]V{}

	componentID, err := unmarshalEntry(componentType, fs, path, &tmpDest)
	noTypeK, err2 := stringToKeyType(discovery.NoType.String(), componentID)
	if err2 != nil {
		return err2
	}
	if err != nil {
		return tmpDest[noTypeK].ErrorF(path, err)
	}

	if componentID == noTypeK {
		return nil
	}

	if componentType == typeService {
		// set directly on target and exit
		typeServiceK, err := stringToKeyType(componentType, componentID)
		if err != nil {
			return err
		}
		serviceEntry := target[typeServiceK].Self()
		tmpDstSM := tmpDest[typeServiceK].ToStringMap()
		for k, v := range tmpDstSM {
			serviceEntry[keyTypeToString(k)] = v
		}
		return nil
	}

	if v, ok := target[componentID]; ok {
		return v.ErrorF(path, fmt.Errorf("duplicate %q", keyTypeToString(componentID)))
	}
	entry := tmpDest[componentID]
	target[componentID] = entry
	return nil
}

func unmarshalEntry[K keyType, V entryType](componentType string, fs fs.FS, path string, dst *map[K]V) (componentID K, err error) {
	if dst == nil {
		err = fmt.Errorf("cannot load %s into nil entry", componentType)
		return
	}

	var unmarshalDst any = dst

	// service is map[string]ServiceEntry{typeService: ServiceEntry} and we want dst to be &ServiceEntry
	if componentType == typeService {
		var s any = typeService
		// service key is always string so this type assertion is safe
		se := (*dst)[s.(K)]
		unmarshalDst = &se
	}

	if err = unmarshalYaml(fs, path, unmarshalDst); err != nil {
		err = fmt.Errorf("failed unmarshalling component %s: %w", componentType, err)
		return
	}

	if componentType == typeService {
		// reset map[string]ServiceEntry dst w/ unmarshalled ServiceEntry and return
		var tService any = typeService
		var serviceEntry any = *(unmarshalDst.(*ServiceEntry))
		(*dst)[tService.(K)] = serviceEntry.(V)
		return tService.(K), nil
	}

	entry := *(unmarshalDst.(*map[K]V))

	if len(entry) == 0 {
		// empty or all-comment files are supported but ignored
		var noTypeK any = discovery.NoType
		// non-service key is always componentID so this type assertion is safe
		return noTypeK.(K), nil
	}
	var componentIDs []K
	var comp V
	for k, v := range entry {
		componentIDs = append(componentIDs, k)
		comp = v
	}

	if len(componentIDs) != 1 {
		// deterministic for testability
		var cids []string
		for _, i := range componentIDs {
			cids = append(cids, keyTypeToString(i))
		}
		sort.Strings(cids)
		err = comp.ErrorF(
			path, fmt.Errorf("must contain a single mapping of ComponentID to component but contained %v", cids),
		)
		return
	}
	return componentIDs[0], nil
}

func unmarshalYaml(fs fs.FS, path string, out any) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	contents, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed reading file %q: %w", path, err)
	}

	if err = yaml.Unmarshal(contents, out); err != nil {
		return fmt.Errorf("failed parsing %q as yaml: %w", path, err)
	}
	return nil
}

func stringToKeyType[K keyType](s string, key K) (K, error) {
	var componentIDK any
	for _, kid := range []K{key} {
		var cidK any = kid
		switch cidK.(type) {
		case string:
			var anyS any = s
			return anyS.(K), nil
		case component.ID:
			if s == discovery.NoType.String() {
				componentIDK = discovery.NoType
			} else {
				var err error
				componentIDK = component.ID{}
				cIDK := componentIDK.(component.ID)
				if err = (&cIDK).UnmarshalText([]byte(s)); err != nil {
					// nolint:gocritic
					return *new(K), err // (gocritic suggestion not valid with type parameter)
				}
			}
		}
		break
	}
	return componentIDK.(K), nil
}

func keyTypeToString[K keyType](key K) string {
	var ret string
	for _, k := range []K{key} {
		var anyK any = k
		switch i := anyK.(type) {
		case string:
			ret = i
		case component.ID:
			ret = i.String()
		}
		break
	}
	return ret
}

var compilablePathSeparator = func() string {
	if os.PathSeparator == '\\' {
		// fs.Stat doesn't use os.PathSeparator so accept '/' as well.
		// TODO: determine if we even need anything but "/"
		return "(\\\\|/)"
	}
	return string(os.PathSeparator)
}()

func mergeConfigWithBundle(userCfg *Config, bundleCfg *Config) error {
	for obs, bundledObs := range bundleCfg.DiscoveryObservers {
		userObs, ok := userCfg.DiscoveryObservers[obs]
		if !ok {
			userCfg.DiscoveryObservers[obs] = bundledObs
			continue
		}
		bundledConfMap := confmap.NewFromStringMap(bundledObs.ToStringMap())
		userConfMap := confmap.NewFromStringMap(userObs.ToStringMap())
		if err := bundledConfMap.Merge(userConfMap); err != nil {
			return fmt.Errorf("failed merged user and bundled observer %q discovery configs: %w", obs, err)
		}
		userCfg.DiscoveryObservers[obs] = ExtensionEntry{Entry: bundledConfMap.ToStringMap()}
	}
	for rec, bundledRec := range bundleCfg.ReceiversToDiscover {
		userRec, ok := userCfg.ReceiversToDiscover[rec]
		if !ok {
			userCfg.ReceiversToDiscover[rec] = bundledRec
			continue
		}
		bundledConfMap := confmap.NewFromStringMap(bundledRec.ToStringMap())
		userConfMap := confmap.NewFromStringMap(userRec.ToStringMap())
		if err := bundledConfMap.Merge(userConfMap); err != nil {
			return fmt.Errorf("failed merged user and bundled receiver %q discovery configs: %w", rec, err)
		}
		receiver := ReceiverToDiscoverEntry{
			Rule: bundledRec.Rule, Config: bundledRec.Config, Entry: bundledConfMap.ToStringMap(),
		}
		for cid, rule := range userRec.Rule {
			receiver.Rule[cid] = rule
		}
		for obs, config := range userRec.Config {
			if bundledConfig, ok := bundledRec.Config[obs]; ok {
				bundledConf := confmap.NewFromStringMap(bundledConfig)
				bundledConf.Merge(confmap.NewFromStringMap(config))
				config = bundledConf.ToStringMap()
			}
			receiver.Config[obs] = config
		}
		userCfg.ReceiversToDiscover[rec] = receiver
	}
	return nil
}
