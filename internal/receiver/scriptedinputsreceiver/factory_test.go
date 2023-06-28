package scriptedinputreceiver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
}
