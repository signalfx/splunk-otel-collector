package modularinput

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TransformerFunc is basically a reducer.. takes in "working" value of modinput string
type TransformerFunc func(value string) (string, error)
type ModinputTransformer struct {
	SchemaName    string
	ModularInputs map[string]ModInput
}

func NewModinputTransformer(schemaName string, inputs map[string]ModInput) *ModinputTransformer {
	return &ModinputTransformer{
		SchemaName:    schemaName,
		ModularInputs: inputs,
	}
}

func (t *ModInput) RegisterTransformer(transformer TransformerFunc) {
	t.Transformers = append(t.Transformers, transformer)
}

func (mit *ModinputTransformer) Transform(modInput *XMLInput) error {

	for _, stanza := range modInput.Configuration.Stanzas {
		stanzaPrefix := fmt.Sprintf("%s://", mit.SchemaName)

		if strings.HasPrefix(stanza.Name, stanzaPrefix) {
			for _, param := range stanza.Params {
				if input, exists := mit.ModularInputs[param.Name]; exists {
					err := input.Transform(param.Value)
					if err != nil {
						return fmt.Errorf("transform failed for input %s: %s", param.Name, err)
					}
				} else {
					return fmt.Errorf("modinput %s does not exist", param.Name)
				}
			}
			break // I believe we should only handle one of these... do we need to look up my process name?
		}
	}
	return nil
}

func (mit *ModinputTransformer) GetFlags() []string {
	var flags []string
	for _, modularInput := range mit.ModularInputs {
		if "" != modularInput.Config.Flag.Name {
			flags = append(flags, fmt.Sprintf("%s%s", flagPrefix, modularInput.Config.Flag.Name))
			if !modularInput.Config.Flag.IsUnary {
				flags = append(flags, modularInput.Value)
			}
		}
	}
	return flags
}

func (mit *ModinputTransformer) GetEnvVars() []string {
	var envVars []string
	for modinputName, modularInput := range mit.ModularInputs {
		if modularInput.Config.PassthroughEnvVar {
			envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(modinputName), modularInput.Value))
		}
	}
	return envVars
}

func DefaultReplaceEnvVarTransformer(original string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Error getting executable path: %v\n", err)
	}
	splunkTaPlatformHome := filepath.Dir(filepath.Dir(execPath)) // ../bin/(windows_x86_64|linux_x86_64)
	splunkTaHome := filepath.Dir(splunkTaPlatformHome)           // ../<Name of TA>
	splunkHome := filepath.Dir(filepath.Dir(splunkTaHome))       // etc/(apps|deployment_apps)/

	replacement := strings.ReplaceAll(original, "$SPLUNK_TA_PLATFORM_HOME", splunkTaPlatformHome)
	replacement = strings.ReplaceAll(original, "$SPLUNK_TA_HOME", splunkTaHome)
	replacement = strings.ReplaceAll(original, "$SPLUNK_HOME", splunkHome)
	return replacement, nil
}
