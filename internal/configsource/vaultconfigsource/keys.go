// Copyright 2020 Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vaultconfigsource

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
