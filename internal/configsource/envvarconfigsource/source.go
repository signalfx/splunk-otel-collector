// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package envvarconfigsource

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// Private error types to help with testability.
type (
	errInvalidRetrieveParams struct{ error }
	errMissingRequiredEnvVar struct{ error }
)

type retrieveParams struct {
	// Optional is used to change the default behavior when an environment variable
	// requested via the config source is not defined. By default the value of this
	// field is 'false' which will cause an error if the specified environment variable
	// is not defined. Set it to 'true' to ignore not defined environment variables.
	Optional bool `mapstructure:"optional"`
}

// envVarConfigSource implements the configprovider.Session interface.
type envVarConfigSource struct {
	defaults map[string]any
}

func newConfigSource(_ configprovider.CreateParams, cfg *Config) configprovider.ConfigSource {
	defaults := make(map[string]any)
	if cfg.Defaults != nil {
		defaults = cfg.Defaults
	}

	return &envVarConfigSource{
		defaults: defaults,
	}
}

func (e *envVarConfigSource) Retrieve(_ context.Context, selector string, paramsConfigMap *confmap.Conf) (configprovider.Retrieved, error) {
	actualParams := retrieveParams{}
	if paramsConfigMap != nil {
		paramsParser := confmap.NewFromStringMap(paramsConfigMap.ToStringMap())
		if err := paramsParser.Unmarshal(&actualParams, confmap.WithErrorUnused()); err != nil {
			return nil, &errInvalidRetrieveParams{fmt.Errorf("failed to unmarshall retrieve params: %w", err)}
		}
	}

	value, ok := os.LookupEnv(selector)
	if ok {
		// Environment variable found, everything is done.
		return configprovider.NewRetrieved(value), nil
	}

	defaultValue, ok := e.defaults[selector]
	if !ok {
		if !actualParams.Optional {
			return nil, &errMissingRequiredEnvVar{fmt.Errorf("env var %q is required but not defined and not present on defaults", selector)}
		}
	}

	return configprovider.NewRetrieved(defaultValue), nil
}

func (e *envVarConfigSource) Close(context.Context) error {
	return nil
}
