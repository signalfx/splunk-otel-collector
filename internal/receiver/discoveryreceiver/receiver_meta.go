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

package discoveryreceiver

func init() {
	receiverMetaMap["prometheus/otelcol"] = ReceiverMeta{
		ServiceType: "prometheus",
		Status: Status{
			Metrics: []Match{
				{
					Status:  "successful",
					Regexp:  "^otelcol_process_uptime$",
					Message: "Successfully connected to prometheus server",
				},
			},
			Statements: []Match{
				{
					Status:  "failed",
					Strict:  "Failed to scrape Prometheus endpoint",
					Message: "Port appears to not be serving prometheus metrics",
				},
				{
					Status:  "failed",
					Regexp:  "Failed to scrape Prometheus endpoint",
					Message: "Port appears to not be serving prometheus metrics",
				},
			},
		},
	}
}
