// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package udpreceiver

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
