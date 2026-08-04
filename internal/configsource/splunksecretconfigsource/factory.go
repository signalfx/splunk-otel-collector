// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package splunksecretconfigsource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
	"github.com/signalfx/splunk-otel-collector/pkg/modularinput"
)

const (
	// The "type" of splunk_secret config sources in configuration.
	typeStr = "splunk_secret"

	defaultTimeout = 10 * time.Second

	// defaultApp and defaultUser are the servicesNS wildcards meaning "all apps" and "all users".
	defaultApp  = "-"
	defaultUser = "-"
)

// Private error types to help with testability.
type (
	errMissingEndpoint    struct{ error }
	errInvalidEndpoint    struct{ error }
	errMissingSessionKey  struct{ error }
	errNonPositiveTimeout struct{ error }
)

type splunkSecretFactory struct{}

func (f *splunkSecretFactory) Type() component.Type {
	return component.MustNewType(typeStr)
}

func (f *splunkSecretFactory) CreateDefaultConfig() configsource.Settings {
	// Defaulting Endpoint/SessionKey to these env vars (set by HandleLaunchAsTA when running
	// as a Splunk TA modular input) lets the config source work out-of-the-box.
	return &Config{
		SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
		Endpoint:       os.Getenv(modularinput.EnvManagementURI),
		SessionKey:     os.Getenv(modularinput.EnvSessionKey),
		App:            defaultApp,
		User:           defaultUser,
		Timeout:        defaultTimeout,
	}
}

func (f *splunkSecretFactory) CreateConfigSource(_ context.Context, settings configsource.Settings, logger *zap.Logger) (configsource.ConfigSource, error) {
	cfg := settings.(*Config)

	if cfg.Endpoint == "" {
		return nil, &errMissingEndpoint{
			fmt.Errorf("'endpoint' must be specified, this value can also be set via the environment variable %s", modularinput.EnvManagementURI),
		}
	}
	if _, err := url.ParseRequestURI(cfg.Endpoint); err != nil {
		return nil, &errInvalidEndpoint{err}
	}
	if cfg.SessionKey == "" {
		return nil, &errMissingSessionKey{
			fmt.Errorf("'session_key' must be specified, this value can also be set via the environment variable %s", modularinput.EnvSessionKey),
		}
	}

	app := cfg.App
	if app == "" {
		app = defaultApp
	}

	user := cfg.User
	if user == "" {
		user = defaultUser
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		if timeout < 0 {
			return nil, &errNonPositiveTimeout{errors.New("'timeout' must be positive")}
		}
		timeout = defaultTimeout
	}

	return newConfigSource(cfg, app, user, timeout, logger), nil
}

// NewFactory creates a new splunkSecretFactory instance.
func NewFactory() configsource.Factory {
	return &splunkSecretFactory{}
}
