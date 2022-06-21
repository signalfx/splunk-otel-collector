// Copyright Splunk, Inc.
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

package configprovider

import (
	"go.opentelemetry.io/collector/config/experimental/configsource"
)

type retrieved struct {
	value any
}

// NewRetrieved is a helper that implements the Retrieved interface.
func NewRetrieved(value any) configsource.Retrieved {
	return &retrieved{
		value,
	}
}

var _ configsource.Retrieved = (*retrieved)(nil)

func (r *retrieved) Value() any {
	return r.value
}

type watchableRetrieved struct {
	retrieved
	watchForUpdateFn func() error
}

// NewWatchableRetrieved is a helper that implements the Watchable interface.
func NewWatchableRetrieved(value any, watchForUpdateFn func() error) configsource.Retrieved {
	return &watchableRetrieved{
		retrieved: retrieved{
			value: value,
		},
		watchForUpdateFn: watchForUpdateFn,
	}
}

var _ configsource.Watchable = (*watchableRetrieved)(nil)
var _ configsource.Retrieved = (*watchableRetrieved)(nil)

func (r *watchableRetrieved) Value() any {
	return r.retrieved.value
}

func (r *watchableRetrieved) WatchForUpdate() error {
	return r.watchForUpdateFn()
}
