package auth

import (
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultCertPoolMustBeSystemPool(t *testing.T) {
	// Get a reference system pool for comparison
	systemPool, sysErr := x509.SystemCertPool()
	require.NoError(t, sysErr)
	require.NotNil(t, systemPool)

	defaultCertPool, err := CertPool()
	require.NoError(t, err)
	require.NotNil(t, defaultCertPool)
	require.Equal(t, systemPool, defaultCertPool)
}
