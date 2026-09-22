// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhome

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// nameRegexp mirrors component/identifiable.go: allows any rune that is not a
// separator, control, or symbol.
var nameRegexp = regexp.MustCompile(`^[^\pZ\pC\pS]+$`)

// TestStableNameInjective proves distinct stanza identities never collapse to
// the same component name, which would collide their checkpoints and receiver
// hashes. slug() alone is not injective; the hash suffix restores it.
func TestStableNameInjective(t *testing.T) {
	raws := []string{
		"monitor:///var/log/a b", // slug collides with next
		"monitor:///var/log/a_b", // -> same slug as above, different hash
		"monitor:///var/log/a-b",
		"monitor:///var/log/foo",
		"monitor:///VAR/LOG/FOO", // case-folds to same slug, different hash
		"tcp://5514",
		"tcp://5515",
		"!!!", // all-symbol target: slug empty, hash carries identity
		"@@@",
	}
	seen := map[string]string{}
	for _, raw := range raws {
		name := stableName("uf", raw)
		require.Truef(t, nameRegexp.MatchString(name), "name %q from %q must satisfy component name regex", name, raw)
		if prev, ok := seen[name]; ok {
			t.Fatalf("collision: %q and %q both map to %q", prev, raw, name)
		}
		seen[name] = raw
	}
}

// TestStableNameDeterministic proves the same stanza always yields the same
// name, which is required for the ID to be a stable checkpoint key across
// reloads.
func TestStableNameDeterministic(t *testing.T) {
	raw := "monitor:///var/log/foo"
	require.Equal(t, stableName("uf", raw), stableName("uf", raw))
}
