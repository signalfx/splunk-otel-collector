package filter

// OverridableStringFilter matches input strings that are positively matched by
// one of the input filters AND are not excluded by any negated filters (they
// work kind of like how .gitignore patterns work), OR are exactly matched by a
// literal filter input (e.g. not a globbed or regex pattern).  Order of the
// items does not matter.
type OverridableStringFilter struct {
	*BasicStringFilter
}

// NewOverridableStringFilter makes a new OverridableStringFilter with the given
// items.
func NewOverridableStringFilter(items []string) (*OverridableStringFilter, error) {
	basic, err := NewBasicStringFilter(items)
	if err != nil {
		return nil, err
	}

	return &OverridableStringFilter{
		BasicStringFilter: basic,
	}, nil
}

// Matches if s is positively matched by the filter items AND is not excluded
// by any, OR if it is postively matched by a non-glob/regex pattern exactly
// and is negated as well.  See the unit tests for examples.
func (f *OverridableStringFilter) Matches(s string) bool {
	// We could use the matcher interface here to reduce LOC but supposedly
	// using interfaces is far slower than using concrete types.  This could be
	// tested for this specific case with benchmarking, but going with what I
	// know is performant for now.
	negated, matched := f.staticSet[s]
	// If a metric is negated and it matched it won't match anything else by
	// definition.
	if matched && negated {
		return false
	}

	for _, reMatch := range f.regexps {
		reMatched, negated := reMatch.Matches(s)
		if reMatched && negated {
			return false
		}
		matched = matched || reMatched
	}

	for _, globMatcher := range f.globs {
		globMatched, negated := globMatcher.Matches(s)
		if globMatched && negated {
			return false
		}
		matched = matched || globMatched
	}
	return matched
}
