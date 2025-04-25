package modularinput

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ModInput struct {
	Config       ModInputConfig
	Transformers []TransformerFunc
	Value        string
}

// TransformerFunc is basically a reducer.. takes in "working" value of modinput string
type TransformerFunc func(value string) (string, error)
type ModinputProcessor struct {
	SchemaName    string
	ModularInputs map[string]ModInput
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

func (mit *ModinputProcessor) ProcessXml(modInput *XMLInput) error {
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

func (mit *ModinputProcessor) GetEnvVars() []string {
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

func (mit *ModinputProcessor) GetMissingRequired(provided map[string]bool) []string {
	var missing []string
	for name, mi := range mit.ModularInputs {
		if _, given := provided[name]; mi.Config.Required && !given {
			missing = append(missing, fmt.Sprintf("modular input %s is required", name))
		}
	}
	return missing
}

func GetBuildDir() string {
	var err error
	buildDir := os.Getenv("BUILD_DIR")
	if buildDir == "" {
		fmt.Println("$BUILD_DIR not set, searching for git root")
		buildDir, err = findGitRoot(".")
		if err != nil || buildDir == "" {
			fmt.Println("no git dir found, defaulting to cwd")
			buildDir, err = os.Getwd()
			if err != nil {
				fmt.Println("couldn't get cwd, defaulting to ./")
				buildDir = "."
			}
		}
		buildDir = filepath.Join(buildDir, "build")
	}
	return buildDir
}

func findGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		_, err := os.Stat(gitPath)
		if err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for .git: %w", err)
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("no git repository found")
		}
		dir = parentDir
	}
}
