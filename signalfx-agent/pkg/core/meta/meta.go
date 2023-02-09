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

package meta

// AgentMeta provides monitors access to global agent metadata.  Putting this
// into a single interface allows easy expansion of metadata without breaking
// backwards-compatibility and without exposing global variables that monitors
// access.
// TODO: get rid of this since it's hacky
type AgentMeta struct {
	InternalStatusHost string
	InternalStatusPort uint16
}
