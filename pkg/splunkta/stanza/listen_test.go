// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package stanza

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseListenAddress(t *testing.T) {
	for _, tt := range []struct {
		name     string
		target   string
		wantAddr string
		wantPort int
	}{
		{
			name:     "address and port",
			target:   "10.0.0.5:5514",
			wantAddr: "10.0.0.5",
			wantPort: 5514,
		},
		{
			name:     "port only means every interface",
			target:   "5514",
			wantAddr: "0.0.0.0",
			wantPort: 5514,
		},
		{
			name:     "empty address means every interface",
			target:   ":5514",
			wantAddr: "0.0.0.0",
			wantPort: 5514,
		},
		{
			// A zero port fails the wrapper's own Validate, so a malformed stanza
			// surfaces at config time rather than binding something arbitrary.
			name:     "out of range port",
			target:   "99999",
			wantAddr: "0.0.0.0",
			wantPort: 0,
		},
		{
			name:     "non-numeric port",
			target:   "0.0.0.0:nope",
			wantAddr: "0.0.0.0",
			wantPort: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			addr, port := ParseListenAddress(tt.target)
			require.Equal(t, tt.wantAddr, addr)
			require.Equal(t, tt.wantPort, port)
		})
	}
}
