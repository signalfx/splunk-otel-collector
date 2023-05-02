package filter

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestBasicStringFilter(t *testing.T) {
	for _, tc := range []struct {
		filter      []string
		input       string
		shouldMatch bool
		shouldError bool
	}{
		{
			filter:      []string{},
			input:       "process_",
			shouldMatch: false,
		},
		{
			filter: []string{
				"!app",
			},
			input:       "app",
			shouldMatch: false,
		},
		{
			filter: []string{
				"!app",
			},
			input:       "something",
			shouldMatch: true,
		},
		{
			filter: []string{
				"other",
				"!app",
			},
			input:       "something",
			shouldMatch: true,
		},
		{
			filter: []string{
				"other",
				"!app",
			},
			input:       "app",
			shouldMatch: false,
		},
		{
			filter: []string{
				"/^process_/",
				"/^node_/",
			},
			input:       "process_",
			shouldMatch: true,
		},
		{
			filter: []string{
				"!/^process_/",
			},
			input:       "process_",
			shouldMatch: false,
		},
		{
			filter: []string{
				"!app",
				"!/^process_/",
			},
			input:       "other",
			shouldMatch: true,
		},
		{
			filter: []string{
				"!other",
				"!/^process_/",
			},
			input: "other",
			// Since "other" is explicitly excluded, it should not ever match.
			shouldMatch: false,
		},
		{
			filter: []string{
				"app",
				"!/^process_/",
			},
			input:       "other",
			shouldMatch: true,
		},
		{
			filter: []string{
				"asdfdfasdf",
				"!/^node_/",
			},
			input:       "process_",
			shouldMatch: true,
		},
		{
			filter: []string{
				"asdfdfasdf",
				"/^node_/",
			},
			input:       "process_",
			shouldMatch: false,
		},
	} {
		f, err := NewBasicStringFilter(tc.filter)
		if tc.shouldError {
			assert.NotNil(t, err, spew.Sdump(tc))
		} else {
			assert.Nil(t, err, spew.Sdump(tc))
		}

		assert.Equal(t, tc.shouldMatch, f.Matches(tc.input), "%s\n%s", spew.Sdump(tc), spew.Sdump(f))
	}
}

func TestStringMapFilter(t *testing.T) {
	for _, tc := range []struct {
		filter      map[string][]string
		input       map[string]string
		shouldMatch bool
		shouldError bool
	}{
		{
			filter: map[string][]string{},
			input:  map[string]string{},
			// Empty map never matches anything, even blank filter
			shouldMatch: false,
		},
		{
			filter: map[string][]string{
				"app": {"test"},
			},
			input:       map[string]string{},
			shouldMatch: false,
		},
		{
			filter: map[string][]string{
				"app?": {"test"},
			},
			input:       map[string]string{},
			shouldMatch: true,
		},
		{
			filter: map[string][]string{
				"app?": {"test"},
			},
			input: map[string]string{
				"version": "latest",
			},
			shouldMatch: true,
		},
		{
			filter: map[string][]string{
				"app":     {"test"},
				"version": {"*"},
			},
			input: map[string]string{
				"app": "test",
			},
			shouldMatch: false,
		},
		{
			filter: map[string][]string{
				"app": {"test"},
			},
			input: map[string]string{
				"app":     "test",
				"version": "2.0",
			},
			shouldMatch: true,
		},
		{
			filter: map[string][]string{
				"version": {`/\d+\.\d+/`},
			},
			input: map[string]string{
				"app":     "test",
				"version": "2.0",
			},
			shouldMatch: true,
		},
		{
			filter: map[string][]string{
				"version": {`/\d+\.\d+/`},
			},
			input: map[string]string{
				"app":     "test",
				"version": "bad",
			},
			shouldMatch: false,
		},
	} {
		f, err := NewStringMapFilter(tc.filter)
		if tc.shouldError {
			assert.NotNil(t, err, spew.Sdump(tc))
		} else {
			assert.Nil(t, err, spew.Sdump(tc))
		}

		assert.Equal(t, tc.shouldMatch, f.Matches(tc.input), "%s\n%s", spew.Sdump(tc), spew.Sdump(f))
	}
}
