// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cyberarkconfigsource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

const (
	// The "type" of CyberArk config sources in configuration.
	typeStr = "cyberark"

	defaultPollInterval = 1 * time.Minute
	defaultBinaryPath   = "CLIPasswordSDK"

	// retrievalModeCP is the Credential Provider backend, accessed via the local
	// CLIPasswordSDK binary. It is currently the only supported mode.
	retrievalModeCP = "cp"
)

// Private error types to help with testability.
type (
	errMissingAppID            struct{ error }
	errMissingSafe             struct{ error }
	errMissingObject           struct{ error }
	errNonPositivePollInterval struct{ error }
	errUnsupportedMode         struct{ error }
)

type cyberarkFactory struct{}

func (c *cyberarkFactory) Type() component.Type {
	return component.MustNewType(typeStr)
}

func (c *cyberarkFactory) CreateDefaultConfig() configsource.Settings {
	return &Config{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
		RetrievalMode:  retrievalModeCP,
		BinaryPath:     defaultBinaryPath,
		PollInterval:   defaultPollInterval,
	}
}

func (c *cyberarkFactory) CreateConfigSource(_ context.Context, settings configsource.Settings, logger *zap.Logger) (configsource.ConfigSource, error) {
	cyberarkCfg := settings.(*Config)

	if cyberarkCfg.RetrievalMode != retrievalModeCP {
		return nil, &errUnsupportedMode{fmt.Errorf("unsupported retrieval_mode %q, only %q is currently supported", cyberarkCfg.RetrievalMode, retrievalModeCP)}
	}

	if cyberarkCfg.AppID == "" {
		return nil, &errMissingAppID{errors.New("app_id cannot be empty")}
	}

	if cyberarkCfg.Safe == "" {
		return nil, &errMissingSafe{errors.New("safe cannot be empty")}
	}

	if cyberarkCfg.Object == "" {
		return nil, &errMissingObject{errors.New("object cannot be empty")}
	}

	if cyberarkCfg.AutoRefresh && cyberarkCfg.PollInterval <= 0 {
		return nil, &errNonPositivePollInterval{errors.New("poll_interval must be positive when auto_refresh is enabled")}
	}

	return newConfigSource(cyberarkCfg, logger)
}

// NewFactory creates a factory for CyberArk ConfigSource objects.
func NewFactory() configsource.Factory {
	return &cyberarkFactory{}
}
