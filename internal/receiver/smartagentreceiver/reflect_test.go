// Copyright 2021 Splunk, Inc.
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

package smartagentreceiver

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type DoublyEmbedded struct {
	Thing string
}

type Embedded struct {
	DoublyEmbedded
	AnotherThing string
}

type Envelope struct {
	Embedded
	YetAnotherThing string
}

var stringType = reflect.TypeOf("")

func TestGetSettableStructFieldValue(t *testing.T) {
	envelope := &Envelope{
		Embedded: Embedded{
			AnotherThing: "another thing",
			DoublyEmbedded: DoublyEmbedded{
				Thing: "thing",
			},
		},
		YetAnotherThing: "yet another thing",
	}
	fieldValue, err := GetSettableStructFieldValue(envelope, "Thing", stringType)
	require.NoError(t, err)
	require.NotNil(t, fieldValue)
	require.Equal(t, "thing", fieldValue.String())
	fieldValue.Set(reflect.ValueOf("something else"))
	require.Equal(t, "something else", envelope.Thing)

	fieldValue, err = GetSettableStructFieldValue(envelope, "AnotherThing", stringType)
	require.NoError(t, err)
	require.NotNil(t, fieldValue)
	require.Equal(t, "another thing", fieldValue.String())
	fieldValue.Set(reflect.ValueOf("used to be 'another thing'"))
	require.Equal(t, "used to be 'another thing'", envelope.AnotherThing)

	// Envelope isn't addressable so its exported fields cannot be set
	fieldValue, err = GetSettableStructFieldValue(*envelope, "YetAnotherThing", stringType)
	require.NoError(t, err)
	require.Nil(t, fieldValue)

	fieldValue, err = GetSettableStructFieldValue(envelope, "YetAnotherThing", stringType)
	require.NoError(t, err)
	require.NotNil(t, fieldValue)
	require.Equal(t, "yet another thing", fieldValue.String())
	fieldValue.Set(reflect.ValueOf("used to be 'yet another thing'"))
	require.Equal(t, "used to be 'yet another thing'", envelope.YetAnotherThing)
}

func TestSetStructFieldWithExplicitType(t *testing.T) {
	envelope := &Envelope{YetAnotherThing: "not zero value"}
	set, err := SetStructFieldWithExplicitType(
		envelope, "YetAnotherThing", "desired value",
		reflect.TypeOf(nil), reflect.TypeOf(1), stringType, reflect.TypeOf(struct{}{}),
	)
	require.NoError(t, err)
	require.True(t, set)
	require.Equal(t, "desired value", envelope.YetAnotherThing)

	set, err = SetStructFieldWithExplicitType(
		envelope, "Thing", "another desired value",
		reflect.TypeOf(nil), reflect.TypeOf(1), stringType, reflect.TypeOf(struct{}{}),
	)
	require.NoError(t, err)
	require.True(t, set)
	require.Equal(t, "another desired value", envelope.Thing)
}

func TestSetStructFieldWithZeroValue(t *testing.T) {
	envelope := &Envelope{YetAnotherThing: "not zero value"}
	set, err := SetStructFieldIfZeroValue(envelope, "Thing", "desired value")
	require.NoError(t, err)
	require.True(t, set)
	require.Equal(t, "desired value", envelope.Thing)

	set, err = SetStructFieldIfZeroValue(envelope, "YetAnotherThing", "should not take")
	require.NoError(t, err)
	require.False(t, set)
	require.Equal(t, "not zero value", envelope.YetAnotherThing)
}
