// The factory file is responsible for providing the required ReceiverFactory object that every
// receiver must return to the opentelemetry receiever.

package githubmetricsreceiver

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

const (
    typeStr = "GithubMetrics"
    defaultInterval = 60 * time.Minute
)

func createDefaultConfig() config.Receiver {
    return &Config{
        ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
        Interval: defaultInterval.String(),
    }
}

func NewFactory() component.ReceiverFactory {
    return nil 
}
