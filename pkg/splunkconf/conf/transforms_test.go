// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTransforms(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "transforms.conf"))
	require.NoError(t, err)
	res, err := ReadTransforms(b)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t,
		Transform{
			Name:   "foo",
			Regex:  "\\s+(?<src>\\S+)\\s+(?:(?<src_host>[^\\s\\[\\]]+)",
			Format: "",
		},
		res[0])
}
