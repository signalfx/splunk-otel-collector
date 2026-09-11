// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package batchreceiver

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/pipeline"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()
	file := fmt.Sprintf("%s%c%s", tempDir, filepath.Separator, "foo.txt")
	cfg := Config{
		Input: conf.Input{
			Configuration: conf.Configuration{
				Stanza: conf.Stanza{
					Name: fmt.Sprintf("batch://%s", file),
					App:  "",
					Params: conf.Params{
						conf.Param{
							Name:  "host",
							Value: "myhost",
						},
					},
				},
			},
		},
	}
	logger, _ := zap.NewDevelopment()
	c := batch{logger: logger}.InputConfig(cfg)
	o, err := c.Build(component.TelemetrySettings{
		Logger:         logger,
		TracerProvider: nooptrace.NewTracerProvider(),
		MeterProvider:  noopmetric.NewMeterProvider(),
		Resource:       pcommon.NewResource(),
	})
	require.NoError(t, err)
	output := testutil.NewFakeOutput(t)
	o.SetOutputIDs([]string{"fake"})
	require.NoError(t, o.SetOutputs([]operator.Operator{
		output,
	}))
	err = o.Start(nil)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, o.Stop())
	}()

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "foo.txt"), []byte("foo\n"), 0o644))
	received := <-output.Received
	require.Equal(t, "foo\n", received.Body)
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		_, err = os.Stat(file)
		require.ErrorIs(tt, err, fs.ErrNotExist)
	}, 1*time.Second, 100*time.Millisecond)
}

func TestRenameMetadata(t *testing.T) {
	ops := renameMetadata()
	output := testutil.NewFakeOutput(t)
	pipe, err := pipeline.Config{
		Operators:     ops,
		DefaultOutput: output,
	}.Build(componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)
	require.NoError(t, pipe.Start(nil))
	defer func() {
		require.NoError(t, pipe.Stop())
	}()
	require.NoError(t, pipe.Operators()[0].Process(context.Background(), &entry.Entry{
		Attributes: map[string]any{
			"source":     "src",
			"sourcetype": "srctype",
			"host":       "foo",
		},
	}))

	result := <-output.Received
	require.Equal(t, "src", result.Attributes["com.splunk.source"])
	require.Equal(t, "srctype", result.Attributes["com.splunk.sourcetype"])
	require.Equal(t, "foo", result.Attributes["host.name"])
}
