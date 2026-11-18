// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/conf"
)

func requireOutputParam(t *testing.T, output *conf.Output, name string) string {
	t.Helper()
	param := output.Configuration.Stanza.Params.Get(name)
	require.NotNil(t, param)
	return param.Value
}

func TestReadOutputs(t *testing.T) {
	tests := []struct {
		expectedErr   error
		name          string
		confRoot      string
		expectedToken string
		expectedURI   string
	}{
		{
			name:          "system_default_only",
			confRoot:      "system_default",
			expectedToken: "token-system-default",
			expectedURI:   "https://system-default:8088/services/collector/event",
		},
		{
			name:          "app_overrides_system_default",
			confRoot:      "app_overrides",
			expectedToken: "token-app-default",
			expectedURI:   "https://app-local:8088/services/collector/event",
		},
		{
			name:          "system_local_wins_over_app",
			confRoot:      "system_local_wins",
			expectedToken: "token-system-local",
			expectedURI:   "https://system-local:8088/services/collector/event",
		},
		{
			name:          "app_local_wins_over_app_default",
			confRoot:      "app_local_wins_over_app_default",
			expectedToken: "token-app-local",
			expectedURI:   "https://app-local:8088/services/collector/event",
		},
		{
			name:        "no_hecout",
			confRoot:    "no_hecout",
			expectedErr: conf.ErrNoHECOut,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merged, err := ReadOutputs(ConfRootDirs(filepath.Join("testdata", "outputs", test.confRoot, "etc")))
			if test.expectedErr != nil {
				require.NoError(t, err) // ReadOutputs itself doesn't error on missing stanza
				_, hecErr := conf.HECOut(merged)
				require.ErrorIs(t, hecErr, test.expectedErr)
				return
			}
			require.NoError(t, err)
			output, err := conf.HECOut(merged)
			require.NoError(t, err)
			require.NotNil(t, output)
			assert.Equal(t, test.expectedToken, requireOutputParam(t, output, "httpEventCollectorToken"))
			assert.Equal(t, test.expectedURI, requireOutputParam(t, output, "uri"))
		})
	}
}

// An app's inputs resolve against the app's own layers merged with system, in
// btool precedence order.
func TestReadInputs(t *testing.T) {
	tests := []struct {
		name               string
		confRoot           string
		expectedStanza     string
		expectedIndex      string
		expectedSourcetype string
		expectedCount      int
	}{
		{
			name:           "ta_default_only",
			confRoot:       "ta_default_only",
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "main",
		},
		{
			name:           "ta_local_overrides_ta_default",
			confRoot:       "ta_local_overrides",
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "local_index",
		},
		{
			name:           "ta_default_wins_over_system_default",
			confRoot:       "system_default_overrides",
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "ta_index",
		},
		{
			name:           "system_local_wins_over_all",
			confRoot:       "system_local_wins",
			expectedCount:  1,
			expectedStanza: "monitor:///var/log/syslog",
			expectedIndex:  "system_local_index",
		},
		{
			name:               "system_default_adds_missing_param",
			confRoot:           "system_adds_param",
			expectedCount:      1,
			expectedStanza:     "monitor:///var/log/syslog",
			expectedSourcetype: "syslog",
			expectedIndex:      "system_index",
		},
		{
			name:          "no_inputs_returns_nil",
			confRoot:      "no_inputs",
			expectedCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confRoot := filepath.Join("testdata", "inputs", test.confRoot, "etc")
			appDir := filepath.Join(confRoot, "apps", "splunk_ta_test")
			inputs, _, err := ReadInputs(AppConfRootDirs(confRoot, appDir))
			require.NoError(t, err)
			require.Len(t, inputs, test.expectedCount)
			if test.expectedCount == 0 {
				return
			}
			assert.Equal(t, test.expectedStanza, inputs[0].Configuration.Stanza.Name)
			if test.expectedIndex != "" {
				param := inputs[0].Configuration.Stanza.Params.Get("index")
				require.NotNil(t, param)
				assert.Equal(t, test.expectedIndex, param.Value)
			}
			if test.expectedSourcetype != "" {
				param := inputs[0].Configuration.Stanza.Params.Get("sourcetype")
				require.NotNil(t, param)
				assert.Equal(t, test.expectedSourcetype, param.Value)
			}
		})
	}
}

func TestReadTransforms(t *testing.T) {
	tests := []struct {
		name          string
		confRoot      string
		expectedName  string
		expectedRegex string
	}{
		{
			name:          "default",
			confRoot:      "splunk_default",
			expectedName:  "example_default",
			expectedRegex: "default",
		},
		{
			name:          "local",
			confRoot:      "splunk_local",
			expectedName:  "example_local",
			expectedRegex: "local",
		},
		{
			name:          "both",
			confRoot:      "splunk_both",
			expectedName:  "example_transform",
			expectedRegex: "local",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transforms, err := ReadTransforms(ConfRootDirs(filepath.Join("testdata", "transforms", test.confRoot, "etc")))
			require.NoError(t, err)
			require.Len(t, transforms, 1)
			require.Equal(t, test.expectedName, transforms[0].Name)
			require.Equal(t, test.expectedRegex, transforms[0].Regex)
		})
	}
}

