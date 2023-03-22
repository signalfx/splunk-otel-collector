package heroku

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
}

// Monitor for Hereoku metadata
type Monitor struct {
	Output types.Output
	cancel context.CancelFunc
	ctx    context.Context
	logger *utils.ThrottledLogger
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Configure monitor
func (m *Monitor) Configure(c *Config) error {
	m.logger = utils.NewThrottledLogger(log.WithFields(log.Fields{"monitorType": "heroku-metadata", "monitorID": c.MonitorID}), 20*time.Second)
	m.ctx, m.cancel = context.WithCancel(context.Background())

	go func() {
		properties := map[string]string{}
		dynoID := os.Getenv("HEROKU_DYNO_ID")

		properties["heroku_release_version"] = os.Getenv("HEROKU_RELEASE_VERSION")
		properties["heroku_app_name"] = os.Getenv("HEROKU_APP_NAME")
		properties["heroku_slug_commit"] = os.Getenv("HEROKU_SLUG_COMMIT")
		properties["heroku_release_creation_timestamp"] = os.Getenv("HEROKU_RELEASE_CREATED_AT")
		properties["heroku_app_id"] = os.Getenv("HEROKU_APP_ID")

		m.Output.SendDimensionUpdate(&types.Dimension{
			Name:              "dyno_id",
			Value:             dynoID,
			Properties:        properties,
			MergeIntoExisting: false,
		})

	}()

	return nil
}
