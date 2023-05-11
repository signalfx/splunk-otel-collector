package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoesServiceMatchRule(t *testing.T) {
	t.Run("Handles parse error in discovery rule", func(t *testing.T) {
		endpoint := NewEndpointCore("abcd", "test", "test", nil)
		endpoint.Host = "10.0.0.1"
		endpoint.AddExtraField("labels", map[string]string{
			"env": "prod",
		})

		assert.True(t, DoesServiceMatchRule(endpoint, `labels["env"] == "prod"`, true))
		assert.False(t, DoesServiceMatchRule(endpoint, `labels["env"] == "dev"`, true))
	})

	t.Run("Handles nested structs in discovery rule", func(t *testing.T) {
		endpoint := NewEndpointCore("abcd", "test", "test", nil)
		endpoint.AddExtraField("a", struct {
			B map[string]bool
		}{
			B: map[string]bool{"c": true},
		})
		assert.True(t, DoesServiceMatchRule(endpoint, `Get(a.B, "c")`, true))
	})
}

func TestRuleEvaluation(t *testing.T) {
	env := map[string]interface{}{
		"container_id":    "abcdef",
		"container_image": "my_app:latest",
		"private_port":    2181,
		"labels": map[string]string{
			"version": "1.5.0",
		},
		"struct": struct {
			B map[string]bool
		}{
			B: map[string]bool{"c": true},
		},
	}

	cases := []struct {
		rule        string
		expected    interface{}
		shouldError bool
	}{
		{
			rule:     `container_image =~ "my_app" && private_port == 2181`,
			expected: true,
		},
		{
			rule:     `container_image=~"my_app" && private_port == 2181`,
			expected: true,
		},
		{
			rule:     `container_image =~ "other_app" && private_port == 2181`,
			expected: false,
		},
		{
			rule:     `container_image=~"other_app" && private_port == 2181`,
			expected: false,
		},
		{
			rule:     `Get(labels, "version") == "1.5.0"`,
			expected: true,
		},
		{
			rule:     `labels["version"] == "1.5.0"`,
			expected: true,
		},
		{
			rule:     `struct.B['c']`,
			expected: true,
		},
		{
			rule:     `labels["version"] == "2.5.0"`,
			expected: false,
		},
		{
			rule: `image =~ "other_app" && private_port == 2181`,
			// Missing a variable
			shouldError: true,
		},
	}

	for i := range cases {
		c := cases[i]

		val, err := ExecuteRule(c.rule, env)
		if c.shouldError {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}

		require.Equal(t, c.expected, val)
	}
}

func TestMapFunctions(t *testing.T) {
	interfacemap := map[interface{}]interface{}{"hello": "world", "good": "bye"}
	stringmap := map[string]interface{}{"hello": "world", "good": "bye"}

	getFunc := ruleFunctions["Get"].(func(...interface{}) interface{})
	containsFunc := ruleFunctions["Contains"].(func(...interface{}) bool)

	t.Run("Get() handles string -> interface{} map type", func(t *testing.T) {
		val := getFunc(stringmap, "hello")
		assert.Equal(t, "world", val, "should return the expected value")
	})

	t.Run("Get() returns string -> interface{} map type", func(t *testing.T) {
		val := getFunc(stringmap, "hello")
		assert.Equal(t, "world", val, "should return the expected value")
	})

	t.Run("Get() returns error on bad map type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				assert.Fail(t, "should error out when wrong map type is used")
			}
		}()
		_ = getFunc("string", 3)
	})

	t.Run("Get() insufficient number of arguments", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				assert.Fail(t, "should error out when not enough arguments are provided")
			}
		}()
		_ = getFunc(interfacemap)
	})

	t.Run("Get() map does not contain key", func(t *testing.T) {
		val := getFunc(interfacemap, "nokey")
		assert.Nil(t, val, "should return nil if the map does not contain the desired value")
	})

	t.Run("Get() returns default if not in map", func(t *testing.T) {
		val := getFunc(interfacemap, "nokey", 50)
		assert.Equal(t, val, 50, "should return default if the map does not contain the desired value")
	})

	t.Run("Get() handles interface{} -> interface{} maps", func(t *testing.T) {
		val := getFunc(interfacemap, "hello")
		assert.Equal(t, "world", val, "should return the expected value")
	})

	t.Run("Contains() map does not contain key", func(t *testing.T) {
		val := containsFunc(interfacemap, "nokey")
		assert.False(t, val, "should only return false if an error occurred")
	})

	t.Run("Contains() incorrect argument types", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				assert.Fail(t, "should error if the supplied arguments are the wrong type")
			}
		}()
		_ = containsFunc(stringmap)
	})

	t.Run("Contains() map contains desired value", func(t *testing.T) {
		val := containsFunc(interfacemap, "good")
		assert.True(t, val, "should return the expected value")
	})
}
