// Copyright 2020 Splunk, Inc.
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

// Package configsources lists all config sources that can be used in the configuration.
package configsources

import (
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/etcd2configsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/vaultconfigsource"
	"github.com/signalfx/splunk-otel-collector/internal/configsource/zookeeperconfigsource"
)

// Get returns the factories to all config sources available to the user.
func Get() []configprovider.Factory {
	return []configprovider.Factory{
		etcd2configsource.NewFactory(),
		vaultconfigsource.NewFactory(),
		zookeeperconfigsource.NewFactory(),
	}
}