// writeApp creates <confRoot>/apps/<name>, with an app.conf in the given layer
// when body is non-empty.
func writeApp(t *testing.T, confRoot, name, layer, body string) {
	t.Helper()
	dir := filepath.Join(confRoot, "apps", name, layer)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	if body != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "app.conf"), []byte(body), 0o600))
	}
}

func baseNames(dirs []string) []string {
	var names []string
	for _, d := range dirs {
		names = append(names, filepath.Base(d))
	}
	return names
}

// A disabled app contributes no config but still has to be watched: enabling it
// is an edit inside its own directory, and the apps watch is not recursive.
func TestAppDirsKeepsDisabledApps(t *testing.T) {
	confRoot := t.TempDir()
	writeApp(t, confRoot, "enabled_app", "default", "[install]\nstate = enabled\n")
	writeApp(t, confRoot, "disabled_app", "default", "[install]\nstate = disabled\n")

	all, err := appDirs(confRoot)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"enabled_app", "disabled_app"}, baseNames(all))

	enabled, err := discoverApps(confRoot)
	require.NoError(t, err)
	require.Equal(t, []string{"enabled_app"}, baseNames(enabled))
}

// splunkd admits every app in the bundle regardless of name, so discovery must
// not filter on one. An app whose inputs.conf is skipped here contributes no
// inputs at all, silently.
func TestDiscoverAppsIgnoresAppNames(t *testing.T) {
	confRoot := t.TempDir()
	for _, name := range []string{
		"Splunk_TA_otel",           // the historical splunk_ta_ shape
		"splunk-connect-for-otlp",  // hyphenated, shipped from packaging/otlpinput
		"my_custom_app",            // a customer app
		"SplunkUniversalForwarder", // stock app
	} {
		writeApp(t, confRoot, name, "local", "")
	}

	apps, err := discoverApps(confRoot)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		"Splunk_TA_otel",
		"splunk-connect-for-otlp",
		"my_custom_app",
		"SplunkUniversalForwarder",
	}, baseNames(apps))
}

func TestDiscoverAppsSkipsDotDirs(t *testing.T) {
	confRoot := t.TempDir()
	writeApp(t, confRoot, "real_app", "local", "")
	writeApp(t, confRoot, ".hidden", "local", "")

	apps, err := discoverApps(confRoot)
	require.NoError(t, err)
	require.Len(t, apps, 1)
	require.Equal(t, "real_app", filepath.Base(apps[0]))
}

// app.conf [install] state gates the app, and local wins over default, so an
// app disabled by its own default can be re-enabled locally.
//
// splunkd allowlists the exact literal "enabled" instead of looking for
// "disabled", so the cases that matter are the ones where a value is neither:
// they exclude the app rather than leaving it in.
func TestDiscoverAppsHonorsDisabledState(t *testing.T) {
	const disabled = "[install]\nstate = disabled\n"
	const enabled = "[install]\nstate = enabled\n"

	for _, tt := range []struct {
		name        string
		defaultConf string
		localConf   string
		wantFound   bool
	}{
		{name: "no app.conf is enabled", wantFound: true},
		{name: "state unset is enabled", defaultConf: "[install]\nis_configured = false\n", wantFound: true},
		{name: "disabled in default", defaultConf: disabled, wantFound: false},
		{name: "disabled in local", localConf: disabled, wantFound: false},
		{name: "local re-enables a disabled default", defaultConf: disabled, localConf: enabled, wantFound: true},
		{name: "local disables an enabled default", defaultConf: enabled, localConf: disabled, wantFound: false},
		{name: "uppercase ENABLED does not enable", defaultConf: "[install]\nstate = ENABLED\n", wantFound: false},
		{name: "a boolean-looking value does not enable", defaultConf: "[install]\nstate = 1\n", wantFound: false},
		{name: "an unrecognized value does not enable", defaultConf: "[install]\nstate = installed\n", wantFound: false},
		{name: "malformed app.conf leaves it enabled", defaultConf: "not a conf file\n", wantFound: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			confRoot := t.TempDir()
			writeApp(t, confRoot, "app_under_test", "default", tt.defaultConf)
			writeApp(t, confRoot, "app_under_test", "local", tt.localConf)

			apps, err := discoverApps(confRoot)
			require.NoError(t, err)
			if tt.wantFound {
				require.Len(t, apps, 1)
				return
			}
			require.Empty(t, apps)
		})
	}
}

func TestDiscoverAppsMissingAppsDir(t *testing.T) {
	apps, err := discoverApps(t.TempDir())
	require.NoError(t, err)
	require.Empty(t, apps)
}

