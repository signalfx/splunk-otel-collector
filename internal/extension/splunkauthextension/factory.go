// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"errors"
	"net/http"
	"os"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension"

	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/modularinput"
)

// NewFactory creates authenticators configured through collector YAML.
func NewFactory() extension.Factory {
	return NewFactoryWithTokenSource(nil)
}

// NewFactoryWithTokenSource creates authenticators that refresh tokens using source.
// A source must return a complete listing on success, including an empty listing
// when no tokens remain. It must honor context cancellation.
func NewFactoryWithTokenSource(source TokenSource) extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		func() component.Config {
			if source != nil {
				return NewDefaultConfig()
			}
			return createDefaultConfig()
		},
		func(ctx context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
			config := cfg.(*Config)
			if err := config.Validate(); err != nil {
				return nil, err
			}
			tokenSource := source
			if tokenSource == nil {
				if config.Endpoint == "" || config.SessionKey == "" {
					return nil, errors.New("endpoint and session_key are required without an injected token source")
				}
				tlsConfig, err := config.TLS.LoadTLSConfig(ctx)
				if err != nil {
					return nil, err
				}
				transport := http.DefaultTransport.(*http.Transport).Clone()
				transport.TLSClientConfig = tlsConfig
				tokenSource = &splunkTokenSource{
					serverURI:  config.Endpoint,
					sessionKey: string(config.SessionKey),
					client:     &http.Client{Transport: transport},
				}
			} else if config.Endpoint != "" || config.SessionKey != "" {
				return nil, errors.New("endpoint and session_key cannot be combined with an injected token source")
			}
			return newSplunkAuth(config, tokenSource, set.Logger), nil
		},
		metadata.ExtensionStability,
	)
}

func createDefaultConfig() component.Config {
	cfg := NewDefaultConfig()
	cfg.Endpoint = os.Getenv(modularinput.EnvManagementURI)
	cfg.SessionKey = configopaque.String(os.Getenv(modularinput.EnvSessionKey))
	return cfg
}
