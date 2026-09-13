// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package script

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func TestDetermineCommandNameWindows(t *testing.T) {
	path, err := DetermineCommandName("", conf.Input{Configuration: conf.Configuration{Stanza: conf.Stanza{Name: "monitor://C:\\foo\\bar.txt"}}})
	require.NoError(t, err)
	require.Equal(t, "C:\\foo\\bar.txt", path)
}

func TestDetermineCommandName(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expected    string
		errExpected string
	}{
		{
			"script",
			"script://./bin/foo.sh",
			filepath.Join("bin", "foo.sh"),
			"",
		},
		{
			"outside the base dir",
			"script://../foo.sh",
			"",
			func() string {
				abs, _ := filepath.Abs("..")
				return fmt.Sprintf("path '%s%cfoo.sh' is outside the base directory", filepath.Clean(abs), filepath.Separator)
			}(),
		},
		{
			"modinput",
			"foo",
			func() string {
				return filepath.Join("bin", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH), "foo")
			}(),
			"",
		},
		{
			"invalid scheme",
			"invalid://foo",
			"",
			`unknown scheme "invalid"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := conf.Input{
				Configuration: conf.Configuration{
					Stanza: conf.Stanza{
						Name: test.command,
					},
				},
			}

			cmd, err := DetermineCommandName("", input)
			if test.errExpected != "" {
				require.Error(t, err)
				require.Equal(t, test.errExpected, err.Error())
			} else {
				abs, _ := filepath.Abs(test.expected)
				require.Equal(t, abs, cmd)
			}
		})
	}
}
