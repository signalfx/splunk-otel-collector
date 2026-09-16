// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package scriptedinput

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
)

func TestBuild(t *testing.T) {
	c := NewConfig()
	o, err := c.Build(componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)
	require.NotNil(t, o)
}
