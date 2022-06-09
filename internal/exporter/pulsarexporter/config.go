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
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)


type Authentication struct {
	TLS *configtls.TLSClientSetting `mapstructure:"tls"`
}
// Config defines configuration for pulsar exporter.
type Config struct {
	config.ExporterSettings        `mapstructure:",squash"`
	exporterhelper.TimeoutSettings `mapstructure:",squash"`
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`

	// The list of pulsar brokers (default localhost:9092)
	Broker string `mapstructure:"broker"`
	// The name of the pulsar topic to export to (default otlp_spans for traces, otlp_metrics for metrics)
	Topic string `mapstructure:"topic"`

	// Encoding of messages (default "otlp_proto")
	Encoding string `mapstructure:"encoding"`

	// Producer is the namespaces for producer properties used only by the Producer
	Producer Producer `mapstructure:"producer"`

	// Authentication defines used authentication mechanism.
	Authentication Authentication `mapstructure:"auth"`

	//Operation time out
	OperationTimeout time.Duration `mapstructure:"operation_timeout"`

	//Connection time out
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}

// Producer defines configuration for producer
type Producer struct {
	// Name specifies a name for the producer.
	// If not assigned, the system will generate a globally unique name which can be access with
	// Producer.ProducerName().
	// When specifying a name, it is up to the user to ensure that, for a given topic, the producer name is unique
	// across all Pulsar's clusters. Brokers will enforce that only a single producer a given name can be publishing on
	// a topic.
	Name string `mapstructure:"producer_name"`

	// Properties specifies a set of application defined properties for the producer.
	// This properties will be visible in the topic stats
	Properties map[string]string `mapstructure:"producer_properties"`

	// SendTimeout specifies the timeout for a message that has not been acknowledged by the server since sent.
	// Send and SendAsync returns an error after timeout.
	// Default is 30 seconds, negative such as -1 to disable.
	SendTimeout time.Duration `mapstructure:"send_timeout"`

	// DisableBlockIfQueueFull controls whether Send and SendAsync block if producer's message queue is full.
	// Default is false, if set to true then Send and SendAsync return error when queue is full.
	DisableBlockIfQueueFull bool `mapstructure:"disable_block_if_queue_full"`

	// MaxPendingMessages specifies the max size of the queue holding the messages pending to receive an
	// acknowledgment from the broker.
	MaxPendingMessages int `mapstructure:"max_pending_messages"`

	// HashingScheme is used to define the partition on where to publish a particular message.
	// Standard hashing functions available are:
	//
	//  - `JavaStringHash` : Java String.hashCode() equivalent
	//  - `Murmur3_32Hash` : Use Murmur3 hashing function.
	// 		https://en.wikipedia.org/wiki/MurmurHash">https://en.wikipedia.org/wiki/MurmurHash
	//
	// Default is `JavaStringHash`.
	HashingScheme string `mapstructure:"hashing_scheme"`

	// CompressionType specifies the compression type for the producer.
	// By default, message payloads are not compressed. Supported compression types are:
	//  - LZ4
	//  - ZLIB
	//  - ZSTD
	//
	// Note: ZSTD is supported since Pulsar 2.3. Consumers will need to be at least at that
	// release in order to be able to receive messages compressed with ZSTD.
	CompressionType string `mapstructure:"compression_type"`

	// CompressionLevel defines the desired compression level. Options:
	// - Default
	// - Faster
	// - Better
	CompressionLevel string `mapstructure:"compression_level"`

	// DisableBatching controls whether automatic batching of messages is enabled for the producer. By default batching
	// is enabled.
	// When batching is enabled, multiple calls to Producer.sendAsync can result in a single batch to be sent to the
	// broker, leading to better throughput, especially when publishing small messages. If compression is enabled,
	// messages will be compressed at the batch level, leading to a much better compression ratio for similar headers or
	// contents.
	// When enabled default batch delay is set to 1 ms and default batch size is 1000 messages
	// Setting `DisableBatching: true` will make the producer to send messages individually
	DisableBatching bool `mapstructure:"disable_batching"`

	// BatchingMaxPublishDelay specifies the time period within which the messages sent will be batched (default: 10ms)
	// if batch messages are enabled. If set to a non zero value, messages will be queued until this time
	// interval or until
	BatchingMaxPublishDelay time.Duration `mapstructure:"batching_max_publish_delay"`

	// BatchingMaxMessages specifies the maximum number of messages permitted in a batch. (default: 1000)
	// If set to a value greater than 1, messages will be queued until this threshold is reached or
	// BatchingMaxSize (see below) has been reached or the batch interval has elapsed.
	BatchingMaxMessages uint `mapstructure:"batching_max_messages"`

	// BatchingMaxSize specifies the maximum number of bytes permitted in a batch. (default 128 KB)
	// If set to a value greater than 1, messages will be queued until this threshold is reached or
	// BatchingMaxMessages (see above) has been reached or the batch interval has elapsed.
	BatchingMaxSize uint `mapstructure:"batching_max_size"`

	// MaxReconnectToBroker specifies the maximum retry number of reconnectToBroker. (default: ultimate)
	MaxReconnectToBroker *uint `mapstructure:"max_reconnect_broker"`

	// BatcherBuilderType sets the batch builder type (default DefaultBatchBuilder)
	// This will be used to create batch container when batching is enabled.
	// Options:
	// - DefaultBatchBuilder
	// - KeyBasedBatchBuilder
	BatcherBuilderType int `mapstructure:"batch_builder_type"`

	// PartitionsAutoDiscoveryInterval is the time interval for the background process to discover new partitions
	// Default is 1 minute
	PartitionsAutoDiscoveryInterval time.Duration `mapstructure:"partitions_auto_discovery_interval"`
}

var _ config.Exporter = (*Config)(nil)

//Validate checks if the exporter configuration is valid
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

	if cfg.Authentication.TLS.InsecureSkipVerify == true {
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

func (cfg *Config) getProducerOptions() (pulsar.ProducerOptions, error) {
	producerOptions := pulsar.ProducerOptions{
		Topic:                           cfg.Topic,
		Name:                            cfg.Producer.Name,
		DisableBatching:                 cfg.Producer.DisableBatching,
		SendTimeout:                     cfg.Producer.SendTimeout,
		DisableBlockIfQueueFull:         cfg.Producer.DisableBlockIfQueueFull,
		MaxPendingMessages:              cfg.Producer.MaxPendingMessages,
		BatchingMaxPublishDelay:         cfg.Producer.BatchingMaxPublishDelay,
		BatchingMaxSize:                 cfg.Producer.BatchingMaxSize,
		BatchingMaxMessages:             cfg.Producer.BatchingMaxMessages,
		PartitionsAutoDiscoveryInterval: cfg.Producer.PartitionsAutoDiscoveryInterval,
		MaxReconnectToBroker:            cfg.Producer.MaxReconnectToBroker,
	}

	compressionType, err := stringToCompressionType(cfg.Producer.CompressionType)
	if err == nil {
		producerOptions.CompressionType = compressionType
	} else {
		log.Info().Msgf("%v",err)
	}

	compressionLevel, err := stringToCompressionLevel(cfg.Producer.CompressionLevel)
	if err == nil {
		producerOptions.CompressionLevel = compressionLevel
	} else {
		log.Info().Msgf("%v",err)
	}

	hashingScheme, err := stringToHashingScheme(cfg.Producer.HashingScheme)
	if err == nil {
		producerOptions.HashingScheme = hashingScheme
	} else {
		log.Info().Msgf("%v",err)
	}

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
		return pulsar.NoCompression, fmt.Errorf("producer.compressionType should be one of 'none', 'lz4', 'zlib', or 'zstd'. configured value %v. Assiging default value as nocompression", compressionType)
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
		return pulsar.Default, fmt.Errorf("producer.compressionLevel should be one of 'none', 'lz4', 'zlib', or 'zstd'. configured value %v. Assiging default value as default", compressionLevel)
	}
}

func stringToHashingScheme(hashingScheme string) (pulsar.HashingScheme, error) {
	switch hashingScheme {
	case "java_string_hash":
		return pulsar.JavaStringHash, nil
	case "murmur3_32hash":
		return pulsar.Murmur3_32Hash, nil
	default:
		return pulsar.JavaStringHash, fmt.Errorf("producer.hashingScheme should be one of 'none', 'lz4', 'zlib', or 'zstd'. configured value %v, Assiging default value as java_string_hash", hashingScheme)
	}
}
