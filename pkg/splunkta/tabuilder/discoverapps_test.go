// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package tabuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeApp creates <home>/etc/apps/<name>, with an app.conf in the given layer
// when body is non-empty.
func writeApp(t *testing.T, home, name, layer, body string) {
	t.Helper()
	dir := filepath.Join(home, "etc", "apps", name, layer)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	if body != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "app.conf"), []byte(body), 0o600))
	}
}

// splunkd admits every app in the bundle regardless of name, so discovery must
// not filter on one. An app whose inputs.conf is skipped here contributes no
// inputs at all, silently.
func TestDiscoverAppsIgnoresAppNames(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{
		"Splunk_TA_otel",           // the historical splunk_ta_ shape
		"splunk-connect-for-otlp",  // hyphenated, shipped from packaging/otlpinput
		"my_custom_app",            // a customer app
		"SplunkUniversalForwarder", // stock app
	} {
		writeApp(t, home, name, "local", "")
	}

	apps, err := DiscoverApps(home)
	require.NoError(t, err)

	var names []string
	for _, a := range apps {
		names = append(names, filepath.Base(a))
	}
	require.ElementsMatch(t, []string{
		"Splunk_TA_otel",
		"splunk-connect-for-otlp",
		"my_custom_app",
		"SplunkUniversalForwarder",
	}, names)
}

func TestDiscoverAppsSkipsDotDirs(t *testing.T) {
	home := t.TempDir()
	writeApp(t, home, "real_app", "local", "")
	writeApp(t, home, ".hidden", "local", "")

	apps, err := DiscoverApps(home)
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
			home := t.TempDir()
			writeApp(t, home, "app_under_test", "default", tt.defaultConf)
			writeApp(t, home, "app_under_test", "local", tt.localConf)

			apps, err := DiscoverApps(home)
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
	apps, err := DiscoverApps(t.TempDir())
	require.NoError(t, err)
	require.Empty(t, apps)
}
