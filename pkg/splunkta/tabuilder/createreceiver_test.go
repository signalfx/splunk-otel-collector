// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package tabuilder

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func stanzaInput(name string) conf.Input {
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{Name: name},
		},
	}
}

func createReceiver(t *testing.T, stanzaName string) (receiver.Logs, error) {
	t.Helper()
	return CreateReceiver(context.Background(), "", consumertest.NewNop(),
		stanzaInput(stanzaName), nil, nil, componenttest.NewNopTelemetrySettings())
}

// inputs.conf.spec spells the event log kind [WinEventLog://<name>], the only
// input kind whose canonical spelling is not lowercase, so a mixed-case kind
// has to reach the wineventlog branch instead of the unsupported default.
func TestCreateReceiverWinEventLogKind(t *testing.T) {
	for _, stanzaName := range []string{"wineventlog://Application", "WinEventLog://Application"} {
		t.Run(stanzaName, func(t *testing.T) {
			r, err := createReceiver(t, stanzaName)
			if runtime.GOOS != "windows" {
				// The underlying receiver is build-tagged. The factory refusing
				// off Windows still proves dispatch reached it rather than
				// returning the nil-receiver-and-nil-error unsupported result.
				require.ErrorContains(t, err, "wineventlog is not supported outside Windows environments")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, r)
		})
	}
}

// Kinds with no mapping return a nil receiver and a nil error, which the caller
// logs and skips. An unprefixed stanza name is one of them: inputs.conf.spec
// defines [<scheme>] as a modular-input scheme defaults stanza whose instances
// are [<scheme>://<name>], so it declares no input to collect.
func TestCreateReceiverUnsupportedKinds(t *testing.T) {
	for _, tt := range []struct {
		name       string
		stanzaName string
	}{
		{name: "unprefixed stanza name", stanzaName: "my_bare_script"},
		{name: "modular input scheme defaults stanza", stanzaName: "Splunk_TA_otel"},
		{name: "unmapped kind", stanzaName: "fschange:///etc/passwd"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := createReceiver(t, tt.stanzaName)
			require.NoError(t, err)
			require.Nil(t, r)
		})
	}
}
