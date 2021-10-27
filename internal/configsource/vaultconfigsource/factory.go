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

package vaultconfigsource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/collector/config"
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	// The "type" of Vault config sources in configuration.
	typeStr = "vault"

	defaultPollInterval = 1 * time.Minute
)

// Private error types to help with testability.
type (
	errEmptyAuth               struct{ error }
	errEmptyToken              struct{ error }
	errInvalidEndpoint         struct{ error }
	errMissingAuthentication   struct{ error }
	errMissingEndpoint         struct{ error }
	errMissingPath             struct{ error }
	errMultipleAuthMethods     struct{ error }
	errNonPositivePollInterval struct{ error }
)

type vaultFactory struct{}

func (v *vaultFactory) Type() config.Type {
	return typeStr
}

func (v *vaultFactory) CreateDefaultConfig() expcfg.Source {
	return &Config{
		SourceSettings: expcfg.NewSourceSettings(config.NewComponentID(typeStr)),
		PollInterval:   defaultPollInterval,
	}
}

func (v *vaultFactory) CreateConfigSource(_ context.Context, params configprovider.CreateParams, cfg expcfg.Source) (configsource.ConfigSource, error) {
	vaultCfg := cfg.(*Config)

	if vaultCfg.Endpoint == "" {
		return nil, &errMissingEndpoint{errors.New("cannot connect to vault with an empty endpoint")}
	}

	if _, err := url.ParseRequestURI(vaultCfg.Endpoint); err != nil {
		return nil, &errInvalidEndpoint{fmt.Errorf("invalid endpoint %q: %w", vaultCfg.Endpoint, err)}
	}

	if vaultCfg.Path == "" {
		return nil, &errMissingPath{errors.New("cannot connect to vault with an empty path")}
	}

	if err := validateAuth(vaultCfg.Authentication); err != nil {
		return nil, err
	}

	if vaultCfg.PollInterval <= 0 {
		return nil, &errNonPositivePollInterval{errors.New("poll_interval must to be positive")}
	}

	return newConfigSource(params, vaultCfg)
}

// NewFactory creates a factory for Vault ConfigSource objects.
func NewFactory() configprovider.Factory {
	return &vaultFactory{}
}

func validateAuth(auth *Authentication) error {
	if auth == nil {
		return &errMissingAuthentication{errors.New("cannot connect to vault without an explicit auth method")}
	}

	countMethods := 0
	if auth.Token != nil {
		countMethods++
		if *auth.Token == "" {
			return &errEmptyToken{errors.New("token cannot be empty")}
		}
	}

	if auth.IAMAuthentication != nil {
		countMethods++
	}

	if auth.GCPAuthentication != nil {
		countMethods++
	}

	if countMethods == 0 {
		return &errEmptyAuth{errors.New("auth cannot be empty, exactly one method must be used")}
	}

	if countMethods > 1 {
		return &errMultipleAuthMethods{errors.New("multiple auth methods were set, use only one")}
	}

	return nil
}
