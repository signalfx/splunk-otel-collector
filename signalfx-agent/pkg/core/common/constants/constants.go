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

package constants

const (
	// CollectdVersionEnvVar is the environment variable name that is set with the collectd version
	CollectdVersionEnvVar = "SIGNALFX_COLLECTD_VERSION"
	// AgentVersionEnvVar is the environment variable name that is set with the agent version
	AgentVersionEnvVar = "SIGNALFX_AGENT_VERSION"
	// BundleDirEnvVar is a path to the root of the collectd/python bundle
	BundleDirEnvVar = "SIGNALFX_BUNDLE_DIR"
)
