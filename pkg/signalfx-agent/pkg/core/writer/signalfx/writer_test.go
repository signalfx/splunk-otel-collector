package signalfx

import (
	"fmt"
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/stretchr/testify/require"
)

var essentialWriterConfig = config.WriterConfig{
	SignalFxAccessToken:                 "11111",
	PropertiesHistorySize:               100,
	PropertiesSendDelaySeconds:          1,
	TraceExportFormat:                   "zipkin",
	TraceHostCorrelationMetricsInterval: timeutil.Duration(1 * time.Second),
	TraceHostCorrelationPurgeInterval:   timeutil.Duration(1 * time.Second),
	StaleServiceTimeout:                 timeutil.Duration(1 * time.Second),
	EventSendIntervalSeconds:            1,
}

func TestWriterSetup(t *testing.T) {
	t.Run("Overrides event URL", func(t *testing.T) {
		t.Parallel()
		conf := essentialWriterConfig
		conf.EventEndpointURL = "http://example.com/v2/event"
		writer, err := New(&conf, nil, nil, nil, nil, nil)

		require.Nil(t, err)
		require.Equal(t, "http://example.com/v2/event", writer.client.EventEndpoint)
	})

	t.Run("Sets default event URL", func(t *testing.T) {
		t.Parallel()
		conf := essentialWriterConfig
		conf.IngestURL = "http://example.com"
		writer, err := New(&conf, nil, nil, nil, nil, nil)
		require.Nil(t, err)
		require.Equal(t, "http://example.com/v2/event", writer.client.EventEndpoint)
	})
}

type tempError struct {
	temporary func() bool
}

func (t tempError) Error() string   { return fmt.Sprintf("%v", t.temporary()) }
func (t tempError) Temporary() bool { return t.temporary() }

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"not a transient error", fmt.Errorf("not_transient"), false},
		{"eof error", io.EOF, true},
		{"econnreset error", syscall.ECONNRESET, true},
		{"temporary error", tempError{func() bool { return true }}, true},
		{"not temporary error", tempError{func() bool { return false }}, false},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(tt *testing.T) {
			require.Equal(tt, test.expected, isTransientError(test.err))

			wrapped := fmt.Errorf("%w", test.err)
			require.Equal(tt, test.expected, isTransientError(wrapped))

			doublyWrapped := fmt.Errorf("%w", wrapped)
			require.Equal(tt, test.expected, isTransientError(doublyWrapped))
		})
	}
}
