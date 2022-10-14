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
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
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
	noType      = config.NewComponentID("")
	defaultType = config.NewComponentID("default")

	discoveryDirRegex = fmt.Sprintf("[^%s]*", compilablePathSeparator)
	serviceEntryRegex = regexp.MustCompile(fmt.Sprintf("%s%sservice\\.(yaml|yml)$", discoveryDirRegex, compilablePathSeparator))

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
	Service map[string]ServiceEntry
	// Exporters is a map of exporters to use in final config.
	// They must be in `config.d/exporters` directory.
	Exporters map[config.ComponentID]ExporterEntry
	// Extensions is a map of extensions to use in final config.
	// They must be in `config.d/extensions` directory.
	Extensions map[config.ComponentID]ExtensionEntry
	// DiscoveryObservers is a map of observer extensions to use in discovery,
	// overriding the default settings. They must be in `config.d/extensions` directory
	// and end with ".discovery.yaml".
	DiscoveryObservers map[config.ComponentID]ExtensionEntry
	// Processors is a map of extensions to use in final config.
	// They must be in `config.d/processors` directory.
	Processors map[config.ComponentID]ProcessorEntry
	// Receivers is a map of receiver entries to use in final config
	// They must be in `config.d/receivers` directory.
	Receivers map[config.ComponentID]ReceiverEntry
	// ReceiversToDiscover is a map of receiver entries to use in discovery mode's
	// underlying discovery receiver. They must be in `config.d/receivers` directory and
	// end with ".discovery.yaml".
	ReceiversToDiscover map[config.ComponentID]ReceiverToDiscoverEntry
}

func NewConfig(logger *zap.Logger) *Config {
	return &Config{
		logger:              logger,
		Service:             map[string]ServiceEntry{},
		Exporters:           map[config.ComponentID]ExporterEntry{},
		Extensions:          map[config.ComponentID]ExtensionEntry{},
		DiscoveryObservers:  map[config.ComponentID]ExtensionEntry{},
		Processors:          map[config.ComponentID]ProcessorEntry{},
		Receivers:           map[config.ComponentID]ReceiverEntry{},
		ReceiversToDiscover: map[config.ComponentID]ReceiverToDiscoverEntry{},
	}
}

func dirAndEntryRegex(dirName string) (*regexp.Regexp, *regexp.Regexp) {
	dirRegex := regexp.MustCompile(fmt.Sprintf("%s%s%s", discoveryDirRegex, compilablePathSeparator, dirName))
	entryRegex := regexp.MustCompile(fmt.Sprintf("%s%s[^%s]*\\.(yaml|yml)$", dirRegex, compilablePathSeparator, compilablePathSeparator))
	return dirRegex, entryRegex
}

type keyType interface {
	string | config.ComponentID
}

type entryType interface {
	ServiceEntry | ExporterEntry | ExtensionEntry | ProcessorEntry | ReceiverEntry | ObserverEntry | ReceiverToDiscoverEntry
	ErrorF(path string, err error) error
	ToStringMap() map[string]any
}

type Entry map[string]any

func (e Entry) ToStringMap() map[string]any {
	return e
}

type ServiceEntry struct {
	Entry `yaml:",inline"`
}

func (ServiceEntry) ErrorF(path string, err error) error {
	return errorF(typeService, path, err)
}

type ExtensionEntry struct {
	Entry `yaml:",inline"`
}

func (ExtensionEntry) ErrorF(path string, err error) error {
	return errorF(typeExtension, path, err)
}

type ExporterEntry struct {
	Entry `yaml:",inline"`
}

func (ExporterEntry) ErrorF(path string, err error) error {
	return errorF(typeExporter, path, err)
}

type ObserverEntry struct {
	Entry `yaml:",inline"`
}

func (ObserverEntry) ErrorF(path string, err error) error {
	return errorF(typeDiscoveryObserver, path, err)
}

type ProcessorEntry struct {
	Entry `yaml:",inline"`
}

func (ProcessorEntry) ErrorF(path string, err error) error {
	return errorF(typeProcessor, path, err)
}

type ReceiverEntry struct {
	Entry `yaml:",inline"`
}

func (ReceiverEntry) ErrorF(path string, err error) error {
	return errorF(typeReceiver, path, err)
}

type ReceiverToDiscoverEntry struct {
	// Receiver creator rules by observer extension ID
	Rule map[config.ComponentID]string
	// Platform/observer specific config by observer extension ID.
	// These are merged w/ "default" component.ID in a "config" map
	Config map[config.ComponentID]map[string]any
	// The remaining items used to merge applicable rule and config
	Entry `yaml:",inline"`
}

