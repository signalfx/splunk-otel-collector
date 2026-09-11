// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadProps(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "props.conf"))
	require.NoError(t, err)
	props, err := ReadProps(b)
	require.NoError(t, err)
	require.Len(t, props, 1)
	require.Equal(t, "scoreboard", props[0].Name)
	require.True(t, props[0].ShouldLineMerge)
}

func TestPropType(t *testing.T) {
	for _, tc := range []struct {
		name     string
		prop     Prop
		expected PropType
	}{
		{
			name:     "source",
			prop:     Prop{Name: "source::foo"},
			expected: Source,
		},
		{
			name:     "sourcetype",
			prop:     Prop{Name: "foo"},
			expected: SourceType,
		},
		{
			name:     "default",
			prop:     Prop{Name: "default"},
			expected: Default,
		},
		{
			name:     "host",
			prop:     Prop{Name: "host::foo"},
			expected: Host,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.prop.Type())
		})
	}
}

func TestOrderProps(t *testing.T) {
	props := []Prop{
		{
			Name: "source::foo",
		},
		{
			Name: "source::bar",
		},
		{
			Name: "host::foo",
		},
		{
			Name: "default",
		},
		{
			Name: "host::bar",
		},
		{
			Name: "foo",
		},
		{
			Name: "bar",
		},
	}

	orderProps(props)

	require.Equal(t, []Prop{
		{
			Name: "source::bar",
		},
		{
			Name: "source::foo",
		},
		{
			Name: "host::bar",
		},
		{
			Name: "host::foo",
		},
		{
			Name: "bar",
		},
		{
			Name: "foo",
		},
		{
			Name: "default",
		},
	}, props)
}

func TestParseFieldAliasExpr(t *testing.T) {
	from, to := parseFieldAliasExpr("foo as bar")
	assert.Equal(t, from, "foo")
	assert.Equal(t, to, "bar")
}

func TestReadFieldAliases(t *testing.T) {
	f := ini.Empty()
	s := f.Section("foo")
	_, _ = s.NewKey("FIELDALIAS-src", "senderIP as src")
	_, _ = s.NewKey("FIELDALIAS-src_user", "sender as src_user")

	aliases := readFieldAliases(s)
	require.Len(t, aliases, 2)
	require.Equal(t, "FIELDALIAS-src", aliases[0].Name)
	require.Equal(t, "senderIP", aliases[0].From)
	require.Equal(t, "src", aliases[0].To)
	require.Equal(t, "sender", aliases[1].From)
	require.Equal(t, "src_user", aliases[1].To)
}
