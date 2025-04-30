package packaging

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSha256(t *testing.T) {
	sum, err := Sha256Sum("./testdata/sampletosha.txt")
	require.NoError(t, err)
	// Expected generated with linux sha256sum utility
	assert.Equal(t, "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f", sum)
}
