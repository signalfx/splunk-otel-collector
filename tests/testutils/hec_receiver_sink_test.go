// Copyright Splunk, Inc.
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

//go:build testutils

package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHECReceiverSink(t *testing.T) {
	hec := NewHECReceiverSink()
	require.NotNil(t, hec)

	require.Empty(t, hec.Endpoint)
	require.Nil(t, hec.Host)
	require.Nil(t, hec.Logger)
	require.Nil(t, hec.logsReceiver)
	require.Nil(t, hec.logsSink)
}

func TestHECReceiverNotBuilt(t *testing.T) {
	hec := NewHECReceiverSink()
	require.Error(t, hec.assertBuilt("NotBuilt"))
	require.Error(t, hec.Start())
	require.Error(t, hec.Shutdown())
	require.Zero(t, hec.LogRecordCount())
	require.Nil(t, hec.AllLogs())
}

func TestHECBuilderMethods(t *testing.T) {
	hec := NewHECReceiverSink()

	withEndpoint := hec.WithEndpoint("myendpoint")
	require.Equal(t, "myendpoint", withEndpoint.Endpoint)
	require.Empty(t, hec.Endpoint)
}

func TestHECBuildDefaults(t *testing.T) {
	hec, err := NewHECReceiverSink().Build()
	require.Error(t, err)
	assert.EqualError(t, err, "must provide an Endpoint for HECReceiverSink")
	assert.Nil(t, hec)

	hec, err = NewHECReceiverSink().WithEndpoint("myEndpoint").Build()
	require.NoError(t, err)
	assert.Equal(t, "myEndpoint", hec.Endpoint)
	assert.NotNil(t, hec.Host)
	assert.NotNil(t, hec.Logger)
	assert.NotNil(t, hec.logsReceiver)
	assert.NotNil(t, hec.logsSink)
}
