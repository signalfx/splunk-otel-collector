package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateMap(t *testing.T) {
	target := map[string]string{"foo": "bar"}
	updates := map[string]string{"baz": "glarch"}
	updateMap(target, updates)
	require.Equal(t, target, map[string]string{"foo": "bar", "baz": "glarch"})
}
