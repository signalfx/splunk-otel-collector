// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamsBool(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  bool
	}{
		{value: "true", want: true},
		{value: "TRUE", want: true},
		{value: "True", want: true},
		{value: "t", want: true},
		{value: "T", want: true},
		{value: "yes", want: true},
		{value: "YES", want: true},
		{value: "1", want: true},
		{value: "false", want: false},
		{value: "FALSE", want: false},
		{value: "f", want: false},
		{value: "no", want: false},
		{value: "0", want: false},
	} {
		t.Run(tt.value, func(t *testing.T) {
			params := Params{{Name: "flag", Value: tt.value}}
			assert.Equal(t, tt.want, params.Bool("flag", true))
			assert.Equal(t, tt.want, params.Bool("flag", false))
		})
	}

	// splunkd leaves the default in place for anything it does not recognize,
	// including a value 0 or 1 is only part of.
	for _, value := range []string{"", "banana", "10", "00", "y", "n", "-1", "2"} {
		t.Run("unrecognized/"+value, func(t *testing.T) {
			params := Params{{Name: "flag", Value: value}}
			assert.True(t, params.Bool("flag", true))
			assert.False(t, params.Bool("flag", false))
		})
	}

	t.Run("missing", func(t *testing.T) {
		assert.True(t, Params{}.Bool("flag", true))
		assert.False(t, Params{}.Bool("flag", false))
		assert.True(t, Params(nil).Bool("flag", true))
	})
}
