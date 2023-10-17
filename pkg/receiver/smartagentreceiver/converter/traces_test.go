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

package converter

import (
	"testing"

	"github.com/signalfx/golib/v3/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

var (
	flse                     = false
	tru                      = true
	oneInt64           int64 = 1
	twoInt64           int64 = 2
	oneHundredInt32    int32 = 100
	twoHundredInt32    int32 = 200
	serviceName              = "some service"
	anotherServiceName       = "another service"
	loopback                 = "127.0.0.1"
	loopbackIPv6             = "::1"
	linkLocal                = "169.254.0.1"
	linkLocalIPv6            = "fe80::1"
)

func TestNilSFxSpanConversion(t *testing.T) {
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{nil}, logger)
	assert.NoError(t, err)
	assert.Zero(t, traces.ResourceSpans().Len())
	assert.Equal(t, 0, logs.Len())
}

func TestInvalidSFxToZipkinSpanModelConversion(t *testing.T) {
	invalidAsZipkinJSON := trace.Span{}
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{&invalidAsZipkinJSON}, logger)
	assert.NoError(t, err)
	assert.Zero(t, traces.ResourceSpans().Len())

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "failed to unmarshall SFx span as Zipkin", entry.Message)
	assert.Equal(t, "valid traceId required", entry.ContextMap()["error"])
}

func TestInvalidSFxToPDataConversion(t *testing.T) {
	invalidAsPData := trace.Span{
		TraceID: "12345",
		ID:      "12345",
		Tags: map[string]string{
			// appears only way to generate error is in link content.
			"otlp.link.0": "|i|n|v|a|l|i|d|",
		},
	}
	obs, logs := observer.New(zap.DebugLevel)
	logger := zap.New(obs)

	traces, err := sfxSpansToPDataTraces([]*trace.Span{&invalidAsPData}, logger)
	assert.EqualError(t, err, "invalid length for ID")
	assert.Equal(t, 0, traces.ResourceSpans().Len())
	require.Equal(t, 0, logs.Len())
}
