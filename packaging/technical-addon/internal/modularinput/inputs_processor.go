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
	"slices"
	"sort"
	"strings"

	"github.com/splunk/splunk-technical-addon/internal/addonruntime"
)

type ModInput struct {
	Value        string
	Transformers []TransformerFunc
	Config       ModInputConfig
}

// TransformerFunc is basically a reducer.. takes in "working" value of modinput string
type (
	TransformerFunc   func(value string) (string, error)
	ModinputProcessor struct {
		ModularInputs map[string]*ModInput
		SchemaName    string
	}
)

func (t *ModInput) TransformInputs(value string) (string, error) {
	t.Value = value
	for _, transformer := range t.Transformers {
		transformed, err := transformer(t.Value)
		if err != nil {
			return "", err
		}
		t.Value = transformed
	}
	return t.Value, nil
}

func NewModinputProcessor(schemaName string, inputs map[string]*ModInput) *ModinputProcessor {
	return &ModinputProcessor{
		SchemaName:    schemaName,
		ModularInputs: inputs,
	}
}

func (t *ModInput) RegisterTransformer(transformer TransformerFunc) {
	t.Transformers = append(t.Transformers, transformer)
}

func (mip *ModinputProcessor) ProcessXML(modInput *XMLInput) error {
	providedInputs := make(map[string]bool)

	for _, stanza := range modInput.Configuration.Stanzas {
		stanzaPrefix := fmt.Sprintf("%s://", mip.SchemaName)

		if strings.HasPrefix(stanza.Name, stanzaPrefix) {
			for _, param := range stanza.Params {
				if input, exists := mip.ModularInputs[param.Name]; exists {
					value, err := input.TransformInputs(param.Value)
					if err != nil {
						return fmt.Errorf("transform failed for input %s: %s", param.Name, err)
					}
					input.Value = value
				}
				providedInputs[param.Name] = true
			}
			break // I believe we should only handle one of these... do we need to look up my process name?
		}
	}
	missing := mip.GetMissingRequired(providedInputs)
	if missing != nil {
		return fmt.Errorf("missing required inputs: %v", missing)
	}
	return nil
}

func (mip *ModinputProcessor) GetFlags() []string {
	var flags []string
	keys := slices.Collect(maps.Keys(mip.ModularInputs))
	sort.Strings(keys)
	for _, modinputName := range keys {
		modularInput := mip.ModularInputs[modinputName]
		if modularInput.Config.Flag.Name != "" {
			flags = append(flags, fmt.Sprintf("--%s", modularInput.Config.Flag.Name))
			if !modularInput.Config.Flag.IsUnary {
				flags = append(flags, modularInput.Value)
			}
		}
	}
	return flags
}

func (mip *ModinputProcessor) GetEnvVars() []string {
	var envVars []string
	keys := slices.Collect(maps.Keys(mip.ModularInputs))
	sort.Strings(keys)
	for _, modinputName := range keys {
		modularInput := mip.ModularInputs[modinputName]
		if modularInput.Config.PassthroughEnvVar {
			envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(modinputName), modularInput.Value))
		}
	}
	return envVars
}

func DefaultReplaceEnvVarTransformer(original string) (string, error) {
	splunkTaPlatformHome, err := addonruntime.GetTaPlatformDir()
	if err != nil {
		return "", err
	}
	splunkTaHome, err := addonruntime.GetTaHome()
	if err != nil {
		return "", err
	}
	splunkHome, err := addonruntime.GetSplunkHome()
	if err != nil {
		return "", err
	}
	replacement := strings.ReplaceAll(original, "$SPLUNK_OTEL_TA_PLATFORM_HOME", splunkTaPlatformHome)
	replacement = strings.ReplaceAll(replacement, "$SPLUNK_OTEL_TA_HOME", splunkTaHome)
	replacement = strings.ReplaceAll(replacement, "$SPLUNK_HOME", splunkHome)
	return replacement, nil
}

func (mip *ModinputProcessor) GetMissingRequired(provided map[string]bool) []string {
	var missing []string
	for name, mi := range mip.ModularInputs {
		if _, given := provided[name]; mi.Config.Required && !given {
			missing = append(missing, fmt.Sprintf("modular input %s is required", name))
		}
	}
	return missing
}
