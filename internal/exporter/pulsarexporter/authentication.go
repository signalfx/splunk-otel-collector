// Copyright  The OpenTelemetry Authors
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

package pulsarexporter

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"

	"go.opentelemetry.io/collector/config/configtls"
)

// Authentication defines authentication.
type Authentication struct {
	TLS *configtls.TLSClientSetting `mapstructure:"tls"`
}

// ConfigureAuthentication configures authentication in sarama.Config.
func ConfigureAuthentication(config Authentication, pulsarClient *pulsar.ClientOptions) error {
	if config.TLS != nil {
		if err := configureTLS(*config.TLS, pulsarClient); err != nil {
			return err
		}
	}
	return nil
}

func configureTLS(config configtls.TLSClientSetting, pulsarClient *pulsar.ClientOptions) error {
	tlsConfig, err := config.LoadTLSConfig()
	if err != nil {
		return fmt.Errorf("error loading tls config: %w", err)
	}

	pulsarClient.TLSAllowInsecureConnection = tlsConfig.InsecureSkipVerify
	return nil
}
