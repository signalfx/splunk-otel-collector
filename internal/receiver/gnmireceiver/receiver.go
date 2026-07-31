// Copyright Splunk, Inc.
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

package gnmireceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

// gnmiReceiver is a skeleton no-op metrics receiver. The gNMI dial-in session
// and metric conversion are implemented in follow-up changes.
type gnmiReceiver struct{}

var _ receiver.Metrics = (*gnmiReceiver)(nil)

func (r *gnmiReceiver) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (r *gnmiReceiver) Shutdown(_ context.Context) error {
	return nil
}
