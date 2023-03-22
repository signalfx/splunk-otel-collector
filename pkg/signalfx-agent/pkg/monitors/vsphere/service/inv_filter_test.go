package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvFilter_Match(t *testing.T) {
	f, err := NewFilter("Datacenter == 'dc0' && Cluster == 'cluster0'")
	require.NoError(t, err)
	dims := pairs{
		pair{dimDatacenter, "dc0"},
		pair{dimCluster, "cluster0"},
	}
	keep, err := f.keep(dims)
	require.NoError(t, err)
	require.True(t, keep)
}

func TestInvFilter_NoMatch(t *testing.T) {
	f, err := NewFilter("Datacenter == 'dc0' && Cluster == 'xyz'")
	require.NoError(t, err)
	dims := pairs{
		pair{dimDatacenter, "dc0"},
		pair{dimCluster, "cluster0"},
	}
	keep, err := f.keep(dims)
	require.NoError(t, err)
	require.False(t, keep)
}

func TestInvFilter_DCMatch(t *testing.T) {
	f, err := NewFilter("Datacenter == 'dc0'")
	require.NoError(t, err)
	dims := pairs{
		pair{dimDatacenter, "dc0"},
		pair{dimCluster, "cluster0"},
	}
	keep, err := f.keep(dims)
	require.NoError(t, err)
	require.True(t, keep)
}

func TestInvFilter_ClusterMatch(t *testing.T) {
	f, err := NewFilter("Cluster == 'cluster0'")
	require.NoError(t, err)
	dims := pairs{
		pair{dimDatacenter, "dc0"},
		pair{dimCluster, "cluster0"},
	}
	keep, err := f.keep(dims)
	require.NoError(t, err)
	require.True(t, keep)
}
