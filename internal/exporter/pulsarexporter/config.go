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
	"errors"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.uber.org/zap"
)

type Authentication struct {
	TLS *configtls.TLSClientSetting `mapstructure:"tls"`
}

// Config defines configuration for pulsar exporter.
type Config struct {
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	Authentication                 Authentication `mapstructure:"auth"`
	Broker                         string         `mapstructure:"broker"`
	Topic                          string         `mapstructure:"topic"`
	Encoding                       string         `mapstructure:"encoding"`
	Producer                       Producer       `mapstructure:"producer"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`
	exporterhelper.TimeoutSettings `mapstructure:",squash"`
	OperationTimeout               time.Duration `mapstructure:"operation_timeout"`
	ConnectionTimeout              time.Duration `mapstructure:"connection_timeout"`
}

// Producer defines configuration for producer
type Producer struct {
	Properties                      map[string]string `mapstructure:"producer_properties"`
	MaxReconnectToBroker            *uint             `mapstructure:"max_reconnect_broker"`
	SendTimeout                     *time.Duration    `mapstructure:"send_timeout"`
	BatcherBuilderType              string            `mapstructure:"batch_builder_type"`
	CompressionType                 string            `mapstructure:"compression_type"`
	CompressionLevel                string            `mapstructure:"compression_level"`
	HashingScheme                   string            `mapstructure:"hashing_scheme"`
	MaxPendingMessages              int               `mapstructure:"max_pending_messages"`
	PartitionsAutoDiscoveryInterval time.Duration     `mapstructure:"partitions_auto_discovery_interval"`
	BatchingMaxPublishDelay         time.Duration     `mapstructure:"batching_max_publish_delay"`
	BatchingMaxMessages             uint              `mapstructure:"batching_max_messages"`
	BatchingMaxSize                 uint              `mapstructure:"batching_max_size"`
	DisableBlockIfQueueFull         bool              `mapstructure:"disable_block_if_queue_full"`
	DisableBatching                 bool              `mapstructure:"disable_batching"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	return nil
}

func (cfg *Config) getClientOptions() (pulsar.ClientOptions, error) {

	options := pulsar.ClientOptions{
		URL:                     cfg.Broker,
		OperationTimeout:        cfg.OperationTimeout,
		ConnectionTimeout:       cfg.ConnectionTimeout,
		MaxConnectionsPerBroker: 1,
	}

	if cfg.Authentication.TLS.InsecureSkipVerify {
		options.TLSAllowInsecureConnection = cfg.Authentication.TLS.InsecureSkipVerify
		return options, nil
	}

	if len(cfg.Authentication.TLS.CAFile) > 0 && len(cfg.Authentication.TLS.CertFile) > 0 && len(cfg.Authentication.TLS.KeyFile) > 0 {
		options.TLSTrustCertsFilePath = cfg.Authentication.TLS.CAFile
		options.Authentication = pulsar.NewAuthenticationTLS(cfg.Authentication.TLS.CertFile, cfg.Authentication.TLS.KeyFile)
	} else {
		return options, errors.New("failed to load TLS config. If certs are not available, set insecure_skip_verify to true for insecure connection")
	}

	return options, nil
}

func (cfg *Config) getProducerOptions(logger *zap.Logger) (pulsar.ProducerOptions, error) {
	// Properties are not used. Issue a warning that these are no longer used.
	if cfg.Producer.Properties != nil {
		logger.Warn("`producer.properties` is no longer used and will be removed in a subsequent release. Please remove this property from the configuration.")
	}
	if cfg.Producer.BatcherBuilderType == "1" {
		logger.Warn("`producer.batch_builder_type` value 1 is deprecated and should use the value `key_based` instead.")
	}
	if cfg.Producer.BatcherBuilderType == "0" {
		logger.Warn("`producer.batch_builder_type` value 0 is deprecated and should use the value `default` instead.")
	}
	timeout := cfg.Timeout
	if cfg.Producer.SendTimeout != nil {
		logger.Warn("`producer.send_timeout` is deprecated and will be removed in a subsequent release. Please use `timeout` instead")
		timeout = *cfg.Producer.SendTimeout
	}

	producerOptions := pulsar.ProducerOptions{
		Topic:                           cfg.Topic,
		DisableBatching:                 cfg.Producer.DisableBatching,
		SendTimeout:                     timeout,
		DisableBlockIfQueueFull:         cfg.Producer.DisableBlockIfQueueFull,
		MaxPendingMessages:              cfg.Producer.MaxPendingMessages,
		BatchingMaxPublishDelay:         cfg.Producer.BatchingMaxPublishDelay,
		BatchingMaxSize:                 cfg.Producer.BatchingMaxSize,
		BatchingMaxMessages:             cfg.Producer.BatchingMaxMessages,
		PartitionsAutoDiscoveryInterval: cfg.Producer.PartitionsAutoDiscoveryInterval,
		MaxReconnectToBroker:            cfg.Producer.MaxReconnectToBroker,
	}

	batchBuilderType, err := stringToBatchBuilderType(cfg.Producer.BatcherBuilderType)
	if err != nil {
		return producerOptions, err
	}
	producerOptions.BatcherBuilderType = batchBuilderType

	compressionType, err := stringToCompressionType(cfg.Producer.CompressionType)
	if err != nil {
		return producerOptions, err
	}
	producerOptions.CompressionType = compressionType

	compressionLevel, err := stringToCompressionLevel(cfg.Producer.CompressionLevel)
	if err != nil {
		return producerOptions, err
	}
	producerOptions.CompressionLevel = compressionLevel

	hashingScheme, err := stringToHashingScheme(cfg.Producer.HashingScheme)
	if err != nil {
		return producerOptions, err
	}
	producerOptions.HashingScheme = hashingScheme

	return producerOptions, nil
}

func stringToCompressionType(compressionType string) (pulsar.CompressionType, error) {
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
		return pulsar.NoCompression, fmt.Errorf("producer.compressionType should be one of 'none', 'lz4', 'zlib', or 'zstd'. configured value %v. Assigning default value as none", compressionType)
	}
}

func stringToCompressionLevel(compressionLevel string) (pulsar.CompressionLevel, error) {
	switch compressionLevel {
	case "default":
		return pulsar.Default, nil
	case "faster":
		return pulsar.Faster, nil
	case "better":
		return pulsar.Better, nil
	default:
		return pulsar.Default, fmt.Errorf("producer.compressionLevel should be one of 'default', 'faster', or 'better'. configured value %v. Assigning default value as default", compressionLevel)
	}
}

func stringToHashingScheme(hashingScheme string) (pulsar.HashingScheme, error) {
	switch hashingScheme {
	case "java_string_hash":
		return pulsar.JavaStringHash, nil
	case "murmur3_32hash":
		return pulsar.Murmur3_32Hash, nil
	default:
		return pulsar.JavaStringHash, fmt.Errorf("producer.hashingScheme should be one of 'java_string_hash' or 'murmur3_32hash'. configured value %v, Assigning default value as java_string_hash", hashingScheme)
	}
}

func stringToBatchBuilderType(builderType string) (pulsar.BatcherBuilderType, error) {
	switch builderType {
	case "0":
		return pulsar.DefaultBatchBuilder, nil
	case "1":
		return pulsar.KeyBasedBatchBuilder, nil
	case "default":
		return pulsar.DefaultBatchBuilder, nil
	case "key_based":
		return pulsar.KeyBasedBatchBuilder, nil
	default:
		return pulsar.DefaultBatchBuilder, fmt.Errorf("producer.batchBuilderType should be one of 'default' or 'key_based'. configured value %v. Assigning default value as default", builderType)
	}
}
