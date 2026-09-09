// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package scriptedinput

import (
	"runtime"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func Test_ScriptedInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows because scripts use bash")
	}

	tests := []struct {
		name      string
		interval  string
		expectMsg bool
	}{
		{
			"always",
			"0",
			true,
		},
		{
			"polling",
			"1",
			true,
		},
		{
			"disabled",
			"-1",
			false,
		},
		{
			"float_interval",
			"60.0",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := NewConfig()
			c.BaseDir = "testdata"
			c.Input = conf.Input{
				Configuration: conf.Configuration{
					Stanza: conf.Stanza{
						Name: "script://./bin/foo.sh",
						Params: []conf.Param{
							{Name: "interval", Value: test.interval},
						},
					},
				},
			}
			settings := componenttest.NewNopTelemetrySettings()
			settings.Logger, _ = zap.NewDevelopment()
			o, err := c.Build(settings)
			assert.NoError(t, err)
			require.NotNil(t, o)
			fo := testutil.NewFakeOutput(t)
			require.NoError(t, fo.Start(nil))
			t.Cleanup(func() {
				require.NoError(t, fo.Stop())
			})
			o.SetOutputIDs([]string{fo.ID()})
			err = o.SetOutputs([]operator.Operator{
				fo,
			})
			require.NoError(t, err)
			err = o.Start(nil)
			require.NoError(t, err)
			if test.expectMsg {
				select {
				case msg := <-fo.Received:
					require.NotNil(t, msg)
					require.Equal(t, "foo\n", msg.Body)
				case <-time.After(5 * time.Second):
					require.Fail(t, "timed out waiting for message")
				}
			} else {
				time.Sleep(time.Millisecond * 100)
				require.Len(t, fo.Received, 0)
			}
			require.NoError(t, o.Stop())
		})
	}
}
