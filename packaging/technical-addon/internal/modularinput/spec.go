// Copyright Splunk, Inc.
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

package modularinput

import (
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Flag struct {
	Name    string `yaml:"name,omitempty"`
	IsUnary bool   `yaml:"is-unary,omitempty"`
}

type ModInputConfig struct {
	Description       string `yaml:"description"`
	Default           string `yaml:"default,omitempty"`
	Flag              Flag   `yaml:"flag,omitempty"`
	Required          bool   `yaml:"required,omitempty"`
	PassthroughEnvVar bool   `yaml:"passthrough,omitempty"`
	ReplaceableEnvVar bool   `yaml:"replaceable,omitempty"`
}

type GenericModularInput struct {
	ModularInputs map[string]*ModInput
	SchemaName    string
}

type TemplateData struct {
	ModularInputs map[string]ModInputConfig `yaml:"modular-inputs"`
	SchemaName    string                    `yaml:"modular-input-schema-name"`
	Version       string                    `yaml:"version"`
}

func LoadConfig(yamlPath string) (*TemplateData, error) {
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config TemplateData
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if config.SchemaName == "" {
		return nil, fmt.Errorf("modular-input-schema-name is empty in file: %s", yamlPath)
	}

	return &config, nil
}

func RenderTemplate(templatePath, outputPath string, data *TemplateData) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
