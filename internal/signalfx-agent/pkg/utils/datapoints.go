package utils

// TruncateDimensionValue truncates the given string to 256 characters.
func TruncateDimensionValue(value string) string {
	// Not sure if our backend enforces character length or byte length.
	// If values include multi-byte unicode chars, this might not work.
	if len(value) > 256 {
		return value[:256]
	}
	return value
}
