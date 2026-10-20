// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputParams(t *testing.T) {
	for _, tt := range []struct {
		name   string
		attrs  ResourceAttrs
		extra  map[string]string
		expect Params
	}{
		{
			name:   "empty",
			expect: Params{},
		},
		{
			name:  "resource attrs in fixed order",
			attrs: ResourceAttrs{Index: "i", Source: "s", Sourcetype: "st", Host: "h"},
			expect: Params{
				{Name: "index", Value: "i"},
				{Name: "source", Value: "s"},
				{Name: "sourcetype", Value: "st"},
				{Name: "host", Value: "h"},
			},
		},
		{
			name:  "unset attrs emit no param",
			attrs: ResourceAttrs{Sourcetype: "st"},
			expect: Params{
				{Name: "sourcetype", Value: "st"},
			},
		},
		{
			name:  "extras sorted by name after attrs",
			attrs: ResourceAttrs{Index: "i"},
			extra: map[string]string{"zeta": "z", "alpha": "a", "mid": "m"},
			expect: Params{
				{Name: "index", Value: "i"},
				{Name: "alpha", Value: "a"},
				{Name: "mid", Value: "m"},
				{Name: "zeta", Value: "z"},
			},
		},
		{
			name:  "extra may carry an empty value",
			extra: map[string]string{"followTail": ""},
			expect: Params{
				{Name: "followTail", Value: ""},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, InputParams(tt.attrs, tt.extra))
		})
	}
}
