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

func TestNewOTLPReceiverSink(t *testing.T) {
	otlp := NewOTLPReceiverSink()
	require.NotNil(t, otlp)

	require.Empty(t, otlp.Endpoint)
	require.Nil(t, otlp.Host)
	require.Nil(t, otlp.Logger)
	require.Nil(t, otlp.logsReceiver)
	require.Nil(t, otlp.logsSink)
	require.Nil(t, otlp.metricsReceiver)
	require.Nil(t, otlp.metricsSink)
	require.Nil(t, otlp.tracesReceiver)
	require.Nil(t, otlp.tracesSink)
}

func TestBuilderMethods(t *testing.T) {
	otlp := NewOTLPReceiverSink()

	withEndpoint := otlp.WithEndpoint("myendpoint")
	require.Equal(t, "myendpoint", withEndpoint.Endpoint)
	require.Empty(t, otlp.Endpoint)
}

func TestBuildDefaults(t *testing.T) {
	otlp, err := NewOTLPReceiverSink().Build()
	require.Error(t, err)
	assert.EqualError(t, err, "must provide an Endpoint for OTLPReceiverSink")
	assert.Nil(t, otlp)

	otlp, err = NewOTLPReceiverSink().WithEndpoint("myEndpoint").Build()
	require.NoError(t, err)
	assert.Equal(t, "myEndpoint", otlp.Endpoint)
	assert.NotNil(t, otlp.Host)
	assert.NotNil(t, otlp.Logger)
	assert.NotNil(t, otlp.logsReceiver)
	assert.NotNil(t, otlp.logsSink)
	assert.NotNil(t, otlp.metricsReceiver)
	assert.NotNil(t, otlp.metricsSink)
	assert.NotNil(t, otlp.tracesReceiver)
	assert.NotNil(t, otlp.tracesSink)
}
