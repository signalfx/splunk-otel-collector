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
	"context"

	"go.opentelemetry.io/collector/confmap"
)

// ConfigSource is the interface to be implemented by objects used by the collector
// to retrieve external configuration information.
//
// ConfigSource object will be used to retrieve full configuration or data to be
// injected into a configuration.
//
// The ConfigSource object should use its creation according to the source needs:
// lock resources, open connections, etc. An implementation, for instance,
// can use the creation time to prevent torn configurations, by acquiring a lock
// (or some other mechanism) that prevents concurrent changes to the configuration
// during time that data is being retrieved from the source.
//
// The code managing the ConfigSource instance must guarantee that the object is not used concurrently.
type ConfigSource interface {
	// Retrieve goes to the configuration source and retrieves the selected data which
	// contains the value to be injected in the configuration and the corresponding watcher that
	// will be used to monitor for updates of the retrieved value. The retrieved value is selected
	// according to the selector and the params passed in the call to Retrieve.
	//
	// The selector is a string that is required on all invocations, the params are optional. Each
	// implementation handles the generic params according to their requirements.
	Retrieve(ctx context.Context, selector string, paramsConfigMap *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error)

	// Shutdown signals that the configuration for which it was used to retrieve values is no longer in use
	// and the object should close and release any watchers that it may have created.
	// This method must be called when the configuration session ends, either in case of success or error.
	Shutdown(ctx context.Context) error
}
