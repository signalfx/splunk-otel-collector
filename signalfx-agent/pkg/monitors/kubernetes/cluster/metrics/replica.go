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

package metrics

import (
	"fmt"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
)

func makeReplicaDPs(resource string, dimensions map[string]string, desired, available int32) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New(
			fmt.Sprintf("kubernetes.%s.desired", resource),
			dimensions,
			datapoint.NewIntValue(int64(desired)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			fmt.Sprintf("kubernetes.%s.available", resource),
			dimensions,
			datapoint.NewIntValue(int64(available)),
			datapoint.Gauge,
			time.Time{}),
	}
}
