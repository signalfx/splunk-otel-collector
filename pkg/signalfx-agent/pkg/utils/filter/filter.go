// Package filter contains common filtering logic that can be used to filter
// datapoints or various resources within other agent components, such as
// monitors.  Filter instances have a Matches function which takes an instance
// of the type that they filter and return whether that instance matches the
// filter.
package filter

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

// StringFilter matches against simple strings
type StringFilter interface {
	Matches(string) bool
}

// StringMapFilter matches against the values of a map[string]string.
type StringMapFilter interface {
	Matches(map[string]string) bool
}

// BasicStringFilter will match if any one of the given strings is a match.
type BasicStringFilter struct {
	staticSet        map[string]bool
	regexps          []regexMatcher
	globs            []globMatcher
	anyStaticNegated bool
}

var _ StringFilter = &BasicStringFilter{}

// Matches returns true if any one of the strings in the filter matches the
// input s
func (f *BasicStringFilter) Matches(s string) bool {
	negated, matched := f.staticSet[s]
	if matched {
		return !negated
	}
	if f.anyStaticNegated {
		return true
	}

	for _, reMatch := range f.regexps {
		if reMatch.re.MatchString(s) != reMatch.negated {
			return true
		}
	}

	for _, globMatch := range f.globs {
		if globMatch.glob.Match(s) != globMatch.negated {
			return true
		}
	}

	return false
}

// NewBasicStringFilter returns a filter that can match against the provided items.
func NewBasicStringFilter(items []string) (*BasicStringFilter, error) {
	staticSet := make(map[string]bool)
	var regexps []regexMatcher
	var globs []globMatcher

	anyStaticNegated := false
	for _, i := range items {
		m, negated := stripNegation(i)
		switch {
		case isRegex(m):
			var re *regexp.Regexp
			var err error

			reText := stripSlashes(m)
			re, err = regexp.Compile(reText)

			if err != nil {
				return nil, err
			}

			regexps = append(regexps, regexMatcher{re: re, negated: negated})
		case isGlobbed(m):
			g, err := glob.Compile(m)
			if err != nil {
				return nil, err
			}

			globs = append(globs, globMatcher{glob: g, negated: negated})
		default:
			staticSet[m] = negated
			if negated {
				anyStaticNegated = true
			}
		}
	}

	return &BasicStringFilter{
		staticSet:        staticSet,
		regexps:          regexps,
		globs:            globs,
		anyStaticNegated: anyStaticNegated,
	}, nil
}

// NewStringMapFilter returns a filter that matches against the provided map.
// All key/value pairs must match the spec given in m for a map to be
// considered a match.
func NewStringMapFilter(m map[string][]string) (StringMapFilter, error) {
	filterMap := map[string]*OverridableStringFilter{}
	okMissing := map[string]bool{}
	for k := range m {
		if len(m[k]) == 0 {
			return nil, errors.New("string map value in filter cannot be empty")
		}

		realKey := strings.TrimSuffix(k, "?")

		var err error
		filterMap[realKey], err = NewOverridableStringFilter(m[k])
		if err != nil {
			return nil, err
		}

		if len(realKey) != len(k) {
			okMissing[realKey] = true
		}
	}

	return &fullStringMapFilter{
		filterMap: filterMap,
		okMissing: okMissing,
	}, nil
}

// Each key/value pair must match the filter for the whole map to match.
type fullStringMapFilter struct {
	filterMap map[string]*OverridableStringFilter
	okMissing map[string]bool
}

func (f *fullStringMapFilter) Matches(m map[string]string) bool {
	// Empty map input never matches
	if len(m) == 0 && len(f.okMissing) == 0 {
		return false
	}

	for k, filter := range f.filterMap {
		if v, ok := m[k]; ok {
			if !filter.Matches(v) {
				return false
			}
		} else {
			return f.okMissing[k]
		}
	}
	return true
}
