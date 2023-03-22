package utils

// MakeRange creates an int slice containing all ints between `min` and `max`
func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

// InterfaceSliceToStringSlice returns a new slice that contains the elements
// of `is` as strings.  Returns nil if any of the elements of `is` are not
// strings.
func InterfaceSliceToStringSlice(is []interface{}) []string {
	var ss []string
	for _, intf := range is {
		if s, ok := intf.(string); ok {
			ss = append(ss, s)
		} else {
			return nil
		}
	}
	return ss
}

// RemoveAllElementsFromStringSlice removes all elements from toRemove that exists
// in inputStrings
func RemoveAllElementsFromStringSlice(inputStrings []string, toRemoveStrings []string) []string {
	inputStringsMap := StringSliceToMap(inputStrings)
	toRemoveStringsMap := StringSliceToMap(toRemoveStrings)

	for key := range toRemoveStringsMap {
		delete(inputStringsMap, key)
	}

	return StringSetToSlice(inputStringsMap)
}
