package main

import (
	"fmt"
	"os"
	"strings"
)

const agentConfigGatewayDisabledPath = "../otelcol/config/collector/agent_config.yaml"
const agentConfigGatewayEnabledPath = "../otelcol/config/collector/agent_to_gateway_config.yaml"
const agentConfigSourcePath = "./agent_config_source.yaml"
const blockEndMarker = "# <---->"
const gatewayDisabledMarker = "# <-- gateway disabled -->"
const gatewayEnabledMarker = "# <-- gateway enabled -->"
const generatedCommentMarker = "# <-- generated comment -->"
const leadingLineBreakCharacters = "\r\n"
const lineBreak = "\n"

type configGenerator struct {
	configSourcePath     string
	sourceConfigContents string
}

func newConfigGenerator(configSourcePath string) *configGenerator {
	return &configGenerator{
		configSourcePath: configSourcePath,
	}
}

func (generator *configGenerator) loadSourceConfigFile() error {
	source, err := os.ReadFile(generator.configSourcePath)
	if err != nil {
		return err
	}
	generator.sourceConfigContents = string(source)
	return nil
}

func (generator *configGenerator) writeGeneratedConfigFile(generatedConfig []byte, filePath string) error {
	return os.WriteFile(filePath, generatedConfig, 0644)
}

func (generator *configGenerator) convertSourceConfig(gatewayEnabled bool) ([]byte, error) {
	var configFileContents strings.Builder
	insideGatewayDisabledOnlyBlock := false
	insideGatewayOnlyBlock := false
	insideCommentBlock := false

	for _, line := range strings.Split(generator.sourceConfigContents, lineBreak) {
		// Need to keep track of whether the current line is in a specialized block
		if strings.Contains(line, blockEndMarker) {
			// Nesting specialized blocks is not allowed, any ending marker means all special sections are over.
			insideGatewayDisabledOnlyBlock = false
			insideGatewayOnlyBlock = false
			insideCommentBlock = false
			continue
		}
		if strings.Contains(line, gatewayDisabledMarker) {
			insideGatewayDisabledOnlyBlock = true
			continue
		}
		if strings.Contains(line, gatewayEnabledMarker) {
			insideGatewayOnlyBlock = true
			continue
		}
		if strings.Contains(line, generatedCommentMarker) {
			insideCommentBlock = true
			continue
		}

		// Cases where we can skip the current line
		if insideCommentBlock {
			continue
		}
		if gatewayEnabled && insideGatewayDisabledOnlyBlock {
			continue
		}
		if !gatewayEnabled && insideGatewayOnlyBlock {
			continue
		}

		configFileContents.WriteString(line + lineBreak)
	}

	return []byte(configFileContents.String()), nil
}

func (generator *configGenerator) validateMatchingConfigMarkers() error {
	if generator.sourceConfigContents == "" {
		return fmt.Errorf("no config source defined")
	}
	markupEndCount := 0
	markupStartCount := 0
	for _, line := range strings.Split(generator.sourceConfigContents, lineBreak) {
		if strings.Contains(line, blockEndMarker) {
			markupEndCount++
		} else if strings.Contains(line, gatewayDisabledMarker) ||
			strings.Contains(line, gatewayEnabledMarker) ||
			strings.Contains(line, generatedCommentMarker) {
			markupStartCount++
		}
	}
	if markupStartCount != markupEndCount {
		return fmt.Errorf("Mismatch count of marker start blocks and marker end blocks: %d != %d", markupStartCount, markupEndCount)
	}
	return nil
}

func (generator *configGenerator) validateSourceConfigFile() error {
	return generator.validateMatchingConfigMarkers()
}

func main() {
	generator := newConfigGenerator(agentConfigSourcePath)
	err := generator.loadSourceConfigFile()
	if err != nil {
		fmt.Printf("Error loading source config file: %v\n", err)
		os.Exit(1)
	}
	err = generator.validateSourceConfigFile()
	if err != nil {
		fmt.Printf("Error validating source config file: %v\n", err)
		os.Exit(1)
	}

	agentWithoutGateway, err := generator.convertSourceConfig(false)
	if err != nil {
		fmt.Printf("Error generating agent without gateway config file: %v\n", err)
		os.Exit(1)
	}
	generator.writeGeneratedConfigFile(agentWithoutGateway, agentConfigGatewayDisabledPath)

	agentToGateway, err := generator.convertSourceConfig(true)
	if err != nil {
		fmt.Printf("Error generating agent to gateway config file: %v\n", err)
		os.Exit(1)
	}
	generator.writeGeneratedConfigFile(agentToGateway, agentConfigGatewayEnabledPath)
}
