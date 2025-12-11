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

package main

import (
	"fmt"
	"strings"
	"text/template"

	"go.opentelemetry.io/collector/component"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery/properties"
)

const defaultValue = "splunk.discovery.default"

func funcMap(componentID component.ID) template.FuncMap {
	return map[string]any{
		"configProperty":       createConfigProperty(componentID),
		"configPropertyEnvVar": createConfigPropertyEnvVar(componentID),
		"defaultValue":         func() string { return defaultValue },
	}
}

func createConfigProperty(componentID component.ID) func(...string) (string, error) {
	return func(args ...string) (string, error) {
		return configPropertyForComponent(componentID, "configProperty", args, func(property *properties.Property) string {
			return fmt.Sprintf("%s=%q", property.Input, property.Val)
		})
	}
}

func createConfigPropertyEnvVar(componentID component.ID) func(...string) (string, error) {
	return func(args ...string) (string, error) {
		return configPropertyForComponent(componentID, "configPropertyEnvVar", args, func(property *properties.Property) string {
			return fmt.Sprintf("%s=%q", property.ToEnvVar(), property.Val)
		})
	}
}

func configPropertyForComponent(componentID component.ID, methodName string, args []string, stringer func(property *properties.Property) string) (string, error) {
	factories, err := components.Get()
	if err != nil {
		return "", fmt.Errorf("failed accessing distribution components: %w", err)
	}

	// Validate component availability based on type
	componentType := ""
	if _, ok := factories.Receivers[componentID.Type()]; ok {
		componentType = "receivers"
	} else if _, ok := factories.Extensions[componentID.Type()]; ok {
		componentType = "extensions"
	} else {
		return "", fmt.Errorf("no receiver or extension %q available in this distribution", componentID.Type())
	}

	prefix, err := configPropertyPrefix(methodName, args, componentID, componentType)
	if err != nil {
		return "", err
	}
	property, err := configProperty(methodName, prefix, args)
	if err != nil {
		return "", err
	}

	return stringer(property), nil
}

func configPropertyPrefix(methodName string, args []string, componentID component.ID, componentType string) (string, error) {
	l := len(args)
	if l < 2 {
		return "", fmt.Errorf("%s takes key+ and value{1} arguments (minimum 2)", methodName)
	}
	prefix := fmt.Sprintf("splunk.discovery.%s.%s.config", componentType, componentID)
	return prefix, nil
}

func configProperty(methodName, prefix string, args []string) (*properties.Property, error) {
	la := len(args)
	if la < 1 {
		return nil, fmt.Errorf("%s requires at least a value", methodName)
	}
	key := ""
	if la > 1 {
		key = "." + strings.Join(args[:la-1], "::")
	}
	property := fmt.Sprintf("%s%s", prefix, key)
	return properties.NewProperty(property, args[la-1])
}
