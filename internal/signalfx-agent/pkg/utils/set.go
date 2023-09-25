package utils

// StringSliceToMap converts a slice of strings into a map with keys from the slice
func StringSliceToMap(strings []string) map[string]bool {
	// Use bool so that the user can do `if setMap[key] { ... }``
	ret := map[string]bool{}
	for _, s := range strings {
		ret[s] = true
	}
	return ret
}

// StringSetToSlice converts a map representing a set into a slice of strings.
// If the value is `false`, the key won't be added.
func StringSetToSlice(set map[string]bool) []string {
	var out []string
	for k, ok := range set {
		if ok {
			out = append(out, k)
		}
	}
	return out
}

// StringSet creates a map set from vararg
func StringSet(strings ...string) map[string]bool {
	return StringSliceToMap(strings)
}
