package vault

import (
	"errors"
	"strings"
)

var errInvalidPathFormat = errors.New("invalid Vault path/key combination, expected format <path>[<key>]")

// Allows key to be dot-delimited to traverse nested maps
func traverseToKey(data map[string]interface{}, key string) interface{} {
	parts := strings.Split(key, ".")

	for i, part := range parts {
		partVal := data[part]
		if i == len(parts)-1 {
			return partVal
		}

		var ok bool
		data, ok = partVal.(map[string]interface{})
		if !ok {
			return nil
		}
	}
	return nil
}

// The config path is of the form `/path/to/secret[keys.to.data]`
func splitConfigPath(pathAndKey string) (string, string, error) {
	if !strings.HasSuffix(pathAndKey, "]") {
		return "", "", errInvalidPathFormat
	}

	parts := strings.SplitN(pathAndKey, "[", 2)
	if len(parts) < 2 || len(parts[0]) == 0 {
		return "", "", errInvalidPathFormat
	}

	return parts[0], strings.TrimSuffix(parts[1], "]"), nil
}
