package utils

import (
	"errors"
	"regexp"
)

var accessTokenSanitizer = regexp.MustCompile(`(?i)x-sf-token:\[[^\s]+\]`)

// SanitizeHTTPError will remove any sensitive data from HTTP client errors.
func SanitizeHTTPError(err error) error {
	return errors.New(accessTokenSanitizer.ReplaceAllString(err.Error(), ""))
}