func (r ReceiverToDiscoverEntry) ToStringMap() map[string]any {
	return r.Entry
}

func (ReceiverToDiscoverEntry) ErrorF(path string, err error) error {
	return errorF(typeReceiverToDiscover, path, err)
}

func (c *Config) Load(discoveryDPath string) error {
	if c == nil {
		return fmt.Errorf("config must not be nil to be loaded (use NewConfig())")
	}
	err := filepath.WalkDir(discoveryDPath, func(path string, d fs.DirEntry, err error) error {
		c.logger.Debug("loading component", zap.String("path", path), zap.String("DirEntry", fmt.Sprintf("%#v", d)), zap.Error(err))
		if err != nil {
			return err
		}
		switch {
		case isServiceEntryPath(path):
			return loadEntry(typeService, path, c.Service)
		case isExporterEntryPath(path):
			return loadEntry(typeExporter, path, c.Exporters)
		case isExtensionEntryPath(path):
			if isDiscoveryObserverEntryPath(path) {
				return loadEntry(typeDiscoveryObserver, path, c.DiscoveryObservers)
			}
			return loadEntry(typeExtension, path, c.Extensions)
		case isProcessorEntryPath(path):
			return loadEntry(typeProcessor, path, c.Processors)
		case isReceiverEntryPath(path):
			if isReceiverToDiscoverEntryPath(path) {
				return loadEntry(typeReceiverToDiscover, path, c.ReceiversToDiscover)
			}
			return loadEntry(typeReceiver, path, c.Receivers)
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
		c.Service = nil
		c.Exporters = nil
		c.Processors = nil
		c.Extensions = nil
	}
	return err
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

func loadEntry[K keyType, V entryType](componentType, path string, target map[K]V) error {
	tmpDest := map[K]V{}

	componentID, err := unmarshalEntry(componentType, path, &tmpDest)
	noTypeK := componentIDToK(noType, componentID)
	if err != nil {
		return tmpDest[noTypeK].ErrorF(path, err)
	}

	if componentID == noTypeK {
		return nil
	}

	if componentType == typeService {
		// set directly on target and exit
		for k, v := range tmpDest {
			target[k] = v
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

func unmarshalEntry[K keyType, V entryType](componentType, path string, dst *map[K]V) (componentID K, err error) {
	if dst == nil {
		err = fmt.Errorf("cannot load %s into nil entry", componentType)
		return
	}
	if err = unmarshalYaml(path, dst); err != nil {
		err = fmt.Errorf("failed unmarshalling component %s: %w", componentType, err)
		return
	}

	// service is marshaled as complete item so return as is
	if componentType == typeService {
		var s any = typeService
		// service key is always string so this type assertion is safe
		return s.(K), nil
	}

	entry := *dst

	if len(entry) == 0 {
		// empty or all-comment files are supported but ignored
		var noTypeK any = noType
		// non-service key is always componentID so this type assertion is safe
		return noTypeK.(K), nil
	}
	var componentIDs []K
	var component V
	for k, v := range entry {
		componentIDs = append(componentIDs, k)
		component = v
	}

	if len(componentIDs) != 1 {
		// deterministic for testability
		var cids []string
		for _, i := range componentIDs {
			cids = append(cids, keyTypeToString(i))
		}
		sort.Strings(cids)
		err = component.ErrorF(
			path, fmt.Errorf("must contain a single mapping of ComponentID to component but contained %v", cids),
		)
		return
	}
	return componentIDs[0], nil
}

func unmarshalYaml(path string, out any) error {
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed reading file %q: %w", path, err)
	}

	if err = yaml.Unmarshal(contents, out); err != nil {
		return fmt.Errorf("failed parsing %q as yaml: %w", path, err)
	}
	return nil
}

func componentIDToK[K keyType](cid config.ComponentID, key K) K {
	var componentIDK any
	for _, kid := range []K{key} {
		var cidK any = kid
		switch cidK.(type) {
		case string:
			componentIDK = cid.String()
		case config.ComponentID:
			componentIDK = cid
		}
		break
	}
	return componentIDK.(K)
}

func keyTypeToString[K keyType](key K) string {
	var ret string
	for _, k := range []K{key} {
		var kk any = k
		switch i := kk.(type) {
		case string:
			ret = i
		case config.ComponentID:
			ret = i.String()
		}
		break
	}
	return ret
}

var compilablePathSeparator = func() string {
	if os.PathSeparator == '\\' {
		return "\\\\"
	}
	return string(os.PathSeparator)
}()
