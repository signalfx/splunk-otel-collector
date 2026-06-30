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

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const generatedCommentMarker = "# <-- generated comment -->"
const generatedCommentEndMarker = "# <-- generated comment end -->"

const gatewayDisabledMarker = "# <-- gateway disabled -->"
const gatewayDisabledEndMarker = "# <-- gateway disabled end -->"

const gatewayEnabledMarker = "# <-- gateway enabled -->"
const gatewayEnabledEndMarker = "# <-- gateway enabled end -->"

const lineBreak = "\n"

type configToGenerate struct {
	destinationPath         string
	generatedConfigContents *bytes.Buffer
	startMarker             string
	endMarker               string
	insideMarker            bool
}

func (cfg *configToGenerate) AppendGeneratedConfigContents(line string) {
	cfg.generatedConfigContents.WriteString(line + lineBreak)
}

func (cfg *configToGenerate) writeGeneratedConfigFile() error {
	return os.WriteFile(cfg.destinationPath, cfg.generatedConfigContents.Bytes(), 0o644)
}

type sourceTemplate struct {
	generateConfigs        []*configToGenerate
	sourceTemplatePath     string
	sourceTemplateContents string
}

func getAgentSourceTemplate() sourceTemplate {
	return sourceTemplate{
		generateConfigs: []*configToGenerate{
			{
				destinationPath:         filepath.Join("..", "otelcol", "config", "collector", "agent_config.yaml"),
				endMarker:               gatewayDisabledEndMarker,
				startMarker:             gatewayDisabledMarker,
				generatedConfigContents: bytes.NewBuffer(make([]byte, 0)),
			},
			{
				destinationPath:         filepath.Join("..", "otelcol", "config", "collector", "agent_to_gateway_config.yaml"),
				endMarker:               gatewayEnabledEndMarker,
				startMarker:             gatewayEnabledMarker,
				generatedConfigContents: bytes.NewBuffer(make([]byte, 0)),
			},
		},
		sourceTemplatePath: filepath.Join("config_templates", "agent_config_source.yaml.tmpl"),
	}
}

func (generator *sourceTemplate) loadSourceConfigFile() error {
	source, err := os.ReadFile(generator.sourceTemplatePath)
	if err != nil {
		return err
	}
	generator.sourceTemplateContents = string(source)
	return nil
}

func (generator *sourceTemplate) convertTemplate() {
	insideCommentBlock := false
	for _, line := range strings.Split(generator.sourceTemplateContents, lineBreak) {
		if strings.Contains(line, generatedCommentEndMarker) {
			// Everything inside of comments is disregarded, can safely reset all status markers
			// when comment is finished
			for _, config := range generator.generateConfigs {
				config.insideMarker = false
			}
			insideCommentBlock = false
			continue
		}
		if insideCommentBlock {
			continue
		}

		// Track start of special blocks
		if strings.Contains(line, generatedCommentMarker) {
			insideCommentBlock = true
			continue
		}

		lineContainsStartMarker := false
		for _, config := range generator.generateConfigs {
			if strings.Contains(line, config.startMarker) {
				config.insideMarker = true
				lineContainsStartMarker = true
				break
			}
		}
		if lineContainsStartMarker {
			continue
		}

		// Track end of special blocks
		lineContainsEndMarker := false
		for _, config := range generator.generateConfigs {
			if strings.Contains(line, config.endMarker) {
				config.insideMarker = false
				lineContainsEndMarker = true
				break
			}
		}
		if lineContainsEndMarker {
			continue
		}

		inSpecializedBlock := false
		for _, config := range generator.generateConfigs {
			if config.insideMarker {
				inSpecializedBlock = true
				config.AppendGeneratedConfigContents(line)
				break // Enforce only being in one special block at a time
			}
		}
		if inSpecializedBlock {
			continue
		}

		for _, config := range generator.generateConfigs {
			config.AppendGeneratedConfigContents(line)
		}
	}
}

func (generator *sourceTemplate) writeAllGeneratedConfigFiles() error {
	for _, config := range generator.generateConfigs {
		if err := config.writeGeneratedConfigFile(); err != nil {
			return err
		}
	}
	return nil
}

func (generator *sourceTemplate) validateSourceConfigFile() error {
	if generator.sourceTemplatePath == "" {
		return fmt.Errorf("source template is empty")
	}

	return nil
}

func main() {
	agentTemplate := getAgentSourceTemplate()

	err := agentTemplate.loadSourceConfigFile()
	if err != nil {
		fmt.Printf("Error loading source config file: %v\n", err)
		os.Exit(1)
	}
	err = agentTemplate.validateSourceConfigFile()
	if err != nil {
		fmt.Printf("Error validating source config file: %v\n", err)
		os.Exit(1)
	}

	agentTemplate.convertTemplate()
	err = agentTemplate.writeAllGeneratedConfigFiles()
	if err != nil {
		fmt.Printf("Error writing all generated config files: %v\n", err)
		os.Exit(1)
	}
}
