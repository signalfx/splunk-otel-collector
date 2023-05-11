package filter

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestOverridableStringFilter(t *testing.T) {
	for _, tc := range []struct {
		filter      []string
		inputs      []string
		shouldMatch []bool
		shouldError bool
	}{
		{
			filter:      []string{},
			inputs:      []string{"process_", "", "asdf"},
			shouldMatch: []bool{false, false, false},
		},
		{
			filter: []string{
				"*",
			},
			inputs:      []string{"app", "asdf", "", "*"},
			shouldMatch: []bool{true, true, true, true},
		},
		{
			filter: []string{
				"!app",
			},
			inputs:      []string{"app", "other"},
			shouldMatch: []bool{false, false},
		},
		{
			filter: []string{
				// A positive and negative literal match cancel each other out
				// and don't match.
				"app",
				"!app",
			},
			inputs:      []string{"app", "other"},
			shouldMatch: []bool{false, false},
		},
		{
			filter: []string{
				"other",
				"!app",
			},
			inputs:      []string{"other", "something", "app"},
			shouldMatch: []bool{true, false, false},
		},
		{
			filter: []string{
				"/^process_/",
				"/^node_/",
			},
			inputs:      []string{"process_", "node_", "process_asdf", "other"},
			shouldMatch: []bool{true, true, true, false},
		},
		{
			filter: []string{
				"!/^process_/",
			},
			inputs:      []string{"process_", "other"},
			shouldMatch: []bool{false, false},
		},
		{
			filter: []string{
				"app",
				"!/^process_/",
				"process_",
			},
			inputs:      []string{"other", "app", "process_cpu", "process_"},
			shouldMatch: []bool{false, true, false, false},
		},
		{
			filter: []string{
				"asdfdfasdf",
				"/^node_/",
			},
			inputs:      []string{"node_test"},
			shouldMatch: []bool{true},
		},
		{
			filter: []string{
				"process_*",
				"!process_cpu",
			},
			inputs:      []string{"process_mem", "process_cpu", "asdf"},
			shouldMatch: []bool{true, false, false},
		},
		{
			filter: []string{
				"*",
				"!process_cpu",
			},
			inputs:      []string{"process_mem", "process_cpu", "asdf"},
			shouldMatch: []bool{true, false, true},
		},
		{
			filter: []string{
				"metric_?",
				"!metric_a",
				"!metric_b",
				"random",
			},
			inputs:      []string{"metric_a", "metric_b", "metric_c", "asdf", "random"},
			shouldMatch: []bool{false, false, true, false, true},
		},
		{
			filter: []string{
				"!process_cpu",
				// Order doesn't matter
				"*",
			},
			inputs:      []string{"process_mem", "process_cpu", "asdf"},
			shouldMatch: []bool{true, false, true},
		},
		{
			filter: []string{
				"/a.*/",
				"!/.*z/",
				"b",
				// Static match should not override the negated regex above
				"alz",
			},
			inputs:      []string{"", "asdf", "asdz", "b", "wrong", "alz"},
			shouldMatch: []bool{false, true, false, true, false, false},
		},
	} {
		f, err := NewOverridableStringFilter(tc.filter)
		if tc.shouldError {
			assert.NotNil(t, err, spew.Sdump(tc))
		} else {
			assert.Nil(t, err, spew.Sdump(tc))
		}
		for i := range tc.inputs {
			assert.Equal(t, tc.shouldMatch[i], f.Matches(tc.inputs[i]), "input[%d] of %s\n%s", i, spew.Sdump(tc), spew.Sdump(f))
		}
	}
}
