// Copyright  Splunk, Inc.
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

package timeutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Duration is a wrapped time.Duration that supports durations as integers
// or a ParseDuration string. If it's an integer it is interpreted in seconds instead
// of being directly cast to time.Duration which would normally make it in nanoseconds.
type Duration time.Duration

// ErrInvalidDuration is returned when the duration can't be interpreted
var ErrInvalidDuration = errors.New("the duration must be a string with time unit specified or an integer as seconds")

// AsDuration returns the type cast to time.Duration
func (d Duration) AsDuration() time.Duration {
	return time.Duration(d)
}

// IsZero returns true if the duration is 0, otherwise true.
func (d Duration) IsZero() bool {
	return time.Duration(d) == 0
}

// UnmarshalJSON unmarshals Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	return d.UnmarshalYAML(func(i interface{}) error {
		return json.Unmarshal(data, i)
	})
}

// UnmarshalYAML unmarshals Duration
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First check if it's an integer. Interpret it as seconds.
	var i int64
	if err := unmarshal(&i); err == nil {
		*d = Duration(time.Duration(i) * time.Second)
		return nil
	}

	var s string

	if err := unmarshal(&s); err == nil {
		// If it's a string but parses as an integer interpret it as seconds.
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			*d = Duration(time.Duration(i) * time.Second)
			return nil
		}

		// If here it's hopefully a string with ParseDuration syntax.
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrInvalidDuration, err)
		}
		*d = Duration(parsed)
		return nil
	}

	return ErrInvalidDuration
}
