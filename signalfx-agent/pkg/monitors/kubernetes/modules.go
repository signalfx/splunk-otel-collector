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

package kubernetes

import (
	// Import the monitors so that they get registered
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/apiserver"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/controllermanager"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/events"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/kubeletmetrics"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/proxy"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/scheduler"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/volumes"
)
