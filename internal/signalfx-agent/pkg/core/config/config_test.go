package config

import (
	"testing"

	"github.com/signalfx/defaults"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/stretchr/testify/require"
)

func TestWriterOutputValidation(t *testing.T) {
	t.Run("one of SignalFx or Splunk output is required", func(t *testing.T) {
		c := &Config{
			Writer: WriterConfig{
				SignalFxEnabled: pointer.Bool(false),
				Splunk:          &SplunkConfig{Enabled: false},
			},
			Monitors: []MonitorConfig{
				{},
				{ProcPath: "/proc"},
			},
		}
		require.NoError(t, defaults.Set(c))

		err := c.validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "output are disabled")
	})

	t.Run("Splunk output can be enabled by itself", func(t *testing.T) {
		c := &Config{
			Writer: WriterConfig{
				Splunk: &SplunkConfig{Enabled: true},
			},
			Monitors: []MonitorConfig{
				{},
				{ProcPath: "/proc"},
			},
		}
		require.NoError(t, defaults.Set(c))

		err := c.validate()
		require.NoError(t, err)
	})
}
