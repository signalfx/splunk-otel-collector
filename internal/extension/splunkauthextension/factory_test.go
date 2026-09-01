// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestFactory_Create(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	ext, err := createExtension(t.Context(), extensiontest.NewNopSettings(extensiontest.NopType), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
}
