package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitConfigPath(t *testing.T) {
	cases := []struct {
		in           string
		expectedPath string
		expectedKey  string
		expectedErr  error
	}{
		{"/secret/data/test[password]", "/secret/data/test", "password", nil},
		{"/secret/data/test[data.password]", "/secret/data/test", "data.password", nil},
		{"secret/data/test[data.password]", "secret/data/test", "data.password", nil},
		{"secret/data/test[]", "secret/data/test", "", nil},
		{"secret/data/test-info[password]", "secret/data/test-info", "password", nil},
		{"/[a]", "/", "a", nil},
		{"secret/data/test", "", "", errInvalidPathFormat},
		{"secret/data/test[", "", "", errInvalidPathFormat},
		{"secret/data/test]", "", "", errInvalidPathFormat},
		{"[]", "", "", errInvalidPathFormat},
		{"", "", "", errInvalidPathFormat},
	}

	for _, tc := range cases {
		path, key, err := splitConfigPath(tc.in)
		assert.Equal(t, tc.expectedPath, path)
		assert.Equal(t, tc.expectedKey, key)
		assert.Equal(t, tc.expectedErr, err)
	}
}

func TestTraverseToKey(t *testing.T) {
	cases := []struct {
		data        map[string]interface{}
		key         string
		expectedVal interface{}
	}{
		{
			map[string]interface{}{
				"data": map[string]interface{}{
					"password": "s3cr3t",
				},
			},
			"data.password",
			"s3cr3t",
		},
		{
			map[string]interface{}{
				"password": "s3cr3t",
			},
			"password",
			"s3cr3t",
		},
		{
			map[string]interface{}{
				"data": map[string]interface{}{
					"password": "s3cr3t",
				},
			},
			"data.password.bad",
			nil,
		},
		{
			map[string]interface{}{
				"password": "s3cr3t",
			},
			"",
			nil,
		},
		{
			map[string]interface{}{
				"password": 5,
			},
			"password",
			5,
		},
	}

	for _, tc := range cases {
		val := traverseToKey(tc.data, tc.key)
		assert.Equal(t, tc.expectedVal, val)
	}
}
