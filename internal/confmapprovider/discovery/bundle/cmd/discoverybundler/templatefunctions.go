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

func funcMap(receiverID component.ID) template.FuncMap {
	return map[string]any{
		"configProperty":       createConfigProperty(receiverID),
		"configPropertyEnvVar": createConfigPropertyEnvVar(receiverID),
		"defaultValue":         func() string { return defaultValue },
	}
}

func createConfigProperty(receiverID component.ID) func(...string) (string, error) {
	return func(args ...string) (string, error) {
		return configPropertyForReceiver(receiverID, "configProperty", args, func(property *properties.Property) string {
			return fmt.Sprintf("%s=%q", property.Input, property.Val)
		})
	}
}

func createConfigPropertyEnvVar(receiverID component.ID) func(...string) (string, error) {
	return func(args ...string) (string, error) {
		return configPropertyForReceiver(receiverID, "configPropertyEnvVar", args, func(property *properties.Property) string {
			return fmt.Sprintf("%s=%q", property.ToEnvVar(), property.Val)
		})
	}
}

func configPropertyForReceiver(receiverID component.ID, methodName string, args []string, stringer func(property *properties.Property) string) (string, error) {
	factories, err := components.Get()
	if err != nil {
		return "", fmt.Errorf("failed accessing distribution components: %w", err)
	}

	// Validate receiver availability
	if _, ok := factories.Receivers[receiverID.Type()]; !ok {
		return "", fmt.Errorf("no receiver %q available in this distribution", receiverID.Type())
	}

	prefix, err := configPropertyPrefix(methodName, args, receiverID)
	if err != nil {
		return "", err
	}
	property, err := configProperty(methodName, prefix, args)
	if err != nil {
		return "", err
	}

	return stringer(property), nil
}

func configPropertyPrefix(methodName string, args []string, receiverID component.ID) (string, error) {
	l := len(args)
	if l < 2 {
		return "", fmt.Errorf("%s takes key+ and value{1} arguments (minimum 2)", methodName)
	}
	prefix := fmt.Sprintf("splunk.discovery.receivers.%s.config", receiverID)
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
