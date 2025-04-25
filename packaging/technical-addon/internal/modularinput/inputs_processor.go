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
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type ModInput struct {
	Value        string
	Transformers []TransformerFunc
	Config       ModInputConfig
}

// TransformerFunc is basically a reducer.. takes in "working" value of modinput string
type TransformerFunc func(value string) (string, error)
type ModinputProcessor struct {
	ModularInputs map[string]ModInput
	SchemaName    string
}

func (t *ModInput) TransformInputs(value string) error {
	t.Value = value
	for _, transformer := range t.Transformers {
		transformed, err := transformer(t.Value)
		if err != nil {
			return err
		}
		t.Value = transformed
	}
	return nil
}

func NewModinputProcessor(schemaName string, inputs map[string]ModInput) *ModinputProcessor {
	return &ModinputProcessor{
		SchemaName:    schemaName,
		ModularInputs: inputs,
	}
}

func (t *ModInput) RegisterTransformer(transformer TransformerFunc) {
	t.Transformers = append(t.Transformers, transformer)
}

func (mit *ModinputProcessor) ProcessXML(modInput *XMLInput) error {
	providedInputs := make(map[string]bool)

	for _, stanza := range modInput.Configuration.Stanzas {
		stanzaPrefix := fmt.Sprintf("%s://", mit.SchemaName)

		if strings.HasPrefix(stanza.Name, stanzaPrefix) {
			for _, param := range stanza.Params {
				if input, exists := mit.ModularInputs[param.Name]; exists {
					err := input.TransformInputs(param.Value)
					if err != nil {
						return fmt.Errorf("transform failed for input %s: %s", param.Name, err)
					}
				}
				providedInputs[param.Name] = true
			}
			break // I believe we should only handle one of these... do we need to look up my process name?
		}
	}
	missing := mit.GetMissingRequired(providedInputs)
	if missing != nil {
		return fmt.Errorf("missing required inputs: %v", missing)
	}
	return nil
}

func (mit *ModinputProcessor) GetFlags() []string {
	var flags []string
	keys := slices.Collect(maps.Keys(mit.ModularInputs))
	sort.Strings(keys)
	for _, modinputName := range keys {
		modularInput := mit.ModularInputs[modinputName]
		if "" != modularInput.Config.Flag.Name {
			flags = append(flags, fmt.Sprintf("--%s", modularInput.Config.Flag.Name))
			if !modularInput.Config.Flag.IsUnary {
				flags = append(flags, modularInput.Value)
			}
		}
	}
	return flags
}

func (mit *ModinputProcessor) GetEnvVars() []string {
	var envVars []string
	keys := slices.Collect(maps.Keys(mit.ModularInputs))
	sort.Strings(keys)
	for _, modinputName := range keys {
		modularInput := mit.ModularInputs[modinputName]
		if modularInput.Config.PassthroughEnvVar {
			envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(modinputName), modularInput.Value))
		}
	}
	return envVars
}

func (mit *ModinputProcessor) GetModularInputs() {

}

func DefaultReplaceEnvVarTransformer(original string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %v", err)
	}
	splunkTaPlatformHome := filepath.Dir(filepath.Dir(execPath)) // ../bin/(windows_x86_64|linux_x86_64)
	splunkTaHome := filepath.Dir(splunkTaPlatformHome)           // ../<Name of TA>
	splunkHome := filepath.Dir(filepath.Dir(splunkTaHome))       // etc/(apps|deployment_apps)/

	replacement := strings.ReplaceAll(original, "$SPLUNK_TA_PLATFORM_HOME", splunkTaPlatformHome)
	replacement = strings.ReplaceAll(replacement, "$SPLUNK_TA_HOME", splunkTaHome)
	replacement = strings.ReplaceAll(replacement, "$SPLUNK_HOME", splunkHome)
	return replacement, nil
}

func (mit *ModinputProcessor) GetMissingRequired(provided map[string]bool) []string {
	var missing []string
	for name, mi := range mit.ModularInputs {
		if _, given := provided[name]; mi.Config.Required && !given {
			missing = append(missing, fmt.Sprintf("modular input %s is required", name))
		}
	}
	return missing
}
