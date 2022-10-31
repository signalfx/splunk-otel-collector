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

package statussources

import (
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestStatementFromZapCoreEntry(t *testing.T) {
	logger := zaptest.NewLogger(t).Named("logger.name")
	encoder := NewZapCoreEncoder()
	ce := logger.Check(zap.DebugLevel, "a.message")
	entry := ce.Entry

	now := time.Now()
	entry.Time = now
	statement, err := StatementFromZapCoreEntry(encoder, entry, []zapcore.Field{
		zap.String("field.one", "field.value"), zap.String("field.two", "another.field.value"),
		zap.Int("int", 1), zap.Bool("bool", true),
	})
	require.NoError(t, err)
	require.NotNil(t, statement)

	require.Equal(t, "a.message", statement.Message)
	require.Equal(t, "debug", statement.Level)
	require.Equal(t, now, statement.Time)
	require.Equal(t, "logger.name", statement.LoggerName)
	require.Equal(t, entry.Caller, statement.Caller)
	require.Equal(t, map[string]any{
		"field.one": "field.value", "field.two": "another.field.value",
		"logger": "logger.name", "bool": true, "int": float64(1), // becomes a float in json unmarshalling
	}, statement.Fields)
}

func TestStatementFromZapCoreEntryUnsupportedEncoder(t *testing.T) {
	logger := zaptest.NewLogger(t).Named("logger.name")
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	ce := logger.Check(zap.DebugLevel, "a.message")
	entry := ce.Entry

	statement, err := StatementFromZapCoreEntry(encoder, entry, nil)
	require.ErrorContains(t, err, "failed representing encoded zapcore.Entry")
	require.Nil(t, statement)
}

func TestStatementToLogRecord(t *testing.T) {
	logger := zaptest.NewLogger(t).Named("logger.name")
	encoder := NewZapCoreEncoder()
	ce := logger.Check(zap.DebugLevel, "a.message")
	entry := ce.Entry

	now := time.Now()
	entry.Time = now
	statement, err := StatementFromZapCoreEntry(encoder, entry, []zapcore.Field{
		zap.String("field.one", "field.value"), zap.String("field.two", "another.field.value"),
		zap.Int("int", 1), zap.Bool("bool", true),
	})
	require.NoError(t, err)
	require.NotNil(t, statement)

	t0 := time.Now()
	lr := statement.ToLogRecord()

	require.Equal(t, "a.message", lr.Body().AsString())
	require.Equal(t, "debug", lr.SeverityText())
	require.Equal(t, now.UTC(), lr.Timestamp().AsTime())
	require.GreaterOrEqual(t, lr.ObservedTimestamp().AsTime(), t0)
	require.Equal(t, map[string]any{
		"field.one": "field.value", "field.two": "another.field.value",
		"logger": "logger.name", "bool": true, "int": float64(1), // becomes a float in json unmarshalling
	}, lr.Attributes().AsRaw())

}

func TestReceiverNameToIDs(t *testing.T) {
	for _, test := range []struct {
		name               string
		receiverName       string
		expectedReceiverID config.ComponentID
		expectedEndpointID observer.EndpointID
	}{
		{name: "happy path",
			receiverName:       `<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
			expectedReceiverID: config.NewComponentIDWithName("<receiver.type>", "<receiver.name>"),
			expectedEndpointID: observer.EndpointID("<Endpoint.ID>"),
		},
		{name: "missing receiver_creator separator",
			receiverName:       `<receiver.type>/<receiver.name>/<receiver-creator.name>{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
			expectedReceiverID: config.NewComponentID(""),
			expectedEndpointID: observer.EndpointID(""),
		},
		{name: "multiple receiver_creator separators",
			receiverName:       `<receiver.type>/<receiver.name>/receiver_creator/receiver_creator/{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
			expectedReceiverID: config.NewComponentID(""),
			expectedEndpointID: observer.EndpointID(""),
		},
		{name: "missing endpoint separator",
			receiverName:       `<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>/<Endpoint.ID>`,
			expectedReceiverID: config.NewComponentID(""),
			expectedEndpointID: observer.EndpointID(""),
		},
		{name: "multiple endpoint separators",
			receiverName:       `<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>/{endpoint="<Endpoint.Target>"}/{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
			expectedReceiverID: config.NewComponentID(""),
			expectedEndpointID: observer.EndpointID(""),
		},
		{name: "missing name with forward slash hostport",
			receiverName:       `debug//receiver_creator/discovery/discovery_name{endpoint="127.0.0.53:53"}/(host_observer/host)127.0.0.53-53-TCP)`,
			expectedReceiverID: config.NewComponentID("debug"),
			expectedEndpointID: observer.EndpointID("(host_observer/host)127.0.0.53-53-TCP)"),
		},
		{name: "missing name without forward slash hostport",
			receiverName:       `debug/receiver_creator/discovery/discovery_name{endpoint="127.0.0.53:53"}/(host_observer/host)127.0.0.53-53-TCP)`,
			expectedReceiverID: config.NewComponentID("debug"),
			expectedEndpointID: observer.EndpointID("(host_observer/host)127.0.0.53-53-TCP)"),
		},
		{name: "docker observer",
			receiverName:       `smartagent/redis/with/additional/slashes/receiver_creator/discovery/discovery_name{endpoint="172.17.0.2:6379"}/d2ee077a262e23bf3fccdd6422f88ce3ec6ed2403bfe67c1d25fb3e5647a0bb7:6379`,
			expectedReceiverID: config.NewComponentIDWithName("smartagent", "redis/with/additional/slashes"),
			expectedEndpointID: observer.EndpointID("d2ee077a262e23bf3fccdd6422f88ce3ec6ed2403bfe67c1d25fb3e5647a0bb7:6379"),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			lr := plog.NewLogRecord()
			lr.Attributes().PutString("name", test.receiverName)
			receiverID, endpointID := ReceiverNameToIDs(lr)
			require.Equal(t, test.expectedReceiverID, receiverID)
			require.Equal(t, test.expectedEndpointID, endpointID)
		})
	}
}

func FuzzReceiverNameToIDs(f *testing.F) {
	for _, receiverName := range []string{
		`invalid`, `<receiver.type>/<receiver.name>/<receiver-creator.name>{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
		`<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>/<Endpoint.ID>`,
		`<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>/{endpoint="<Endpoint.Target>"}/{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
		`<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>/{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
		`<receiver.type>/<receiver.name>/receiver_creator/receiver_creator/{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`,
		`debug//receiver_creator/discovery/discovery_name{endpoint="127.0.0.53:53"}/(host_observer/host)127.0.0.53-53-TCP)`,
		`debug/receiver_creator/discovery/discovery_name{endpoint="127.0.0.53:53"}/(host_observer/host)127.0.0.53-53-TCP`,
		`smartagent/redis/receiver_creator/discovery/discovery_name{endpoint="172.17.0.2:6379"}/d2ee077a262e23bf3fccdd6422f88ce3ec6ed2403bfe67c1d25fb3e5647a0bb7:6379`,
	} {
		f.Add(receiverName)
	}
	f.Fuzz(func(t *testing.T, receiverName string) {
		require.NotPanics(t, func() {
			lr := plog.NewLogRecord()
			lr.Attributes().PutString("name", receiverName)
			receiverID, endpointID := ReceiverNameToIDs(lr)
			// if we can't find a receiver we should never return an EndpointID
			if receiverID == discovery.NoType {
				require.Equal(t, observer.EndpointID(""), endpointID)
			} else if receiverID.Type() == config.Type("") {
				// if the receiver type is empty the name should also be empty
				require.Equal(t, "", receiverID.Name())
			}
		})
	})
}
