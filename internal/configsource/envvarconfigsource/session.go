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

	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"

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

// envVarSession implements the configsource.Session interface.
type envVarSession struct {
	defaults map[string]interface{}
}

var _ configsource.Session = (*envVarSession)(nil)

func (e *envVarSession) Retrieve(_ context.Context, selector string, params interface{}) (configsource.Retrieved, error) {
	actualParams := retrieveParams{}
	if params != nil {
		paramsParser := config.NewParserFromStringMap(cast.ToStringMap(params))
		if err := paramsParser.UnmarshalExact(&actualParams); err != nil {
			return nil, &errInvalidRetrieveParams{fmt.Errorf("failed to unmarshall retrieve params: %w", err)}
		}
	}

	value, ok := os.LookupEnv(selector)
	if ok {
		// Environment variable found, everything is done.
		return configprovider.NewRetrieved(value, configprovider.WatcherNotSupported), nil
	}

	defaultValue, ok := e.defaults[selector]
	if !ok {
		if !actualParams.Optional {
			return nil, &errMissingRequiredEnvVar{fmt.Errorf("env var %q is required but not defined and not present on defaults", selector)}
		}

		// To keep with default behavior for env vars not defined set the value to empty string
		defaultValue = ""
	}

	return configprovider.NewRetrieved(defaultValue, configprovider.WatcherNotSupported), nil
}

func (e *envVarSession) RetrieveEnd(context.Context) error {
	return nil
}

func (e *envVarSession) Close(context.Context) error {
	return nil
}

func newSession(defaults map[string]interface{}) (*envVarSession, error) {
	return &envVarSession{
		defaults: defaults,
	}, nil
}
