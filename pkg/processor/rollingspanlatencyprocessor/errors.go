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

import "errors"

var (
	errInvalidHalfLife            = errors.New("half_life must be a positive duration")
	errInvalidSlowThreshold       = errors.New("slow_threshold must be positive")
	errVerySlowMustExceedSlow     = errors.New("very_slow_threshold must be greater than slow_threshold")
	errEmptyAttributeKey          = errors.New("attribute_key must not be empty")
	errEmptyResourceKeyAttributes = errors.New("resource_key_attributes must contain at least one entry")
	errInvalidIdleTimeout         = errors.New("idle_timeout must be a positive duration")
	errInvalidEvictionInterval    = errors.New("eviction_interval must be a positive duration")
	errNegativeMaxBaselines       = errors.New("max_baselines must be >= 0 (0 means unlimited)")
	errInvalidChurnWarningRatio   = errors.New("churn_warning_ratio must be in the range (0, 1]")
	errInvalidWarmupCount         = errors.New("warmup_count must be > 0")
	errInvalidMinStddev           = errors.New("min_stddev must be >= 0")
)
