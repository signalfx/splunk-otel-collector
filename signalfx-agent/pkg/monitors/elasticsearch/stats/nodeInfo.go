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

package stats

// NodeInfoOutput holds output from "_nodes/_local" endpoint
type NodeInfoOutput struct {
	ClusterName *string             `json:"cluster_name"`
	Nodes       map[string]NodeInfo `json:"nodes"`
}

// NodeInfo contains basic information from the nodes
type NodeInfo struct {
	Name             *string `json:"name"`
	TransportAddress *string `json:"transport_address"`
	Host             *string `json:"host"`
	IP               *string `json:"ip"`
	Version          *string `json:"version"`
}

// MasterInfoOutput holds output from "_cluster/state/master_node"
type MasterInfoOutput struct {
	ClusterName *string `json:"cluster_name"`
	MasterNode  *string `json:"master_node"`
}
