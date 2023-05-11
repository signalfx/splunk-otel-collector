package config

import (
	"time"

	"github.com/signalfx/signalfx-agent/pkg/apm/correlations"
)

func ClientConfigFromWriterConfig(conf *WriterConfig) correlations.ClientConfig {
	return correlations.ClientConfig{
		Config: correlations.Config{
			MaxRequests:     conf.PropertiesMaxRequests,
			MaxBuffered:     conf.PropertiesMaxBuffered,
			MaxRetries:      conf.TraceHostCorrelationMaxRequestRetries,
			LogUpdates:      conf.LogDimensionUpdates,
			RetryDelay:      time.Duration(conf.PropertiesSendDelaySeconds) * time.Second,
			CleanupInterval: conf.TraceHostCorrelationPurgeInterval.AsDuration(),
		},
		AccessToken: conf.SignalFxAccessToken,
		URL:         conf.ParsedAPIURL(),
	}
}
