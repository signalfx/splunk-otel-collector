package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncateDimensionValuesInPlace(t *testing.T) {
	d := map[string]string{
		"a": strings.Repeat("a", 300),
		"b": strings.Repeat("b", 253),
	}
	TruncateDimensionValuesInPlace(d)
	require.Len(t, d["a"], 256)
	require.Len(t, d["b"], 253)
}
