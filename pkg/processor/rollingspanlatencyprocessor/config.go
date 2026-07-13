// Copyright  Splunk, Inc.
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

package rollingspanlatencyprocessor // import "github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor"

import (
	"time"

	"go.opentelemetry.io/collector/component"
)

var _ component.Config = (*Config)(nil)

// Config holds configuration for the rolling span latency processor.
type Config struct {
	// HalfLife controls the EWMA decay. The default (2h) means the effective
	// weight of a sample halves every 2 hours, biasing the baseline toward
	// recent data while retaining long-term signal.
	HalfLife time.Duration `mapstructure:"half_life"`

	// SlowThreshold is the number of standard deviations above the EWMA mean
	// at which a span is labeled "slow". Default: 3.
	SlowThreshold float64 `mapstructure:"slow_threshold"`

	// VerySlowThreshold is the number of standard deviations above the EWMA
	// mean at which a span is labeled "very_slow". Default: 4.
	VerySlowThreshold float64 `mapstructure:"very_slow_threshold"`

	// AttributeKey is the span attribute key written when a span is slow or
	// very slow. Default: "latency.category".
	AttributeKey string `mapstructure:"attribute_key"`

	// ResourceKeyAttributes is the ordered list of resource attribute keys
	// whose values are included in the baseline key, alongside the span name.
	// Spans that share the same values for all of these attributes and the same
	// span name share a single EWMA baseline.
	//
	// Default: ["service.namespace", "service.name", "deployment.environment.name"]
	ResourceKeyAttributes []string `mapstructure:"resource_key_attributes"`

	// IdleTimeout is how long a baseline key must go without any observations
	// before it is evicted from memory. This prevents unbounded growth when
	// services are retired or span names change.
	// Default: 4 × HalfLife (8h with the default 2h half-life).
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`

	// EvictionInterval controls how often the background eviction sweep runs.
	// Shorter intervals reduce peak memory at the cost of more frequent lock
	// contention on the stats map. Default: 10m.
	EvictionInterval time.Duration `mapstructure:"eviction_interval"`

	// MaxBaselines is the maximum number of baseline entries held in memory at
	// once. When the cap is reached, new keys are dropped and a warning is
	// logged. 0 means unlimited. Default: 0.
	MaxBaselines int `mapstructure:"max_baselines"`

	// ChurnWarningRatio is the fraction of the active baseline count that, when
	// exceeded by a single eviction sweep's evicted count, triggers a warning
	// log indicating high key churn. For example, 0.5 warns when more than 50%
	// of active baselines were replaced within one eviction interval. Default: 0.5.
	ChurnWarningRatio float64 `mapstructure:"churn_warning_ratio"`

	// WarmupCount is the minimum number of observations required before a
	// baseline key is eligible for labeling. Until this count is reached the
	// mean and variance are still accumulated, but no attribute is written.
	// Default: 30.
	WarmupCount int `mapstructure:"warmup_count"`

	// MinStddev is the minimum standard deviation (in nanoseconds) used when
	// scoring a span. When the EWMA stddev falls below this floor the floor is
	// used instead, preventing near-zero variance from turning small OS jitter
	// into false positives. Default: 1ms (1_000_000 ns).
	MinStddev float64 `mapstructure:"min_stddev"`
}

var defaultResourceKeyAttributes = []string{
	"service.namespace",
	"service.name",
	"deployment.environment.name",
}

func defaultConfig() Config {
	attrs := make([]string, len(defaultResourceKeyAttributes))
	copy(attrs, defaultResourceKeyAttributes)
	return Config{
		HalfLife:              2 * time.Hour,
		SlowThreshold:         3.0,
		VerySlowThreshold:     4.0,
		AttributeKey:          "latency.category",
		ResourceKeyAttributes: attrs,
		IdleTimeout:           8 * time.Hour,
		EvictionInterval:      10 * time.Minute,
		MaxBaselines:          0,
		ChurnWarningRatio:     0.5,
		WarmupCount:           30,
		MinStddev:             1e6, // 1ms in nanoseconds
	}
}

func (c *Config) Validate() error {
	if c.HalfLife <= 0 {
		return errInvalidHalfLife
	}
	if c.SlowThreshold <= 0 {
		return errInvalidSlowThreshold
	}
	if c.VerySlowThreshold <= c.SlowThreshold {
		return errVerySlowMustExceedSlow
	}
	if c.AttributeKey == "" {
		return errEmptyAttributeKey
	}
	if len(c.ResourceKeyAttributes) == 0 {
		return errEmptyResourceKeyAttributes
	}
	if c.IdleTimeout <= 0 {
		return errInvalidIdleTimeout
	}
	if c.EvictionInterval <= 0 {
		return errInvalidEvictionInterval
	}
	if c.MaxBaselines < 0 {
		return errNegativeMaxBaselines
	}
	if c.ChurnWarningRatio <= 0 || c.ChurnWarningRatio > 1 {
		return errInvalidChurnWarningRatio
	}
	if c.WarmupCount <= 0 {
		return errInvalidWarmupCount
	}
	if c.MinStddev < 0 {
		return errInvalidMinStddev
	}
	return nil
}
