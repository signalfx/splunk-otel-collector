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

// StringSet creates a map set from vararg
func StringSet(strings ...string) map[string]bool {
	return StringSliceToMap(strings)
}
