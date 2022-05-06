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

package pulsarexporter

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"time"
)

// Config defines configuration for pulsar exporter.
type Config struct {
	config.ExporterSettings        `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	exporterhelper.TimeoutSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`

	//// The list of pulsar brokers (default localhost:9092)
	Brokers string `mapstructure:"brokers"`
	//// The name of the pulsar topic to export to (default otlp_spans for traces, otlp_metrics for metrics)
	Topic string `mapstructure:"topic"`

	// Encoding of messages (default "otlp_proto")
	Encoding string `mapstructure:"encoding"`

	//// Metadata is the namespace for metadata management properties used by the
	//// Client, and shared by the Producer/Consumer.
	Metadata Metadata `mapstructure:"metadata"`

	// Producer is the namespaces for producer properties used only by the Producer
	Producer Producer `mapstructure:"producer"`

	// Authentication defines used authentication mechanism.
	Authentication Authentication `mapstructure:"auth"`
}

// Metadata defines configuration for retrieving metadata from the broker.
type Metadata struct {
}

// Producer defines configuration for producer
type Producer struct {
}

// MetadataRetry defines retry configuration for Metadata.
type MetadataRetry struct {
}

var _ config.Exporter = (*Config)(nil)

//Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	return nil
}

func (cfg *Config) getClientOptions() (*pulsar.ClientOptions, error) {

	return &pulsar.ClientOptions{URL: "pulsar+ssl://10.176.29.102:6651",
		OperationTimeout:           30 * time.Second,
		ConnectionTimeout:          30 * time.Second,
		TLSAllowInsecureConnection: true,
		TLSTrustCertsFilePath:      "",
		Authentication:             pulsar.NewAuthenticationTLS("my-cert.pem", "my-key.pem"),
		TLSValidateHostname:        false,
		MaxConnectionsPerBroker:    1}, nil
}

func pulsarProducerCompressionType(compressionType string) (pulsar.CompressionType, error) {
	switch compressionType {
	case "none":
		return pulsar.NoCompression, nil
	case "lz4":
		return pulsar.LZ4, nil
	case "zlib":
		return pulsar.ZLib, nil
	case "zstd":
		return pulsar.ZSTD, nil
	default:
		return pulsar.NoCompression, fmt.Errorf("producer.compressionType should be one of 'none', 'lz4', 'zlib', or 'zstd'. configured value %v", compressionType)
	}
}
