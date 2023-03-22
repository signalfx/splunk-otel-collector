package config

import (
	"testing"

	"github.com/signalfx/defaults"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/stretchr/testify/require"
)

func TestConfigPropagation(t *testing.T) {
	t.Run("ProcPath is passed down if not present", func(t *testing.T) {
		c := &Config{
			ProcPath: "/hostfs/proc",
			Monitors: []MonitorConfig{
				{},
				{ProcPath: "/proc"},
			},
		}
		c.propagateValuesDown()

		require.Equal(t, c.Monitors[0].ProcPath, "/hostfs/proc")
		require.Equal(t, c.Monitors[1].ProcPath, "/proc")
	})
}

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
		require.Nil(t, defaults.Set(c))

		err := c.validate()
		require.NotNil(t, err)
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
		require.Nil(t, defaults.Set(c))

		err := c.validate()
		require.Nil(t, err)
	})
}
