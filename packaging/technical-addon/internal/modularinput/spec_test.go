package modularinput

import (
	"github.com/splunk/splunk-technical-addon/internal/testcommon"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	actual, err := LoadConfig("./testdata/sample-modular-inputs.yaml")
	require.NoError(t, err)
	require.EqualValues(t,
		&TemplateData{
			ModularInputs: map[string]ModInputConfig{
				"everything_set":                 {Description: "SET ALL THE THINGS", Default: "$SPLUNK_OTEL_TA_HOME/local/access_token", Flag: Flag{Name: "test-flag", IsUnary: false}, Required: false, PassthroughEnvVar: true, ReplaceableEnvVar: true},
				"minimal_set":                    {Description: "This is all you need", Default: "", Flag: Flag{Name: "", IsUnary: false}, Required: false, PassthroughEnvVar: false, ReplaceableEnvVar: false},
				"minimal_set_required":           {Description: "hello", Default: "", Flag: Flag{Name: "", IsUnary: false}, Required: true, PassthroughEnvVar: false, ReplaceableEnvVar: false},
				"unary_flag_with_everything_set": {Description: "Unary flags don't take arguments/values and are either present or not", Default: "$SPLUNK_OTEL_TA_HOME/local/access_token", Flag: Flag{Name: "test-flag", IsUnary: true}, Required: false, PassthroughEnvVar: true, ReplaceableEnvVar: true},
			},
			SchemaName: "Sample_Addon",
			Version:    "1.2.3",
		}, actual)

}

func TestRenderTemplate(t *testing.T) {
	sampleTemplateData, err := LoadConfig("./testdata/sample-modular-inputs.yaml")
	require.NoError(t, err)
	actualRender := filepath.Join(t.TempDir(), "render_template.txt")
	err = RenderTemplate("./testdata/sample-template.tmpl", actualRender, sampleTemplateData)
	testcommon.AssertFilesMatch(t, "./testdata/expected-template-rendered.txt", actualRender)
}
