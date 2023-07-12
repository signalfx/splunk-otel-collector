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

package bundle

import (
	"fmt"
	"strings"
	"text/template"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/properties"
)

const defaultValue = "splunk.discovery.default"

func FuncMap() template.FuncMap {
	dc := newDiscoveryConfig()
	return map[string]any{
		"configProperty":       dc.configProperty,
		"configPropertyEnvVar": dc.configPropertyEnvVar,
		"extension":            dc.extension,
		"receiver":             dc.receiver,
		"defaultValue":         func() string { return defaultValue },
	}
}

type discoveryConfig struct {
	factories     otelcol.Factories
	componentID   component.ID
	componentKind component.Kind
}

func newDiscoveryConfig() *discoveryConfig {
	factories, err := components.Get()
	if err != nil {
		panic(fmt.Errorf("failed accessing distribution components: %w", err))
	}
	return &discoveryConfig{
		factories: factories,
	}
}

func (dc *discoveryConfig) extension(id string) (string, error) {
	return dc.setComponentType(id, component.KindExtension)
}

func (dc *discoveryConfig) receiver(id string) (string, error) {
	return dc.setComponentType(id, component.KindReceiver)
}

func (dc *discoveryConfig) setComponentType(id string, kind component.Kind) (string, error) {
	cid := &component.ID{}
	if err := cid.UnmarshalText([]byte(id)); err != nil {
		return "", err
	}
	dc.componentKind = kind
	dc.componentID = *cid
	switch kind {
	case component.KindExtension:
		if _, ok := dc.factories.Extensions[cid.Type()]; !ok {
			return "", fmt.Errorf("no extension %q available in this distribution", cid.Type())
		}
	case component.KindReceiver:
		if _, ok := dc.factories.Receivers[cid.Type()]; !ok {
			return "", fmt.Errorf("no receiver %q available in this distribution", cid.Type())
		}
	default:
		return "", fmt.Errorf("unsupported discovery config component kind %#v", kind)
	}
	return dc.componentID.String(), nil
}

func (dc *discoveryConfig) configProperty(args ...string) (string, error) {
	return dc.configPropertyWithStringer(args, "configProperty", func(property *properties.Property) string {
		return fmt.Sprintf("%s=%q", property.Input, property.Val)
	})
}

func (dc *discoveryConfig) configPropertyEnvVar(args ...string) (string, error) {
	return dc.configPropertyWithStringer(args, "configPropertyEnvVar", func(property *properties.Property) string {
		return fmt.Sprintf("%s=%q", property.ToEnvVar(), property.Val)
	})
}

func (dc *discoveryConfig) configPropertyWithStringer(args []string, methodName string, stringer func(property *properties.Property) string) (string, error) {
	prefix, err := dc.configPropertyPrefix(methodName, args)
	if err != nil {
		return "", err
	}
	property, err := configProperty(methodName, prefix, args)
	if err != nil {
		return "", err
	}

	return stringer(property), nil
}

func (dc *discoveryConfig) configPropertyPrefix(methodName string, args []string) (string, error) {
	l := len(args)
	if l < 2 {
		return "", fmt.Errorf("%s takes key+ and value{1} arguments (minimum 2)", methodName)
	}
	var prefix string
	switch dc.componentKind {
	case component.KindReceiver:
		prefix = fmt.Sprintf("splunk.discovery.receivers.%s.config", dc.componentID)
	case component.KindExtension:
		prefix = fmt.Sprintf("splunk.discovery.extensions.%s.config", dc.componentID)
	default:
		return "", fmt.Errorf("invalid discovery config component type %d", dc.componentKind)
	}
	return prefix, nil
}

func configProperty(methodName, prefix string, args []string) (*properties.Property, error) {
	la := len(args)
	if la < 1 {
		return nil, fmt.Errorf("%s requires at least a value", methodName)
	}
	key := ""
	if la > 1 {
		key = fmt.Sprintf(".%s", strings.Join(args[:la-1], "::"))
	}
	property := fmt.Sprintf("%s%s", prefix, key)
	return properties.NewProperty(property, args[la-1])
}
