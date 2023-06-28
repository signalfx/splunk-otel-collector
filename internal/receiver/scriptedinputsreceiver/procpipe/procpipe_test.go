package procpipe

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)

	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, builded, "failed to create default config")
}

func TestCreateDefaultWithSmallLogSize(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, err := config.Build(sugar)
	assert.Nil(t, builded, "failed to create default config")
	assert.Equal(t, err.Error(), "invalid value for parameter 'max_log_size', must be equal to or greater than 65536 bytes")
}

func TestCreateDefaultWithMissingExecFile(t *testing.T) {
	config := Config{}
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, err := config.Build(sugar)
	assert.Nil(t, builded, "failed to create default config")
	assert.Equal(t, err.Error(), "'exec_file' must be specified")
}

func TestCreateDefaultWithNonEmptyMultiline(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	config.Multiline = helper.MultilineConfig{
		LineStartPattern: "a",
		LineEndPattern:   "",
	}

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)
	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, builded, "failed to create default config")
}

func TestCreateTicker(t *testing.T) {
	ticker, err := createTicker("20s")

	assert.NotNil(t, ticker, "failed to create default config")
	assert.Nil(t, err, "failed to create default config")
}

func TestReadOutput(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"
	config.CollectionInterval = "60s"

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)
	builded.Start(nil)
	builded.Stop()

	assert.NotNil(t, builded, "failed to create default config")
}
