package utils

// RemoveAllElementsFromStringSlice removes all elements from toRemove that exists
// in inputStrings
func RemoveAllElementsFromStringSlice(inputStrings, toRemoveStrings []string) []string {
	inputStringsMap := StringSliceToMap(inputStrings)
	toRemoveStringsMap := StringSliceToMap(toRemoveStrings)

	for key := range toRemoveStringsMap {
		delete(inputStringsMap, key)
	}

	return StringSetToSlice(inputStringsMap)
}
