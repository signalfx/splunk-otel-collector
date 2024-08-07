// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Copied from https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/internal/sharedcomponent

package sharedcomponent

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
