// The factory file is responsible for providing the required ReceiverFactory object that every
// receiver must return to the opentelemetry receiever.

package githubmetricsreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
)

// define default values as constants here
const (
    typeStr         = "GithubMetrics"
    defaultInterval = 60 * time.Minute
    defaultEndpoint = "https://api.github.com"
    defaultTimeout  = 10 * time.Second
)

func createDefaultConfig() config.Receiver {
    return &Config{
        HTTPClientSettings: confighttp.HTTPClientSettings{
            Endpoint: defaultEndpoint, 
            Timeout: defaultTimeout,
        },
        ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
        Interval: defaultInterval.String(),
    }
}

func NewFactory() component.ReceiverFactory {
    return component.NewReceiverFactory(
        typeStr,
        createDefaultConfig,
        component.WithMetricsReceiver(createMetricsReceiver, component.StabilityLevelAlpha))
}

func createMetricsReceiver(_ context.Context,
    params component.ReceiverCreateSettings,
    baseCfg config.Receiver,
    consumer consumer.Metrics) (component.MetricsReceiver, error) {
    return nil, nil
}