// writeSpec declares schemes in dir's README/inputs.conf.spec, between the prose
// and the non-modular headers a real spec file carries.
func writeSpec(t *testing.T, dir string, schemes ...string) {
	t.Helper()
	var body strings.Builder
	body.WriteString("# A comment.\n[default]\n* Not a scheme.\n\n[splunktcp]\n[tcp:<port>]\n")
	for _, s := range schemes {
		body.WriteString("[" + s + "://<name>]\n* A scheme.\n\nparam = <value>\n* A setting.\n\n")
	}
	readme := filepath.Join(dir, "README")
	require.NoError(t, os.MkdirAll(readme, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(readme, "inputs.conf.spec"), []byte(body.String()), 0o600))
}

// A scheme belongs to the first app that declares it, reading system and then
// the apps in byte order, and a disabled app that wins blocks the scheme.
func TestSchemeAppDirs(t *testing.T) {
	confRoot := t.TempDir()
	app := func(name string) string { return filepath.Join(confRoot, "apps", name) }

	writeSpec(t, filepath.Join(confRoot, "system"), "WinEventLog", "shared")
	// system is never disabled, whatever its app.conf says.
	sysDefault := filepath.Join(confRoot, "system", "default")
	require.NoError(t, os.MkdirAll(sysDefault, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sysDefault, "app.conf"), []byte("[install]\nstate = disabled\n"), 0o600))
	writeSpec(t, app("my_ta"), "mycollector", "shared")

	// Byte order puts uppercase before lowercase and "app10" before "app9".
	writeSpec(t, app("aapp"), "casefirst")
	writeSpec(t, app("Zapp"), "casefirst")
	writeSpec(t, app("app9"), "numeric")
	writeSpec(t, app("app10"), "numeric")

	writeSpec(t, app("a_disabled"), "blocked")
	writeApp(t, confRoot, "a_disabled", "local", "[install]\nstate = disabled\n")
	writeSpec(t, app("b_enabled"), "blocked")

	// A dotfile directory sorts first and is never disabled, whatever its
	// app.conf says.
	writeSpec(t, app(".hidden"), "dotfirst")
	writeApp(t, confRoot, ".hidden", "local", "[install]\nstate = disabled\n")
	writeSpec(t, app("aaa"), "dotfirst")

	writeApp(t, confRoot, "no_readme", "local", "")
	require.NoError(t, os.WriteFile(app("not_a_dir"), nil, 0o600))

	got, err := schemeAppDirs(confRoot)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"WinEventLog": filepath.Join(confRoot, "system"),
		"shared":      filepath.Join(confRoot, "system"),
		"mycollector": app("my_ta"),
		"casefirst":   app("Zapp"),
		"numeric":     app("app10"),
		"blocked":     "",
		"dotfirst":    app(".hidden"),
	}, got)
}

// A spec file is read the way splunkd reads one, which decides what counts as a
// declaration.
func TestDeclaredSchemes(t *testing.T) {
	for _, tt := range []struct {
		name string
		spec string
		want []string
	}{
		{name: "a header and a setting", spec: "[a://<name>]\nparam = <value>\n", want: []string{"a"}},
		{name: "a header alone declares nothing", spec: "[a://<name>]\n* Prose.\n"},
		{
			name: "only an invalid parameter name declares nothing",
			spec: "[a://<name>]\npython.version = python3\n* Defaults to x = 1.\nmy param = 1\n",
		},
		{name: "a commented setting declares nothing", spec: "[a://<name>]\n# param = 1\n; param = 1\n"},
		{name: "an indented setting counts", spec: "[a://<name>]\n   param = 1\n", want: []string{"a"}},
		{
			name: "an indented header leaves its settings with the stanza before",
			spec: "[a://<name>]\n  [b://<name>]\nparam = 1\n",
			want: []string{"a"},
		},
		{name: "text after the closing bracket", spec: "[a://<name>] trailing\nparam = 1\n", want: []string{"a"}},
		{name: "a hyphenated scheme", spec: "[my-scheme://default]\nparam = 1\n", want: []string{"my-scheme"}},
		{name: "CRLF line endings", spec: "[a://<name>]\r\nparam = 1\r\n", want: []string{"a"}},
		{
			name: "a name cut before :// is not a scheme",
			spec: "[a.b://x]\np = 1\n[a<b://x]\np = 1\n[a:b://x]\np = 1\n[_a://x]\np = 1\n[://x]\np = 1\n",
		},
		{name: "not a scheme stanza", spec: "[default]\np = 1\n[tcp:<port>]\np = 1\n[splunktcp]\np = 1\n"},
		{
			name: "every scheme in file order, once each",
			spec: "[b://<name>]\np = 1\nq = 2\n\n[a://<name>]\np = 1\n",
			want: []string{"b", "a"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "inputs.conf.spec")
			require.NoError(t, os.WriteFile(path, []byte(tt.spec), 0o600))
			assert.Equal(t, tt.want, declaredSchemes(path))
		})
	}
}

func TestSchemeAppDirsEmptyRoot(t *testing.T) {
	got, err := schemeAppDirs(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, got)
}
